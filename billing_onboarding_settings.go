package plugipay

import "context"

// BillingResource — /api/v1/billing (merchant subscription to Plugipay itself)
type BillingResource struct{ c *Client }

type BillingRefreshResult struct {
	Refreshed bool `json:"refreshed"`
}

func (r *BillingResource) ListTiers(ctx context.Context) ([]BillingTier, error) {
	var out []BillingTier
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/billing/tiers",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BillingResource) ListPlans(ctx context.Context) ([]BillingPlan, error) {
	var out []BillingPlan
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/billing/plans",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BillingResource) RefreshTiers(ctx context.Context) (*BillingRefreshResult, error) {
	var out BillingRefreshResult
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/billing/tiers/refresh", Body: map[string]any{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// OnboardingResource — /api/v1/onboarding
type OnboardingResource struct{ c *Client }

// ProvisionManagedResult is the managed (xenPlatform) sub-account provisioned.
type ProvisionManagedResult struct {
	SubAccountID string `json:"subAccountId"`
}

// ProvisionManaged provisions the managed sub-account for the key's workspace. It takes
// no input (the API reads none).
func (r *OnboardingResource) ProvisionManaged(ctx context.Context) (*ProvisionManagedResult, error) {
	var out ProvisionManagedResult
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/onboarding/provision-managed",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutSettingsResource — /api/v1/checkout/settings
type CheckoutSettingsResource struct{ c *Client }

// CheckoutSettingsUpdateInput changes the hosted checkout's settings; only the fields
// set change.
type CheckoutSettingsUpdateInput struct {
	EnabledMethods   []string          `json:"enabledMethods,omitempty"`
	MethodOrder      []string          `json:"methodOrder,omitempty"`
	MethodAdapter    map[string]string `json:"methodAdapter,omitempty"`
	BrandName        *string           `json:"brandName,omitempty"`
	BrandLogoURL     *string           `json:"brandLogoUrl,omitempty"`
	BrandAccentColor *string           `json:"brandAccentColor,omitempty"` // #RRGGBB
	BrandTagline     *string           `json:"brandTagline,omitempty"`
	BusinessPhone    *string           `json:"businessPhone,omitempty"`
	BusinessEmail    *string           `json:"businessEmail,omitempty"`
	BusinessAddress  *string           `json:"businessAddress,omitempty"`
	BusinessTaxID    *string           `json:"businessTaxId,omitempty"`
	ReceiptTemplate  map[string]any    `json:"receiptTemplate,omitempty"`
}

func (r *CheckoutSettingsResource) Get(ctx context.Context) (*CheckoutSettings, error) {
	var out CheckoutSettings
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/checkout/settings",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CheckoutSettingsResource) Update(ctx context.Context, patch CheckoutSettingsUpdateInput) (*CheckoutSettings, error) {
	var out CheckoutSettings
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/checkout/settings", Body: patch,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
