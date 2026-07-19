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

// UpdateSpendLimit sets or clears an app's monthly transcoding spend cap from an
// explicit params struct. Set Params.MonthlySpendLimitEur (via proto.Float64) to
// a value greater than 0 to set the cap, or leave it nil to clear it and return
// the app to unlimited. SetSpendLimit and ClearSpendLimit are the ergonomic
// shorthands for the two cases.
func (a *Apps) UpdateSpendLimit(ctx context.Context, params *AppUpdateSpendLimitParams) (*App, error) {
	if params == nil {
		params = &AppUpdateSpendLimitParams{}
	}
	resp, err := a.client.UpdateSpendLimit(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApp(), nil
}

// SetSpendLimit sets the app's monthly transcoding spend cap in EUR (must be
// greater than 0). Once recorded spend for the current billing period reaches
// the cap, new jobs are rejected with the "limit_exceeded" error code; in-flight
// jobs are never stopped. Use ClearSpendLimit to return the app to unlimited.
func (a *Apps) SetSpendLimit(ctx context.Context, id string, limitEUR float64) (*App, error) {
	return a.UpdateSpendLimit(ctx, &AppUpdateSpendLimitParams{
		AppId:                id,
		MonthlySpendLimitEur: proto.Float64(limitEUR),
	})
}

// ClearSpendLimit removes the app's monthly spend cap, returning it to unlimited
// (the default). It omits the optional limit field, which the server treats as
// "clear any existing cap".
func (a *Apps) ClearSpendLimit(ctx context.Context, id string) (*App, error) {
	// MonthlySpendLimitEur left nil (absent) = clear the cap.
	return a.UpdateSpendLimit(ctx, &AppUpdateSpendLimitParams{AppId: id})
}

// GetSpend returns the app's current-period transcoding spend against its limit:
// the billing-period bounds, the EUR spent so far, the cap (if set), and whether
// the 80% warning and 100% breach events have fired this period.
func (a *Apps) GetSpend(ctx context.Context, id string) (*AppSpend, error) {
	resp, err := a.client.GetSpend(ctx, connect.NewRequest(&v1.GetSpendRequest{AppId: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}
