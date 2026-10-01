package plugipay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Every hand-written method calls a route the backend really has, and sends what that
// route requires. The spec (backend/openapi.json) is made from the backend's own code
// by scripts/apigen.sh, so a method pointing at a route that was renamed or never
// existed, one that sends no Idempotency-Key where the route takes one, or one whose
// body the route's schema refuses (a required field it can never send, a field a strict
// schema doesn't know) fails here instead of 4xx-ing for a customer. Each method is
// called with every field of its inputs filled in, so the request recorded is the most
// a caller can ever send.

type specSchema struct {
	Properties           map[string]json.RawMessage `json:"properties"`
	Required             []string                   `json:"required"`
	AdditionalProperties *bool                      `json:"additionalProperties"`
	AnyOf                []specSchema               `json:"anyOf"`
}

type specOperation struct {
	XForjio struct {
		Guards []string `json:"guards"`
		Body   string   `json:"body"`
	} `json:"x-forjio"`
	RequestBody *struct {
		Content map[string]struct {
			Schema *specSchema `json:"schema"`
		} `json:"content"`
	} `json:"requestBody"`
}

// multiRequest: methods that make more than one request, and why.
var multiRequest = map[string]string{
	// the route copies the template as "<name> (copy)"; a name of your own is a rename after it
	"Templates.Duplicate": "duplicate, then PATCH the name",
}

type sentRequest struct {
	route string
	idem  string
	body  map[string]any
}

