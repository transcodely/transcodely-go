package transcodely

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/transcodely/transcodely-go/internal/codec"
	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// Webhook HTTP headers Transcodely sets on every delivery.
const (
	// SignatureHeader carries the signature: `t=<unix>,v1=<hex>[,v1=<hex>]`.
	// Pass its value to [ConstructEvent] / [VerifyWebhookSignature].
	SignatureHeader = "Transcodely-Signature"
	// EventIDHeader carries the event ID (`evt_*`). It equals the envelope's
	// `id` and is stable across retries — use it to deduplicate deliveries.
	EventIDHeader = "Webhook-Id"
)

// DefaultWebhookTolerance is the maximum clock skew accepted between the
// signature timestamp and now. Override per-call with [WithWebhookTolerance].
const DefaultWebhookTolerance = 5 * time.Minute

// EventType is a webhook event type such as "job.succeeded". The wire value is
// a plain string; the typed constants below cover the catalog the API emits.
type EventType string

// The 16 event types the platform emits. Mirrors the API's
// domain.WebhookEventTypes(). The "*" wildcard is a subscription value only and
// is intentionally absent — it is never the type of a delivered event. Event.Type
// is a plain string so an older SDK still decodes a type the API adds later.
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
	EventTypeVideoReady     EventType = "video.ready"
	EventTypeVideoFailed    EventType = "video.failed"
	EventTypeVideoDeleted   EventType = "video.deleted"
	// EventTypeVideoSourceScheduledForDeletion fires when a hosted video's
	// original source file is scheduled for deletion by the app's
	// delete_source_after_days lifecycle rule, at least 72 hours before the
	// deletion. Renditions and playback are never affected.
	EventTypeVideoSourceScheduledForDeletion EventType = "video.source_scheduled_for_deletion"
	EventTypeAppCreated                      EventType = "app.created"
	EventTypeAppUpdated                      EventType = "app.updated"
)

// WebhookEventTypes is the full catalog of emittable event types (excludes the
// "*" subscription wildcard).
var WebhookEventTypes = []EventType{
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
	EventTypeVideoReady,
	EventTypeVideoFailed,
	EventTypeVideoDeleted,
	EventTypeVideoSourceScheduledForDeletion,
	EventTypeAppCreated,
	EventTypeAppUpdated,
}

// HealthWindow is the rolling window for [WebhookEndpoints.GetHealth].
type HealthWindow string

// HealthWindow values.
const (
	HealthWindow24h HealthWindow = "24h"
	HealthWindow7d  HealthWindow = "7d"
	HealthWindow30d HealthWindow = "30d"
)

// webhookCodec decodes the inner resource snapshot (`data`) with the same
// snake_case + lowercase-enum JSON semantics the rest of the SDK uses on the
// wire. encoding/json alone would mis-handle enums, int64-as-string, and
// timestamps.
var webhookCodec = codec.NewProtoJSONCodec()

// ---------- Errors ----------

type webhookErr struct {
	msg   string
	cause error
}

func (e *webhookErr) Error() string {
	if e.cause != nil {
		return "transcodely: " + e.msg + ": " + e.cause.Error()
	}
	return "transcodely: " + e.msg
}

func (e *webhookErr) Unwrap() error { return e.cause }

// WebhookSignatureError means the signature header was missing/malformed or no
// v1 entry matched the computed HMAC over the raw body.
type WebhookSignatureError struct{ webhookErr }

// WebhookTimestampError means the signature timestamp fell outside the
// tolerance window (default 5 minutes).
type WebhookTimestampError struct{ webhookErr }

// WebhookPayloadError means the body was not valid JSON or did not match the
// event-envelope shape.
type WebhookPayloadError struct{ webhookErr }

func newSignatureError(msg string) *WebhookSignatureError {
	return &WebhookSignatureError{webhookErr{msg: msg}}
}

func newTimestampError(msg string) *WebhookTimestampError {
	return &WebhookTimestampError{webhookErr{msg: msg}}
}

