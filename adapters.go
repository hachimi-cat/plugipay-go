package plugipay

import "context"

// AdaptersResource — /api/v1/adapters (payment provider config)
type AdaptersResource struct{ c *Client }

func (r *AdaptersResource) List(ctx context.Context) ([]AdapterConfig, error) {
	var out []AdapterConfig
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/adapters",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AdaptersResource) UpdateXendit(ctx context.Context, config map[string]any) (*AdapterConfig, error) {
	return r.updateKind(ctx, "xendit", config)
}

func (r *AdaptersResource) UpdatePaypal(ctx context.Context, config map[string]any) (*AdapterConfig, error) {
	return r.updateKind(ctx, "paypal", config)
}

func (r *AdaptersResource) UpdateMidtrans(ctx context.Context, config map[string]any) (*AdapterConfig, error) {
	return r.updateKind(ctx, "midtrans", config)
}

func (r *AdaptersResource) UpdateManual(ctx context.Context, config map[string]any) (*AdapterConfig, error) {
	return r.updateKind(ctx, "manual", config)
}

func (r *AdaptersResource) updateKind(ctx context.Context, kind string, config map[string]any) (*AdapterConfig, error) {
	var out AdapterConfig
	err := r.c.Do(ctx, RequestOptions{
		Method: "PUT", Path: "/api/v1/adapters/" + kind, Body: config,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AdaptersResource) ManagedOnboardingState(ctx context.Context) (*ManagedOnboardingState, error) {
	var out ManagedOnboardingState
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/adapters/managed/onboarding",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type ManagedOnboardingStartInput struct {
	Kind    AdapterKind    `json:"kind"`
	Details map[string]any `json:"details,omitempty"`
}

func (r *AdaptersResource) StartManagedOnboarding(ctx context.Context, in ManagedOnboardingStartInput) (*ManagedOnboardingState, error) {
	if in.Kind == "" {
		in.Kind = AdapterKindXendit
	}
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

type ManagedOnboardingSimulateInput struct {
	Result string `json:"result"` // verified | failed
}

func (r *AdaptersResource) SimulateManagedOnboarding(ctx context.Context, in ManagedOnboardingSimulateInput) (*ManagedOnboardingState, error) {
	if in.Result == "" {
		in.Result = "verified"
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
