package plugipay

import "context"

// CustomersResource — /api/v1/customers
type CustomersResource struct{ c *Client }

type CustomerCreateInput struct {
	Email      *string `json:"email,omitempty"`
	Name       *string `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	ExternalID *string `json:"externalId,omitempty"`
	// TaxID is the customer's tax id (NPWP in Indonesia).
	TaxID                 *string           `json:"taxId,omitempty"`
	DefaultPaymentTokenID *string           `json:"defaultPaymentTokenId,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

// CustomerUpdateInput changes a customer; only the fields set change.
type CustomerUpdateInput struct {
	Email                 *string           `json:"email,omitempty"`
	Name                  *string           `json:"name,omitempty"`
	Phone                 *string           `json:"phone,omitempty"`
	ExternalID            *string           `json:"externalId,omitempty"`
	TaxID                 *string           `json:"taxId,omitempty"`
	DefaultPaymentTokenID *string           `json:"defaultPaymentTokenId,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

type CustomerListParams struct {
	Limit        *int    `json:"limit,omitempty"`
	Cursor       *string `json:"cursor,omitempty"`
	Order        *string `json:"order,omitempty"` // asc or desc (the default)
	Email        *string `json:"email,omitempty"`
	ExternalID   *string `json:"externalId,omitempty"`
	CreatedAfter *string `json:"createdAfter,omitempty"` // ISO-8601
}

func (r *CustomersResource) Create(ctx context.Context, in CustomerCreateInput) (*Customer, error) {
	var out Customer
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/customers",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CustomersResource) Get(ctx context.Context, id string) (*Customer, error) {
	var out Customer
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/customers/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CustomersResource) List(ctx context.Context, params CustomerListParams) (Page[Customer], error) {
	return DoList[Customer](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/customers" + qs(map[string]any{
			"limit":        params.Limit,
			"cursor":       params.Cursor,
			"order":        params.Order,
			"email":        params.Email,
			"externalId":   params.ExternalID,
			"createdAfter": params.CreatedAfter,
		}),
	})
}

func (r *CustomersResource) Update(ctx context.Context, id string, patch CustomerUpdateInput) (*Customer, error) {
	var out Customer
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/customers/" + id, Body: patch, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a customer with no money history (409 while subscriptions, invoices, paid
// or in-flight checkout sessions, or gift cards hold it).
func (r *CustomersResource) Delete(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/customers/" + id, IdempotencyKey: genIdem(),
	}, nil)
}
