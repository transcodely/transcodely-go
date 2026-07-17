package transcodely

import (
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
	if len(WebhookEventTypes) != 15 {
		t.Fatalf("WebhookEventTypes has %d entries, want 15", len(WebhookEventTypes))
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
		{"app.updated", "*App"},
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
