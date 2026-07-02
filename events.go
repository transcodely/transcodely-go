package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Events is the Stripe-style namespace for querying and replaying events.
// Every event it returns is the same unified [Event] that [ConstructEvent]
// produces, so a handler tested against an event from [Events.Retrieve]
// behaves identically to one driven by a live webhook delivery.
type Events struct {
	client transcodelyv1connect.WebhookServiceClient
}

func newEvents(c transcodelyv1connect.WebhookServiceClient) *Events {
	return &Events{client: c}
}

// Retrieve fetches a single event by ID (`evt_*`).
func (e *Events) Retrieve(ctx context.Context, id string) (*Event, error) {
	resp, err := e.client.RetrieveEvent(ctx, connect.NewRequest(&v1.RetrieveEventRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return eventFromProto(resp.Msg.GetEvent()), nil
}

// List returns an iterator over an app's events, newest first.
func (e *Events) List(ctx context.Context, params *EventListParams) *Iter[*Event] {
	if params == nil {
		params = &EventListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Event, string, error) {
		req := proto.Clone(params).(*EventListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := e.client.ListEvents(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		protos := resp.Msg.GetEvents()
		events := make([]*Event, 0, len(protos))
		for _, p := range protos {
			events = append(events, eventFromProto(p))
		}
		return events, resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Resend re-queues an existing event for delivery, creating new pending
// delivery records (one per target endpoint) the worker picks up immediately.
// Pass endpointIDs to target a subset; omit them to resend to every
// currently-subscribed enabled endpoint.
func (e *Events) Resend(ctx context.Context, id string, endpointIDs ...string) ([]*WebhookDelivery, error) {
	resp, err := e.client.ResendEvent(ctx, connect.NewRequest(&v1.ResendEventRequest{
		Id:          id,
		EndpointIds: endpointIDs,
	}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetDeliveries(), nil
}
