package transcodely

import (
	"encoding/json"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

func TestNew_WiresWebhookNamespaces(t *testing.T) {
	c, err := New("ak_test_123")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.WebhookEndpoints == nil {
		t.Error("client.WebhookEndpoints is nil")
	}
	if c.Events == nil {
		t.Error("client.Events is nil")
	}
}

func TestWebhookEventCatalog(t *testing.T) {
	if len(WebhookEventTypes) != 18 {
		t.Fatalf("WebhookEventTypes has %d entries, want 18", len(WebhookEventTypes))
	}
	for _, et := range WebhookEventTypes {
		if et == "job.updated" {
			t.Error(`catalog contains "job.updated", which does not exist`)
		}
		if et == "*" {
			t.Error(`catalog contains the "*" wildcard, which is subscription-only`)
		}
	}
	// Spot-check the constant values match their wire strings.
	if EventTypeJobSucceeded != "job.succeeded" {
		t.Errorf("EventTypeJobSucceeded = %q", EventTypeJobSucceeded)
	}
	if EventTypeOutputReady != "output.ready" {
		t.Errorf("EventTypeOutputReady = %q", EventTypeOutputReady)
	}
	if EventTypeAppSpendLimitWarning != "app.spend_limit_warning" {
		t.Errorf("EventTypeAppSpendLimitWarning = %q", EventTypeAppSpendLimitWarning)
	}
	if EventTypeAppSpendLimitExceeded != "app.spend_limit_exceeded" {
		t.Errorf("EventTypeAppSpendLimitExceeded = %q", EventTypeAppSpendLimitExceeded)
	}
}

func TestResourceForEventType(t *testing.T) {
	cases := []struct {
		typ  string
		want string // concrete type name, or "" for nil
	}{
		{"job.succeeded", "*Job"},
		{"job.progress", "*Job"},
		{"output.ready", "*JobOutput"},
		{"output.progress", "*JobOutput"},
		{"video.uploaded", "*Video"},
		{"video.ready", "*Video"},
		{"video.failed", "*Video"},
		{"video.source_scheduled_for_deletion", "*Video"},
		{"app.updated", "*App"},
		// Spend-limit events share the "app." prefix but are notification
		// payloads, not App snapshots — they must NOT decode to *App.
		{"app.spend_limit_warning", ""},
		{"app.spend_limit_exceeded", ""},
		{"subscription.created", ""},
		{"", ""},
	}
	for _, c := range cases {
		got := resourceForEventType(c.typ)
		if c.want == "" {
			if got != nil {
				t.Errorf("resourceForEventType(%q) = %T, want nil", c.typ, got)
			}
			continue
		}
		if name := typeName(got); name != c.want {
			t.Errorf("resourceForEventType(%q) = %s, want %s", c.typ, name, c.want)
		}
	}
}

func typeName(v any) string {
	switch v.(type) {
	case *v1.Job:
		return "*Job"
	case *v1.JobOutput:
		return "*JobOutput"
	case *v1.Video:
		return "*Video"
	case *v1.App:
		return "*App"
	}
	return "unknown"
}

func TestWebhookEvent_SpendLimit(t *testing.T) {
	for _, et := range []EventType{EventTypeAppSpendLimitWarning, EventTypeAppSpendLimitExceeded} {
		ev := &WebhookEvent{
			Type: et,
			raw:  json.RawMessage(`{"app_id":"app_demo","period_start":"2026-01-01","period_end":"2026-02-01","limit_eur":100,"spent_eur":82.5,"threshold_pct":80,"currency":"EUR"}`),
		}
		// decodeData must NOT populate data with a mis-decoded *App.
		ev.decodeData()
		if ev.data != nil {
			t.Errorf("%s: decodeData populated data=%T, want nil (notification payload)", et, ev.data)
		}
		if _, isApp := ev.App(); isApp {
			t.Errorf("%s: App() returned true, want false for a notification payload", et)
		}

		n, ok := ev.SpendLimit()
		if !ok {
			t.Fatalf("%s: SpendLimit() ok=false, want true", et)
		}
		if n.AppID != "app_demo" || n.LimitEUR != 100 || n.SpentEUR != 82.5 ||
			n.ThresholdPct != 80 || n.Currency != "EUR" ||
			n.PeriodStart != "2026-01-01" || n.PeriodEnd != "2026-02-01" {
			t.Errorf("%s: SpendLimit() = %+v, unexpected fields", et, n)
		}
	}

	// SpendLimit() returns (nil, false) for a non-spend-limit event.
	if _, ok := (&WebhookEvent{Type: EventTypeAppUpdated}).SpendLimit(); ok {
		t.Error("SpendLimit() ok=true for app.updated, want false")
	}
}

func TestWebhookEventFromProto_LedgerBridge(t *testing.T) {
	created, err := time.Parse(time.RFC3339, "2026-05-24T10:55:08Z")
	if err != nil {
		t.Fatal(err)
	}
	pe := &v1.Event{
		Id:              "evt_a1b2c3d4e5f6g7h8",
		Type:            "job.succeeded",
		Object:          "event",
		ApiVersion:      "2026-05-23",
		Data:            `{"id":"job_1","object":"job","status":"completed"}`,
		RequestId:       proto.String("req_abc123"),
		PendingWebhooks: 0,
		CreatedAt:       timestamppb.New(created),
	}
	ev := webhookEventFromProto(pe)
	if ev.ID != "evt_a1b2c3d4e5f6g7h8" {
		t.Errorf("ID = %q", ev.ID)
	}
	if ev.Type != EventTypeJobSucceeded {
		t.Errorf("Type = %q", ev.Type)
	}
	if ev.Created != "2026-05-24T10:55:08Z" {
		t.Errorf("Created = %q", ev.Created)
	}
	if ev.Request.ID != "req_abc123" {
		t.Errorf("Request.ID = %q", ev.Request.ID)
	}
	job, ok := ev.Job()
	if !ok {
		t.Fatal("Job() ok = false")
	}
	if job.GetId() != "job_1" || job.GetStatus() != JobStatusCompleted {
		t.Errorf("job = %v / %v", job.GetId(), job.GetStatus())
	}
}

func TestWebhookEventFromProto_NilTolerant(t *testing.T) {
	ev := webhookEventFromProto(nil)
	if ev == nil {
		t.Fatal("webhookEventFromProto(nil) = nil")
	}
	if _, ok := ev.Job(); ok {
		t.Error("Job() ok = true on empty event")
	}
}

// TestMembershipUserStatusReExports ensures the previously-missing enum
// re-exports are usable from outside the internal package.
func TestMembershipUserStatusReExports(t *testing.T) {
	var ms MembershipStatus = MembershipStatusActive
	if ms != v1.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE {
		t.Error("MembershipStatusActive mismatch")
	}
	var us UserStatus = UserStatusSuspended
	if us != v1.UserStatus_USER_STATUS_SUSPENDED {
		t.Error("UserStatusSuspended mismatch")
	}
	// They must be settable on the List params.
	_ = &MembershipListParams{Status: &ms}
	_ = &UserListParams{Status: &us}
}
