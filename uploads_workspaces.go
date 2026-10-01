package plugipay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
)

// UploadsResource — /api/v1/uploads
type UploadsResource struct{ c *Client }

// UploadImageInput is an image for UploadsResource.Image: its bytes (File), or —
// the 0.2 shape — base64 text (Base64 with Filename and Mime).
type UploadImageInput struct {
	// File is the image (PNG, JPEG or WEBP, at most 5 MB).
	File io.Reader
	// Filename is the name it is sent under; "image" when empty.
	Filename string
	// ContentType of the file part; optional — Plugipay tells the type from the bytes.
	ContentType string
	// Base64 is the image as base64 text, used when File is nil.
	Base64 string
	// Mime is ContentType's 0.2 name.
	Mime string
}

// Image uploads an image, sent as multipart/form-data with the file in the
// field "file". It returns the image's URL (relative, served by plugipay.com).
func (r *UploadsResource) Image(ctx context.Context, in UploadImageInput) (*UploadedFile, error) {
	content := in.File
	if content == nil && in.Base64 != "" {
		raw, err := base64.StdEncoding.DecodeString(in.Base64)
		if err != nil {
			return nil, newErr(0, "invalid_request", "UploadImageInput.Base64 is not base64: "+err.Error())
		}
		content = strings.NewReader(string(raw))
	}
	if content == nil {
		return nil, newErr(0, "invalid_request", "UploadImageInput needs File (or Base64)")
	}
	name := in.Filename
	if name == "" {
		name = "image"
	}
	ct := in.ContentType
	if ct == "" {
		ct = in.Mime
	}
	data, err := r.c.sendForm(ctx, "POST", "/api/v1/uploads/image", nil,
		[]formPart{{field: "file", filename: name, contentType: ct, content: content}})
	if err != nil {
		return nil, err
	}
	var out UploadedFile
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, newErr(0, "invalid_response", "failed to decode the upload: "+err.Error())
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
