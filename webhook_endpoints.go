package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// WebhookEndpoints is the Stripe-style namespace for managing signed webhook
// endpoints: registering HTTPS URLs, rotating signing secrets, sending test
// events, and inspecting delivery history and health.
type WebhookEndpoints struct {
	client transcodelyv1connect.WebhookServiceClient
}

func newWebhookEndpoints(c transcodelyv1connect.WebhookServiceClient) *WebhookEndpoints {
	return &WebhookEndpoints{client: c}
}

// Create registers a new webhook endpoint. The returned endpoint's
// GetSecret() carries the signing secret in plain text — this is the only
// response that ever exposes it, so store it securely on receipt.
func (w *WebhookEndpoints) Create(ctx context.Context, params *WebhookEndpointCreateParams) (*WebhookEndpoint, error) {
	if params == nil {
		params = &WebhookEndpointCreateParams{}
	}
	resp, err := w.client.CreateWebhookEndpoint(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEndpoint(), nil
}

// Retrieve fetches a webhook endpoint by ID (`whe_*`). The secret is never
// populated here; use [WebhookEndpoints.RotateSecret] to obtain a new one.
func (w *WebhookEndpoints) Retrieve(ctx context.Context, id string) (*WebhookEndpoint, error) {
	resp, err := w.client.RetrieveWebhookEndpoint(ctx, connect.NewRequest(&v1.RetrieveWebhookEndpointRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEndpoint(), nil
}

// Update mutates an endpoint's URL, description, enabled events, status, or
// metadata. Only the fields set on params are changed.
func (w *WebhookEndpoints) Update(ctx context.Context, params *WebhookEndpointUpdateParams) (*WebhookEndpoint, error) {
	if params == nil {
		params = &WebhookEndpointUpdateParams{}
	}
	resp, err := w.client.UpdateWebhookEndpoint(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEndpoint(), nil
}

// Delete soft-deletes a webhook endpoint.
func (w *WebhookEndpoints) Delete(ctx context.Context, id string) error {
	_, err := w.client.DeleteWebhookEndpoint(ctx, connect.NewRequest(&v1.DeleteWebhookEndpointRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}

// List returns an iterator over the webhook endpoints registered for an app.
func (w *WebhookEndpoints) List(ctx context.Context, params *WebhookEndpointListParams) *Iter[*WebhookEndpoint] {
	if params == nil {
		params = &WebhookEndpointListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*WebhookEndpoint, string, error) {
		req := proto.Clone(params).(*WebhookEndpointListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := w.client.ListWebhookEndpoints(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetEndpoints(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// RotateSecret rotates the endpoint's signing secret. The previous secret
// stays valid for 24h so in-flight deliveries still verify. The returned
// endpoint's GetSecret() carries the new plain-text secret — the only
// response that exposes it.
func (w *WebhookEndpoints) RotateSecret(ctx context.Context, id string) (*WebhookEndpoint, error) {
	resp, err := w.client.RotateWebhookSecret(ctx, connect.NewRequest(&v1.RotateWebhookSecretRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEndpoint(), nil
}

// SendTest delivers a synthetic event of the given type to a single endpoint
// through the normal delivery pipeline, returning the delivery record to
// inspect. The synthetic event is invisible to [Events.List] and never bumps
// pending-webhook counters. Rate-limited to 10/min per endpoint.
func (w *WebhookEndpoints) SendTest(ctx context.Context, endpointID string, eventType EventType) (*WebhookDelivery, error) {
	resp, err := w.client.SendTestWebhook(ctx, connect.NewRequest(&v1.SendTestWebhookRequest{
		EndpointId: endpointID,
		EventType:  string(eventType),
	}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetDelivery(), nil
}

// ListDeliveries returns an iterator over delivery attempts. Set EndpointId
// (deliveries for one endpoint), EventId (deliveries for one event across all
// subscribers), or both on params; at least one is required server-side.
func (w *WebhookEndpoints) ListDeliveries(ctx context.Context, params *WebhookDeliveryListParams) *Iter[*WebhookDelivery] {
	if params == nil {
		params = &WebhookDeliveryListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*WebhookDelivery, string, error) {
		req := proto.Clone(params).(*WebhookDeliveryListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := w.client.ListWebhookDeliveries(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetDeliveries(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// HealthWindow selects the rolling window for [WebhookEndpoints.GetHealth].
type HealthWindow string

const (
	HealthWindow24h HealthWindow = "24h"
	HealthWindow7d  HealthWindow = "7d"
	HealthWindow30d HealthWindow = "30d"
)

// GetHealth returns aggregate delivery stats for one endpoint over a rolling
// window (defaults to 24h when window is ""). The response is cached
// server-side for ~30s.
func (w *WebhookEndpoints) GetHealth(ctx context.Context, endpointID string, window HealthWindow) (*EndpointHealth, error) {
	req := &v1.GetEndpointHealthRequest{EndpointId: endpointID}
	if window != "" {
		s := string(window)
		req.Window = &s
	}
	resp, err := w.client.GetEndpointHealth(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}
