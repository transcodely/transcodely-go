package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Origins is the Stripe-style namespace for storage origins (GCS, S3, HTTP).
type Origins struct {
	client transcodelyv1connect.OriginServiceClient
}

func newOrigins(c transcodelyv1connect.OriginServiceClient) *Origins {
	return &Origins{client: c}
}

// Create registers a new origin.
func (o *Origins) Create(ctx context.Context, params *OriginCreateParams) (*Origin, error) {
	if params == nil {
		params = &OriginCreateParams{}
	}
	resp, err := o.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetOrigin(), nil
}

// Get fetches an origin by ID (`ori_*`).
func (o *Origins) Get(ctx context.Context, id string) (*Origin, error) {
	resp, err := o.client.Get(ctx, connect.NewRequest(&v1.GetOriginRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetOrigin(), nil
}

// List returns an iterator over origins in the active app.
func (o *Origins) List(ctx context.Context, params *OriginListParams) *Iter[*Origin] {
	if params == nil {
		params = &OriginListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Origin, string, error) {
		req := proto.Clone(params).(*OriginListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := o.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetOrigins(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Update mutates an origin's metadata or credentials.
func (o *Origins) Update(ctx context.Context, params *OriginUpdateParams) (*Origin, error) {
	if params == nil {
		params = &OriginUpdateParams{}
	}
	resp, err := o.client.Update(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetOrigin(), nil
}

// Validate runs a connectivity / permissions check against the origin.
func (o *Origins) Validate(ctx context.Context, id string) (*ValidationResult, error) {
	resp, err := o.client.Validate(ctx, connect.NewRequest(&v1.ValidateOriginRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetValidation(), nil
}

// Archive soft-deletes an origin.
func (o *Origins) Archive(ctx context.Context, id string) error {
	_, err := o.client.Archive(ctx, connect.NewRequest(&v1.ArchiveOriginRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}
