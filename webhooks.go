package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// WebhookEndpoints is the Stripe-style namespace for signed webhook endpoints,
// test deliveries, delivery history, and endpoint health. Reach it via
// client.WebhookEndpoints.
//
// To verify inbound deliveries, use the package-level [ConstructEvent].
type WebhookEndpoints struct {
	client transcodelyv1connect.WebhookServiceClient
}

func newWebhookEndpoints(c transcodelyv1connect.WebhookServiceClient) *WebhookEndpoints {
	return &WebhookEndpoints{client: c}
}

// Create registers a signed HTTPS endpoint on an app. The returned endpoint's
// GetSecret() carries the `whsec_…` signing secret — this is the only response
// that exposes it, so store it now.
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

// Retrieve fetches an endpoint by ID (`whe_*`). The secret is never included.
func (w *WebhookEndpoints) Retrieve(ctx context.Context, id string) (*WebhookEndpoint, error) {
	resp, err := w.client.RetrieveWebhookEndpoint(ctx, connect.NewRequest(&v1.RetrieveWebhookEndpointRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEndpoint(), nil
}

// Update mutates an endpoint. Only the fields you set are applied; the server
// treats zero-values as "no change" (an empty EnabledEvents is rejected rather
// than clearing the subscription). Pause an endpoint by setting Status to
// "disabled"; re-enable with "enabled".
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

// Delete soft-deletes an endpoint. Delivery history is preserved.
func (w *WebhookEndpoints) Delete(ctx context.Context, id string) error {
	_, err := w.client.DeleteWebhookEndpoint(ctx, connect.NewRequest(&v1.DeleteWebhookEndpointRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}

// List returns an auto-paging iterator over the app's endpoints (secrets
// omitted). Mutate params.Pagination to override the page size or cursor.
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

// RotateSecret issues a new signing secret for an endpoint. The returned
// endpoint's GetSecret() carries the new secret. The previous secret keeps
// signing for 24 hours (both appear as v1 entries in the signature) — verify
// against both through the overlap with [ConstructEventWithSecrets].
func (w *WebhookEndpoints) RotateSecret(ctx context.Context, id string) (*WebhookEndpoint, error) {
	resp, err := w.client.RotateWebhookSecret(ctx, connect.NewRequest(&v1.RotateWebhookSecretRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEndpoint(), nil
}

// SendTest delivers a synthetic event of the given type to one endpoint through
// the normal signed pipeline and returns the resulting delivery to inspect. The
// event type must be concrete — the "*" wildcard is rejected. Rate-limited to
// 10 calls per minute per endpoint; disabled endpoints reject with a
// [PreconditionError].
func (w *WebhookEndpoints) SendTest(ctx context.Context, endpointID string, eventType EventType) (*WebhookDelivery, error) {
	req := &v1.SendTestWebhookRequest{EndpointId: endpointID, EventType: string(eventType)}
	resp, err := w.client.SendTestWebhook(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetDelivery(), nil
}

// ListDeliveries returns an auto-paging iterator over delivery attempts. Set
// params.EndpointId, params.EventId, or both — at least one is required.
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

// GetHealth returns aggregate delivery stats for one endpoint over a rolling
// window ([HealthWindow24h], [HealthWindow7d], or [HealthWindow30d]). An empty
// window defaults to 24h server-side. The response is cached for ~30s.
func (w *WebhookEndpoints) GetHealth(ctx context.Context, endpointID string, window HealthWindow) (*EndpointHealth, error) {
	req := &v1.GetEndpointHealthRequest{EndpointId: endpointID}
	if window != "" {
		req.Window = proto.String(string(window))
	}
	resp, err := w.client.GetEndpointHealth(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}