func requestProblems(op *specOperation, s sentRequest) []string {
	if op == nil {
		return []string{s.route + ": no such route"}
	}
	var out []string
	for _, g := range op.XForjio.Guards {
		if g == "idempotency" && s.idem == "" {
			out = append(out, s.route+": sends no Idempotency-Key (the route takes one)")
		}
	}
	if s.body == nil {
		return out
	}
	keys := make([]string, 0, len(s.body))
	for k := range s.body {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var schema *specSchema
	if op.RequestBody != nil {
		schema = op.RequestBody.Content["application/json"].Schema
	}
	switch {
	case op.XForjio.Body == "validated" && schema != nil:
		branches := schema.AnyOf
		if len(branches) == 0 {
			branches = []specSchema{*schema}
		}
		check := func(b specSchema) (missing, unknown []string) {
			for _, r := range b.Required {
				if _, ok := s.body[r]; !ok {
					missing = append(missing, r)
				}
			}
			if b.AdditionalProperties != nil && !*b.AdditionalProperties {
				for _, k := range keys {
					if _, ok := b.Properties[k]; !ok {
						unknown = append(unknown, k)
					}
				}
			}
			return
		}
		fits := false
		for _, b := range branches {
			if m, u := check(b); len(m) == 0 && len(u) == 0 {
				fits = true
			}
		}
		if !fits {
			m, u := check(branches[0])
			msg := s.route + ": body the route refuses"
			if len(m) > 0 {
				msg += " — can never send required " + strings.Join(m, ", ")
			}
			if len(u) > 0 {
				msg += " — sends " + strings.Join(u, ", ") + ", which the schema doesn't allow"
			}
			out = append(out, msg)
		}
	case op.XForjio.Body == "none" && len(keys) > 0:
		out = append(out, s.route+": sends "+strings.Join(keys, ", ")+", but the route reads no body")
	case op.XForjio.Body == "read-unvalidated" && schema != nil && len(schema.Properties) > 0:
		var ignored []string
		for _, k := range keys {
			if _, ok := schema.Properties[k]; !ok {
				ignored = append(ignored, k)
			}
		}
		if len(ignored) > 0 {
			out = append(out, s.route+": sends "+strings.Join(ignored, ", ")+", which the route never reads")
		}
	}
	return out
}

var readerType = reflect.TypeOf((*io.Reader)(nil)).Elem()

// filled returns a value of t with every field set: strings "x", numbers 1, true,
// one element in a slice or map, pointers to filled values, readers that read "x".
func filled(t reflect.Type, depth int) reflect.Value {
	v := reflect.New(t).Elem()
	if t == readerType {
		v.Set(reflect.ValueOf(strings.NewReader("x")))
		return v
	}
	switch t.Kind() {
	case reflect.String:
		v.SetString("x")
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(1)
	case reflect.Ptr:
		if depth > 0 {
			v.Set(filled(t.Elem(), depth-1).Addr())
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if f := t.Field(i); f.IsExported() && depth > 0 {
				v.Field(i).Set(filled(f.Type, depth-1))
			}
		}
	case reflect.Slice:
		if depth > 0 {
			s := reflect.MakeSlice(t, 1, 1)
			s.Index(0).Set(filled(t.Elem(), depth-1))
			v.Set(s)
		}
	case reflect.Map:
		m := reflect.MakeMap(t)
		if t.Key().Kind() == reflect.String && depth > 0 {
			m.SetMapIndex(reflect.ValueOf("x").Convert(t.Key()), filled(t.Elem(), depth-1))
		}
		v.Set(m)
	case reflect.Interface:
		if t.NumMethod() == 0 {
			v.Set(reflect.ValueOf("x"))
		}
	}
	return v
}

func TestEveryMethodCallsARouteInTheSpec(t *testing.T) {
	raw, err := os.ReadFile("../../backend/openapi.json")
	if os.IsNotExist(err) {
		t.Skip("no backend/openapi.json beside this SDK (a public mirror of sdk/go)")
	}
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var spec struct {
		Paths map[string]map[string]*specOperation `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	param := regexp.MustCompile(`\{[^}]+\}`)
	routes := map[string]*specOperation{}
	for path, ops := range spec.Paths {
		for method, op := range ops {
			routes[strings.ToUpper(method)+" "+param.ReplaceAllString(path, "{}")] = op
		}
	}

	var sent []sentRequest
	placeholder := regexp.MustCompile(`__\d+__`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := sentRequest{
			route: r.Method + " " + placeholder.ReplaceAllString(r.URL.Path, "{}"),
			idem:  r.Header.Get("Idempotency-Key"),
		}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			b, _ := io.ReadAll(r.Body)
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				s.body = m
			}
		}
		sent = append(sent, s)
		// a created object's id ("__9__") stands for a path parameter too (duplicate + rename)
		_, _ = w.Write([]byte(`{"data":{"id":"__9__"},"error":null,"meta":{"requestId":"r"}}`))
	}))
	defer srv.Close()
	c, err := NewClient(ClientOptions{KeyID: "ak_test_x", Secret: "sk", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	var found []string
	count := 0
	call := func(name string, m reflect.Value) {
		args := []reflect.Value{}
		for i := 0; i < m.Type().NumIn(); i++ {
			in := m.Type().In(i)
			switch {
			case in == reflect.TypeOf((*context.Context)(nil)).Elem():
				args = append(args, reflect.ValueOf(context.Background()))
			case in.Kind() == reflect.String:
				// Ids in the path: "__1__", "__2__", ... stand for its parameters.
				args = append(args, reflect.ValueOf(fmt.Sprintf("__%d__", i)).Convert(in))
			case in.Kind() == reflect.Map:
				// a map that IS the body declares no fields of its own: an empty one
				args = append(args, reflect.MakeMap(in))
			default:
				args = append(args, filled(in, 3))
			}
		}
		sent = sent[:0]
		m.Call(args)
		count++
		if len(sent) == 0 || (len(sent) > 1 && multiRequest[name] == "") {
			found = append(found, fmt.Sprintf("%s made %d requests", name, len(sent)))
			return
		}
		for _, s := range sent {
			for _, p := range requestProblems(routes[s.route], s) {
				found = append(found, name+": "+p)
			}
		}
	}
	// Methods of every resource namespace (c.Customers.List, c.Billing.ListTiers, …).
	var walk func(name string, v reflect.Value)
	walk = func(name string, v reflect.Value) {
		for i := 0; i < v.NumMethod(); i++ {
			call(strings.TrimPrefix(name+"."+v.Type().Method(i).Name, "."), v.Method(i))
		}
		elem := v
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < elem.NumField(); i++ {
			if f := elem.Type().Field(i); f.IsExported() {
				walk(name+"."+f.Name, elem.Field(i))
			}
		}
	}
	cv := reflect.ValueOf(c).Elem()
	for i := 0; i < cv.NumField(); i++ {
		// API (api_generated.go) is made from the spec itself, so it cannot drift.
		if f := cv.Type().Field(i); f.IsExported() && f.Name != "API" {
			walk(f.Name, cv.Field(i))
		}
	}
	if count < 80 {
		t.Fatalf("only %d methods found", count)
	}
	if len(found) > 0 {
		seen := map[string]bool{}
		var uniq []string
		for _, f := range found {
			if !seen[f] {
				seen[f] = true
				uniq = append(uniq, f)
			}
		}
		t.Fatalf("hand-written methods the API refuses:\n%s", strings.Join(uniq, "\n"))
	}
}
