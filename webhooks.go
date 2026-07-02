package transcodely

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/transcodely/transcodely-go/internal/codec"
	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// EventType is the discriminator on a webhook [Event]. Switch on it to handle
// the event, then pull the decoded resource with the matching accessor
// ([Event.Job], [Event.Video], etc.).
type EventType string

// The canonical event types the platform emits. An endpoint subscribes to a
// subset of these (or "*" for all). A future API version may add types not in
// this list; an [Event] still decodes, with Data carrying the raw JSON — see
// [Event.RawData].
const (
	EventTypeJobCreated     EventType = "job.created"
	EventTypeJobSucceeded   EventType = "job.succeeded"
	EventTypeJobFailed      EventType = "job.failed"
	EventTypeJobCanceled    EventType = "job.canceled"
	EventTypeJobProgress    EventType = "job.progress"
	EventTypeOutputCreated  EventType = "output.created"
	EventTypeOutputReady    EventType = "output.ready"
	EventTypeOutputFailed   EventType = "output.failed"
	EventTypeOutputProgress EventType = "output.progress"
	EventTypeVideoUploaded  EventType = "video.uploaded"
	EventTypeVideoDeleted   EventType = "video.deleted"
	EventTypeAppCreated     EventType = "app.created"
	EventTypeAppUpdated     EventType = "app.updated"
)

// AllEventTypes returns every concrete event type the platform emits, in a
// stable order. It is the SDK's mirror of the API's canonical catalog and is
// the single source of truth a catalog test asserts against. It deliberately
// excludes the "*" wildcard, which is only valid as an endpoint subscription
// value (see [EventTypeWildcard]), never as the type of an emitted event.
func AllEventTypes() []EventType {
	return []EventType{
		EventTypeJobCreated,
		EventTypeJobSucceeded,
		EventTypeJobFailed,
		EventTypeJobCanceled,
		EventTypeJobProgress,
		EventTypeOutputCreated,
		EventTypeOutputReady,
		EventTypeOutputFailed,
		EventTypeOutputProgress,
		EventTypeVideoUploaded,
		EventTypeVideoDeleted,
		EventTypeAppCreated,
		EventTypeAppUpdated,
	}
}

// EventTypeWildcard is the "*" subscription token meaning "all current and
// future event types". It is valid only in a [WebhookEndpoint]'s enabled
// events; the platform never emits an event whose Type is "*".
const EventTypeWildcard EventType = "*"

// Event is the unified, customer-facing webhook event. It is what both
// [ConstructEvent] (decoding an HTTP delivery envelope) and the events
// resource ([Events.Retrieve], [Events.List], decoding a proto Event) return,
// so a handler behaves identically regardless of how the event reached it.
//
// Data holds the decoded resource snapshot: a *[Job], *[JobOutput], *[Video],
// or *[App] for a known [EventType], or a [json.RawMessage] for a type this
// SDK version does not recognize. Use the typed accessors to pull it out:
//
//	switch event.Type {
//	case transcodely.EventTypeJobSucceeded:
//	    if job, ok := event.Job(); ok {
//	        log.Printf("job %s %s", job.GetId(), job.GetStatus())
//	    }
//	}
type Event struct {
	// ID is the event identifier (e.g. "evt_...").
	ID string
	// Object is always "event".
	Object string
	// APIVersion is the API version the event was emitted under; frozen for
	// the event's lifetime.
	APIVersion string
	// Created is when the event was created (UTC).
	Created time.Time
	// Type is the event-type discriminator.
	Type EventType
	// Data is the decoded resource: *Job, *JobOutput, *Video, *App, or a
	// json.RawMessage for an unrecognized type.
	Data any
	// Livemode is true for production events.
	Livemode bool
	// PendingWebhooks is the number of delivery attempts still pending across
	// all subscribed endpoints.
	PendingWebhooks int
	// Request describes the API request that triggered the event.
	Request EventRequest
}

// EventRequest identifies the API request that produced an event.
type EventRequest struct {
	// ID is the originating request ID, or "" if unknown.
	ID string
	// IdempotencyKey is the idempotency key the originating request set, or ""
	// if none was set.
	IdempotencyKey string
}

// Job returns the event's resource as a *[Job] and true when the event carries
// a job (the "job.*" types); otherwise it returns (nil, false).
func (e *Event) Job() (*Job, bool) {
	j, ok := e.Data.(*Job)
	return j, ok
}

// JobOutput returns the event's resource as a *[JobOutput] and true when the
// event carries an output (the "output.*" types); otherwise (nil, false).
func (e *Event) JobOutput() (*JobOutput, bool) {
	o, ok := e.Data.(*JobOutput)
	return o, ok
}

