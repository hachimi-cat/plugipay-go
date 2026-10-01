package plugipay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ─── helpers ────────────────────────────────────────────────────

func newTestClient(t *testing.T, srvURL string) *Client {
	t.Helper()
	c, err := NewClient(ClientOptions{
		KeyID:   "ak_test_key",
		Secret:  "sk_test_secret_supersecret",
		BaseURL: srvURL,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func envelopeBody(t *testing.T, data any, meta *EnvelopeMeta) []byte {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	env := APIEnvelope{Data: raw, Meta: meta}
	out, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return out
}

func errorEnvelopeBody(t *testing.T, code, msg, requestID string) []byte {
	t.Helper()
	env := APIEnvelope{
		Error: &APIError{Code: code, Message: msg},
	}
	if requestID != "" {
		env.Meta = &EnvelopeMeta{RequestID: requestID}
	}
	out, _ := json.Marshal(env)
	return out
}

// ─── constructor / env defaults ─────────────────────────────────

func TestNewClient_RejectsMissingCreds(t *testing.T) {
	// clear env so the defaults can't accidentally rescue us
	t.Setenv("PLUGIPAY_KEY_ID", "")
	t.Setenv("PLUGIPAY_SECRET", "")
	if _, err := NewClient(ClientOptions{}); err == nil {
		t.Fatal("expected error when KeyID + Secret missing")
	}

	if _, err := NewClient(ClientOptions{KeyID: "x"}); err == nil {
		t.Fatal("expected error when Secret missing")
	}
	if _, err := NewClient(ClientOptions{Secret: "x"}); err == nil {
		t.Fatal("expected error when KeyID missing")
	}
}

func TestNewClient_ReadsFromEnv(t *testing.T) {
	t.Setenv("PLUGIPAY_KEY_ID", "env_key_id")
	t.Setenv("PLUGIPAY_SECRET", "env_secret")
	t.Setenv("PLUGIPAY_BASE_URL", "https://env.example.com/")
	t.Setenv("PLUGIPAY_ON_BEHALF_OF", "acc_env")

	c, err := NewClient(ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.KeyID() != "env_key_id" {
		t.Errorf("KeyID: want env_key_id, got %s", c.KeyID())
	}
	if c.BaseURL() != "https://env.example.com" {
		t.Errorf("BaseURL trim: got %q", c.BaseURL())
	}
	if c.OnBehalfOf() != "acc_env" {
		t.Errorf("OnBehalfOf: got %q", c.OnBehalfOf())
	}
}

func TestNewClient_DefaultBaseURLAndTimeout(t *testing.T) {
	c, err := NewClient(ClientOptions{KeyID: "k", Secret: "s"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.BaseURL() != DefaultBaseURL {
		t.Errorf("BaseURL default: got %s, want %s", c.BaseURL(), DefaultBaseURL)
	}
	if c.http.Timeout != DefaultTimeout {
		t.Errorf("Timeout default: got %s, want %s", c.http.Timeout, DefaultTimeout)
	}
}

func TestForMerchant_OverridesOnBehalfOf(t *testing.T) {
	// ensure no on-behalf env leak
	_ = os.Unsetenv("PLUGIPAY_ON_BEHALF_OF")
	c, _ := NewClient(ClientOptions{KeyID: "k", Secret: "s"})
	if c.OnBehalfOf() != "" {
		t.Fatalf("expected empty default OBO, got %q", c.OnBehalfOf())
	}
	scoped := c.ForMerchant("acc_xyz")
	if scoped.OnBehalfOf() != "acc_xyz" {
		t.Errorf("ForMerchant: got %q", scoped.OnBehalfOf())
	}
	if c.OnBehalfOf() != "" {
		t.Error("ForMerchant mutated parent")
	}
}

// ─── HMAC signing round-trip ────────────────────────────────────

func TestSign_RoundTrip(t *testing.T) {
	secret := "sk_test_supersecret"
	ts := int64(1_700_000_000)
	body := `{"hello":"world"}`

	got := Sign(secret, SignInput{
		Method: "POST", Path: "/api/v1/customers",
		Body: body, Timestamp: ts,
	})

	// independently re-derive
	bodyHash := sha256.Sum256([]byte(body))
	stringToSign := "POST\n/api/v1/customers\n" +
		strconv.FormatInt(ts, 10) + "\n" + hex.EncodeToString(bodyHash[:])
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	wantSig := hex.EncodeToString(mac.Sum(nil))

	if got.Signature != wantSig {
		t.Fatalf("signature mismatch:\n got=%s\nwant=%s", got.Signature, wantSig)
	}
	if got.Timestamp != strconv.FormatInt(ts, 10) {
		t.Fatalf("timestamp: got %s", got.Timestamp)
	}
}

func TestSign_IncludesIdempotencyKey(t *testing.T) {
	secret := "sk_x"
	ts := int64(1_700_000_000)

	withoutKey := Sign(secret, SignInput{Method: "POST", Path: "/x", Body: "{}", Timestamp: ts})
	withKey := Sign(secret, SignInput{Method: "POST", Path: "/x", Body: "{}",
		IdempotencyKey: "idem_abc", Timestamp: ts})

	if withoutKey.Signature == withKey.Signature {
		t.Fatal("idempotency key not included in signature")
	}

	// re-derive the with-key path
	bodyHash := sha256.Sum256([]byte("{}"))
	stringToSign := "POST\n/x\n" + strconv.FormatInt(ts, 10) + "\n" +
		hex.EncodeToString(bodyHash[:]) + "\nidem_abc"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	want := hex.EncodeToString(mac.Sum(nil))
	if withKey.Signature != want {
		t.Fatalf("with-key sig mismatch: got %s want %s", withKey.Signature, want)
	}
}

func TestAuthorizationHeader_Format(t *testing.T) {
	h := AuthorizationHeader("ak_live_123", "deadbeef")
	want := "Plugipay-HMAC-SHA256 keyId=ak_live_123, scope=*, signature=deadbeef"
	if h != want {
		t.Fatalf("header format: got %q want %q", h, want)
	}
}

// ─── round-trip tests via httptest ──────────────────────────────

// assertSignedRequest verifies that the incoming HTTP request carries
// the right Authorization, Timestamp, and (when applicable) idempotency
// + content-type headers, AND that re-deriving the signature from the
// server side produces the same value the SDK sent.
func assertSignedRequest(t *testing.T, r *http.Request, secret, keyID string) []byte {
	t.Helper()

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Plugipay-HMAC-SHA256 keyId="+keyID) {
		t.Errorf("Authorization prefix bad: %q", auth)
	}

	tsStr := r.Header.Get("X-Plugipay-Timestamp")
	if tsStr == "" {
		t.Error("missing X-Plugipay-Timestamp")
	}

	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()

	// extract signature from Authorization
	var sig string
	for _, p := range strings.Split(auth, ",") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "signature=") {
			sig = strings.TrimPrefix(p, "signature=")
		}
	}
	if sig == "" {
		t.Fatal("no signature= in Authorization")
	}

	idem := r.Header.Get("Idempotency-Key")
	bodyHash := sha256.Sum256(body)
	stringToSign := r.Method + "\n" + r.URL.RequestURI() + "\n" + tsStr + "\n" +
		hex.EncodeToString(bodyHash[:])
	if idem != "" {
		stringToSign += "\n" + idem
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	want := hex.EncodeToString(mac.Sum(nil))
	if want != sig {
		t.Errorf("signature mismatch:\n got=%s\nwant=%s\nstringToSign=%q",
			sig, want, stringToSign)
	}

	if len(body) > 0 && r.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type: want application/json, got %q", r.Header.Get("Content-Type"))
	}
	return body
}

func TestCustomers_Create_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/customers" {
			t.Errorf("bad route: %s %s", r.Method, r.URL.Path)
		}
		body := assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")

		// ensure idempotency-key header was sent
		if r.Header.Get("Idempotency-Key") == "" {
			t.Error("missing Idempotency-Key")
		}

		// echo back as Customer
		var in CustomerCreateInput
		if err := json.Unmarshal(body, &in); err != nil {
			t.Fatalf("server unmarshal: %v", err)
		}
		out := Customer{
			ID:        "cus_test_1",
			AccountID: "acc_1",
			Email:     in.Email,
			Name:      in.Name,
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t, out, &EnvelopeMeta{RequestID: "req_1"}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	email := "buyer@example.com"
	got, err := c.Customers.Create(context.Background(), CustomerCreateInput{Email: &email})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != "cus_test_1" || got.Email == nil || *got.Email != email {
		t.Fatalf("bad response: %+v", got)
	}
}

func TestCustomers_Get_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/customers/cus_42" {
			t.Errorf("bad path: %s", r.URL.Path)
		}
		assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t, Customer{ID: "cus_42", AccountID: "acc_1",
			CreatedAt: "x", UpdatedAt: "y"}, nil))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.Customers.Get(context.Background(), "cus_42")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "cus_42" {
		t.Fatalf("got %+v", got)
	}
}

