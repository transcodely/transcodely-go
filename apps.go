package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Apps is the Stripe-style namespace for application scopes within an org.
type Apps struct {
	client transcodelyv1connect.AppServiceClient
}

func newApps(c transcodelyv1connect.AppServiceClient) *Apps {
	return &Apps{client: c}
}

// Create registers a new app.
func (a *Apps) Create(ctx context.Context, params *AppCreateParams) (*App, error) {
	if params == nil {
		params = &AppCreateParams{}
	}
	resp, err := a.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApp(), nil
}

// Get fetches an app by ID (`app_*`).
func (a *Apps) Get(ctx context.Context, id string) (*App, error) {
	resp, err := a.client.Get(ctx, connect.NewRequest(&v1.GetAppRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApp(), nil
}

// Update mutates app metadata.
func (a *Apps) Update(ctx context.Context, params *AppUpdateParams) (*App, error) {
	if params == nil {
		params = &AppUpdateParams{}
	}
	resp, err := a.client.Update(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApp(), nil
}

// List returns an iterator over apps the caller can see.
func (a *Apps) List(ctx context.Context, params *AppListParams) *Iter[*App] {
	if params == nil {
		params = &AppListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*App, string, error) {
		req := proto.Clone(params).(*AppListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := a.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetApps(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Archive soft-deletes an app.
func (a *Apps) Archive(ctx context.Context, id string) error {
	_, err := a.client.Archive(ctx, connect.NewRequest(&v1.ArchiveAppRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}

// EnableHosting turns on the managed hosting experience for an app.
func (a *Apps) EnableHosting(ctx context.Context, id string) (*App, error) {
	resp, err := a.client.EnableHosting(ctx, connect.NewRequest(&v1.EnableHostingRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApp(), nil
}

// UpdateHostingConfig mutates the managed-hosting configuration of an app.
func (a *Apps) UpdateHostingConfig(ctx context.Context, params *AppUpdateHostingConfigParams) (*App, error) {
	if params == nil {
		params = &AppUpdateHostingConfigParams{}
	}
	resp, err := a.client.UpdateHostingConfig(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApp(), nil
}
