package plugipay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// 0.4.0: hand-written methods the API refused before. live_app_test.go drives the same
// methods against a running API.

type recorded struct {
	method, path, idem string
	body               map[string]any
}

func recordingServer(t *testing.T, reply func(r *http.Request, w http.ResponseWriter)) (*Client, *[]recorded) {
	t.Helper()
	var sent []recorded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recorded{method: r.Method, path: r.URL.Path, idem: r.Header.Get("Idempotency-Key")}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &rec.body)
		sent = append(sent, rec)
		reply(r, w)
	}))
	t.Cleanup(srv.Close)
	return newTestClient(t, srv.URL), &sent
}

func jsonReply(data string) func(*http.Request, http.ResponseWriter) {
	return func(_ *http.Request, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":` + data + `,"error":null}`))
	}
}

func TestDelete_204NoContent(t *testing.T) {
	c, _ := recordingServer(t, func(_ *http.Request, w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) })
	if err := c.Templates.Delete(context.Background(), "tpl_1"); err != nil {
		t.Fatalf("a 204 is success, got %v", err)
	}
}

func TestCustomersUpdate_SendsIdempotencyKey(t *testing.T) {
	c, sent := recordingServer(t, jsonReply(`{"id":"cus_1"}`))
	name := "Ada"
	if _, err := c.Customers.Update(context.Background(), "cus_1", CustomerUpdateInput{Name: &name}); err != nil {
		t.Fatal(err)
	}
	if (*sent)[0].idem == "" {
		t.Fatal("no Idempotency-Key")
	}
}

func TestPlansCreate_FlatShorthand(t *testing.T) {
	c, sent := recordingServer(t, jsonReply(`{"id":"pln_1"}`))
	if _, err := c.Plans.Create(context.Background(), PlanCreateInput{Name: "Pro", Interval: "month", Currency: "IDR", Amount: 150000}); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"name": "Pro", "interval": "month",
		"portalFeatures": map[string]any{"selfServeCancel": true, "selfServePause": false, "selfServeUpgrade": false, "selfServeDowngrade": false, "updatePaymentMethod": true},
		"prices":         []any{map[string]any{"currency": "IDR", "model": "flat", "unitAmount": float64(150000)}},
	}
	if got := (*sent)[0].body; !reflect.DeepEqual(got, want) {
		t.Fatalf("body:\n got %v\nwant %v", got, want)
	}
}

func TestTemplatesPreview_ReturnsThePage(t *testing.T) {
	c, sent := recordingServer(t, func(_ *http.Request, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><p>hi</p>"))
	})
	out, err := c.Templates.Preview(context.Background(), TemplatePreviewInput{Kind: "receipt"})
	if err != nil {
		t.Fatal(err)
	}
	if out.HTML != "<!doctype html><p>hi</p>" {
		t.Fatalf("html: %q", out.HTML)
	}
	if got := (*sent)[0].body; !reflect.DeepEqual(got, map[string]any{"kind": "receipt", "config": map[string]any{}}) {
		t.Fatalf("body: %v", got)
	}
}

func TestTemplatesDuplicate_RenamesTheCopy(t *testing.T) {
	c, sent := recordingServer(t, jsonReply(`{"id":"tpl_2","name":"B"}`))
	name := "B"
	if _, err := c.Templates.Duplicate(context.Background(), "tpl_1", &name); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range *sent {
		got = append(got, s.method+" "+s.path)
	}
	if strings.Join(got, ", ") != "POST /api/v1/templates/tpl_1/duplicate, PATCH /api/v1/templates/tpl_2" {
		t.Fatalf("requests: %v", got)
	}
}

func TestAdaptersList_FromTheObjectKeyedByKind(t *testing.T) {
	c, _ := recordingServer(t, jsonReply(`{"xendit":{"kind":"xendit","status":"active"},"manual":{"kind":"manual","status":"active"}}`))
	list, err := c.Adapters.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].Kind != "manual" || list[1].Kind != "xendit" {
		t.Fatalf("list: %+v", list)
	}
}

func TestAdaptersUpdate_SendsIdempotencyKey(t *testing.T) {
	c, sent := recordingServer(t, jsonReply(`{"kind":"xendit"}`))
	if _, err := c.Adapters.UpdateXendit(context.Background(), XenditAdapterInput{SecretKey: "xnd_sk_test_1"}); err != nil {
		t.Fatal(err)
	}
	if s := (*sent)[0]; s.idem == "" || !reflect.DeepEqual(s.body, map[string]any{"secretKey": "xnd_sk_test_1"}) {
		t.Fatalf("sent: %+v", s)
	}
}
