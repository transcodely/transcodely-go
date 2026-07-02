package transcodely

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// webhookEnvelopeMap returns a valid flat delivery envelope as a map so tests
// can tweak or delete individual fields before signing.
func webhookEnvelopeMap() map[string]any {
	return map[string]any{
		"id":          "evt_abc123",
		"object":      "event",
		"api_version": "2026-05-23",
		"created":     "2026-05-24T10:55:08Z",
		"type":        "job.succeeded",
		"data": map[string]any{
			"id":        "job_abc",
			"object":    "job",
			"status":    "completed",
			"input_url": "https://x/in.mp4",
		},
		"livemode":         true,
		"pending_webhooks": 0,
		"request":          map[string]any{"id": "req_xyz", "idempotency_key": nil},
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

func webhookHeader(body string, ts int64, secret string) string {
	return fmt.Sprintf("t=%d,v1=%s", ts, signHMAC(ts, body, secret))
}

func TestConstructEvent_HappyPaths(t *testing.T) {
	now := fixedNow(testTS)

	t.Run("decodes a job.succeeded event with Job data", func(t *testing.T) {
		body := mustJSON(t, webhookEnvelopeMap())
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Type != EventTypeJobSucceeded {
			t.Errorf("Type = %q, want job.succeeded", event.Type)
		}
		if event.ID != "evt_abc123" {
			t.Errorf("ID = %q", event.ID)
		}
		if event.Object != "event" {
			t.Errorf("Object = %q", event.Object)
		}
		if event.APIVersion != "2026-05-23" {
			t.Errorf("APIVersion = %q", event.APIVersion)
		}
		wantCreated := time.Date(2026, 5, 24, 10, 55, 8, 0, time.UTC)
		if !event.Created.Equal(wantCreated) {
			t.Errorf("Created = %v, want %v", event.Created, wantCreated)
		}
		if !event.Livemode {
			t.Errorf("Livemode = false")
		}
		if event.PendingWebhooks != 0 {
			t.Errorf("PendingWebhooks = %d", event.PendingWebhooks)
		}
		if event.Request != (EventRequest{ID: "req_xyz"}) {
			t.Errorf("Request = %+v", event.Request)
		}
		job, ok := event.Job()
		if !ok {
			t.Fatalf("event.Job() ok = false, Data is %T", event.Data)
		}
		if job.GetId() != "job_abc" {
			t.Errorf("job.GetId() = %q", job.GetId())
		}
		if job.GetInputUrl() != "https://x/in.mp4" {
			t.Errorf("job.GetInputUrl() = %q", job.GetInputUrl())
		}
		// Proves the enum-expanding codec path: "completed" → JOB_STATUS_COMPLETED.
		if got := job.GetStatus().String(); got != "JOB_STATUS_COMPLETED" {
			t.Errorf("job.GetStatus() = %s, want JOB_STATUS_COMPLETED", got)
		}
	})

	t.Run("decodes an output.ready event with JobOutput data", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "output.ready"
		m["data"] = map[string]any{"id": "jot_abc", "output_url": "https://cdn/out.m3u8", "progress": 100}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Type != EventTypeOutputReady {
			t.Errorf("Type = %q", event.Type)
		}
		out, ok := event.JobOutput()
		if !ok {
			t.Fatalf("event.JobOutput() ok = false, Data is %T", event.Data)
		}
		if out.GetId() != "jot_abc" {
			t.Errorf("out.GetId() = %q", out.GetId())
		}
		if out.GetOutputUrl() != "https://cdn/out.m3u8" {
			t.Errorf("out.GetOutputUrl() = %q", out.GetOutputUrl())
		}
	})

	t.Run("decodes a video.uploaded event with Video data", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "video.uploaded"
		m["data"] = map[string]any{"id": "vid_abc", "object": "video"}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		vid, ok := event.Video()
		if !ok {
			t.Fatalf("event.Video() ok = false, Data is %T", event.Data)
		}
		if vid.GetId() != "vid_abc" {
			t.Errorf("vid.GetId() = %q", vid.GetId())
		}
	})

	t.Run("decodes an app.created event with App data", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "app.created"
		m["data"] = map[string]any{"id": "app_abc", "object": "app", "name": "My App"}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		app, ok := event.App()
		if !ok {
			t.Fatalf("event.App() ok = false, Data is %T", event.Data)
		}
		if app.GetId() != "app_abc" {
			t.Errorf("app.GetId() = %q", app.GetId())
		}
		if app.GetName() != "My App" {
			t.Errorf("app.GetName() = %q", app.GetName())
		}
	})

	t.Run("preserves a non-null request.idempotency_key", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["request"] = map[string]any{"id": "req_xyz", "idempotency_key": "user_supplied_key"}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Request.IdempotencyKey != "user_supplied_key" {
			t.Errorf("IdempotencyKey = %q", event.Request.IdempotencyKey)
		}
	})

	t.Run("accepts an empty request.id (worker-originated and test deliveries)", func(t *testing.T) {
		// The API sends "request":{"id":"","idempotency_key":null} for every
		// event emitted outside request scope — the worker callbacks
		// (job.succeeded/failed/canceled/progress, all output.*) and every
		// SendTestWebhook delivery. These are the highest-value events, so
		// ConstructEvent must accept an empty request.id, not reject it.
		m := webhookEnvelopeMap()
		m["request"] = map[string]any{"id": "", "idempotency_key": nil}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error for empty request.id: %v", err)
		}
		if event.Request.ID != "" {
			t.Errorf("Request.ID = %q, want empty", event.Request.ID)
		}
		if _, ok := event.Job(); !ok {
			t.Errorf("expected Job decode, Data is %T", event.Data)
		}
	})

	t.Run("unknown type with a known prefix still decodes (job.scheduled → Job)", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "job.scheduled"
		m["data"] = map[string]any{"id": "job_future", "object": "job"}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Type != "job.scheduled" {
			t.Errorf("Type = %q", event.Type)
		}
		if _, ok := event.Job(); !ok {
			t.Errorf("expected Job decode for job. prefix, Data is %T", event.Data)
		}
	})

	t.Run("genuinely unknown type leaves data as raw JSON (forward-compat)", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "future.thing"
		m["data"] = map[string]any{"id": "ftr_abc", "object": "future"}
		body := mustJSON(t, m)
		event, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Type != "future.thing" {
			t.Errorf("Type = %q", event.Type)
		}
		if _, ok := event.Job(); ok {
			t.Errorf("did not expect a Job for unknown prefix")
		}
		raw, ok := event.RawData()
		if !ok {
			t.Fatalf("RawData() ok = false, Data is %T", event.Data)
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("raw data not valid JSON: %v", err)
		}
		if got["id"] != "ftr_abc" {
			t.Errorf("raw data id = %v", got["id"])
		}
	})
}

