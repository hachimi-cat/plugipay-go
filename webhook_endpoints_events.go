package plugipay

import "context"

// WebhookEndpointsResource — /api/v1/webhooks
type WebhookEndpointsResource struct{ c *Client }

type WebhookEndpointCreateInput struct {
	URL         string   `json:"url"`
	Events      []string `json:"events,omitempty"`
	Description *string  `json:"description,omitempty"`
}

func (r *WebhookEndpointsResource) List(ctx context.Context) ([]WebhookEndpoint, error) {
	var out []WebhookEndpoint
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/webhooks",
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *WebhookEndpointsResource) Create(ctx context.Context, in WebhookEndpointCreateInput) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/webhooks",
		Body: in, IdempotencyKey: genIdem(),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *WebhookEndpointsResource) Delete(ctx context.Context, id string) error {
	return r.c.Do(ctx, RequestOptions{
		Method: "DELETE", Path: "/api/v1/webhooks/" + id,
	}, nil)
}

// EventsResource — /api/v1/events
type EventsResource struct{ c *Client }

type EventListParams struct {
	Limit          *int    `json:"limit,omitempty"`
	Cursor         *string `json:"cursor,omitempty"`
	Order          *string `json:"order,omitempty"`
	Type           *string `json:"type,omitempty"`
	OccurredAfter  *string `json:"occurredAfter,omitempty"`
	OccurredBefore *string `json:"occurredBefore,omitempty"`
}

func (r *EventsResource) List(ctx context.Context, params EventListParams) (Page[EventRecord], error) {
	return DoList[EventRecord](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/events" + qs(map[string]any{
			"limit":          params.Limit,
			"cursor":         params.Cursor,
			"order":          params.Order,
			"type":           params.Type,
			"occurredAfter":  params.OccurredAfter,
			"occurredBefore": params.OccurredBefore,
		}),
	})
}

func (r *EventsResource) Get(ctx context.Context, id string) (*EventRecord, error) {
	var out EventRecord
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/events/" + id,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
