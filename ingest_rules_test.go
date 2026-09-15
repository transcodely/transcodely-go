package transcodely

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/transcodely/transcodely-go/internal/codec"
	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// The facade re-exports every ingest message, and each alias is the generated
// type itself rather than a distinct look-alike. These declarations fail to
// compile if an alias is missing or points somewhere else.
var (
	_ *v1.IngestRule        = (*IngestRule)(nil)
	_ *v1.IngestRuleFilters = (*IngestRuleFilters)(nil)
	_ *v1.IngestRuleAction  = (*IngestRuleAction)(nil)
	_ *v1.StorageEvent      = (*StorageEvent)(nil)

	_ *v1.CreateIngestRuleRequest = (*IngestRuleCreateParams)(nil)
	_ *v1.UpdateIngestRuleRequest = (*IngestRuleUpdateParams)(nil)
	_ *v1.ListIngestRulesRequest  = (*IngestRuleListParams)(nil)
	_ *v1.ListIngestEventsRequest = (*IngestEventListParams)(nil)
	_ *v1.TestIngestRuleRequest   = (*IngestRuleTestParams)(nil)
	_ *v1.TestIngestRuleResponse  = (*IngestRuleTestResult)(nil)

	_ v1.StorageEventSource = StorageEventSourceS3SNS
	_ v1.StorageEventStatus = StorageEventStatusSkipped

	_ *IngestRuleFilters = (&IngestRule{}).GetFilters()
	_ *IngestRuleAction  = (&IngestRule{}).GetAction()
)

// Every RPC on the generated service client has a method of the same name on
// the facade. This is the completeness guarantee: adding an RPC to
// IngestRuleService and regenerating, without touching ingest_rules.go, fails
// here rather than shipping a client that cannot reach it.
func TestIngestRules_FacadeCoversEveryRPC(t *testing.T) {
	iface := reflect.TypeOf((*transcodelyv1connect.IngestRuleServiceClient)(nil)).Elem()
	if iface.NumMethod() == 0 {
		t.Fatal("generated client interface has no methods")
	}

	var want []string
	for i := range iface.NumMethod() {
		want = append(want, iface.Method(i).Name)
	}
	sort.Strings(want)

	facade := reflect.TypeOf(&IngestRules{})
	var got []string
	for i := range facade.NumMethod() {
		got = append(got, facade.Method(i).Name)
	}
	sort.Strings(got)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("facade methods = %v, want %v (the RPC set on IngestRuleServiceClient)", got, want)
	}
}

