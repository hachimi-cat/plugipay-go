package plugipay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// serverSignature is what the backend checks (middleware/hmac-auth.ts):
// method, the path with its query, the timestamp, the hash of the body bytes
// sent and the idempotency key.
func serverSignature(r *http.Request, body []byte, secret string) string {
	sum := sha256.Sum256(body)
	toSign := r.Method + "\n" + r.URL.RequestURI() + "\n" + r.Header.Get("X-Plugipay-Timestamp") + "\n" + hex.EncodeToString(sum[:])
	if idem := r.Header.Get("Idempotency-Key"); idem != "" {
		toSign += "\n" + idem
	}
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(toSign))
	return hex.EncodeToString(m.Sum(nil))
}

// client.API (api_generated.go) goes through the client's own request path:
// the same HMAC signing, idempotency key on writes and envelope handling as
// every hand-written call.
func TestAPISendsPathQueryBodySigned(t *testing.T) {
	var got struct {
		method, path, query, auth, idem, want string
		body                                  map[string]any
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		got.method, got.path, got.query = r.Method, r.URL.Path, r.URL.RawQuery
		got.auth, got.idem = r.Header.Get("Authorization"), r.Header.Get("Idempotency-Key")
		got.want = "Plugipay-HMAC-SHA256 keyId=ak_test, scope=*, signature=" + serverSignature(r, raw, "sk_test")
		got.body = nil
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &got.body)
		}
		_, _ = w.Write([]byte(`{"data":{"id":"cus_1"},"error":null,"meta":{"requestId":"r"}}`))
	}))
	defer srv.Close()
	c, err := NewClient(ClientOptions{KeyID: "ak_test", Secret: "sk_test", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	out, err := c.API.CustomersCreate(context.Background(), &CustomersCreateArgs{Email: Ptr("ada@example.com"), Name: Ptr("Ada <&>")})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"id":"cus_1"}` {
		t.Fatalf("data = %s", out)
	}
	if got.method != "POST" || got.path != "/api/v1/customers" || got.idem == "" || got.auth != got.want {
		t.Fatalf("sent %+v", got)
	}
	if got.body["email"] != "ada@example.com" || got.body["name"] != "Ada <&>" {
		t.Fatalf("body %v", got.body)
	}

	if _, err := c.API.DashboardMetrics(context.Background(), &DashboardMetricsArgs{Window: "7d"}); err != nil {
		t.Fatal(err)
	}
	if got.method != "GET" || got.path != "/api/v1/dashboard/metrics" || got.query != "window=7d" || got.idem != "" || got.auth != got.want {
		t.Fatalf("sent %+v", got)
	}
}

func TestAPIKeyIsSentAsBearer(t *testing.T) {
	var auth, ts string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, ts = r.Header.Get("Authorization"), r.Header.Get("X-Plugipay-Timestamp")
		_, _ = w.Write([]byte(`{"data":[],"error":null,"meta":{"requestId":"r"}}`))
	}))
	defer srv.Close()
	c, err := NewClient(ClientOptions{APIKey: "pk_live_abc", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.API.CustomersList(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer pk_live_abc" || ts != "" {
		t.Fatalf("auth=%q ts=%q", auth, ts)
	}
}

func TestAPIReportsTheErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"data":null,"error":{"code":"not_found","message":"no such customer"},"meta":{"requestId":"r"}}`))
	}))
	defer srv.Close()
	c, _ := NewClient(ClientOptions{KeyID: "ak_test", Secret: "sk_test", BaseURL: srv.URL})
	_, err := c.API.CustomersGet(context.Background(), "cus_missing")
	e, ok := err.(*Error)
	if !ok || e.Status != 404 || e.Code != "not_found" {
		t.Fatalf("err = %#v", err)
	}
}
