package plugipay

import "context"

// PlansResource — /api/v1/plans
type PlansResource struct{ c *Client }

// PortalFeatures is what a plan's customers may do for themselves in the billing portal.
type PortalFeatures struct {
	SelfServeCancel     bool `json:"selfServeCancel"`
	SelfServePause      bool `json:"selfServePause"`
	SelfServeUpgrade    bool `json:"selfServeUpgrade"`
	SelfServeDowngrade  bool `json:"selfServeDowngrade"`
	UpdatePaymentMethod bool `json:"updatePaymentMethod"`
}

// DefaultPortalFeatures is what Plans.Create sends when PlanCreateInput.PortalFeatures
// is nil: the dashboard's defaults (cancel, update the payment method).
var DefaultPortalFeatures = PortalFeatures{SelfServeCancel: true, UpdatePaymentMethod: true}

// PriceTier is one tier of a tiered or volume price. UpTo is a number, or "inf" for
// the last tier.
type PriceTier struct {
	UpTo       any   `json:"upTo"`
	UnitAmount int64 `json:"unitAmount"`
	FlatAmount int64 `json:"flatAmount"`
}

// PriceInput is a price as Plans.Create takes it: UnitAmount (minor units) for "flat"
// and "usage", Tiers for "tiered" and "volume".
type PriceInput struct {
	Currency   CurrencyCode `json:"currency"`
	Model      string       `json:"model"` // flat | tiered | volume | usage
	UnitAmount *int64       `json:"unitAmount,omitempty"`
	Tiers      []PriceTier  `json:"tiers,omitempty"`
	TaxMode    *string      `json:"taxMode,omitempty"` // inclusive | exclusive (default)
	Active     *bool        `json:"active,omitempty"`
}

// PlanCreateInput creates a plan with its prices. For one flat price, leave Prices
// empty and set Currency + Amount (minor units): they become
// Prices: []PriceInput{{Currency, Model: "flat", UnitAmount: &Amount}}.
type PlanCreateInput struct {
	Name     string       `json:"name"`
	Interval string       `json:"interval"` // day | week | month | year
	Prices   []PriceInput `json:"prices"`
	// Shorthand for a single flat price when Prices is empty.
	Currency CurrencyCode `json:"-"`
	Amount   int64        `json:"-"`

	Description   *string `json:"description,omitempty"`
	IntervalCount *int    `json:"intervalCount,omitempty"`
	TrialDays     *int    `json:"trialDays,omitempty"`
	// Nil sends DefaultPortalFeatures.
	PortalFeatures  *PortalFeatures   `json:"portalFeatures,omitempty"`
	DunningPolicyID *string           `json:"dunningPolicyId,omitempty"`
	UsageAggregate  *string           `json:"usageAggregate,omitempty"` // sum | max | last
	MeteredUnit     *string           `json:"meteredUnit,omitempty"`
	Active          *bool             `json:"active,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type PlanUpdateInput struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Active      *bool          `json:"active,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type PlanListParams struct {
	Limit  *int    `json:"limit,omitempty"`
	Active *bool   `json:"active,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
	Order  *string `json:"order,omitempty"` // asc | desc
}

func (r *PlansResource) Create(ctx context.Context, in PlanCreateInput) (*Plan, error) {
	if in.PortalFeatures == nil {
		pf := DefaultPortalFeatures
		in.PortalFeatures = &pf
	}
	if len(in.Prices) == 0 && in.Currency != "" {
		amount := in.Amount
		in.Prices = []PriceInput{{Currency: in.Currency, Model: "flat", UnitAmount: &amount}}
	}
	if in.Prices == nil {
		in.Prices = []PriceInput{}
	}
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/plans", Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PlansResource) Get(ctx context.Context, id string) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/plans/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PlansResource) List(ctx context.Context, params PlanListParams) (Page[Plan], error) {
	return DoList[Plan](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/plans" + qs(map[string]any{
			"limit":  params.Limit,
			"active": params.Active,
			"cursor": params.Cursor,
			"order":  params.Order,
		}),
	})
}

func (r *PlansResource) Update(ctx context.Context, id string, patch PlanUpdateInput) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/plans/" + id, Body: patch, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PlansResource) Archive(ctx context.Context, id string) (*Plan, error) {
	var out Plan
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/plans/" + id + "/archive",
		Body: map[string]any{}, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