func newPayloadError(msg string, cause error) *WebhookPayloadError {
	return &WebhookPayloadError{webhookErr{msg: msg, cause: cause}}
}

// ---------- Verification options ----------

type webhookVerifyConfig struct {
	tolerance time.Duration
	now       func() time.Time
}

// WebhookVerifyOption tunes signature verification.
type WebhookVerifyOption func(*webhookVerifyConfig)

// WithWebhookTolerance overrides the default clock-skew tolerance
// ([DefaultWebhookTolerance]). Non-positive values are ignored.
func WithWebhookTolerance(d time.Duration) WebhookVerifyOption {
	return func(c *webhookVerifyConfig) {
		if d > 0 {
			c.tolerance = d
		}
	}
}

// withWebhookClock overrides the clock. Unexported — for tests only.
func withWebhookClock(now func() time.Time) WebhookVerifyOption {
	return func(c *webhookVerifyConfig) {
		if now != nil {
			c.now = now
		}
	}
}

// ---------- WebhookEvent ----------

// WebhookEventRequest identifies the API request that triggered an event.
type WebhookEventRequest struct {
	// ID is the request ID (`req_*`), or "" for events emitted outside a
	// request scope (worker-callback events like job.succeeded, and test sends).
	ID string
	// IdempotencyKey is the key the originating request carried. Reserved;
	// always nil today.
	IdempotencyKey *string
}

// WebhookEvent is a decoded, verified webhook event. It is produced both by
// [ConstructEvent] (from a raw HTTP delivery) and by the [Events] resource
// (from the event ledger), so a handler dispatched on Type behaves identically
// either way.
//
// Dispatch on Type, then pull the typed resource with one of the accessors:
//
//	switch event.Type {
//	case transcodely.EventTypeJobSucceeded:
//	    if job, ok := event.Job(); ok { /* ... */ }
//	}
type WebhookEvent struct {
	// ID is the event ID (`evt_*`), equal to the Webhook-Id header. Stable
	// across retries and resends — use it for idempotency.
	ID string
	// Type is the event type, e.g. EventTypeJobSucceeded.
	Type EventType
	// Object is the envelope discriminator; always "event".
	Object string
	// APIVersion is the API version frozen at emit time (e.g. "2026-05-23").
	APIVersion string
	// Created is the RFC 3339 UTC timestamp of when the event fired.
	Created string
	// PendingWebhooks counts delivery attempts still pending across all
	// subscribed endpoints.
	PendingWebhooks int
	// Request identifies the originating API request.
	Request WebhookEventRequest

	data proto.Message
	raw  json.RawMessage
}

// Job returns the decoded Job and true for job.* events, else (nil, false).
func (e *WebhookEvent) Job() (*Job, bool) {
	j, ok := e.data.(*Job)
	return j, ok
}

// JobOutput returns the decoded JobOutput and true for output.* events, else
// (nil, false).
func (e *WebhookEvent) JobOutput() (*JobOutput, bool) {
	o, ok := e.data.(*JobOutput)
	return o, ok
}

// Video returns the decoded Video and true for video.* events, else (nil, false).
func (e *WebhookEvent) Video() (*Video, bool) {
	v, ok := e.data.(*Video)
	return v, ok
}

// App returns the decoded App and true for app.* events, else (nil, false).
func (e *WebhookEvent) App() (*App, bool) {
	a, ok := e.data.(*App)
	return a, ok
}

// Data returns the decoded resource snapshot (a *Job, *JobOutput, *Video, or
// *App), or nil for an unrecognized event type. Use [WebhookEvent.RawData] to
// inspect an unknown type's payload.
func (e *WebhookEvent) Data() proto.Message { return e.data }

// RawData returns the raw JSON bytes of the event's `data` field, useful for a
// future event type this SDK version cannot yet decode.
func (e *WebhookEvent) RawData() json.RawMessage { return e.raw }

