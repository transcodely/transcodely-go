package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// SlugCheckResult describes the outcome of a slug availability check.
type SlugCheckResult struct {
	Available bool
	Reason    string
}

// Organizations is the Stripe-style namespace for organizations.
type Organizations struct {
	client transcodelyv1connect.OrganizationServiceClient
}

func newOrganizations(c transcodelyv1connect.OrganizationServiceClient) *Organizations {
	return &Organizations{client: c}
}

// Create provisions a new organization. The caller becomes its first owner.
func (o *Organizations) Create(ctx context.Context, params *OrgCreateParams) (*Organization, error) {
	if params == nil {
		params = &OrgCreateParams{}
	}
	resp, err := o.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetOrganization(), nil
}

// Get fetches an organization by ID (`org_*`) or slug.
func (o *Organizations) Get(ctx context.Context, idOrSlug string) (*Organization, error) {
	resp, err := o.client.Get(ctx, connect.NewRequest(&v1.GetOrganizationRequest{IdOrSlug: idOrSlug}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetOrganization(), nil
}

// Update mutates organization metadata.
func (o *Organizations) Update(ctx context.Context, params *OrgUpdateParams) (*Organization, error) {
	if params == nil {
		params = &OrgUpdateParams{}
	}
	resp, err := o.client.Update(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetOrganization(), nil
}

// List returns an iterator over organizations the caller can see.
func (o *Organizations) List(ctx context.Context, params *OrgListParams) *Iter[*Organization] {
	if params == nil {
		params = &OrgListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Organization, string, error) {
		req := proto.Clone(params).(*OrgListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := o.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetOrganizations(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// CheckSlug reports whether a candidate slug is free.
func (o *Organizations) CheckSlug(ctx context.Context, slug string) (*SlugCheckResult, error) {
	resp, err := o.client.CheckSlug(ctx, connect.NewRequest(&v1.CheckSlugRequest{Slug: slug}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return &SlugCheckResult{
		Available: resp.Msg.GetAvailable(),
		Reason:    resp.Msg.GetReason(),
	}, nil
}
