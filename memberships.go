package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// MembershipRole values (re-exported here for proximity).
type MembershipRole = v1.MembershipRole

const (
	MembershipRoleOwner  = v1.MembershipRole_MEMBERSHIP_ROLE_OWNER
	MembershipRoleAdmin  = v1.MembershipRole_MEMBERSHIP_ROLE_ADMIN
	MembershipRoleMember = v1.MembershipRole_MEMBERSHIP_ROLE_MEMBER
	MembershipRoleViewer = v1.MembershipRole_MEMBERSHIP_ROLE_VIEWER
)

// Memberships is the Stripe-style namespace for org memberships.
type Memberships struct {
	client transcodelyv1connect.MembershipServiceClient
}

func newMemberships(c transcodelyv1connect.MembershipServiceClient) *Memberships {
	return &Memberships{client: c}
}

// List returns an iterator over memberships within the active organization.
// Items include the linked user via MembershipWithUser.
func (m *Memberships) List(ctx context.Context, params *MembershipListParams) *Iter[*MembershipWithUser] {
	if params == nil {
		params = &MembershipListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*MembershipWithUser, string, error) {
		req := proto.Clone(params).(*MembershipListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := m.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetMemberships(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Get fetches a membership by ID (`mem_*`). The result includes the linked user.
func (m *Memberships) Get(ctx context.Context, id string) (*MembershipWithUser, error) {
	resp, err := m.client.Get(ctx, connect.NewRequest(&v1.GetMembershipRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetMembership(), nil
}

// UpdateRole changes a member's role.
func (m *Memberships) UpdateRole(ctx context.Context, id string, role MembershipRole) (*MembershipWithUser, error) {
	req := &v1.UpdateMembershipRoleRequest{Id: id, Role: role}
	resp, err := m.client.UpdateRole(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetMembership(), nil
}

// Remove deletes a membership.
func (m *Memberships) Remove(ctx context.Context, id string) error {
	_, err := m.client.Remove(ctx, connect.NewRequest(&v1.RemoveMembershipRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}
