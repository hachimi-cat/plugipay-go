package plugipay

import "context"

// PayoutsResource — /api/v1/payouts
type PayoutsResource struct{ c *Client }

type PayoutCreateInput struct {
	Amount            int64        `json:"amount"`
	Currency          CurrencyCode `json:"currency"`
	BankCode          *string      `json:"bankCode,omitempty"`
	BankName          *string      `json:"bankName,omitempty"`
	BankAccountNumber *string      `json:"bankAccountNumber,omitempty"`
	BankAccountHolder *string      `json:"bankAccountHolder,omitempty"`
	Note              *string      `json:"note,omitempty"`
}

type PayoutListParams struct {
	Limit  *int          `json:"limit,omitempty"`
	Cursor *string       `json:"cursor,omitempty"`
	Status *PayoutStatus `json:"status,omitempty"`
}

type BankAccountInput struct {
	BankCode          *string `json:"bankCode,omitempty"`
	BankName          string  `json:"bankName"`
	BankAccountNumber string  `json:"bankAccountNumber"`
	BankAccountHolder string  `json:"bankAccountHolder"`
}

func (r *PayoutsResource) Create(ctx context.Context, in PayoutCreateInput) (*Payout, error) {
	var out Payout
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/payouts",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) Get(ctx context.Context, id string) (*Payout, error) {
	var out Payout
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/payouts/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) List(ctx context.Context, params PayoutListParams) (Page[Payout], error) {
	var statusVal any
	if params.Status != nil {
		statusVal = string(*params.Status)
	}
	return DoList[Payout](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/payouts" + qs(map[string]any{
			"limit":  params.Limit,
			"cursor": params.Cursor,
			"status": statusVal,
		}),
	})
}

func (r *PayoutsResource) Cancel(ctx context.Context, id string) (*Payout, error) {
	var out Payout
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/payouts/" + id + "/cancel",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) MarkInTransit(ctx context.Context, id string, reference *string) (*Payout, error) {
	body := map[string]any{"reference": nil}
	if reference != nil {
		body["reference"] = *reference
	}
	var out Payout
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/payouts/" + id + "/mark-in-transit",
		Body: body, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) MarkPaid(ctx context.Context, id string, reference *string) (*Payout, error) {
	body := map[string]any{"reference": nil}
	if reference != nil {
		body["reference"] = *reference
	}
	var out Payout
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/payouts/" + id + "/mark-paid",
		Body: body, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) MarkFailed(ctx context.Context, id string, failureReason string) (*Payout, error) {
	var out Payout
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/payouts/" + id + "/mark-failed",
		Body:           map[string]any{"failureReason": failureReason},
		IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) Balance(ctx context.Context) (*AvailableBalance, error) {
	var out AvailableBalance
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/payouts/balance",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) GetBankAccount(ctx context.Context) (*BankAccount, error) {
	var out BankAccount
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/payouts/bank-account",
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PayoutsResource) UpdateBankAccount(ctx context.Context, in BankAccountInput) (*BankAccount, error) {
	var out BankAccount
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/payouts/bank-account", Body: in,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
