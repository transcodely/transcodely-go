package codec

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMoneyLoopEnumRoundTrip extends the wire-format conformance matrix to the
// trust-ladder and dunning enums. The multi-word values are the interesting
// ones: the prefix is stripped and the remainder stays snake_case, so
// EXPOSURE_THRESHOLD_SOURCE_TRUST_TIER is "trust_tier" on the wire and not
// "trusttier".
func TestMoneyLoopEnumRoundTrip(t *testing.T) {
	balance := (&v1.OutstandingBalance{}).ProtoReflect().Descriptor()
	tierEnum := balance.Fields().ByName("tier").Enum()
	sourceEnum := balance.Fields().ByName("threshold_source").Enum()

	treatmentEnum := (&v1.Organization{}).ProtoReflect().Descriptor().
		Fields().ByName("billing_treatment").Enum()

	// DunningStage rides on admin-only messages, which the public SDK does not
	// vendor, so there is no field to reach it through — take the enum
	// descriptor directly.
	stageEnum := v1.DunningStage_DUNNING_STAGE_NONE.Descriptor()

	cases := []struct {
		canonical, simple string
		enum              protoreflect.EnumDescriptor
	}{
		{"TRUST_TIER_NEW", "new", tierEnum},
		{"TRUST_TIER_ESTABLISHED", "established", tierEnum},
		{"TRUST_TIER_PROVEN", "proven", tierEnum},

		{"EXPOSURE_THRESHOLD_SOURCE_OVERRIDE", "override", sourceEnum},
		{"EXPOSURE_THRESHOLD_SOURCE_TRUST_TIER", "trust_tier", sourceEnum},
		{"EXPOSURE_THRESHOLD_SOURCE_ORG_PLAN", "org_plan", sourceEnum},
		{"EXPOSURE_THRESHOLD_SOURCE_PLATFORM_DEFAULT", "platform_default", sourceEnum},
		{"EXPOSURE_THRESHOLD_SOURCE_UNBOUNDED", "unbounded", sourceEnum},

		{"BILLING_TREATMENT_NORMAL", "normal", treatmentEnum},
		{"BILLING_TREATMENT_TRUSTED", "trusted", treatmentEnum},
		{"BILLING_TREATMENT_EXEMPT", "exempt", treatmentEnum},

		{"DUNNING_STAGE_NONE", "none", stageEnum},
		{"DUNNING_STAGE_PAST_DUE", "past_due", stageEnum},
		{"DUNNING_STAGE_SOFT_LIMITED", "soft_limited", stageEnum},
		{"DUNNING_STAGE_SUSPENDED", "suspended", stageEnum},
		{"DUNNING_STAGE_DELETION_WARNED", "deletion_warned", stageEnum},
		{"DUNNING_STAGE_WRITTEN_OFF", "written_off", stageEnum},
	}

	for _, c := range cases {
		if got := simplifyEnumValue(c.canonical, c.enum); got != c.simple {
			t.Errorf("simplify %q: got %q, want %q", c.canonical, got, c.simple)
		}
		if got := expandEnumValue(c.simple, c.enum); got != c.canonical {
			t.Errorf("expand %q: got %q, want %q", c.simple, got, c.canonical)
		}
	}
}

