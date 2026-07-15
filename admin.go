package plugipay

import (
	"context"
	"encoding/json"
)

// AdminPortalResource — /api/v1/admin-portal (Plugipay internal operators)
type AdminPortalResource struct{ c *Client }

func (r *AdminPortalResource) Me(ctx context.Context) (*AdminPortalIdentity, error) {
	var out AdminPortalIdentity
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/admin-portal/me",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AdminPortalResource) ListBillingAccounts(ctx context.Context) ([]json.RawMessage, error) {
	var out []json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/admin-portal/billing-accounts",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AdminPortalResource) UpdateBillingAccount(ctx context.Context, accountID string, patch map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/admin-portal/billing-accounts/" + accountID, Body: patch,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AdminPortalResource) ListPartners(ctx context.Context) ([]json.RawMessage, error) {
	var out []json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/admin-portal/partners",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AdminPortalResource) CreatePartner(ctx context.Context, in map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/admin-portal/partners", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AdminPortalResource) UpdatePartner(ctx context.Context, id string, patch map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/admin-portal/partners/" + id, Body: patch,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AdminPortalResource) DeletePartner(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/admin-portal/partners/" + id,
	}, nil)
}

// AdminResource — /api/v1/admin (requires plugipay:platform:admin scope)
type AdminResource struct{ c *Client }

type AdminProvisionWorkspaceInput struct {
	AccountID     string  `json:"accountId"`
	Partner       string  `json:"partner"` // storlaunch | fulkruma | ripllo
	DiscountRate  float64 `json:"discountRate"`
	BrandName     *string `json:"brandName,omitempty"`
	BusinessEmail *string `json:"businessEmail,omitempty"`
}

type AdminPartnerUsageParams struct {
	Partner string // storlaunch | fulkruma | ripllo
	From    string
	To      string
}

func (r *AdminResource) ProvisionWorkspace(ctx context.Context, in AdminProvisionWorkspaceInput) (*PartnerWorkspace, error) {
	var out PartnerWorkspace
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/admin/workspaces", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AdminResource) GetWorkspace(ctx context.Context, accountID string) (*PartnerWorkspace, error) {
	var out PartnerWorkspace
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/admin/workspaces/" + accountID,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AdminResource) PartnerUsage(ctx context.Context, p AdminPartnerUsageParams) (*PartnerUsageSummary, error) {
	var out PartnerUsageSummary
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET",
		Path: "/api/v1/admin/partner/usage" + qs(map[string]any{
			"partner": p.Partner,
			"from":    p.From,
			"to":      p.To,
		}),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
