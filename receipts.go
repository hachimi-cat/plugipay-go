package plugipay

import (
	"context"
	"encoding/json"
)

// ReceiptsResource — /api/v1/receipts
type ReceiptsResource struct{ c *Client }

type ReceiptListParams struct {
	Limit        *int    `json:"limit,omitempty"`
	Cursor       *string `json:"cursor,omitempty"`
	SourceType   *string `json:"sourceType,omitempty"` // checkout_session | invoice
	CustomerID   *string `json:"customerId,omitempty"`
	IssuedAfter  *string `json:"issuedAfter,omitempty"`
	IssuedBefore *string `json:"issuedBefore,omitempty"`
}

func (r *ReceiptsResource) List(ctx context.Context, params ReceiptListParams) (Page[ReceiptSummary], error) {
	return DoList[ReceiptSummary](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/receipts" + qs(map[string]any{
			"limit":        params.Limit,
			"cursor":       params.Cursor,
			"sourceType":   params.SourceType,
			"customerId":   params.CustomerID,
			"issuedAfter":  params.IssuedAfter,
			"issuedBefore": params.IssuedBefore,
		}),
	})
}

// Get returns the raw receipt payload — the backend shape here is
// richer than ReceiptSummary and intentionally opaque to the SDK.
func (r *ReceiptsResource) Get(ctx context.Context, id string) (json.RawMessage, error) {
	var out json.RawMessage
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/receipts/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
