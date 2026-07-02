package transcodely

import (
	"strings"
	"testing"
)

// canonicalEventCatalog is the exact set of concrete event types the API emits,
// transcribed from the API's source of truth (api/internal/domain/webhook.go:
// WebhookEventTypes). It deliberately excludes the "*" wildcard. This list is
// the contract; AllEventTypes() must equal it value-for-value and order-for-
// order. job.updated is intentionally absent — the API dropped it (terminal-
// only job-event philosophy) and asserts it INVALID in domain/webhook_test.go.
var canonicalEventCatalog = []EventType{
	"job.created",
	"job.succeeded",
	"job.failed",
	"job.canceled",
	"job.progress",
	"output.created",
	"output.ready",
	"output.failed",
	"output.progress",
	"video.uploaded",
	"video.deleted",
	"app.created",
	"app.updated",
}

func TestAllEventTypes_MatchesAPICatalog(t *testing.T) {
	got := AllEventTypes()
	if len(got) != len(canonicalEventCatalog) {
		t.Fatalf("AllEventTypes() has %d entries, want %d:\n got=%v\nwant=%v",
			len(got), len(canonicalEventCatalog), got, canonicalEventCatalog)
	}
	for i, want := range canonicalEventCatalog {
		if got[i] != want {
			t.Errorf("AllEventTypes()[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// TestEventCatalog_NoJobUpdated is the targeted regression guard for the
// drift this audit fixed: the API removed job.updated, so it must appear
// nowhere in the SDK's emitted-event catalog.
func TestEventCatalog_NoJobUpdated(t *testing.T) {
	for _, et := range AllEventTypes() {
		if et == "job.updated" {
			t.Fatalf("job.updated is back in AllEventTypes(); the API does not emit it")
		}
	}
}

// TestEventCatalog_WildcardNotEmitted asserts the "*" wildcard is a
// subscription-only token: it is the value of EventTypeWildcard but is never
// part of the emitted-event catalog.
func TestEventCatalog_WildcardNotEmitted(t *testing.T) {
	if EventTypeWildcard != "*" {
		t.Errorf("EventTypeWildcard = %q, want %q", EventTypeWildcard, "*")
	}
	for _, et := range AllEventTypes() {
		if et == EventTypeWildcard {
			t.Fatalf("wildcard %q must not appear in AllEventTypes()", EventTypeWildcard)
		}
	}
}

// TestEventCatalog_EveryTypeDecodes asserts each catalog entry's prefix maps to
// a resource decoder, so the catalog and the resource-decode table can't drift
// apart (a new event family added to one but not the other is caught here).
func TestEventCatalog_EveryTypeDecodes(t *testing.T) {
	for _, et := range AllEventTypes() {
		matched := false
		for _, d := range resourceDecoders {
			if strings.HasPrefix(string(et), d.prefix) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("event type %q has no matching resource decoder prefix", et)
		}
	}
}
