// Package plugipay is the official Go SDK for the Plugipay payments
// API. It mirrors the public surface of @forjio/plugipay-node and the
// Python `plugipay` package: HMAC-signed REST requests, a single
// resource namespace per concept (customers, plans, checkout-sessions,
// invoices, subscriptions, refunds, adapters, payouts, ledger, reports,
// webhooks, etc.), and stdlib-only deps.
//
// Quickstart:
//
//	c, err := plugipay.NewClient(plugipay.ClientOptions{
//	    KeyID:  "ak_live_...",
//	    Secret: "sk_...",
//	})
//	if err != nil { /* missing creds */ }
//
//	cust, err := c.Customers.Create(ctx, plugipay.CustomerCreateInput{
//	    Email: ptr("buyer@example.com"),
//	})
package plugipay

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// DefaultBaseURL is the production Plugipay endpoint. Override via
// ClientOptions.BaseURL or PLUGIPAY_BASE_URL.
const DefaultBaseURL = "https://plugipay.com"

// DefaultTimeout is the per-request timeout when ClientOptions.Timeout
// is zero.
const DefaultTimeout = 30 * time.Second

// ClientOptions configures a Client. KeyID + Secret (or APIKey) are
// required; the rest default from env vars (PLUGIPAY_KEY_ID,
// PLUGIPAY_SECRET, PLUGIPAY_API_KEY, PLUGIPAY_BASE_URL,
// PLUGIPAY_ON_BEHALF_OF) so a zero-value ClientOptions{} works in
// production environments where the standard env vars are set.
type ClientOptions struct {
	// HMAC access key id (e.g. "ak_live_..."). Defaults to PLUGIPAY_KEY_ID.
	KeyID string
	// HMAC secret. Defaults to PLUGIPAY_SECRET.
	Secret string
	// A key minted in the dashboard (Settings → API keys), "pk_live_…" /
	// "pk_test_…": sent as "Authorization: Bearer <key>" instead of signing
	// with KeyID + Secret. Defaults to PLUGIPAY_API_KEY.
	APIKey string
	// Base URL. Defaults to PLUGIPAY_BASE_URL or DefaultBaseURL.
	BaseURL string
	// Optional merchant accountId to scope calls against. Forwarded as
	// X-Plugipay-On-Behalf-Of. Only allowed for platform-admin keys.
	// Defaults to PLUGIPAY_ON_BEHALF_OF.
	OnBehalfOf string
	// Per-request timeout. Zero falls back to DefaultTimeout.
	Timeout time.Duration
	// HTTP client. Nil falls back to a fresh http.Client (does NOT use
	// http.DefaultClient — we want our own Timeout default).
	HTTP *http.Client
}

// Client is the main SDK entry point. Construct with NewClient. Safe
// for concurrent use by multiple goroutines.
type Client struct {
	keyID             string
	secret            string
	apiKey            string
	baseURL           string
	defaultOnBehalfOf string
	timeout           time.Duration
	http              *http.Client

	// Resource groups — exposed as fields so callers do
	// c.Customers.Get(...) just like the Node SDK does c.customers.get(...).
	Customers        *CustomersResource
	Plans            *PlansResource
	CheckoutSessions *CheckoutSessionsResource
	Invoices         *InvoicesResource
	Subscriptions    *SubscriptionsResource
	PortalSessions   *PortalSessionsResource
	Receipts         *ReceiptsResource
	Payouts          *PayoutsResource
	Ledger           *LedgerResource
	Reports          *ReportsResource
	WebhookEndpoints *WebhookEndpointsResource
	Events           *EventsResource
	Refunds          *RefundsResource
	Adapters         *AdaptersResource
	ApiKeys          *ApiKeysResource
	Billing          *BillingResource
	Onboarding       *OnboardingResource
	CheckoutSettings *CheckoutSettingsResource
	Templates        *TemplatesResource
	Uploads          *UploadsResource
	Workspaces       *WorkspacesResource
	Account          *AccountResource
	AdminPortal      *AdminPortalResource
	Admin            *AdminResource

	// API has every feature route, one method each (generated from the API
	// spec: api_generated.go), sent like every other call.
	API *GeneratedAPI
}