// decodeData populates data from raw based on the event type's resource group.
// Best-effort: an unknown type or a decode failure leaves data nil (raw is
// always retained), so unrecognized future events still surface.
func (e *WebhookEvent) decodeData() {
	msg := resourceForEventType(string(e.Type))
	if msg == nil || len(e.raw) == 0 {
		return
	}
	if err := webhookCodec.Unmarshal(e.raw, msg); err != nil {
		return
	}
	e.data = msg
}

// resourceForEventType returns a fresh resource message for the event type's
// prefix, or nil if the type is unrecognized.
func resourceForEventType(t string) proto.Message {
	switch {
	case strings.HasPrefix(t, "output."):
		return &v1.JobOutput{}
	case strings.HasPrefix(t, "job."):
		return &v1.Job{}
	case strings.HasPrefix(t, "video."):
		return &v1.Video{}
	case strings.HasPrefix(t, "app."):
		return &v1.App{}
	}
	return nil
}

// ---------- ConstructEvent ----------

// ConstructEvent verifies a signed webhook delivery and decodes it into a typed
// [WebhookEvent]. Pass the raw request body (never a re-serialized object), the
// value of the [SignatureHeader] header, and the endpoint's signing secret.
//
// It returns a [WebhookTimestampError] when the signature timestamp is outside
// the tolerance window, a [WebhookSignatureError] when no signature matches, and
// a [WebhookPayloadError] when the body is not a valid event envelope.
func ConstructEvent(payload []byte, sigHeader, secret string, opts ...WebhookVerifyOption) (*WebhookEvent, error) {
	return ConstructEventWithSecrets(payload, sigHeader, []string{secret}, opts...)
}

// ConstructEventWithSecrets is [ConstructEvent] with multiple candidate secrets.
// During a secret rotation the signature carries a v1 entry for each active
// secret; pass []string{newSecret, previousSecret} to accept either through the
// 24-hour overlap.
func ConstructEventWithSecrets(payload []byte, sigHeader string, secrets []string, opts ...WebhookVerifyOption) (*WebhookEvent, error) {
	if err := VerifyWebhookSignature(payload, sigHeader, secrets, opts...); err != nil {
		return nil, err
	}
	return buildWebhookEvent(payload)
}

// VerifyWebhookSignature verifies a Transcodely-Signature header against the raw
// body without decoding the envelope. ConstructEvent calls this first; use it
// directly only if you decode the body yourself.
func VerifyWebhookSignature(payload []byte, sigHeader string, secrets []string, opts ...WebhookVerifyOption) error {
	cfg := webhookVerifyConfig{tolerance: DefaultWebhookTolerance, now: time.Now}
	for _, o := range opts {
		o(&cfg)
	}

	ts, sigs, err := parseSignatureHeader(sigHeader)
	if err != nil {
		return err
	}
	if len(secrets) == 0 {
		return newSignatureError("no signing secrets provided")
	}

	skew := cfg.now().Unix() - ts
	if skew < 0 {
		skew = -skew
	}
	if time.Duration(skew)*time.Second > cfg.tolerance {
		return newTimestampError("signature timestamp is outside the tolerance window")
	}

	signed := append([]byte(strconv.FormatInt(ts, 10)+"."), payload...)
	for _, secret := range secrets {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(signed)
		expected := hex.EncodeToString(mac.Sum(nil))
		for _, sig := range sigs {
			// expected is lowercase hex; normalize the candidate so a
			// mixed-case digest still matches with a constant-time compare.
			if hmac.Equal([]byte(expected), []byte(strings.ToLower(sig))) {
				return nil
			}
		}
	}
	return newSignatureError("no signatures matched the expected value")
}

