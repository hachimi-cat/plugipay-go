package plugipay

import "context"

// TemplatesResource — /api/v1/templates
type TemplatesResource struct{ c *Client }

type TemplateListParams struct {
	Kind *TemplateKind `json:"kind,omitempty"`
}

// TemplateCreateInput creates a receipt, invoice or checkout template. Config holds the
// kind's settings, e.g. {"accentColor": "#0F766E", "footerText": "…"}; nil sends {}
// (every setting at its default).
type TemplateCreateInput struct {
	Kind      TemplateKind   `json:"kind"`
	Name      string         `json:"name"`
	IsDefault *bool          `json:"isDefault,omitempty"`
	Config    map[string]any `json:"config"`
}

// TemplateUpdateInput renames a template or replaces its Config (the whole config:
// settings left out reset to their defaults).
type TemplateUpdateInput struct {
	Name   *string        `json:"name,omitempty"`
	Config map[string]any `json:"config,omitempty"`
}

// TemplatePreviewInput renders a kind + config as HTML with sample data.
type TemplatePreviewInput struct {
	Kind   TemplateKind   `json:"kind"`
	Config map[string]any `json:"config"`
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
	if in.Config == nil {
		in.Config = map[string]any{}
	}
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
		Method: "PATCH", Path: "/api/v1/templates/" + id, Body: patch, IdempotencyKey: genIdem(),
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
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Duplicate copies a template. The API names the copy "<name> (copy)"; a non-empty
// name renames it straight after (a second request).
func (r *TemplatesResource) Duplicate(ctx context.Context, id string, name *string) (*Template, error) {
	var out Template
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/templates/" + id + "/duplicate",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	if name == nil || *name == "" {
		return &out, nil
	}
	return r.Update(ctx, out.ID, TemplateUpdateInput{Name: name})
}

// Preview renders a kind + config as HTML with sample data, without saving anything.
func (r *TemplatesResource) Preview(ctx context.Context, in TemplatePreviewInput) (*TemplatePreviewResult, error) {
	if in.Config == nil {
		in.Config = map[string]any{}
	}
	// The route answers with the page itself (text/html), which Do hands back as a string.
	var html string
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/templates/preview", Body: in,
	}, &html)
	if err != nil {
		return nil, err
	}
	return &TemplatePreviewResult{HTML: html}, nil
}

func (r *TemplatesResource) Delete(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/templates/" + id,
	}, nil)
}