func TestCustomers_List_PaginationMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit qs: %s", r.URL.RawQuery)
		}
		assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")
		cursor := "cur_next"
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t,
			[]Customer{{ID: "cus_1", AccountID: "a", CreatedAt: "x", UpdatedAt: "y"}},
			&EnvelopeMeta{Cursor: &cursor, HasMore: true},
		))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	limit := 5
	page, err := c.Customers.List(context.Background(), CustomerListParams{Limit: &limit})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Data) != 1 || page.Data[0].ID != "cus_1" {
		t.Fatalf("data: %+v", page.Data)
	}
	if !page.HasMore || page.Cursor == nil || *page.Cursor != "cur_next" {
		t.Fatalf("meta: cursor=%v hasMore=%v", page.Cursor, page.HasMore)
	}
}

func TestCheckoutSessions_Create_NormalizesLineItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")
		// confirm lineItems is present as an empty array, not missing.
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		li, ok := parsed["lineItems"]
		if !ok {
			t.Errorf("lineItems missing from body: %s", body)
		}
		if arr, ok2 := li.([]any); !ok2 || len(arr) != 0 {
			t.Errorf("lineItems not normalized to []: %T %v", li, li)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t, CheckoutSession{
			ID: "cs_1", AccountID: "acc_1",
			Status: "open", SuccessURL: "https://s/", CancelURL: "https://c/",
			HostedURL: "https://h/", ExpiresAt: "2099-01-01T00:00:00Z",
			CreatedAt: "x", UpdatedAt: "y",
		}, nil))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.CheckoutSessions.Create(context.Background(), CheckoutSessionCreateInput{
		Amount: 10000, Currency: CurrencyIDR,
		Methods:    []CheckoutMethod{CheckoutMethodQRIS},
		SuccessURL: "https://s/", CancelURL: "https://c/",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != "cs_1" {
		t.Fatal("bad echo")
	}
}

func TestRefunds_Create_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/refunds" {
			t.Errorf("bad route: %s %s", r.Method, r.URL.Path)
		}
		body := assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")
		var in RefundCreateInput
		_ = json.Unmarshal(body, &in)
		if in.SourceType != SourceTypeCheckoutSession || in.SourceID != "cs_1" {
			t.Errorf("body: %+v", in)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t, Refund{
			ID: "rf_1", AccountID: "a", Amount: 5000, Currency: CurrencyIDR,
			Status: RefundStatusPending, ChargeID: "chg_1", Reason: "requested_by_customer",
			CreatedAt: "x", UpdatedAt: "y",
		}, nil))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	amt := int64(5000)
	got, err := c.Refunds.Create(context.Background(), RefundCreateInput{
		SourceType: SourceTypeCheckoutSession, SourceID: "cs_1", Amount: &amt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Status != RefundStatusPending {
		t.Fatalf("status: %s", got.Status)
	}
}

func TestWebhookEndpoints_Delete_NoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Path != "/api/v1/webhooks/we_1" {
			t.Errorf("path: %s", r.URL.Path)
		}
		assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":null}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.WebhookEndpoints.Delete(context.Background(), "we_1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestPayouts_Balance_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/payouts/balance" {
			t.Errorf("bad route: %s %s", r.Method, r.URL.Path)
		}
		assertSignedRequest(t, r, "sk_test_secret_supersecret", "ak_test_key")
		cur := "IDR"
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t, AvailableBalance{
			LedgerBalance: 100000, Locked: 1000, Available: 99000, Currency: &cur,
		}, nil))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	bal, err := c.Payouts.Balance(context.Background())
	if err != nil {
		t.Fatalf("Balance: %v", err)
	}
	if bal.Available != 99000 {
		t.Fatalf("Available: %d", bal.Available)
	}
}