// A create response decodes through the SDK's own codec with the wire's
// snake_case field names and lowercase enum values, and the reveal-once secret
// is readable off it.
func TestIngestRule_CreateDecodesWithSecret(t *testing.T) {
	payload := []byte(`{"rule":{
		"id":"ing_a1b2c3d4e5f6",
		"app_id":"app_k1l2m3n4o5",
		"origin_id":"ori_a1b2c3d4e5f6",
		"name":"Watch uploads/",
		"enabled":true,
		"filters":{"prefix":"uploads/","suffixes":[".mp4",".mov"],"min_bytes":1024,"max_bytes":0},
		"action":{"managed":true,"priority":"standard","output_path_template":"{input_dir}/{input_name}/{resolution}"},
		"endpoint_url":"https://api.transcodely.com/ingest/ing_a1b2c3d4e5f6",
		"secret_prefix":"ings_a1b",
		"secret_hint":"z9y8",
		"events_received":0,
		"jobs_created":0,
		"created_at":"2026-09-14T10:00:00Z",
		"updated_at":"2026-09-14T10:00:00Z"
	},"secret":"ings_a1b2c3d4e5f6g7h8i9j0z9y8"}`)

	var resp v1.CreateIngestRuleResponse
	if err := codec.NewProtoJSONCodec().Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got, want := resp.GetSecret(), "ings_a1b2c3d4e5f6g7h8i9j0z9y8"; got != want {
		t.Errorf("secret = %q, want %q", got, want)
	}
	rule := resp.GetRule()
	if got, want := rule.GetId(), "ing_a1b2c3d4e5f6"; got != want {
		t.Errorf("rule.id = %q, want %q", got, want)
	}
	if got, want := rule.GetEndpointUrl(), "https://api.transcodely.com/ingest/ing_a1b2c3d4e5f6"; got != want {
		t.Errorf("rule.endpoint_url = %q, want %q", got, want)
	}
	if got, want := rule.GetSecretPrefix(), "ings_a1b"; got != want {
		t.Errorf("rule.secret_prefix = %q, want %q", got, want)
	}
	if got, want := rule.GetSecretHint(), "z9y8"; got != want {
		t.Errorf("rule.secret_hint = %q, want %q", got, want)
	}
	if !rule.GetEnabled() {
		t.Error("rule.enabled = false, want true")
	}
	if got, want := rule.GetFilters().GetPrefix(), "uploads/"; got != want {
		t.Errorf("filters.prefix = %q, want %q", got, want)
	}
	if got, want := rule.GetFilters().GetSuffixes(), []string{".mp4", ".mov"}; !reflect.DeepEqual(got, want) {
		t.Errorf("filters.suffixes = %v, want %v", got, want)
	}
	if got, want := rule.GetFilters().GetMinBytes(), int64(1024); got != want {
		t.Errorf("filters.min_bytes = %d, want %d", got, want)
	}
	if !rule.GetAction().GetManaged() {
		t.Error("action.managed = false, want true")
	}
	// The lowercase enum on the wire expands to the generated constant.
	if got, want := rule.GetAction().GetPriority(), JobPriorityStandard; got != want {
		t.Errorf("action.priority = %v, want %v", got, want)
	}
	if got, want := rule.GetAction().GetOutputPathTemplate(), "{input_dir}/{input_name}/{resolution}"; got != want {
		t.Errorf("action.output_path_template = %q, want %q", got, want)
	}
}

// A read of the same rule carries no secret — only the prefix and hint. The
// secret exists on the create response and nowhere else.
func TestIngestRule_GetCarriesNoSecret(t *testing.T) {
	payload := []byte(`{"rule":{
		"id":"ing_a1b2c3d4e5f6","app_id":"app_k1l2m3n4o5","origin_id":"ori_a1b2c3d4e5f6",
		"name":"Watch uploads/","enabled":true,
		"secret_prefix":"ings_a1b","secret_hint":"z9y8",
		"events_received":42,"jobs_created":40,
		"last_event_at":"2026-09-14T11:00:00Z",
		"created_at":"2026-09-14T10:00:00Z","updated_at":"2026-09-14T10:00:00Z"
	}}`)

	var resp v1.GetIngestRuleResponse
	if err := codec.NewProtoJSONCodec().Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// GetIngestRuleResponse has no secret field at all; assert the shape the
	// rule does expose, and that the counters survive the round trip.
	rule := resp.GetRule()
	if got, want := rule.GetEventsReceived(), int64(42); got != want {
		t.Errorf("events_received = %d, want %d", got, want)
	}
	if got, want := rule.GetJobsCreated(), int64(40); got != want {
		t.Errorf("jobs_created = %d, want %d", got, want)
	}
	if rule.GetLastEventAt() == nil {
		t.Error("last_event_at is absent, want a timestamp")
	}
	if rule.GetSecretRotatedAt() != nil {
		t.Error("secret_rotated_at is set on a never-rotated rule")
	}
	if fd := rule.ProtoReflect().Descriptor().Fields().ByName("secret"); fd != nil {
		t.Error("IngestRule has a `secret` field; the full secret must never ride on a read")
	}
}