// NewClient builds a configured client. Returns an error if KeyID or
// Secret cannot be resolved from options or env.
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.KeyID == "" {
		opts.KeyID = os.Getenv("PLUGIPAY_KEY_ID")
	}
	if opts.Secret == "" {
		opts.Secret = os.Getenv("PLUGIPAY_SECRET")
	}
	if opts.APIKey == "" && opts.KeyID == "" {
		opts.APIKey = os.Getenv("PLUGIPAY_API_KEY")
	}
	if opts.BaseURL == "" {
		opts.BaseURL = os.Getenv("PLUGIPAY_BASE_URL")
	}
	if opts.OnBehalfOf == "" {
		opts.OnBehalfOf = os.Getenv("PLUGIPAY_ON_BEHALF_OF")
	}
	if opts.APIKey == "" && opts.KeyID == "" {
		return nil, newErr(0, "missing_key_id",
			"set PLUGIPAY_KEY_ID env or ClientOptions.KeyID (or ClientOptions.APIKey)")
	}
	if opts.APIKey == "" && opts.Secret == "" {
		return nil, newErr(0, "missing_secret",
			"set PLUGIPAY_SECRET env or ClientOptions.Secret")
	}
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.HTTP == nil {
		opts.HTTP = &http.Client{Timeout: opts.Timeout}
	}
	c := &Client{
		keyID:             opts.KeyID,
		secret:            opts.Secret,
		apiKey:            opts.APIKey,
		baseURL:           strings.TrimRight(opts.BaseURL, "/"),
		defaultOnBehalfOf: opts.OnBehalfOf,
		timeout:           opts.Timeout,
		http:              opts.HTTP,
	}
	c.installResources()
	return c, nil
}

// apigenRequest is the call behind Client.API (api_generated.go): the same
// signing (or dashboard key), idempotency key on writes and envelope handling
// as every hand-written call.
func (c *Client) apigenRequest(ctx context.Context, method, path string, query url.Values, body map[string]any) (json.RawMessage, error) {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	opts := RequestOptions{Method: strings.ToUpper(method), Path: path}
	if body != nil {
		opts.Body = body
	}
	if opts.Method != http.MethodGet {
		opts.IdempotencyKey = genIdem()
	}
	env, err := c.do(ctx, opts)
	if err != nil {
		return nil, err
	}
	return env.Data, nil
}

// apigenForm is the call behind a generated file-upload method (api_generated.go):
// the form fields and the files as multipart/form-data, signed like every other
// call (over an empty body: the server does not hash a multipart body), with an
// idempotency key.
func (c *Client) apigenForm(ctx context.Context, method, path string, query url.Values, form map[string]string, files map[string]FormFile) (json.RawMessage, error) {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	parts := make([]formPart, 0, len(files))
	for field, f := range files {
		parts = append(parts, formPart{field: field, filename: f.Name, content: f.Content})
	}
	return c.sendForm(ctx, strings.ToUpper(method), path, form, parts)
}

// formPart is one file of a multipart upload.
type formPart struct {
	field, filename, contentType string
	content                      io.Reader
}

func (c *Client) sendForm(ctx context.Context, method, path string, form map[string]string, files []formPart) (json.RawMessage, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := w.WriteField(k, form[k]); err != nil {
			return nil, newErr(0, "serialize_failed", err.Error())
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].field < files[j].field })
	for _, f := range files {
		name := f.filename
		if name == "" {
			name = f.field
		}
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, quoteEscaper.Replace(f.field), quoteEscaper.Replace(name)))
		ct := f.contentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		h.Set("Content-Type", ct)
		part, err := w.CreatePart(h)
		if err != nil {
			return nil, newErr(0, "serialize_failed", err.Error())
		}
		if f.content != nil {
			if _, err := io.Copy(part, f.content); err != nil {
				return nil, newErr(0, "serialize_failed", "failed to read the file: "+err.Error())
			}
		}
	}
	if err := w.Close(); err != nil {
		return nil, newErr(0, "serialize_failed", err.Error())
	}
	opts := RequestOptions{Method: method, Path: path}
	if method != http.MethodGet {
		opts.IdempotencyKey = genIdem()
	}
	env, err := c.send(ctx, opts, buf.Bytes(), "", w.FormDataContentType())
	if err != nil {
		return nil, err
	}
	return env.Data, nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// ForMerchant returns a shallow clone scoped to a different on-behalf-of
// account id — useful when a platform-admin key is reused across many
// merchant accounts. The clone shares the underlying *http.Client.
func (c *Client) ForMerchant(accountID string) *Client {
	clone := &Client{
		keyID:             c.keyID,
		secret:            c.secret,
		apiKey:            c.apiKey,
		baseURL:           c.baseURL,
		defaultOnBehalfOf: accountID,
		timeout:           c.timeout,
		http:              c.http,
	}
	clone.installResources()
	return clone
}

// KeyID returns the access key id used by this client.
func (c *Client) KeyID() string { return c.keyID }

