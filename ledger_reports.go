package plugipay

import "context"

// LedgerResource — /api/v1/ledger
type LedgerResource struct{ c *Client }

type LedgerListParams struct {
	Limit      *int    `json:"limit,omitempty"`
	Cursor     *string `json:"cursor,omitempty"`
	Order      *string `json:"order,omitempty"`
	TxID       *string `json:"txId,omitempty"`
	Code       *string `json:"code,omitempty"`
	SourceType *string `json:"sourceType,omitempty"`
	SourceID   *string `json:"sourceId,omitempty"`
}

func (r *LedgerResource) List(ctx context.Context, params LedgerListParams) (Page[LedgerEntry], error) {
	return DoList[LedgerEntry](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/ledger" + qs(map[string]any{
			"limit":      params.Limit,
			"cursor":     params.Cursor,
			"order":      params.Order,
			"txId":       params.TxID,
			"code":       params.Code,
			"sourceType": params.SourceType,
			"sourceId":   params.SourceID,
		}),
	})
}

func (r *LedgerResource) Balances(ctx context.Context) ([]LedgerBalance, error) {
	var out []LedgerBalance
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/ledger/balances",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReportsResource — /api/v1/reports
type ReportsResource struct{ c *Client }

type ReportRangeParams struct {
	From     string  // ISO date
	To       string  // ISO date
	Currency *string // optional
}

func (r *ReportsResource) PnL(ctx context.Context, p ReportRangeParams) (*PnLReport, error) {
	var out PnLReport
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET",
		Path: "/api/v1/reports/pnl" + qs(map[string]any{
			"from":     p.From,
			"to":       p.To,
			"currency": p.Currency,
		}),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ReportsResource) CashFlow(ctx context.Context, p ReportRangeParams) (*CashFlowReport, error) {
	var out CashFlowReport
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET",
		Path: "/api/v1/reports/cash-flow" + qs(map[string]any{
			"from":     p.From,
			"to":       p.To,
			"currency": p.Currency,
		}),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
