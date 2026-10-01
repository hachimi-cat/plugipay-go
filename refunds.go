package plugipay

import "context"

// RefundsResource — /api/v1/refunds
type RefundsResource struct{ c *Client }

type RefundCreateInput struct {
	SourceType SourceType `json:"sourceType"` // checkout_session | invoice
	SourceID   string     `json:"sourceId"`
	Amount     *int64     `json:"amount,omitempty"`
	Reason     *string    `json:"reason,omitempty"`
}

// RefundListParams filters Refunds.List. ChargeID lists one payment's refunds (a checkout
// session's PaymentID, or its own ID for a manual payment).
type RefundListParams struct {
	Limit    *int          `json:"limit,omitempty"`
	Cursor   *string       `json:"cursor,omitempty"`
	Order    *string       `json:"order,omitempty"` // asc or desc (the default)
	Status   *RefundStatus `json:"status,omitempty"`
	ChargeID *string       `json:"chargeId,omitempty"`
}

func (r *RefundsResource) Create(ctx context.Context, in RefundCreateInput) (*Refund, error) {
	var out Refund
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/refunds",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RefundsResource) Get(ctx context.Context, id string) (*Refund, error) {
	var out Refund
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/refunds/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RefundsResource) List(ctx context.Context, params RefundListParams) (Page[Refund], error) {
	var statusVal any
	if params.Status != nil {
		statusVal = string(*params.Status)
	}
	return DoList[Refund](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/refunds" + qs(map[string]any{
			"limit":    params.Limit,
			"cursor":   params.Cursor,
			"order":    params.Order,
			"status":   statusVal,
			"chargeId": params.ChargeID,
		}),
	})
}
