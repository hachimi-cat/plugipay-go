package plugipay

import "context"

// ApiKeysResource — /api/v1/api-keys
type ApiKeysResource struct{ c *Client }

// ApiKeyCreateInput: a name, the environment ("test", the default, or "live") and the
// scopes (all merchant scopes when empty).
type ApiKeyCreateInput struct {
	Name        string   `json:"name"`
	Environment *string  `json:"environment,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

func (r *ApiKeysResource) List(ctx context.Context) ([]ApiKey, error) {
	var out []ApiKey
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/api-keys",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ApiKeysResource) Create(ctx context.Context, in ApiKeyCreateInput) (*ApiKey, error) {
	var out ApiKey
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/api-keys",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ApiKeysResource) Revoke(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/api-keys/" + id,
	}, nil)
}
