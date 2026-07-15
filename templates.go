package plugipay

import "context"

// TemplatesResource — /api/v1/templates
type TemplatesResource struct{ c *Client }

type TemplateListParams struct {
	Kind *TemplateKind `json:"kind,omitempty"`
}

type TemplateCreateInput struct {
	Kind     TemplateKind   `json:"kind"`
	Name     string         `json:"name"`
	Document map[string]any `json:"document"`
}

type TemplateUpdateInput struct {
	Name     *string        `json:"name,omitempty"`
	Document map[string]any `json:"document,omitempty"`
}

type TemplatePreviewInput struct {
	Kind       TemplateKind   `json:"kind"`
	Document   map[string]any `json:"document"`
	SampleData map[string]any `json:"sampleData,omitempty"`
}

type TemplatePreviewResult struct {
	HTML string `json:"html"`
}

func (r *TemplatesResource) List(ctx context.Context, params TemplateListParams) ([]Template, error) {
	var kindVal any
	if params.Kind != nil {
		kindVal = string(*params.Kind)
	}
	var out []Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET",
		Path: "/api/v1/templates" + qs(map[string]any{
			"kind": kindVal,
		}),
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TemplatesResource) Get(ctx context.Context, id string) (*Template, error) {
	var out Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/templates/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TemplatesResource) Create(ctx context.Context, in TemplateCreateInput) (*Template, error) {
	var out Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/templates",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TemplatesResource) Update(ctx context.Context, id string, patch TemplateUpdateInput) (*Template, error) {
	var out Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/templates/" + id, Body: patch,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TemplatesResource) MakeDefault(ctx context.Context, id string) (*Template, error) {
	var out Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/templates/" + id + "/make-default",
		Body: map[string]any{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TemplatesResource) Duplicate(ctx context.Context, id string, name *string) (*Template, error) {
	body := map[string]any{}
	if name != nil && *name != "" {
		body["name"] = *name
	}
	var out Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/templates/" + id + "/duplicate",
		Body: body, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TemplatesResource) Preview(ctx context.Context, in TemplatePreviewInput) (*TemplatePreviewResult, error) {
	var out TemplatePreviewResult
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/templates/preview", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TemplatesResource) Delete(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/templates/" + id,
	}, nil)
}