// Video returns the event's resource as a *[Video] and true when the event
// carries a video (the "video.*" types); otherwise (nil, false).
func (e *Event) Video() (*Video, bool) {
	v, ok := e.Data.(*Video)
	return v, ok
}

// App returns the event's resource as an *[App] and true when the event
// carries an app (the "app.*" types); otherwise (nil, false).
func (e *Event) App() (*App, bool) {
	a, ok := e.Data.(*App)
	return a, ok
}

// RawData returns the undecoded resource JSON and true when the event type is
// one this SDK version does not recognize. For known types Data holds a
// decoded message and RawData reports false.
func (e *Event) RawData() (json.RawMessage, bool) {
	r, ok := e.Data.(json.RawMessage)
	return r, ok
}

// webhookCodec decodes the resource snapshot in an event's data. It is the
// same JSON codec the transport uses, so it accepts the simplified lowercase
// enum values (e.g. "completed") that appear on the wire.
var webhookCodec = codec.NewProtoJSONCodec()

// resourceDecoders maps an event-type prefix to a factory for the proto
// message its data decodes into. Prefixes are disjoint; "output." precedes
// "job." to make the intent explicit. Shared by [ConstructEvent] (envelope
// path) and [eventFromProto] (events-resource path).
var resourceDecoders = []struct {
	prefix string
	newMsg func() proto.Message
}{
	{"output.", func() proto.Message { return &v1.JobOutput{} }},
	{"job.", func() proto.Message { return &v1.Job{} }},
	{"video.", func() proto.Message { return &v1.Video{} }},
	{"app.", func() proto.Message { return &v1.App{} }},
}

// decodeResource decodes a resource snapshot for the given event type. Known
// type prefixes decode into the matching proto message; unknown types (and any
// decode failure on a known type — never drop a legitimate event over a server
// payload hiccup) return the raw JSON unchanged.
func decodeResource(eventType EventType, rawData []byte) any {
	for _, d := range resourceDecoders {
		if strings.HasPrefix(string(eventType), d.prefix) {
			msg := d.newMsg()
			if err := webhookCodec.Unmarshal(rawData, msg); err != nil {
				return json.RawMessage(cloneBytes(rawData))
			}
			return msg
		}
	}
	return json.RawMessage(cloneBytes(rawData))
}

func cloneBytes(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

// ConstructEvent verifies a signed webhook delivery and decodes it into a
// typed [Event]. payload is the raw HTTP request body, sigHeader the value of
// the [SignatureHeader] header, and secret the endpoint's signing secret
// (whsec_...).
//
// It returns a [*WebhookSignatureError] on a bad signature, a
// [*WebhookTimestampError] when the timestamp is outside the tolerance window,
// and a [*WebhookPayloadError] when the body is not valid JSON or the envelope
// shape is wrong. Use [ConstructEventWithSecrets] to accept more than one
// secret during a rotation window.
func ConstructEvent(payload []byte, sigHeader, secret string, opts ...VerifyOption) (*Event, error) {
	return ConstructEventWithSecrets(payload, sigHeader, []string{secret}, opts...)
}

// ConstructEventWithSecrets is [ConstructEvent] for the secret-rotation case:
// it accepts deliveries signed under any of secrets, so a receiver can verify
// against both the new and previous secret during the 24h overlap window.
func ConstructEventWithSecrets(payload []byte, sigHeader string, secrets []string, opts ...VerifyOption) (*Event, error) {
	if err := VerifySignature(payload, sigHeader, secrets, opts...); err != nil {
		return nil, err
	}
	return buildEvent(payload)
}

// buildEvent validates a flat delivery-envelope body and decodes it into an
// [Event]. Every field is type-checked individually so a malformed envelope
// fails with a precise [*WebhookPayloadError] rather than a silent zero value.
func buildEvent(body []byte) (*Event, error) {
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, newWebhookPayloadError("webhook body is not valid JSON", err)
	}
	m, ok := parsed.(map[string]any)
	if !ok {
		return nil, newWebhookPayloadError("webhook body must be a JSON object", nil)
	}

	id, err := requireEnvString(m, "id")
	if err != nil {
		return nil, err
	}
	if obj, _ := m["object"].(string); obj != "event" {
		return nil, newWebhookPayloadError(`webhook envelope "object" must be "event"`, nil)
	}
	apiVersion, err := requireEnvString(m, "api_version")
	if err != nil {
		return nil, err
	}
	created, err := requireEnvString(m, "created")
	if err != nil {
		return nil, err
	}
	eventType, err := requireEnvString(m, "type")
	if err != nil {
		return nil, err
	}
	livemode, err := requireEnvBool(m, "livemode")
	if err != nil {
		return nil, err
	}
	pendingWebhooks, err := requireEnvNumber(m, "pending_webhooks")
	if err != nil {
		return nil, err
	}

	dataRaw, ok := m["data"].(map[string]any)
	if !ok {
		return nil, newWebhookPayloadError("webhook envelope field `data` must be a JSON object", nil)
	}
	requestRaw, ok := m["request"].(map[string]any)
	if !ok {
		return nil, newWebhookPayloadError("webhook envelope field `request` must be a JSON object", nil)
	}
	// request.id is always present on the wire but is the empty string for
	// events emitted outside request scope — worker callbacks (job.succeeded,
	// job.failed, job.canceled, job.progress, all output.*) and every
	// SendTestWebhook delivery. So it must be allowed to be empty; it must
	// still be a string (the field is never null or absent).
	requestID, ok := requestRaw["id"].(string)
	if !ok {
		return nil, newWebhookPayloadError("webhook envelope field `request.id` must be a string", nil)
	}
	idempotencyKey := ""
	switch v := requestRaw["idempotency_key"].(type) {
	case nil:
		// null or absent → no key
	case string:
		idempotencyKey = v
	default:
		return nil, newWebhookPayloadError("webhook envelope field `request.idempotency_key` must be a string or null", nil)
	}

	dataBytes, err := json.Marshal(dataRaw)
	if err != nil {
		return nil, newWebhookPayloadError("webhook envelope field `data` could not be re-encoded", err)
	}

	return &Event{
		ID:              id,
		Object:          "event",
		APIVersion:      apiVersion,
		Created:         parseEventTime(created),
		Type:            EventType(eventType),
		Data:            decodeResource(EventType(eventType), dataBytes),
		Livemode:        livemode,
		PendingWebhooks: pendingWebhooks,
		Request:         EventRequest{ID: requestID, IdempotencyKey: idempotencyKey},
	}, nil
}

