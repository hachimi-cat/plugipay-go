package plugipay

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// AdaptersResource — /api/v1/adapters (payment provider config)
type AdaptersResource struct{ c *Client }

// XenditAdapterInput connects Xendit: the secret key (xnd_…) and, optionally, the
// callback token Xendit signs its webhooks with.
type XenditAdapterInput struct {
	SecretKey     string  `json:"secretKey"`
	CallbackToken *string `json:"callbackToken,omitempty"`
}

// PaypalAdapterInput connects PayPal: the REST app's client id and secret; Mode is
// "live" (default) or "sandbox".
type PaypalAdapterInput struct {
	ClientID string  `json:"clientId"`
	Secret   string  `json:"secret"`
	Mode     *string `json:"mode,omitempty"`
}

// MidtransAdapterInput connects Midtrans; Env is "sandbox" (default) or "production".
type MidtransAdapterInput struct {
	ServerKey  string  `json:"serverKey"`
	ClientKey  string  `json:"clientKey"`
	MerchantID string  `json:"merchantId"`
	Env        *string `json:"env,omitempty"`
}

// ManualBankAccount is one account buyers transfer to.
type ManualBankAccount struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountHolder string `json:"accountHolder"`
}

// ManualAdapterInput sets up manual (merchant-confirmed) payments.
type ManualAdapterInput struct {
	BankAccounts     []ManualBankAccount `json:"bankAccounts,omitempty"`
	StaticQRImageURL *string             `json:"staticQrImageUrl,omitempty"`
	Instructions     *string             `json:"instructions,omitempty"`
}

// List returns the connected providers in the key's mode, one per kind, sorted by
// kind. (The API answers an object keyed by kind.)
func (r *AdaptersResource) List(ctx context.Context) ([]AdapterConfig, error) {
	var byKind map[string]AdapterConfig
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/adapters",
	}, &byKind)
	if err != nil {
		return nil, err
	}
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	out := make([]AdapterConfig, 0, len(kinds))
	for _, k := range kinds {
		a := byKind[k]
		if a.Kind == "" {
			a.Kind = k
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *AdaptersResource) UpdateXendit(ctx context.Context, in XenditAdapterInput) (*AdapterConfig, error) {
	return r.updateKind(ctx, "xendit", in)
}

func (r *AdaptersResource) UpdatePaypal(ctx context.Context, in PaypalAdapterInput) (*AdapterConfig, error) {
	return r.updateKind(ctx, "paypal", in)
}

func (r *AdaptersResource) UpdateMidtrans(ctx context.Context, in MidtransAdapterInput) (*AdapterConfig, error) {
	return r.updateKind(ctx, "midtrans", in)
}

func (r *AdaptersResource) UpdateManual(ctx context.Context, in ManualAdapterInput) (*AdapterConfig, error) {
	return r.updateKind(ctx, "manual", in)
}

func (r *AdaptersResource) updateKind(ctx context.Context, kind string, body any) (*AdapterConfig, error) {
	var out AdapterConfig
	err := r.c.Do(ctx, RequestOptions{
		Method: "PUT", Path: "/api/v1/adapters/" + kind, Body: body, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ManagedOnboardingState returns the managed sub-account, or nil before onboarding started.
func (r *AdaptersResource) ManagedOnboardingState(ctx context.Context) (*ManagedOnboardingState, error) {
	var raw json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/adapters/managed/onboarding",
	}, &raw)
	if err != nil || len(raw) == 0 || string(raw) == "null" {
		return nil, err
	}
	var out ManagedOnboardingState
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, newErr(0, "invalid_response", fmt.Sprintf("failed to decode response body: %s", err.Error()))
	}
	return &out, nil
}

// ManagedOnboardingStartInput starts managed payments for the payout account Email.
type ManagedOnboardingStartInput struct {
	Email string `json:"email"`
}

func (r *AdaptersResource) StartManagedOnboarding(ctx context.Context, in ManagedOnboardingStartInput) (*ManagedOnboardingState, error) {
	var out ManagedOnboardingState
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/adapters/managed/onboarding",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ManagedOnboardingSimulateInput sets the sub-account's verification state (staging
// only). All nil means fully verified: KybStatus and CapabilitiesStatus "live",
// PayoutsReady true.
type ManagedOnboardingSimulateInput struct {
	KybStatus          *string `json:"kybStatus,omitempty"`          // not_started | invited | registered | live | rejected
	CapabilitiesStatus *string `json:"capabilitiesStatus,omitempty"` // pending | live | declined | resubmission_required
	PayoutsReady       *bool   `json:"payoutsReady,omitempty"`
}

// SimulateManagedOnboarding is staging only: plugipay.com answers 404.
func (r *AdaptersResource) SimulateManagedOnboarding(ctx context.Context, in ManagedOnboardingSimulateInput) (*ManagedOnboardingState, error) {
	if in.KybStatus == nil && in.CapabilitiesStatus == nil && in.PayoutsReady == nil {
		live, ready := "live", true
		in = ManagedOnboardingSimulateInput{KybStatus: &live, CapabilitiesStatus: &live, PayoutsReady: &ready}
	}
	var out ManagedOnboardingState
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/adapters/managed/onboarding/_simulate",
		Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
