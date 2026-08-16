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
// Invoices are read-only. An invoice is the record of what happened in a
// billing period, generated automatically when the period closes; there is no
// API to create, edit, or delete one. The two writes this namespace does have
// do not author statements either — [Billing.UpdateBudget] sets the customer's
// own alerting threshold, and [Billing.SettleOutstandingBalance] closes the
// period early so the balance owed can be charged now.
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

// GetBudget returns the organization's monthly budget together with the spend
// it is measured against — everything a budget card needs, in one call.
//
// A budget is telemetry the customer sets for themselves: crossing 100% sends
// an email and changes nothing else. No job is refused, no video stops playing.
// The hard cap is the per-app spend limit ([Apps.SetSpendLimit]), which is a
// different number on a different scope.
//
// Always returns a budget. An organization that has never set one gets AmountEur
// absent with SpentEur still populated, so the current period's spend can be
// shown before a budget exists.
func (b *Billing) GetBudget(ctx context.Context) (*Budget, error) {
	resp, err := b.client.GetBudget(ctx, connect.NewRequest(&v1.GetBudgetRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetBudget(), nil
}

// UpdateBudget sets or clears the organization's monthly budget and returns it
// with current-period spend recomputed. Omitting params.AmountEur clears the
// budget; prefer [Billing.SetBudget] and [Billing.ClearBudget] for that.
//
// Changing the amount never re-sends an alert step already sent this period —
// see Budget.NotifiedSteps.
func (b *Billing) UpdateBudget(ctx context.Context, params *BudgetUpdateParams) (*Budget, error) {
	if params == nil {
		params = &BudgetUpdateParams{}
	}
	resp, err := b.client.UpdateBudget(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetBudget(), nil
}

// SetBudget sets the organization's monthly budget in EUR (must be greater than
// 0). Alert emails go out at 50%, 80% and 100% of it, once each per billing
// period. Nothing is enforced at any step.
func (b *Billing) SetBudget(ctx context.Context, amountEUR float64) (*Budget, error) {
	return b.UpdateBudget(ctx, &BudgetUpdateParams{AmountEur: proto.Float64(amountEUR)})
}

// ClearBudget removes the organization's monthly budget, which is the only way
// to turn the alert emails off. It omits the optional amount field, which the
// server treats as "clear any existing budget". The returned Budget still
// carries the period's spend, with AmountEur absent.
func (b *Billing) ClearBudget(ctx context.Context) (*Budget, error) {
	return b.UpdateBudget(ctx, &BudgetUpdateParams{})
}

// GetOutstandingBalance returns usage that has been accrued but not yet billed,
// together with the threshold it is measured against.
//
// Distinct from [Billing.GetUpcomingInvoice], which shows what the CURRENT
// PERIOD has accrued: this shows what is UNSETTLED, which also carries anything
// an earlier period left uncaptured, and it is the number that decides whether
// new jobs are admitted. It drops to zero when a statement is paid; Budget's
// SpentEur does not.
//
// Reminder emails go out at 80%, 100%, 125%, 150% and 175% of the threshold and
// nothing is restricted at any of them. Only at HardStopCents — twice the
// threshold — does Blocked become true and new jobs are refused with error code
// "outstanding_balance_exceeded". Even then queued work finishes, videos keep
// playing, uploads keep working, and jobs can still be canceled.
//
// Always returns a balance. An organization that owes nothing gets one with
// OutstandingCents 0 and its threshold still populated.
func (b *Billing) GetOutstandingBalance(ctx context.Context) (*OutstandingBalance, error) {
	resp, err := b.client.GetOutstandingBalance(ctx, connect.NewRequest(&v1.GetOutstandingBalanceRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetBalance(), nil
}

// SettleOutstandingBalance pays the outstanding balance now, without waiting for
// the period to end. It closes the current period at this instant and produces a
// real statement for everything owed, which the payment provider then charges.
//
// It takes no amount on purpose: the figure is whatever the ledger says when the
// settlement runs, so a stale page cannot pay less than is owed. On success the
// balance is zero, any admission block is lifted immediately, and the returned
// result carries both the new statement and the new balance — no second call.
//
// Check OutstandingBalance.SettlementAvailable before offering this: a
// deployment with the settlement rail switched off returns a
// [PreconditionError] with error code "settlement_unavailable", and one with
// nothing owed returns "nothing_outstanding".
func (b *Billing) SettleOutstandingBalance(ctx context.Context) (*SettlementResult, error) {
	resp, err := b.client.SettleOutstandingBalance(ctx, connect.NewRequest(&v1.SettleOutstandingBalanceRequest{}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}
