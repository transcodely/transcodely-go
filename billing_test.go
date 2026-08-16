package transcodely

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// fakeBillingClient stubs BillingServiceClient the same way fakeJobClient does:
// it embeds the generated interface so an unimplemented call nil-panics instead
// of silently passing.
type fakeBillingClient struct {
	transcodelyv1connect.BillingServiceClient

	gotList     *v1.ListInvoicesRequest
	gotGet      *v1.GetInvoiceRequest
	listPages   []*v1.ListInvoicesResponse
	listCalls   int
	invoice     *v1.Invoice
	upcoming    *v1.Invoice
	getInvErr   error
	upcomingErr error
	profile     *v1.BillingProfile
	portal      *v1.BillingPortalSession

	gotBudget  *v1.UpdateBudgetRequest
	budget     *v1.Budget
	balance    *v1.OutstandingBalance
	settlement *v1.SettleOutstandingBalanceResponse
	settleErr  error
}

func (f *fakeBillingClient) ListInvoices(_ context.Context, req *connect.Request[v1.ListInvoicesRequest]) (*connect.Response[v1.ListInvoicesResponse], error) {
	f.gotList = req.Msg
	page := f.listPages[f.listCalls]
	f.listCalls++
	return connect.NewResponse(page), nil
}

func (f *fakeBillingClient) GetInvoice(_ context.Context, req *connect.Request[v1.GetInvoiceRequest]) (*connect.Response[v1.GetInvoiceResponse], error) {
	f.gotGet = req.Msg
	if f.getInvErr != nil {
		return nil, f.getInvErr
	}
	return connect.NewResponse(&v1.GetInvoiceResponse{Invoice: f.invoice}), nil
}

func (f *fakeBillingClient) GetUpcomingInvoice(_ context.Context, _ *connect.Request[v1.GetUpcomingInvoiceRequest]) (*connect.Response[v1.GetUpcomingInvoiceResponse], error) {
	if f.upcomingErr != nil {
		return nil, f.upcomingErr
	}
	return connect.NewResponse(&v1.GetUpcomingInvoiceResponse{Invoice: f.upcoming}), nil
}

func (f *fakeBillingClient) GetBillingProfile(_ context.Context, _ *connect.Request[v1.GetBillingProfileRequest]) (*connect.Response[v1.GetBillingProfileResponse], error) {
	return connect.NewResponse(&v1.GetBillingProfileResponse{Profile: f.profile}), nil
}

func (f *fakeBillingClient) CreateBillingPortalSession(_ context.Context, _ *connect.Request[v1.CreateBillingPortalSessionRequest]) (*connect.Response[v1.CreateBillingPortalSessionResponse], error) {
	return connect.NewResponse(&v1.CreateBillingPortalSessionResponse{Session: f.portal}), nil
}

func (f *fakeBillingClient) GetBudget(_ context.Context, _ *connect.Request[v1.GetBudgetRequest]) (*connect.Response[v1.GetBudgetResponse], error) {
	return connect.NewResponse(&v1.GetBudgetResponse{Budget: f.budget}), nil
}

func (f *fakeBillingClient) UpdateBudget(_ context.Context, req *connect.Request[v1.UpdateBudgetRequest]) (*connect.Response[v1.UpdateBudgetResponse], error) {
	f.gotBudget = req.Msg
	return connect.NewResponse(&v1.UpdateBudgetResponse{Budget: f.budget}), nil
}

func (f *fakeBillingClient) GetOutstandingBalance(_ context.Context, _ *connect.Request[v1.GetOutstandingBalanceRequest]) (*connect.Response[v1.GetOutstandingBalanceResponse], error) {
	return connect.NewResponse(&v1.GetOutstandingBalanceResponse{Balance: f.balance}), nil
}

