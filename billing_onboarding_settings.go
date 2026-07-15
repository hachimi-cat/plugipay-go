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

func (r *BillingResource) ListPlans(ctx context.Context) ([]Plan, error) {
	var out []Plan
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

type OnboardingProvisionManagedInput struct {
	BusinessEmail string  `json:"businessEmail"`
	BrandName     *string `json:"brandName,omitempty"`
}

func (r *OnboardingResource) ProvisionManaged(ctx context.Context, in OnboardingProvisionManagedInput) (*Workspace, error) {
	var out Workspace
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/onboarding/provision-managed",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutSettingsResource — /api/v1/checkout/settings
type CheckoutSettingsResource struct{ c *Client }

type CheckoutSettingsUpdateInput struct {
	BrandLogoURL      *string `json:"brandLogoUrl,omitempty"`
	BrandColor        *string `json:"brandColor,omitempty"`
	DefaultTemplateID *string `json:"defaultTemplateId,omitempty"`
	TermsURL          *string `json:"termsUrl,omitempty"`
	PrivacyURL        *string `json:"privacyUrl,omitempty"`
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