func TestConstructEvent_ErrorPaths(t *testing.T) {
	now := fixedNow(testTS)

	t.Run("WebhookSignatureError when the body is tampered", func(t *testing.T) {
		body := mustJSON(t, webhookEnvelopeMap())
		header := webhookHeader(body, testTS, testSecret)
		tampered := strings.Replace(body, "evt_abc123", "evt_other000", 1)
		_, err := ConstructEvent([]byte(tampered), header, testSecret, now)
		assertSignatureError(t, err)
	})

	t.Run("WebhookSignatureError on wrong secret", func(t *testing.T) {
		body := mustJSON(t, webhookEnvelopeMap())
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), "whsec_other", now)
		assertSignatureError(t, err)
	})

	t.Run("WebhookTimestampError on expired timestamp", func(t *testing.T) {
		body := mustJSON(t, webhookEnvelopeMap())
		header := webhookHeader(body, testTS-600, testSecret)
		_, err := ConstructEvent([]byte(body), header, testSecret, now)
		assertTimestampError(t, err)
	})

	t.Run("WebhookPayloadError on malformed JSON", func(t *testing.T) {
		body := "{not valid json"
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		assertPayloadError(t, err)
	})

	t.Run("WebhookPayloadError when object != event", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["object"] = "something_else"
		body := mustJSON(t, m)
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		assertPayloadError(t, err)
	})

	t.Run("WebhookPayloadError when data is a string instead of object", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["data"] = "not an object"
		body := mustJSON(t, m)
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		assertPayloadError(t, err)
	})

	t.Run("WebhookPayloadError when a required field is missing", func(t *testing.T) {
		m := webhookEnvelopeMap()
		delete(m, "id")
		body := mustJSON(t, m)
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		assertPayloadError(t, err)
	})

	t.Run("WebhookPayloadError when request.idempotency_key is the wrong type", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["request"] = map[string]any{"id": "req_x", "idempotency_key": 42}
		body := mustJSON(t, m)
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		assertPayloadError(t, err)
	})

	t.Run("WebhookPayloadError when body is a JSON array, not an object", func(t *testing.T) {
		body := "[]"
		_, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		assertPayloadError(t, err)
	})
}

// TestEventFromProto exercises the proto→unified bridge and proves the two
// wire shapes converge: a proto Event (data as a JSON string, created_at as a
// Timestamp, request_id flat) decodes into the same typed Event a delivery
// envelope would.
func TestEventFromProto(t *testing.T) {
	created := time.Date(2026, 5, 24, 10, 55, 8, 0, time.UTC)
	requestID := "req_proto"
	p := &v1.Event{
		Id:              "evt_proto1",
		Type:            "job.succeeded",
		Data:            `{"id":"job_proto","object":"job","status":"completed"}`,
		ApiVersion:      "2026-05-23",
		Livemode:        true,
		PendingWebhooks: 2,
		Object:          "event",
		CreatedAt:       timestamppb.New(created),
		RequestId:       &requestID,
	}

	ev := eventFromProto(p)
	if ev.ID != "evt_proto1" || ev.Type != EventTypeJobSucceeded || ev.APIVersion != "2026-05-23" {
		t.Fatalf("envelope fields wrong: %+v", ev)
	}
	if !ev.Livemode || ev.PendingWebhooks != 2 {
		t.Errorf("livemode/pending wrong: %+v", ev)
	}
	if !ev.Created.Equal(created) {
		t.Errorf("Created = %v, want %v", ev.Created, created)
	}
	if ev.Request.ID != "req_proto" || ev.Request.IdempotencyKey != "" {
		t.Errorf("Request = %+v", ev.Request)
	}
	job, ok := ev.Job()
	if !ok {
		t.Fatalf("ev.Job() ok = false, Data is %T", ev.Data)
	}
	if job.GetId() != "job_proto" {
		t.Errorf("job.GetId() = %q", job.GetId())
	}
	if got := job.GetStatus().String(); got != "JOB_STATUS_COMPLETED" {
		t.Errorf("job.GetStatus() = %s, want JOB_STATUS_COMPLETED", got)
	}
}