func (f *fakeBillingClient) SettleOutstandingBalance(_ context.Context, _ *connect.Request[v1.SettleOutstandingBalanceRequest]) (*connect.Response[v1.SettleOutstandingBalanceResponse], error) {
	if f.settleErr != nil {
		return nil, f.settleErr
	}
	return connect.NewResponse(f.settlement), nil
}

// ListInvoices auto-pages: the cursor from page one is sent on the next fetch,
// and every invoice is yielded exactly once across the page boundary.
func TestBilling_ListInvoices_AutoPages(t *testing.T) {
	next := "cursor_page2"
	fake := &fakeBillingClient{
		listPages: []*v1.ListInvoicesResponse{
			{
				Invoices:   []*v1.Invoice{{Id: "inv_one"}, {Id: "inv_two"}},
				Pagination: &v1.PaginationResponse{NextCursor: next},
			},
			{
				Invoices:   []*v1.Invoice{{Id: "inv_three"}},
				Pagination: &v1.PaginationResponse{},
			},
		},
	}
	b := newBilling(fake)

	var ids []string
	iter := b.ListInvoices(context.Background(), nil)
	for iter.Next() {
		ids = append(ids, iter.Current().GetId())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iteration error = %v", err)
	}

	want := []string{"inv_one", "inv_two", "inv_three"}
	if len(ids) != len(want) {
		t.Fatalf("got %d invoices %v, want %d", len(ids), ids, len(want))
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("invoice[%d] = %q, want %q", i, ids[i], want[i])
		}
	}
	if fake.listCalls != 2 {
		t.Errorf("ListInvoices called %d times, want 2", fake.listCalls)
	}
	if got := fake.gotList.GetPagination().GetCursor(); got != next {
		t.Errorf("second page cursor = %q, want %q", got, next)
	}
}

// GetInvoice forwards the id and unwraps the invoice with its line items.
func TestBilling_GetInvoice(t *testing.T) {
	fake := &fakeBillingClient{
		invoice: &v1.Invoice{
			Id:         "inv_a1b2c3d4e5f6",
			Object:     "invoice",
			Status:     InvoiceStatusOpen,
			TotalCents: 1250,
			LineItems: []*v1.InvoiceLineItem{
				{Id: "li_x", LineType: InvoiceLineTypeUsage, AmountCents: 1300},
				{Id: "li_y", LineType: InvoiceLineTypeAdjustment, AmountCents: -50},
			},
		},
	}
	b := newBilling(fake)

	inv, err := b.GetInvoice(context.Background(), "inv_a1b2c3d4e5f6")
	if err != nil {
		t.Fatalf("GetInvoice() error = %v", err)
	}
	if fake.gotGet.GetId() != "inv_a1b2c3d4e5f6" {
		t.Errorf("forwarded id = %q, want inv_a1b2c3d4e5f6", fake.gotGet.GetId())
	}
	if inv.GetTotalCents() != 1250 {
		t.Errorf("total_cents = %d, want 1250", inv.GetTotalCents())
	}
	if len(inv.GetLineItems()) != 2 {
		t.Fatalf("line items = %d, want 2", len(inv.GetLineItems()))
	}
	// The adjustment line is signed and routinely negative.
	if got := inv.GetLineItems()[1].GetAmountCents(); got != -50 {
		t.Errorf("adjustment amount_cents = %d, want -50", got)
	}
}

// The upcoming statement is computed, not stored: no id, status draft.
func TestBilling_GetUpcomingInvoice(t *testing.T) {
	fake := &fakeBillingClient{
		upcoming: &v1.Invoice{Status: InvoiceStatusDraft, TotalCents: 400},
	}
	b := newBilling(fake)

	inv, err := b.GetUpcomingInvoice(context.Background())
	if err != nil {
		t.Fatalf("GetUpcomingInvoice() error = %v", err)
	}
	if inv.GetId() != "" {
		t.Errorf("upcoming invoice id = %q, want empty", inv.GetId())
	}
	if inv.GetStatus() != InvoiceStatusDraft {
		t.Errorf("status = %v, want draft", inv.GetStatus())
	}
}

