package plugipay

import "context"

// PortalSessionsResource — /api/v1/portal-sessions
type PortalSessionsResource struct{ c *Client }

type PortalSessionCreateInput struct {
	CustomerID string `json:"customerId"`
	ReturnURL  string `json:"returnUrl"`
}

func (r *PortalSessionsResource) Create(ctx context.Context, in PortalSessionCreateInput) (*PortalSession, error) {
	var out PortalSession
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/portal-sessions",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