// The event log decodes with lowercase source and status on the wire, and a
// skipped event carries the slug that says why.
func TestIngestEvents_DecodeWithLowercaseEnums(t *testing.T) {
	payload := []byte(`{"events":[
		{"id":"sev_a1b2c3d4e5f6g7","rule_id":"ing_a1b2c3d4e5f6","app_id":"app_k1l2m3n4o5",
		 "bucket":"my-uploads","object_key":"uploads/my clip.mp4","etag":"d41d8cd98f00b204",
		 "size_bytes":10485760,"content_type":"video/mp4",
		 "source":"s3_sns","status":"created","job_id":"job_a1b2c3d4e5f6",
		 "received_at":"2026-09-14T11:00:00Z","processed_at":"2026-09-14T11:00:02Z"},
		{"id":"sev_b2c3d4e5f6g7h8","rule_id":"ing_a1b2c3d4e5f6","app_id":"app_k1l2m3n4o5",
		 "bucket":"my-uploads","object_key":"uploads/notes.txt",
		 "source":"gcs_pubsub","status":"skipped","reason":"filter_suffix",
		 "received_at":"2026-09-14T11:05:00Z","processed_at":"2026-09-14T11:05:00Z"}
	],"pagination":{"next_cursor":""}}`)

	var resp v1.ListIngestEventsResponse
	if err := codec.NewProtoJSONCodec().Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	events := resp.GetEvents()
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}

	if got, want := events[0].GetSource(), StorageEventSourceS3SNS; got != want {
		t.Errorf("events[0].source = %v, want %v", got, want)
	}
	if got, want := events[0].GetStatus(), StorageEventStatusCreated; got != want {
		t.Errorf("events[0].status = %v, want %v", got, want)
	}
	// The key is stored URL-decoded: S3 writes a space as "+".
	if got, want := events[0].GetObjectKey(), "uploads/my clip.mp4"; got != want {
		t.Errorf("events[0].object_key = %q, want %q", got, want)
	}
	if got, want := events[0].GetJobId(), "job_a1b2c3d4e5f6"; got != want {
		t.Errorf("events[0].job_id = %q, want %q", got, want)
	}

	if got, want := events[1].GetSource(), StorageEventSourceGCSPubSub; got != want {
		t.Errorf("events[1].source = %v, want %v", got, want)
	}
	if got, want := events[1].GetStatus(), StorageEventStatusSkipped; got != want {
		t.Errorf("events[1].status = %v, want %v", got, want)
	}
	if got, want := events[1].GetReason(), "filter_suffix"; got != want {
		t.Errorf("events[1].reason = %q, want %q", got, want)
	}
	if events[1].GetProcessedAt() == nil {
		t.Error("a terminal event has no processed_at")
	}
	if events[1].GetJobId() != "" {
		t.Errorf("a skipped event names job %q, want none", events[1].GetJobId())
	}
}

// fakeIngestClient stubs IngestRuleServiceClient the way fakeBillingClient
// does: it embeds the generated interface so an unimplemented call nil-panics
// rather than silently passing.
type fakeIngestClient struct {
	transcodelyv1connect.IngestRuleServiceClient

	gotCreate *v1.CreateIngestRuleRequest
	gotUpdate *v1.UpdateIngestRuleRequest
	gotReplay *v1.ReplayIngestEventRequest

	create *v1.CreateIngestRuleResponse
	update *v1.UpdateIngestRuleResponse
	replay *v1.ReplayIngestEventResponse
}

func (f *fakeIngestClient) Create(_ context.Context, req *connect.Request[v1.CreateIngestRuleRequest]) (*connect.Response[v1.CreateIngestRuleResponse], error) {
	f.gotCreate = req.Msg
	return connect.NewResponse(f.create), nil
}

func (f *fakeIngestClient) Update(_ context.Context, req *connect.Request[v1.UpdateIngestRuleRequest]) (*connect.Response[v1.UpdateIngestRuleResponse], error) {
	f.gotUpdate = req.Msg
	return connect.NewResponse(f.update), nil
}

func (f *fakeIngestClient) ReplayEvent(_ context.Context, req *connect.Request[v1.ReplayIngestEventRequest]) (*connect.Response[v1.ReplayIngestEventResponse], error) {
	f.gotReplay = req.Msg
	return connect.NewResponse(f.replay), nil
}

