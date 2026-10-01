package plugipay

import "context"

// AccountResource — /api/v1/account (merchant profile + sessions + linked accounts)
type AccountResource struct{ c *Client }

type AccountUpdateInput struct {
	Name   *string `json:"name,omitempty"`
	Locale *string `json:"locale,omitempty"`
}

// AccountEmailChangeInput starts an email change: a link goes to Email. Password is the
// current one (ignored for a person without a password).
type AccountEmailChangeInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountEmailChangeResult struct {
	Pending  bool   `json:"pending"`
	NewEmail string `json:"newEmail"`
}

// AccountPasswordChangeInput: CurrentPassword is required unless the person has none yet
// (signs in with Google / Apple only).
type AccountPasswordChangeInput struct {
	CurrentPassword string `json:"currentPassword,omitempty"`
	NewPassword     string `json:"newPassword"`
}

type AccountRevokeAllResult struct {
	RevokedCount int64 `json:"revokedCount"`
}

func (r *AccountResource) Get(ctx context.Context) (*AccountProfile, error) {
	var out AccountProfile
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/account",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes the name or locale and answers with those fields.
func (r *AccountResource) Update(ctx context.Context, patch AccountUpdateInput) (*AccountProfileUpdate, error) {
	var out AccountProfileUpdate
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/account", Body: patch,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AccountResource) ListSessions(ctx context.Context) ([]BrowserSession, error) {
	var out []BrowserSession
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/account/sessions",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AccountResource) RevokeSession(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/account/sessions/" + id + "/revoke",
		Body: map[string]any{},
	}, nil)
}

func (r *AccountResource) RevokeAllSessions(ctx context.Context) (*AccountRevokeAllResult, error) {
	var out AccountRevokeAllResult
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/account/sessions/revoke-all", Body: map[string]any{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListLinked returns the linked Google / Apple sign-ins and whether the person also has
// a password.
func (r *AccountResource) ListLinked(ctx context.Context) (*LinkedAccounts, error) {
	var out LinkedAccounts
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/account/linked-accounts",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AccountResource) Unlink(ctx context.Context, provider string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/account/linked-accounts/" + provider,
	}, nil)
}

func (r *AccountResource) ChangeEmail(ctx context.Context, in AccountEmailChangeInput) (*AccountEmailChangeResult, error) {
	var out AccountEmailChangeResult
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/account/email-change", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AccountResource) ChangePassword(ctx context.Context, in AccountPasswordChangeInput) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/account/password-change", Body: in,
	}, nil)
}

func (r *AccountResource) ListMembers(ctx context.Context) ([]WorkspaceMember, error) {
	var out []WorkspaceMember
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/account/members",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