// A card the provider will not describe is still a payment method: the state
// says on_file even though brand and last4 are absent, and that state — not the
// missing digits — is what callers branch on.
func TestBilling_GetBillingProfile_OnFileWithoutCardDetails(t *testing.T) {
	fake := &fakeBillingClient{
		profile: &v1.BillingProfile{
			Object:             "billing_profile",
			OrgId:              "org_f6g7h8i9j0",
			PaymentMethodState: PaymentMethodStateOnFile,
			PaymentMethods:     []*v1.BillingPaymentMethod{{Id: "pm_1", Type: "card"}},
		},
	}
	b := newBilling(fake)

	profile, err := b.GetBillingProfile(context.Background())
	if err != nil {
		t.Fatalf("GetBillingProfile() error = %v", err)
	}
	if profile.GetPaymentMethodState() != PaymentMethodStateOnFile {
		t.Errorf("state = %v, want on_file", profile.GetPaymentMethodState())
	}
	if len(profile.GetPaymentMethods()) != 1 {
		t.Fatalf("payment methods = %d, want 1", len(profile.GetPaymentMethods()))
	}
	if got := profile.GetPaymentMethods()[0]; got.GetBrand() != "" || got.GetLast4() != "" {
		t.Errorf("brand/last4 = %q/%q, want both empty", got.GetBrand(), got.GetLast4())
	}
}

// An org that has never touched billing reports "none" and no methods, without
// the read creating anything provider-side.
func TestBilling_GetBillingProfile_NeverBilled(t *testing.T) {
	fake := &fakeBillingClient{
		profile: &v1.BillingProfile{PaymentMethodState: PaymentMethodStateNone},
	}
	b := newBilling(fake)

	profile, err := b.GetBillingProfile(context.Background())
	if err != nil {
		t.Fatalf("GetBillingProfile() error = %v", err)
	}
	if profile.GetPaymentMethodState() != PaymentMethodStateNone {
		t.Errorf("state = %v, want none", profile.GetPaymentMethodState())
	}
	if len(profile.GetPaymentMethods()) != 0 {
		t.Errorf("payment methods = %d, want 0", len(profile.GetPaymentMethods()))
	}
}

func TestBilling_CreateBillingPortalSession(t *testing.T) {
	fake := &fakeBillingClient{
		portal: &v1.BillingPortalSession{
			Object: "billing_portal_session",
			Url:    "https://portal.example/session/abc",
		},
	}
	b := newBilling(fake)

	session, err := b.CreateBillingPortalSession(context.Background())
	if err != nil {
		t.Fatalf("CreateBillingPortalSession() error = %v", err)
	}
	if session.GetUrl() != "https://portal.example/session/abc" {
		t.Errorf("url = %q, want the provider session URL", session.GetUrl())
	}
}

// An org with no budget set still reports the period's spend, so a budget card
// can be rendered before a budget exists.
func TestBilling_GetBudget_NoBudgetStillReportsSpend(t *testing.T) {
	fake := &fakeBillingClient{
		budget: &v1.Budget{Object: "budget", OrgId: "org_f6g7h8i9j0", SpentEur: 12.5, Currency: "EUR"},
	}
	b := newBilling(fake)

	budget, err := b.GetBudget(context.Background())
	if err != nil {
		t.Fatalf("GetBudget() error = %v", err)
	}
	if budget.AmountEur != nil {
		t.Errorf("amount_eur = %v, want absent", budget.GetAmountEur())
	}
	if budget.GetSpentEur() != 12.5 {
		t.Errorf("spent_eur = %v, want 12.5", budget.GetSpentEur())
	}
}

