package plugipay

import "context"

// PlansResource — /api/v1/plans
type PlansResource struct{ c *Client }

type PlanCreateInput struct {
	Name     string       `json:"name"`
	Currency CurrencyCode `json:"currency"`
	Amount   int64        `json:"amount"`
	Interval string       `json:"interval"` // day | week | month | year
}

type PlanUpdateInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Active      *bool          `json:"active,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type PlanListParams struct {
	Limit  *int    `json:"limit,omitempty"`
	Active *bool   `json:"active,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
	Order  *string `json:"order,omitempty"` // asc | desc
}

func (r *PlansResource) Create(ctx context.Context, in PlanCreateInput) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/plans", Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PlansResource) Get(ctx context.Context, id string) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/plans/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PlansResource) List(ctx context.Context, params PlanListParams) (Page[Plan], error) {
	return DoList[Plan](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/plans" + qs(map[string]any{
			"limit":  params.Limit,
			"active": params.Active,
			"cursor": params.Cursor,
			"order":  params.Order,
		}),
	})
}

func (r *PlansResource) Update(ctx context.Context, id string, patch PlanUpdateInput) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/plans/" + id, Body: patch, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PlansResource) Archive(ctx context.Context, id string) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/plans/" + id + "/archive",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
