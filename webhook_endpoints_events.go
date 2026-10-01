package plugipay

import (
	"context"
	"net/url"
)

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

// WebhookEndpointUpdateInput is a PATCH: only the fields set change. Active false pauses
// the endpoint (its queued deliveries fail); Active true re-enables it — also after
// Plugipay switched it off for failing — and clears its failure streak.
type WebhookEndpointUpdateInput struct {
	URL         *string  `json:"url,omitempty"`
	Events      []string `json:"events,omitempty"`
	Description *string  `json:"description,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

func (r *WebhookEndpointsResource) Update(ctx context.Context, id string, in WebhookEndpointUpdateInput) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	err := r.c.Do(ctx, RequestOptions{
		Method: "PATCH", Path: "/api/v1/webhooks/" + url.PathEscape(id), Body: in,
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

// WebhookDeliveryListParams filters the delivery log. Status is pending, succeeded or
// failed; page with Cursor while HasMore.
type WebhookDeliveryListParams struct {
	Limit      *int
	Cursor     *string
	EndpointID *string
	Status     *WebhookDeliveryStatus
	Type       *string
}

// ListDeliveries lists the delivery log, newest first: one row per event per endpoint,
// in the key's mode, with every attempt.
func (r *WebhookEndpointsResource) ListDeliveries(ctx context.Context, params WebhookDeliveryListParams) (Page[WebhookDelivery], error) {
	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}
	return DoList[WebhookDelivery](ctx, r.c, RequestOptions{
		Method: "GET",
		Path: "/api/v1/webhooks/deliveries" + qs(map[string]any{
			"limit":      params.Limit,
			"cursor":     params.Cursor,
			"endpointId": params.EndpointID,
			"status":     status,
			"type":       params.Type,
		}),
	})
}

func (r *WebhookEndpointsResource) GetDelivery(ctx context.Context, id string) (*WebhookDelivery, error) {
	var out WebhookDelivery
	err := r.c.Do(ctx, RequestOptions{
		Method: "GET", Path: "/api/v1/webhooks/deliveries/" + url.PathEscape(id),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RetryDelivery queues one more attempt now at a failed delivery (or sends a succeeded
// one again); it goes out within seconds. A 409 (already_queued, endpoint_disabled) when
// it is queued already or its endpoint is off.
func (r *WebhookEndpointsResource) RetryDelivery(ctx context.Context, id string) (*WebhookDelivery, error) {
	var out WebhookDelivery
	err := r.c.Do(ctx, RequestOptions{
		Method: "POST", Path: "/api/v1/webhooks/deliveries/" + url.PathEscape(id) + "/retry", Body: map[string]any{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
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
