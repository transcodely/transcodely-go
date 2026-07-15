package transcodely

import (
	"context"
	"encoding/json"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Events is the Stripe-style namespace for the event ledger. Reach it via
// client.Events. Each event it returns is the same [WebhookEvent] shape that
// [ConstructEvent] produces from a live delivery, so a handler dispatched on
// Type behaves identically whether the event arrived over HTTP or was pulled
// from the ledger.
type Events struct {
	client transcodelyv1connect.WebhookServiceClient
}

func newEvents(c transcodelyv1connect.WebhookServiceClient) *Events {
	return &Events{client: c}
}

// List returns an auto-paging iterator over the app's events, newest first.
// Filter with params.Type, params.CreatedAfter, and params.CreatedBefore.
func (e *Events) List(ctx context.Context, params *EventListParams) *Iter[*WebhookEvent] {
	if params == nil {
		params = &EventListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*WebhookEvent, string, error) {
		req := proto.Clone(params).(*EventListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := e.client.ListEvents(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		events := make([]*WebhookEvent, 0, len(resp.Msg.GetEvents()))
		for _, pe := range resp.Msg.GetEvents() {
			events = append(events, webhookEventFromProto(pe))
		}
		return events, resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Retrieve fetches a single event by ID (`evt_*`).
func (e *Events) Retrieve(ctx context.Context, id string) (*WebhookEvent, error) {
	resp, err := e.client.RetrieveEvent(ctx, connect.NewRequest(&v1.RetrieveEventRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return webhookEventFromProto(resp.Msg.GetEvent()), nil
}

// Resend re-queues an existing event for delivery. With no endpointIDs it
// resends to every currently-subscribed enabled endpoint; otherwise it targets
// the given subset (max 100). The resend reuses the original event ID, so a
// correct receiver deduplicates it. Returns one delivery per targeted endpoint.
func (e *Events) Resend(ctx context.Context, id string, endpointIDs ...string) ([]*WebhookDelivery, error) {
	req := &v1.ResendEventRequest{Id: id, EndpointIds: endpointIDs}
	resp, err := e.client.ResendEvent(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetDeliveries(), nil
}

// webhookEventFromProto bridges a ledger Event (data as a JSON string,
// created_at as a Timestamp) to the unified [WebhookEvent]. It is tolerant: a
// malformed inner snapshot leaves the typed accessors empty rather than failing
// the whole page.
func webhookEventFromProto(pe *v1.Event) *WebhookEvent {
	if pe == nil {
		return &WebhookEvent{Object: "event"}
	}
	ev := &WebhookEvent{
		ID:              pe.GetId(),
		Type:            EventType(pe.GetType()),
		Object:          "event",
		APIVersion:      pe.GetApiVersion(),
		PendingWebhooks: int(pe.GetPendingWebhooks()),
		Request:         WebhookEventRequest{ID: pe.GetRequestId()},
	}
	if ca := pe.GetCreatedAt(); ca != nil {
		ev.Created = ca.AsTime().UTC().Format(time.RFC3339)
	}
	if data := pe.GetData(); data != "" {
		ev.raw = json.RawMessage(data)
		ev.decodeData()
	}
	return ev
}