// TestMarshalOutstandingBalance pins the money-card payload: snake_case fields,
// lowercase enums, and int64 cents as JSON strings per the protobuf 64-bit
// mapping. A client that parses outstanding_cents as a number will break.
func TestMarshalOutstandingBalance(t *testing.T) {
	c := NewProtoJSONCodec()
	resp := &v1.GetOutstandingBalanceResponse{
		Balance: &v1.OutstandingBalance{
			Object:              "outstanding_balance",
			OrgId:               "org_f6g7h8i9j0",
			OutstandingCents:    7500,
			Tier:                v1.TrustTier_TRUST_TIER_ESTABLISHED,
			SettledPayments:     2,
			ThresholdCents:      proto.Int64(5000),
			ThresholdSource:     v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_TRUST_TIER,
			HardStopCents:       proto.Int64(10000),
			UsedPercent:         proto.Float64(150),
			AlertSteps:          []int32{80, 100, 125, 150, 175, 200},
			NotifiedSteps:       []int32{80, 100, 125},
			Currency:            "EUR",
			SettlementAvailable: true,
		},
	}

	data, err := c.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Balance struct {
			OutstandingCents    string  `json:"outstanding_cents"`
			Tier                string  `json:"tier"`
			ThresholdCents      string  `json:"threshold_cents"`
			ThresholdSource     string  `json:"threshold_source"`
			HardStopCents       string  `json:"hard_stop_cents"`
			UsedPercent         float64 `json:"used_percent"`
			NotifiedSteps       []int32 `json:"notified_steps"`
			SettlementAvailable bool    `json:"settlement_available"`
		} `json:"balance"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	if out.Balance.OutstandingCents != "7500" {
		t.Errorf("outstanding_cents = %q, want \"7500\"", out.Balance.OutstandingCents)
	}
	if out.Balance.Tier != "established" {
		t.Errorf("tier = %q, want established", out.Balance.Tier)
	}
	if out.Balance.ThresholdSource != "trust_tier" {
		t.Errorf("threshold_source = %q, want trust_tier", out.Balance.ThresholdSource)
	}
	if out.Balance.HardStopCents != "10000" {
		t.Errorf("hard_stop_cents = %q, want \"10000\"", out.Balance.HardStopCents)
	}
	if out.Balance.UsedPercent != 150 {
		t.Errorf("used_percent = %v, want 150 (not capped at 100)", out.Balance.UsedPercent)
	}
	if len(out.Balance.NotifiedSteps) != 3 {
		t.Errorf("notified_steps = %v, want 3 entries", out.Balance.NotifiedSteps)
	}
	if !out.Balance.SettlementAvailable {
		t.Error("settlement_available = false, want true")
	}
}

// TestUnmarshalOutstandingBalance reads the same shape back off the wire in its
// simplified form, which is what the server actually sends.
func TestUnmarshalOutstandingBalance(t *testing.T) {
	c := NewProtoJSONCodec()
	payload := []byte(`{"balance":{"object":"outstanding_balance","outstanding_cents":"7500",` +
		`"tier":"proven","threshold_source":"unbounded","blocked":false,"currency":"EUR"}}`)

	var resp v1.GetOutstandingBalanceResponse
	if err := c.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	b := resp.GetBalance()
	if b.GetOutstandingCents() != 7500 {
		t.Errorf("outstanding_cents = %d, want 7500", b.GetOutstandingCents())
	}
	if b.GetTier() != v1.TrustTier_TRUST_TIER_PROVEN {
		t.Errorf("tier = %v, want PROVEN", b.GetTier())
	}
	if b.GetThresholdSource() != v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_UNBOUNDED {
		t.Errorf("threshold_source = %v, want UNBOUNDED", b.GetThresholdSource())
	}
	// An absent threshold stays absent rather than defaulting to zero, which
	// would read as "no headroom" instead of "no limit".
	if b.ThresholdCents != nil {
		t.Errorf("threshold_cents = %d, want absent", b.GetThresholdCents())
	}
}

// TestUpdateBudgetClearOmitsAmount is the clear-the-budget contract: the field
// must be absent from the request body, not sent as 0, because 0 is a value the
// server would have to reject rather than read as "clear".
func TestUpdateBudgetClearOmitsAmount(t *testing.T) {
	c := NewProtoJSONCodec()
	data, err := c.Marshal(&v1.UpdateBudgetRequest{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if v, present := m["amount_eur"]; present {
		t.Errorf("amount_eur present as %v; clearing a budget must omit the field", v)
	}

	data, err = c.Marshal(&v1.UpdateBudgetRequest{AmountEur: proto.Float64(250)})
	if err != nil {
		t.Fatalf("marshal set: %v", err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse set: %v", err)
	}
	if m["amount_eur"] != float64(250) {
		t.Errorf("amount_eur = %v, want 250", m["amount_eur"])
	}
}
