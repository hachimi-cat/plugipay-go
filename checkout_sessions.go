package plugipay

import "context"

// CheckoutSessionsResource — /api/v1/checkout-sessions
type CheckoutSessionsResource struct{ c *Client }

type CheckoutSessionLineItem struct {
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	UnitAmount int64  `json:"unitAmount"`
}

type CheckoutSessionCreateInput struct {
	Amount       int64                     `json:"amount"`
	Currency     CurrencyCode              `json:"currency"`
	Methods      []CheckoutMethod          `json:"methods"`
	SuccessURL   string                    `json:"successUrl"`
	CancelURL    string                    `json:"cancelUrl"`
	CustomerID   *string                   `json:"customerId,omitempty"`
	LineItems    []CheckoutSessionLineItem `json:"lineItems"`
	ExpiresInSec *int                      `json:"expiresInSec,omitempty"`
	Metadata     map[string]string         `json:"metadata,omitempty"`
	TemplateID   *string                   `json:"templateId,omitempty"`
}

type CheckoutSessionListParams struct {
	Limit      *int    `json:"limit,omitempty"`
	Status     *string `json:"status,omitempty"`
	CustomerID *string `json:"customerId,omitempty"`
}

func (r *CheckoutSessionsResource) Create(ctx context.Context, in CheckoutSessionCreateInput) (*CheckoutSession, error) {
	// Node SDK normalizes lineItems to [] (never undefined). Do the same.
	if in.LineItems == nil {
		in.LineItems = []CheckoutSessionLineItem{}
	}
	var out CheckoutSession
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/checkout-sessions",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CheckoutSessionsResource) Get(ctx context.Context, id string) (*CheckoutSession, error) {
	var out CheckoutSession
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/checkout-sessions/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CheckoutSessionsResource) List(ctx context.Context, params CheckoutSessionListParams) (Page[CheckoutSession], error) {
	return DoList[CheckoutSession](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/checkout-sessions" + qs(map[string]any{
			"limit":      params.Limit,
			"status":     params.Status,
			"customerId": params.CustomerID,
		}),
	})
}

func (r *CheckoutSessionsResource) Cancel(ctx context.Context, id string) (*CheckoutSession, error) {
	var out CheckoutSession
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/checkout-sessions/" + id + "/cancel",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CheckoutSessionsResource) Confirm(ctx context.Context, id string) (*CheckoutSession, error) {
	var out CheckoutSession
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/checkout-sessions/" + id + "/confirm",
		Body: map[string]any{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