// eventFromProto bridges a proto v1.Event (returned by the events resource,
// where data is a JSON string, created_at is a Timestamp, and request_id is
// flat) into the same unified [Event] that [ConstructEvent] produces.
func eventFromProto(p *v1.Event) *Event {
	if p == nil {
		return nil
	}
	var data any
	if raw := p.GetData(); raw != "" {
		data = decodeResource(EventType(p.GetType()), []byte(raw))
	} else {
		data = json.RawMessage("{}")
	}
	var created time.Time
	if ts := p.GetCreatedAt(); ts != nil {
		created = ts.AsTime()
	}
	return &Event{
		ID:              p.GetId(),
		Object:          "event",
		APIVersion:      p.GetApiVersion(),
		Created:         created,
		Type:            EventType(p.GetType()),
		Data:            data,
		Livemode:        p.GetLivemode(),
		PendingWebhooks: int(p.GetPendingWebhooks()),
		// The proto Event carries only request_id; idempotency_key is reserved
		// (always unset) until JobService.Create propagates it.
		Request: EventRequest{ID: p.GetRequestId()},
	}
}

func requireEnvString(m map[string]any, key string) (string, error) {
	v, ok := m[key].(string)
	if !ok || v == "" {
		return "", newWebhookPayloadError(fmt.Sprintf("webhook envelope is missing required string field `%s`", key), nil)
	}
	return v, nil
}

func requireEnvBool(m map[string]any, key string) (bool, error) {
	v, ok := m[key].(bool)
	if !ok {
		return false, newWebhookPayloadError(fmt.Sprintf("webhook envelope field `%s` must be a boolean", key), nil)
	}
	return v, nil
}

func requireEnvNumber(m map[string]any, key string) (int, error) {
	v, ok := m[key].(float64)
	if !ok {
		return 0, newWebhookPayloadError(fmt.Sprintf("webhook envelope field `%s` must be a number", key), nil)
	}
	return int(v), nil
}

func parseEventTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

// Webhooks groups the stateless webhook helpers under a Stripe-style
// namespace, reachable as client.Webhooks for discoverability. It holds no
// state — the package-level [ConstructEvent] and [VerifySignature] are
// equivalent and need no client.
type Webhooks struct{}

// ConstructEvent is the namespace form of [ConstructEvent].
func (Webhooks) ConstructEvent(payload []byte, sigHeader, secret string, opts ...VerifyOption) (*Event, error) {
	return ConstructEvent(payload, sigHeader, secret, opts...)
}

// ConstructEventWithSecrets is the namespace form of [ConstructEventWithSecrets].
func (Webhooks) ConstructEventWithSecrets(payload []byte, sigHeader string, secrets []string, opts ...VerifyOption) (*Event, error) {
	return ConstructEventWithSecrets(payload, sigHeader, secrets, opts...)
}

// VerifySignature is the namespace form of [VerifySignature].
func (Webhooks) VerifySignature(payload []byte, sigHeader string, secrets []string, opts ...VerifyOption) error {
	return VerifySignature(payload, sigHeader, secrets, opts...)
}