// SetBudget sends the amount; ClearBudget omits it, which is what the server
// reads as "clear the budget and stop alerting".
func TestBilling_SetAndClearBudget(t *testing.T) {
	fake := &fakeBillingClient{budget: &v1.Budget{AmountEur: proto.Float64(250)}}
	b := newBilling(fake)

	if _, err := b.SetBudget(context.Background(), 250); err != nil {
		t.Fatalf("SetBudget() error = %v", err)
	}
	if fake.gotBudget.AmountEur == nil {
		t.Fatal("SetBudget must send amount_eur")
	}
	if got := fake.gotBudget.GetAmountEur(); got != 250 {
		t.Errorf("forwarded amount_eur = %v, want 250", got)
	}

	if _, err := b.ClearBudget(context.Background()); err != nil {
		t.Fatalf("ClearBudget() error = %v", err)
	}
	if fake.gotBudget.AmountEur != nil {
		t.Errorf("ClearBudget sent amount_eur = %v, want it omitted", fake.gotBudget.GetAmountEur())
	}
}

// Past the threshold but under the hard stop, nothing is refused: used_percent
// runs past 100 and Blocked stays false. Only twice the threshold blocks.
func TestBilling_GetOutstandingBalance_OverThresholdIsNotBlocked(t *testing.T) {
	fake := &fakeBillingClient{
		balance: &v1.OutstandingBalance{
			Object:              "outstanding_balance",
			OutstandingCents:    7500,
			Tier:                TrustTierEstablished,
			ThresholdCents:      proto.Int64(5000),
			ThresholdSource:     ExposureThresholdSourceTrustTier,
			HardStopCents:       proto.Int64(10000),
			UsedPercent:         proto.Float64(150),
			Blocked:             false,
			SettlementAvailable: true,
			Currency:            "EUR",
		},
	}
	b := newBilling(fake)

	balance, err := b.GetOutstandingBalance(context.Background())
	if err != nil {
		t.Fatalf("GetOutstandingBalance() error = %v", err)
	}
	if balance.GetBlocked() {
		t.Error("blocked = true at 150% of threshold; admission only stops at the hard stop")
	}
	if got := balance.GetUsedPercent(); got != 150 {
		t.Errorf("used_percent = %v, want 150 (not capped at 100)", got)
	}
	if got := balance.GetThresholdSource(); got != ExposureThresholdSourceTrustTier {
		t.Errorf("threshold_source = %v, want trust_tier", got)
	}
}

// A long-standing account can have no threshold at all: both threshold and hard
// stop are absent together, and the source says unbounded.
func TestBilling_GetOutstandingBalance_Unbounded(t *testing.T) {
	fake := &fakeBillingClient{
		balance: &v1.OutstandingBalance{
			OutstandingCents: 42000,
			Tier:             TrustTierProven,
			ThresholdSource:  ExposureThresholdSourceUnbounded,
		},
	}
	b := newBilling(fake)

	balance, err := b.GetOutstandingBalance(context.Background())
	if err != nil {
		t.Fatalf("GetOutstandingBalance() error = %v", err)
	}
	if balance.ThresholdCents != nil || balance.HardStopCents != nil {
		t.Error("threshold_cents and hard_stop_cents must be absent together when unbounded")
	}
	if balance.GetBlocked() {
		t.Error("blocked = true with no threshold; admission is never refused for this reason")
	}
}

// Settling returns both halves: the statement produced, and the balance as it
// now stands — zero, unblocked — so no second call is needed.
func TestBilling_SettleOutstandingBalance(t *testing.T) {
	fake := &fakeBillingClient{
		settlement: &v1.SettleOutstandingBalanceResponse{
			Settlement: &v1.Settlement{
				Object:      "settlement",
				InvoiceId:   "inv_a1b2c3d4e5f6",
				AmountCents: 7500,
				Currency:    "EUR",
			},
			Balance: &v1.OutstandingBalance{OutstandingCents: 0, Blocked: false},
		},
	}
	b := newBilling(fake)

	result, err := b.SettleOutstandingBalance(context.Background())
	if err != nil {
		t.Fatalf("SettleOutstandingBalance() error = %v", err)
	}
	if got := result.GetSettlement().GetInvoiceId(); got != "inv_a1b2c3d4e5f6" {
		t.Errorf("invoice_id = %q, want inv_a1b2c3d4e5f6", got)
	}
	if got := result.GetBalance().GetOutstandingCents(); got != 0 {
		t.Errorf("balance after settling = %d, want 0", got)
	}
	if result.GetBalance().GetBlocked() {
		t.Error("blocked stayed true after settling; paying lifts the block immediately")
	}
}