// BaseURL returns the resolved base URL (no trailing slash).
func (c *Client) BaseURL() string { return c.baseURL }

// OnBehalfOf returns the default on-behalf-of merchant id, if any.
func (c *Client) OnBehalfOf() string { return c.defaultOnBehalfOf }

// ─────────────────────────────────────────────────────────────────
// Low-level request
// ─────────────────────────────────────────────────────────────────

// RequestOptions tunes a single call. Body, IdempotencyKey, OnBehalfOf
// are all optional. Query is appended as a URL querystring (sorted-key
// order is not guaranteed; signatures cover the path-as-sent).
type RequestOptions struct {
	Method         string // GET, POST, PATCH, PUT, DELETE
	Path           string // e.g. "/api/v1/customers"
	Body           any    // serialized via json.Marshal; nil → no body
	IdempotencyKey string
	OnBehalfOf     string // overrides the client default
}

// Do executes a signed request and unmarshals the envelope's `data`
// into out. Returns *Error on any failure. If out is nil, the response
// body is decoded but discarded — useful for endpoints that return
// void/empty payloads (DELETE, revoke, etc).
func (c *Client) Do(ctx context.Context, opts RequestOptions, out any) error {
	envelope, err := c.do(ctx, opts)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return newErr(0, "invalid_response",
			fmt.Sprintf("failed to decode response body into %T: %s",
				out, err.Error()))
	}
	return nil
}

// DoList executes a signed request that returns a list envelope and
// re-shapes it into a Page[T]. Mirrors PlugipayClient#requestList in
// the Node SDK.
func DoList[T any](ctx context.Context, c *Client, opts RequestOptions) (Page[T], error) {
	envelope, err := c.do(ctx, opts)
	if err != nil {
		return Page[T]{}, err
	}
	var data []T
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return Page[T]{}, newErr(0, "invalid_response",
				fmt.Sprintf("failed to decode list body: %s", err.Error()))
		}
	}
	page := Page[T]{Data: data}
	if envelope.Meta != nil {
		page.Cursor = envelope.Meta.Cursor
		page.HasMore = envelope.Meta.HasMore
	}
	return page, nil
}

func (c *Client) do(ctx context.Context, opts RequestOptions) (*APIEnvelope, error) {
	var bodyBytes []byte
	if opts.Body != nil {
		// nil interface guard: a typed nil (e.g. (*Foo)(nil)) still has
		// opts.Body != nil per Go's interface semantics, which is fine —
		// json.Marshal will emit "null".
		b, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, newErr(0, "serialize_failed",
				"failed to marshal request body: "+err.Error())
		}
		bodyBytes = b
	}
	contentType := ""
	if bodyBytes != nil {
		contentType = "application/json"
	}
	return c.send(ctx, opts, bodyBytes, string(bodyBytes), contentType)
}

// send makes one request: body is what goes on the wire, signedBody what the
// signature covers (the JSON body; "" for a multipart upload, whose bytes the
// server does not hash). Each call is signed with a fresh timestamp when it is
// sent, so a retry made by calling again is signed anew.
func (c *Client) send(ctx context.Context, opts RequestOptions, body []byte, signedBody, contentType string) (*APIEnvelope, error) {
	url := c.baseURL + opts.Path
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, opts.Method, url, reqBody)
	if err != nil {
		return nil, newErr(0, "invalid_request", err.Error())
	}

	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		// A dashboard key: sent as is, nothing to sign.
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else {
		sig := Sign(c.secret, SignInput{
			Method:         opts.Method,
			Path:           opts.Path,
			Body:           signedBody,
			IdempotencyKey: opts.IdempotencyKey,
		})
		req.Header.Set("Authorization", AuthorizationHeader(c.keyID, sig.Signature))
		req.Header.Set("X-Plugipay-Timestamp", sig.Timestamp)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if opts.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", opts.IdempotencyKey)
	}
	effectiveOnBehalf := opts.OnBehalfOf
	if effectiveOnBehalf == "" {
		effectiveOnBehalf = c.defaultOnBehalfOf
	}
	if effectiveOnBehalf != "" {
		req.Header.Set("X-Plugipay-On-Behalf-Of", effectiveOnBehalf)
	}

	res, err := c.http.Do(req)
	if err != nil {
		// distinguish context cancellation/timeout from transport error
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, newErr(0, "timeout",
				fmt.Sprintf("plugipay request timed out (timeout=%s)", c.timeout))
		}
		if errors.Is(err, context.Canceled) {
			return nil, newErr(0, "canceled", "request canceled")
		}
		return nil, newErr(0, "network_error", err.Error())
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, newErr(res.StatusCode, "network_error",
			"failed to read response: "+err.Error())
	}

	// DELETE-style routes answer 204 with no body: nothing to decode.
	if res.StatusCode < 400 && len(bytes.TrimSpace(raw)) == 0 {
		return &APIEnvelope{}, nil
	}
	// A route that renders a page (POST /templates/preview) answers 2xx with the HTML
	// itself, not an envelope: it becomes the data, as a JSON string.
	if res.StatusCode < 400 && strings.HasPrefix(strings.ToLower(res.Header.Get("Content-Type")), "text/html") {
		text, _ := json.Marshal(string(raw))
		return &APIEnvelope{Data: text}, nil
	}

	var env APIEnvelope
	if jsonErr := json.Unmarshal(raw, &env); jsonErr != nil {
		snippet := string(raw)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return nil, newErr(res.StatusCode, "invalid_response",
			"non-JSON response: "+snippet)
	}

	if res.StatusCode >= 400 || env.Error != nil {
		code := "unknown"
		msg := fmt.Sprintf("HTTP %d", res.StatusCode)
		if env.Error != nil {
			if env.Error.Code != "" {
				code = env.Error.Code
			}
			if env.Error.Message != "" {
				msg = env.Error.Message
			}
		}
		e := newErr(res.StatusCode, code, msg)
		if env.Meta != nil {
			e.RequestID = env.Meta.RequestID
		}
		return nil, e
	}

	return &env, nil
}

