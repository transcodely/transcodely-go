package transcodely

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"connectrpc.com/connect"

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