// Create hands back the secret; Update surfaces both the rotated secret and
// the backlog an un-pause exposes; ReplayEvent addresses the event by ID.
func TestIngestRules_FacadeSurfacesSecretsAndBacklog(t *testing.T) {
	fake := &fakeIngestClient{
		create: &v1.CreateIngestRuleResponse{
			Rule:   &v1.IngestRule{Id: "ing_a1b2c3d4e5f6", SecretPrefix: "ings_a1b"},
			Secret: "ings_full_secret_value",
		},
		update: &v1.UpdateIngestRuleResponse{
			Rule:                       &v1.IngestRule{Id: "ing_a1b2c3d4e5f6", Enabled: true},
			Secret:                     "ings_rotated_secret",
			EventsSkippedWhileDisabled: 7,
		},
		replay: &v1.ReplayIngestEventResponse{
			Event: &v1.StorageEvent{Id: "sev_a1b2c3d4e5f6g7", Status: v1.StorageEventStatus_STORAGE_EVENT_STATUS_RECEIVED},
		},
	}
	rules := newIngestRules(fake)
	ctx := context.Background()

	created, err := rules.Create(ctx, &IngestRuleCreateParams{OriginId: "ori_a1b2c3d4e5f6", Name: "Watch uploads/"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got, want := created.Secret, "ings_full_secret_value"; got != want {
		t.Errorf("created.Secret = %q, want %q", got, want)
	}
	if got, want := created.Rule.GetId(), "ing_a1b2c3d4e5f6"; got != want {
		t.Errorf("created.Rule.Id = %q, want %q", got, want)
	}
	if got, want := fake.gotCreate.GetOriginId(), "ori_a1b2c3d4e5f6"; got != want {
		t.Errorf("sent origin_id = %q, want %q", got, want)
	}

	updated, err := rules.Update(ctx, &IngestRuleUpdateParams{Id: "ing_a1b2c3d4e5f6", RotateSecret: true})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got, want := updated.Secret, "ings_rotated_secret"; got != want {
		t.Errorf("updated.Secret = %q, want %q", got, want)
	}
	if got, want := updated.EventsSkippedWhileDisabled, int64(7); got != want {
		t.Errorf("updated.EventsSkippedWhileDisabled = %d, want %d", got, want)
	}
	if !fake.gotUpdate.GetRotateSecret() {
		t.Error("rotate_secret was not sent")
	}

	event, err := rules.ReplayEvent(ctx, "sev_a1b2c3d4e5f6g7")
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if got, want := event.GetStatus(), StorageEventStatusReceived; got != want {
		t.Errorf("replayed status = %v, want %v", got, want)
	}
	if got, want := fake.gotReplay.GetEventId(), "sev_a1b2c3d4e5f6g7"; got != want {
		t.Errorf("sent event_id = %q, want %q", got, want)
	}
}

// Update merges, so removing a filter or part of an action takes an explicit
// flag rather than an empty value. Both flags ride on the params struct and
// reach the wire under their snake_case names — without them a caller has no
// way to widen a rule or drop an action's thumbnails.
func TestIngestRules_UpdateCarriesTheClearFlags(t *testing.T) {
	fake := &fakeIngestClient{
		update: &v1.UpdateIngestRuleResponse{
			Rule: &v1.IngestRule{Id: "ing_a1b2c3d4e5f6"},
		},
	}
	rules := newIngestRules(fake)

	params := &IngestRuleUpdateParams{
		Id:           "ing_a1b2c3d4e5f6",
		ClearFilters: true,
		ClearAction:  true,
		Action: &IngestRuleAction{
			Managed: true,
			Outputs: []*OutputSpec{{Type: OutputFormatHLS}},
		},
	}
	if _, err := rules.Update(context.Background(), params); err != nil {
		t.Fatalf("update: %v", err)
	}
	if !fake.gotUpdate.GetClearFilters() {
		t.Error("clear_filters was not sent")
	}
	if !fake.gotUpdate.GetClearAction() {
		t.Error("clear_action was not sent")
	}

	encoded, err := codec.NewProtoJSONCodec().Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, field := range []string{`"clear_filters":true`, `"clear_action":true`} {
		if !strings.Contains(string(encoded), field) {
			t.Errorf("encoded request %s is missing %s", encoded, field)
		}
	}
}
