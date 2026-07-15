package plugipay

import "context"

// UploadsResource — /api/v1/uploads
type UploadsResource struct{ c *Client }

type UploadImageInput struct {
	Filename string `json:"filename"`
	Mime     string `json:"mime"`
	Base64   string `json:"base64"`
}

// Image uploads a base64-encoded image. Pass raw bytes + filename + mime.
func (r *UploadsResource) Image(ctx context.Context, in UploadImageInput) (*UploadedFile, error) {
	var out UploadedFile
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/uploads/image", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// WorkspacesResource — /api/v1/workspaces (merchant-facing CRUD)
type WorkspacesResource struct{ c *Client }

type WorkspaceCreateInput struct {
	BrandName     *string `json:"brandName,omitempty"`
	BusinessEmail *string `json:"businessEmail,omitempty"`
}

type WorkspaceUpdateInput struct {
	BrandName     *string `json:"brandName,omitempty"`
	BusinessEmail *string `json:"businessEmail,omitempty"`
}

func (r *WorkspacesResource) List(ctx context.Context) ([]Workspace, error) {
	var out []Workspace
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/workspaces",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *WorkspacesResource) Create(ctx context.Context, in WorkspaceCreateInput) (*Workspace, error) {
	var out Workspace
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/workspaces",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *WorkspacesResource) Update(ctx context.Context, id string, patch WorkspaceUpdateInput) (*Workspace, error) {
	var out Workspace
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/workspaces/" + id, Body: patch,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *WorkspacesResource) Delete(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/workspaces/" + id,
	}, nil)
}
