package plugipay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Every hand-written method calls a route the backend really has. The spec
// (backend/openapi.json) is made from the backend's own code by
// scripts/apigen.sh, so a method pointing at a route that was renamed or never
// existed fails here instead of 404ing for a customer.
func TestEveryMethodCallsARouteInTheSpec(t *testing.T) {
	raw, err := os.ReadFile("../../backend/openapi.json")
	if os.IsNotExist(err) {
		t.Skip("no backend/openapi.json beside this SDK (a public mirror of sdk/go)")
	}
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	param := regexp.MustCompile(`\{[^}]+\}`)
	routes := map[string]bool{}
	for path, ops := range spec.Paths {
		for method := range ops {
			routes[strings.ToUpper(method)+" "+param.ReplaceAllString(path, "{}")] = true
		}
	}

	var calls []string
	placeholder := regexp.MustCompile(`__\d+__`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+placeholder.ReplaceAllString(r.URL.Path, "{}"))
		_, _ = w.Write([]byte(`{"data":null,"error":null,"meta":{"requestId":"r"}}`))
	}))
	defer srv.Close()
	c, err := NewClient(ClientOptions{KeyID: "ak_test_x", Secret: "sk", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
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
				args = append(args, reflect.ValueOf("__"+string(rune('0'+i))+"__").Convert(in))
			default:
				args = append(args, reflect.Zero(in))
			}
		}
		calls = calls[:0]
		m.Call(args)
		count++
		if len(calls) != 1 {
			t.Errorf("%s made %d requests", name, len(calls))
			return
		}
		if !routes[calls[0]] {
			missing = append(missing, name+": "+calls[0])
		}
	}
	// Methods of every resource namespace (c.Customers.List, c.Billing.ListTiers, …).
	var walk func(name string, v reflect.Value)
	walk = func(name string, v reflect.Value) {
		for i := 0; i < v.NumMethod(); i++ {
			call(name+"."+v.Type().Method(i).Name, v.Method(i))
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
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("methods calling routes the backend does not have:\n%s", strings.Join(missing, "\n"))
	}
}