// TestWebhookData_ExpandsLowercaseEnums guards against the TS-style webhook
// enum-decoding bug: the server serializes the inner `data` resource with
// simplified lowercase enum values (e.g. JobStatus JOB_STATUS_COMPLETED → wire
// "completed"), identical to Get-RPC responses. The decode must route through
// the SDK's enum-expanding codec, NOT raw protojson (which rejects a lowercase
// enum string). This test fails if the bug is present — a lowercase enum would
// fail to decode and the typed field would never reach the expected constant.
// It asserts the concrete enum CONSTANTS (not their string names) on both the
// ConstructEvent (HTTP delivery) and eventFromProto (events resource) paths.
func TestWebhookData_ExpandsLowercaseEnums(t *testing.T) {
	now := fixedNow(testTS)

	t.Run("constructEvent: Job status+priority", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "job.succeeded"
		m["data"] = map[string]any{"id": "job_abc", "object": "job", "status": "completed", "priority": "premium"}
		body := mustJSON(t, m)
		ev, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		job, ok := ev.Job()
		if !ok {
			t.Fatalf("Data is %T, want *Job", ev.Data)
		}
		if job.GetStatus() != v1.JobStatus_JOB_STATUS_COMPLETED {
			t.Errorf("status = %v (%d), want JOB_STATUS_COMPLETED", job.GetStatus(), job.GetStatus())
		}
		if job.GetPriority() != v1.JobPriority_JOB_PRIORITY_PREMIUM {
			t.Errorf("priority = %v (%d), want JOB_PRIORITY_PREMIUM", job.GetPriority(), job.GetPriority())
		}
	})

	t.Run("constructEvent: OutputStatus", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "output.ready"
		m["data"] = map[string]any{"id": "jot_abc", "status": "completed"}
		body := mustJSON(t, m)
		ev, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out, ok := ev.JobOutput()
		if !ok {
			t.Fatalf("Data is %T, want *JobOutput", ev.Data)
		}
		if out.GetStatus() != v1.OutputStatus_OUTPUT_STATUS_COMPLETED {
			t.Errorf("status = %v (%d), want OUTPUT_STATUS_COMPLETED", out.GetStatus(), out.GetStatus())
		}
	})

	t.Run("constructEvent: AppStatus", func(t *testing.T) {
		m := webhookEnvelopeMap()
		m["type"] = "app.created"
		m["data"] = map[string]any{"id": "app_abc", "object": "app", "status": "active"}
		body := mustJSON(t, m)
		ev, err := ConstructEvent([]byte(body), webhookHeader(body, testTS, testSecret), testSecret, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		app, ok := ev.App()
		if !ok {
			t.Fatalf("Data is %T, want *App", ev.Data)
		}
		if app.GetStatus() != v1.AppStatus_APP_STATUS_ACTIVE {
			t.Errorf("status = %v (%d), want APP_STATUS_ACTIVE", app.GetStatus(), app.GetStatus())
		}
	})

	t.Run("eventFromProto (events resource): Job status+priority", func(t *testing.T) {
		// The proto Event path carries `data` as a JSON string — same lowercase
		// enums — and must converge on the same decoded enum constants.
		p := &v1.Event{
			Id:         "evt_x",
			Type:       "job.succeeded",
			Object:     "event",
			ApiVersion: "2026-05-23",
			Data:       `{"id":"job_abc","object":"job","status":"completed","priority":"premium"}`,
		}
		ev := eventFromProto(p)
		job, ok := ev.Job()
		if !ok {
			t.Fatalf("Data is %T, want *Job", ev.Data)
		}
		if job.GetStatus() != v1.JobStatus_JOB_STATUS_COMPLETED {
			t.Errorf("status = %v, want JOB_STATUS_COMPLETED", job.GetStatus())
		}
		if job.GetPriority() != v1.JobPriority_JOB_PRIORITY_PREMIUM {
			t.Errorf("priority = %v, want JOB_PRIORITY_PREMIUM", job.GetPriority())
		}
	})
}

func assertPayloadError(t *testing.T, err error) {
	t.Helper()
	var target *WebhookPayloadError
	if !errors.As(err, &target) {
		t.Fatalf("expected *WebhookPayloadError, got %T (%v)", err, err)
	}
}