// ─────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────

// genIdem mints a fresh idempotency key. Falls back to a timestamp-based
// key if crypto/rand somehow fails (it shouldn't, but we don't want SDK
// callers to ever see that error).
func genIdem() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("idem_%d", time.Now().UnixNano())
	}
	return "idem_" + hex.EncodeToString(b[:])
}

// qs builds a querystring from a map of values. Skips nil pointers and
// empty strings. Numeric/bool values are formatted with %v. Returns ""
// when no params survive the filter, or "?k=v&..." otherwise.
//
// Order is map-iteration-defined, which is fine — Plugipay's HMAC signs
// the URL path as-sent, so as long as we use the same string for the
// signature and the request, key order doesn't matter.
func qs(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	q := url.Values{}
	for k, v := range params {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if t == "" {
				continue
			}
			q.Set(k, t)
		case *string:
			if t == nil || *t == "" {
				continue
			}
			q.Set(k, *t)
		case int:
			q.Set(k, fmt.Sprintf("%d", t))
		case *int:
			if t == nil {
				continue
			}
			q.Set(k, fmt.Sprintf("%d", *t))
		case int64:
			q.Set(k, fmt.Sprintf("%d", t))
		case *int64:
			if t == nil {
				continue
			}
			q.Set(k, fmt.Sprintf("%d", *t))
		case bool:
			q.Set(k, fmt.Sprintf("%t", t))
		case *bool:
			if t == nil {
				continue
			}
			q.Set(k, fmt.Sprintf("%t", *t))
		default:
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

func (c *Client) installResources() {
	c.API = &GeneratedAPI{c: c}
	c.Customers = &CustomersResource{c: c}
	c.Plans = &PlansResource{c: c}
	c.CheckoutSessions = &CheckoutSessionsResource{c: c}
	c.Invoices = &InvoicesResource{c: c}
	c.Subscriptions = &SubscriptionsResource{c: c}
	c.PortalSessions = &PortalSessionsResource{c: c}
	c.Receipts = &ReceiptsResource{c: c}
	c.Payouts = &PayoutsResource{c: c}
	c.Ledger = &LedgerResource{c: c}
	c.Reports = &ReportsResource{c: c}
	c.WebhookEndpoints = &WebhookEndpointsResource{c: c}
	c.Events = &EventsResource{c: c}
	c.Refunds = &RefundsResource{c: c}
	c.Adapters = &AdaptersResource{c: c}
	c.ApiKeys = &ApiKeysResource{c: c}
	c.Billing = &BillingResource{c: c}
	c.Onboarding = &OnboardingResource{c: c}
	c.CheckoutSettings = &CheckoutSettingsResource{c: c}
	c.Templates = &TemplatesResource{c: c}
	c.Uploads = &UploadsResource{c: c}
	c.Workspaces = &WorkspacesResource{c: c}
	c.Account = &AccountResource{c: c}
	c.AdminPortal = &AdminPortalResource{c: c}
	c.Admin = &AdminResource{c: c}
}
