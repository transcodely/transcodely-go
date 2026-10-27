package transcodely

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// fakeAppPricingClient embeds the generated interface so an unimplemented call
// nil-panics instead of silently passing.
type fakeAppPricingClient struct {
	transcodelyv1connect.AppServiceClient

	got  *v1.GetPricingRequest
	resp *v1.GetPricingResponse
	err  error
}

func (f *fakeAppPricingClient) GetPricing(_ context.Context, req *connect.Request[v1.GetPricingRequest]) (*connect.Response[v1.GetPricingResponse], error) {
	f.got = req.Msg
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(f.resp), nil
}

func TestAppsGetPricing_PassesAppIDAndReturnsEntries(t *testing.T) {
	fake := &fakeAppPricingClient{resp: &AppPricing{
		AppId:    "app_abc",
		Currency: "EUR",
		Entries: []*PricingEntry{{
			RuleType: "codec", RuleKey: "h264", Label: "H.264", GroupLabel: "Encoding",
			Unit: "eur_per_output_minute", BasePrice: proto.Float64(0.004), Custom: true,
			ListBasePrice: proto.Float64(0.005),
		}},
	}}
	got, err := newApps(fake).GetPricing(context.Background(), "app_abc")
	if err != nil {
		t.Fatalf("GetPricing: %v", err)
	}
	if fake.got.AppId == nil || *fake.got.AppId != "app_abc" {
		t.Fatalf("app_id sent = %v, want app_abc", fake.got.AppId)
	}
	if got.Currency != "EUR" || len(got.Entries) != 1 || got.Entries[0].GetBasePrice() != 0.004 || !got.Entries[0].Custom {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestAppsGetPricing_EmptyIDOmitsAppID(t *testing.T) {
	fake := &fakeAppPricingClient{resp: &AppPricing{AppId: "app_default", Currency: "EUR"}}
	if _, err := newApps(fake).GetPricing(context.Background(), ""); err != nil {
		t.Fatalf("GetPricing: %v", err)
	}
	if fake.got.AppId != nil {
		t.Fatalf("app_id should be absent for empty id, got %q", *fake.got.AppId)
	}
}

func TestAppsGetPricing_MapsErrors(t *testing.T) {
	fake := &fakeAppPricingClient{err: connect.NewError(connect.CodePermissionDenied, errors.New("other app"))}
	_, err := newApps(fake).GetPricing(context.Background(), "app_other")
	var apiErr *PermissionError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *PermissionError, got %T %v", err, err)
	}
}
