package plugipay

import "context"

// SubscriptionsResource — /api/v1/subscriptions
type SubscriptionsResource struct{ c *Client }

type SubscriptionCreateInput struct {
	CustomerID       string            `json:"customerId"`
	PlanID           string            `json:"planId"`
	PriceID          string            `json:"priceId"`
	TrialDays        *int              `json:"trialDays,omitempty"`
	PaymentTokenID   *string           `json:"paymentTokenId,omitempty"`
	CollectionMethod *string           `json:"collectionMethod,omitempty"` // charge_automatically | send_invoice
	Metadata         map[string]string `json:"metadata,omitempty"`
	InitialDiscount  *int64            `json:"initialDiscount,omitempty"`
}

type SubscriptionListParams struct {
	Limit      *int    `json:"limit,omitempty"`
	Status     *string `json:"status,omitempty"`
	CustomerID *string `json:"customerId,omitempty"`
	PlanID     *string `json:"planId,omitempty"`
}

func (r *SubscriptionsResource) Create(ctx context.Context, in SubscriptionCreateInput) (*Subscription, error) {
	var out Subscription
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/subscriptions",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SubscriptionsResource) Get(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/subscriptions/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SubscriptionsResource) List(ctx context.Context, params SubscriptionListParams) (Page[Subscription], error) {
	return DoList[Subscription](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/subscriptions" + qs(map[string]any{
			"limit":      params.Limit,
			"status":     params.Status,
			"customerId": params.CustomerID,
			"planId":     params.PlanID,
		}),
	})
}

// Cancel — pass "now" or "period_end" (default).
func (r *SubscriptionsResource) Cancel(ctx context.Context, id string, at string) (*Subscription, error) {
	if at == "" {
		at = "period_end"
	}
	var out Subscription
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/subscriptions/" + id + "/cancel",
		Body: map[string]any{"at": at}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Pause — optional resumeAt ISO timestamp.
func (r *SubscriptionsResource) Pause(ctx context.Context, id string, resumeAt *string) (*Subscription, error) {
	body := map[string]any{}
	if resumeAt != nil && *resumeAt != "" {
		body["resumeAt"] = *resumeAt
	}
	var out Subscription
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/subscriptions/" + id + "/pause",
		Body: body, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SubscriptionsResource) Resume(ctx context.Context, id string) (*Subscription, error) {
	var out Subscription
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/subscriptions/" + id + "/resume",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
