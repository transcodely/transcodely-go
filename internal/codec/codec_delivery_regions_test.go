package codec

import (
	"testing"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalUpdateHostingConfigDeliveryRegions pins HostingConfig.delivery_regions
// as a snake_case array of lowercase region strings on the wire.
func TestMarshalUpdateHostingConfigDeliveryRegions(t *testing.T) {
	c := NewProtoJSONCodec()

	m := marshalToMap(t, c, &v1.UpdateHostingConfigRequest{
		Id:            "app_abc",
		HostingConfig: &v1.HostingConfig{DeliveryRegions: []string{"europe"}},
	})
	cfg := marshalToMap(t, c, &v1.HostingConfig{DeliveryRegions: []string{"europe"}})
	if got := string(cfg["delivery_regions"]); got != `["europe"]` {
		t.Errorf("delivery_regions = %s; payload: %v", got, cfg)
	}
	if _, ok := cfg["deliveryRegions"]; ok {
		t.Errorf("camelCase key present: %v", cfg)
	}
	if _, ok := m["hosting_config"]; !ok {
		t.Errorf("hosting_config missing: %v", m)
	}
}

// TestUnmarshalUsageSummaryEgressByRegion verifies a GetUsage payload with the
// per-region egress breakdown (summary and per-day) decodes.
func TestUnmarshalUsageSummaryEgressByRegion(t *testing.T) {
	c := NewProtoJSONCodec()
	var resp v1.GetUsageResponse
	body := `{"usage":{"egress_gb":3,"egress_by_region":[` +
		`{"region":"europe","egress_bytes":"2147483648","egress_gb":2},` +
		`{"region":"","egress_bytes":"1073741824","egress_gb":1}],` +
		`"daily":[{"date":"2026-11-01","egress_bytes":"3221225472","egress_by_region":[` +
		`{"region":"europe","egress_bytes":"2147483648"},{"egress_bytes":"1073741824"}]}]}}`
	if err := c.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	by := resp.GetUsage().GetEgressByRegion()
	if len(by) != 2 || by[0].GetRegion() != "europe" || by[0].GetEgressBytes() != 2147483648 || by[0].GetEgressGb() != 2 || by[1].GetRegion() != "" {
		t.Fatalf("egress_by_region = %+v", by)
	}
	days := resp.GetUsage().GetDaily()
	if len(days) != 1 || len(days[0].GetEgressByRegion()) != 2 || days[0].GetEgressByRegion()[0].GetRegion() != "europe" {
		t.Fatalf("daily egress_by_region = %+v", days)
	}
}