// A deployment with the settlement rail switched off answers failed_precondition
// with error code "settlement_unavailable"; the SDK types it as a
// PreconditionError and preserves the code.
func TestBilling_SettleOutstandingBalance_Unavailable(t *testing.T) {
	fake := &fakeBillingClient{
		settleErr: connect.NewError(connect.CodeFailedPrecondition,
			errors.New(`{"type":"invalid_request_error","code":"settlement_unavailable","message":"mid-cycle settlement is not enabled"}`)),
	}
	b := newBilling(fake)

	_, err := b.SettleOutstandingBalance(context.Background())
	if err == nil {
		t.Fatal("SettleOutstandingBalance() error = nil, want PreconditionError")
	}
	var precondErr *PreconditionError
	if !errors.As(err, &precondErr) {
		t.Fatalf("error is %T, want *PreconditionError", err)
	}
	if precondErr.ErrorCode() != "settlement_unavailable" {
		t.Errorf("error code = %q, want settlement_unavailable", precondErr.ErrorCode())
	}
}

// An API-key caller is refused by the server; the SDK surfaces that as a typed
// PermissionError rather than a bare Connect error.
func TestBilling_APIKeyCaller_IsPermissionError(t *testing.T) {
	fake := &fakeBillingClient{
		upcomingErr: connect.NewError(connect.CodePermissionDenied,
			errors.New("billing is scoped to an organization and is not available to API keys")),
	}
	b := newBilling(fake)

	_, err := b.GetUpcomingInvoice(context.Background())
	if err == nil {
		t.Fatal("GetUpcomingInvoice() error = nil, want PermissionError")
	}
	var permErr *PermissionError
	if !errors.As(err, &permErr) {
		t.Fatalf("error is %T, want *PermissionError", err)
	}
}

// WithOrganization sets X-Organization-ID, which the org-scoped surfaces need.
// Without the option the header is absent entirely rather than sent empty.
func TestSetStandardHeaders_Organization(t *testing.T) {
	withOrg := &config{apiKey: "ak_test", userAgent: "ua", apiVersion: "2026-05-03", orgID: "org_f6g7h8i9j0"}
	h := http.Header{}
	setStandardHeaders(h, withOrg)
	if got := h.Get("X-Organization-ID"); got != "org_f6g7h8i9j0" {
		t.Errorf("X-Organization-ID = %q, want org_f6g7h8i9j0", got)
	}
	if got := h.Get("Authorization"); got != "Bearer ak_test" {
		t.Errorf("Authorization = %q, want Bearer ak_test", got)
	}

	withoutOrg := &config{apiKey: "ak_test", userAgent: "ua", apiVersion: "2026-05-03"}
	h2 := http.Header{}
	setStandardHeaders(h2, withoutOrg)
	if _, present := h2["X-Organization-Id"]; present {
		t.Error("X-Organization-ID must not be sent when no organization is configured")
	}
}

// The option plumbs through New into the constructed client's config.
func TestNew_WithOrganization(t *testing.T) {
	c, err := New("ak_test", WithOrganization("org_f6g7h8i9j0"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if c.cfg.orgID != "org_f6g7h8i9j0" {
		t.Errorf("cfg.orgID = %q, want org_f6g7h8i9j0", c.cfg.orgID)
	}
	if c.Billing == nil {
		t.Error("client.Billing is nil; the namespace must be wired in New")
	}
}
