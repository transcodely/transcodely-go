package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Billing is the Stripe-style namespace for an organization's billing
// statements. Reach it via client.Billing.
//
// Unlike every other namespace, billing is settled for a whole organization
// rather than a single app, so it is NOT available to API-key clients: an API
// key is scoped to one app, and there is no app-scoped subset of an invoice
// worth serving. Calls made with an API key are rejected with a
// [PermissionError].
//
// To read invoices you need a dashboard session token for an organization
// OWNER, plus the organization the request is for:
//
//	client, err := transcodely.New(sessionToken,
//	    transcodely.WithOrganization("org_f6g7h8i9j0"))
//
// Everything here is read-only. An invoice is the record of what happened in a
// billing period, generated automatically when the period closes; there is no
// API to create, edit, or delete one.
type Billing struct {
	client transcodelyv1connect.BillingServiceClient
}

func newBilling(c transcodelyv1connect.BillingServiceClient) *Billing {
	return &Billing{client: c}
}

// ListInvoices returns an auto-paging iterator over the organization's
// finalized invoices, newest period first. Line items are omitted; fetch a
// single invoice with [Billing.GetInvoice] to get them.
//
// A statement still being generated for a just-ended period is never returned.
// For the period currently accruing use [Billing.GetUpcomingInvoice].
func (b *Billing) ListInvoices(ctx context.Context, params *InvoiceListParams) *Iter[*Invoice] {
	if params == nil {
		params = &InvoiceListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Invoice, string, error) {
		req := proto.Clone(params).(*InvoiceListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := b.client.ListInvoices(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetInvoices(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// GetInvoice fetches one invoice by ID (`inv_*`), including its line items.
func (b *Billing) GetInvoice(ctx context.Context, id string) (*Invoice, error) {
	resp, err := b.client.GetInvoice(ctx, connect.NewRequest(&v1.GetInvoiceRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetInvoice(), nil
}

// GetUpcomingInvoice returns the statement for the period currently accruing.
//
// It is computed live from settled jobs rather than stored, so its Id is empty,
// its Status is draft, and its totals move as jobs finish. Jobs still running
// are not included at any price — a job is billed only once it settles.
func (b *Billing) GetUpcomingInvoice(ctx context.Context) (*Invoice, error) {
	resp, err := b.client.GetUpcomingInvoice(ctx, connect.NewRequest(&v1.GetUpcomingInvoiceRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetInvoice(), nil
}

// GetBillingProfile reports whether the organization has a payment method on
// file, along with whatever the payment provider will say about it for display.
//
// Read-only and side-effect free: it never creates provider resources. An
// organization that has never touched billing reports
// [PaymentMethodStateNone] and no payment methods.
//
// The state is the only reliable signal. Card brand and last4 are often absent
// even for a working card, so branch on PaymentMethodState rather than on
// whether the digits arrived.
func (b *Billing) GetBillingProfile(ctx context.Context) (*BillingProfile, error) {
	resp, err := b.client.GetBillingProfile(ctx, connect.NewRequest(&v1.GetBillingProfileRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetProfile(), nil
}

// CreateBillingPortalSession returns a short-lived URL for the payment
// provider's hosted billing portal — where a customer adds or replaces a
// payment method and finds their receipts.
//
// The first call for an organization also links it to the payment provider, so
// the portal comes back already attached to this organization's billing
// account. Safe to call repeatedly.
//
// The session is single-use and expires; request a fresh one per visit instead
// of storing the URL.
func (b *Billing) CreateBillingPortalSession(ctx context.Context) (*BillingPortalSession, error) {
	resp, err := b.client.CreateBillingPortalSession(ctx, connect.NewRequest(&v1.CreateBillingPortalSessionRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetSession(), nil
}
