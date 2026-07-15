package plugipay

import "context"

// InvoicesResource — /api/v1/invoices
type InvoicesResource struct{ c *Client }

type InvoiceCreateLine struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	UnitAmount  int64  `json:"unitAmount"`
}

type InvoiceCreateInput struct {
	CustomerID string              `json:"customerId"`
	Currency   CurrencyCode        `json:"currency"`
	Lines      []InvoiceCreateLine `json:"lines"`
	Discount   *int64              `json:"discount,omitempty"`
	Tax        *int64              `json:"tax,omitempty"`
	DueAt      *string             `json:"dueAt,omitempty"`
	Status     *string             `json:"status,omitempty"` // draft | open
	Memo       *string             `json:"memo,omitempty"`
}

type InvoiceListParams struct {
	Limit      *int    `json:"limit,omitempty"`
	Cursor     *string `json:"cursor,omitempty"`
	Status     *string `json:"status,omitempty"`
	CustomerID *string `json:"customerId,omitempty"`
}

func (r *InvoicesResource) Create(ctx context.Context, in InvoiceCreateInput) (*Invoice, error) {
	var out Invoice
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/invoices", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *InvoicesResource) Get(ctx context.Context, id string) (*Invoice, error) {
	var out Invoice
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/invoices/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *InvoicesResource) List(ctx context.Context, params InvoiceListParams) (Page[Invoice], error) {
	return DoList[Invoice](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/invoices" + qs(map[string]any{
			"limit":      params.Limit,
			"cursor":     params.Cursor,
			"status":     params.Status,
			"customerId": params.CustomerID,
		}),
	})
}

func (r *InvoicesResource) Finalize(ctx context.Context, id string) (*Invoice, error) {
	var out Invoice
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/invoices/" + id + "/finalize",
		Body: map[string]any{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *InvoicesResource) Pay(ctx context.Context, id string) (*Invoice, error) {
	var out Invoice
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/invoices/" + id + "/pay",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *InvoicesResource) Void(ctx context.Context, id string) (*Invoice, error) {
	var out Invoice
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/invoices/" + id + "/void",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// InvoiceSendEmailResult — what /send-email returns.
type InvoiceSendEmailResult struct {
	Sent bool   `json:"sent"`
	To   string `json:"to"`
}

func (r *InvoicesResource) SendEmail(ctx context.Context, id string, to *string) (*InvoiceSendEmailResult, error) {
	body := map[string]any{}
	if to != nil && *to != "" {
		body["to"] = *to
	}
	var out InvoiceSendEmailResult
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/invoices/" + id + "/send-email",
		Body: body,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