// parseSignatureHeader extracts the timestamp and the v1 HMAC entries. Unknown
// keys are ignored so a future scheme version does not break older receivers.
func parseSignatureHeader(header string) (int64, []string, error) {
	var ts int64
	haveTS := false
	var sigs []string
	for _, part := range strings.Split(header, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
				ts = n
				haveTS = true
			}
		case "v1":
			sigs = append(sigs, strings.TrimSpace(v))
		}
	}
	if !haveTS {
		return 0, nil, newSignatureError("signature header is missing the timestamp (t=) component")
	}
	if len(sigs) == 0 {
		return 0, nil, newSignatureError("signature header has no v1 entries")
	}
	return ts, sigs, nil
}

// buildWebhookEvent parses and validates a verified envelope, decoding the inner
// resource snapshot.
func buildWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(payload, &top); err != nil {
		return nil, newPayloadError("webhook body is not a JSON object", err)
	}

	id, err := envString(top, "id")
	if err != nil {
		return nil, err
	}
	if obj, _ := envString(top, "object"); obj != "event" {
		return nil, newPayloadError(`webhook envelope "object" must be "event"`, nil)
	}
	apiVersion, err := envString(top, "api_version")
	if err != nil {
		return nil, err
	}
	created, err := envString(top, "created")
	if err != nil {
		return nil, err
	}
	typ, err := envString(top, "type")
	if err != nil {
		return nil, err
	}
	pending, err := envInt(top, "pending_webhooks")
	if err != nil {
		return nil, err
	}

	dataRaw, ok := top["data"]
	if !ok || !isJSONObject(dataRaw) {
		return nil, newPayloadError("webhook envelope field `data` must be a JSON object", nil)
	}

	reqRaw, ok := top["request"]
	if !ok || !isJSONObject(reqRaw) {
		return nil, newPayloadError("webhook envelope field `request` must be a JSON object", nil)
	}
	req, err := parseEventRequest(reqRaw)
	if err != nil {
		return nil, err
	}

	ev := &WebhookEvent{
		ID:              id,
		Type:            EventType(typ),
		Object:          "event",
		APIVersion:      apiVersion,
		Created:         created,
		PendingWebhooks: pending,
		Request:         req,
		raw:             append(json.RawMessage(nil), dataRaw...),
	}
	ev.decodeData()
	return ev, nil
}

func parseEventRequest(raw json.RawMessage) (WebhookEventRequest, error) {
	var r struct {
		ID             *string `json:"id"`
		IdempotencyKey *string `json:"idempotency_key"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return WebhookEventRequest{}, newPayloadError("webhook envelope field `request` is malformed", err)
	}
	// request.id is null for events emitted outside a request scope; accept
	// null or a non-empty string, but reject "".
	if r.ID != nil && *r.ID == "" {
		return WebhookEventRequest{}, newPayloadError("webhook envelope field `request.id` must be a non-empty string or null", nil)
	}
	out := WebhookEventRequest{IdempotencyKey: r.IdempotencyKey}
	if r.ID != nil {
		out.ID = *r.ID
	}
	return out, nil
}

func envString(top map[string]json.RawMessage, key string) (string, error) {
	raw, ok := top[key]
	if !ok {
		return "", newPayloadError("webhook envelope is missing required string field `"+key+"`", nil)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil || s == "" {
		return "", newPayloadError("webhook envelope field `"+key+"` must be a non-empty string", err)
	}
	return s, nil
}

func envInt(top map[string]json.RawMessage, key string) (int, error) {
	raw, ok := top[key]
	if !ok {
		return 0, newPayloadError("webhook envelope is missing required numeric field `"+key+"`", nil)
	}
	// Decode into any and require a JSON number (float64) so a JSON bool or
	// string is rejected, matching the JS/Python helpers.
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0, newPayloadError("webhook envelope field `"+key+"` must be a number", err)
	}
	f, ok := v.(float64)
	if !ok {
		return 0, newPayloadError("webhook envelope field `"+key+"` must be a number", nil)
	}
	return int(f), nil
}

func isJSONObject(raw json.RawMessage) bool {
	for _, b := range raw {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return true
		default:
			return false
		}
	}
	return false
}