func TestForMerchant_SendsOnBehalfOfHeader(t *testing.T) {
	var gotOBO string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOBO = r.Header.Get("X-Plugipay-On-Behalf-Of")
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelopeBody(t, AccountProfile{
			ID: "u_1", Email: "x@x.x", EmailVerified: true,
		}, nil))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	scoped := c.ForMerchant("acc_target")
	if _, err := scoped.Account.Get(context.Background()); err != nil {
		t.Fatalf("Account.Get: %v", err)
	}
	if gotOBO != "acc_target" {
		t.Fatalf("OBO header: got %q", gotOBO)
	}
}

// ─── error handling ─────────────────────────────────────────────

func TestDo_ApiError_PopulatesErrorFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		w.Write(errorEnvelopeBody(t, "rate_limited", "slow down", "req_xyz"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Customers.Get(context.Background(), "cus_1")
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *Error
	if !errors.As(err, &pe) {
		t.Fatalf("not a *plugipay.Error: %T", err)
	}
	if pe.Status != 429 || pe.Code != "rate_limited" || pe.RequestID != "req_xyz" {
		t.Errorf("got %+v", pe)
	}
	if !strings.Contains(pe.Error(), "req_xyz") {
		t.Errorf("Error() should mention request id: %s", pe.Error())
	}
}

func TestDo_NonJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(502)
		w.Write([]byte("<html>bad gateway</html>"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Customers.Get(context.Background(), "cus_1")
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *Error
	if !errors.As(err, &pe) || pe.Code != "invalid_response" {
		t.Fatalf("got %+v", err)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// hold the connection long enough for ctx to fire
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":null}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Customers.Get(ctx, "cus_1")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var pe *Error
	if !errors.As(err, &pe) {
		t.Fatalf("not *Error: %T", err)
	}
	if pe.Code != "timeout" && pe.Code != "network_error" && pe.Code != "canceled" {
		t.Errorf("expected timeout/network_error/canceled, got %s", pe.Code)
	}
}

// ─── errors.go behavior ─────────────────────────────────────────

func TestError_Format(t *testing.T) {
	e := newErr(429, "rate_limited", "too fast")
	if !strings.Contains(e.Error(), "rate_limited") {
		t.Error("missing code in Error()")
	}
	if !strings.Contains(e.Error(), "429") {
		t.Error("missing status in Error()")
	}

	e2 := newErr(0, "timeout", "took too long")
	if strings.Contains(e2.Error(), "status=0") {
		t.Errorf("status=0 should not appear: %s", e2.Error())
	}
}
