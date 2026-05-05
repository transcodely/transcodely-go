package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Users is the Stripe-style namespace for user accounts.
type Users struct {
	client transcodelyv1connect.UserServiceClient
}

func newUsers(c transcodelyv1connect.UserServiceClient) *Users {
	return &Users{client: c}
}

// GetMe returns the user identified by the current API key, including the
// organizations they belong to.
func (u *Users) GetMe(ctx context.Context) (*UserWithOrganizations, error) {
	resp, err := u.client.GetMe(ctx, connect.NewRequest(&v1.GetMeRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetUser(), nil
}

// Get fetches a user by ID (admin-only in most contexts).
func (u *Users) Get(ctx context.Context, id string) (*User, error) {
	resp, err := u.client.Get(ctx, connect.NewRequest(&v1.GetUserRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetUser(), nil
}

// UpdateMe mutates the current user's profile.
func (u *Users) UpdateMe(ctx context.Context, params *UserUpdateMeParams) (*User, error) {
	if params == nil {
		params = &UserUpdateMeParams{}
	}
	resp, err := u.client.UpdateMe(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetUser(), nil
}

// List returns an iterator over users (admin-only in most contexts).
func (u *Users) List(ctx context.Context, params *UserListParams) *Iter[*User] {
	if params == nil {
		params = &UserListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*User, string, error) {
		req := proto.Clone(params).(*UserListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := u.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetUsers(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}
