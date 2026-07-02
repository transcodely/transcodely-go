package transcodely

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DefaultToleranceSeconds is the default signature-timestamp tolerance window
// in seconds (Stripe parity). A delivery whose timestamp differs from the
// receiver's clock by more than this is rejected as a replay or clock skew.
const DefaultToleranceSeconds = 300

// SignatureHeader is the HTTP header that carries the webhook signature.
// Header lookups are case-insensitive, so r.Header.Get(SignatureHeader) works
// regardless of the casing the platform sent.
const SignatureHeader = "Transcodely-Signature"

// EventIDHeader is the HTTP header carrying the event ID ("evt_...") on every
// webhook delivery. It mirrors the envelope's top-level id, so a receiver can
// deduplicate retried deliveries straight from the header without parsing the
// body. Like [SignatureHeader], lookups are case-insensitive.
const EventIDHeader = "Webhook-Id"

// VerifyOption configures signature verification and event construction.
type VerifyOption func(*verifyOptions)

type verifyOptions struct {
	tolerance time.Duration
	now       func() int64 // unix seconds; nil → time.Now
}

// WithTolerance overrides the default timestamp tolerance window
// ([DefaultToleranceSeconds]). A delivery whose timestamp differs from the
// current time by more than d is rejected with a [*WebhookTimestampError].
func WithTolerance(d time.Duration) VerifyOption {
	return func(o *verifyOptions) { o.tolerance = d }
}

// withNow overrides the clock used for the tolerance check. Test-only.
func withNow(fn func() int64) VerifyOption {
	return func(o *verifyOptions) { o.now = fn }
}

func resolveVerifyOptions(opts []VerifyOption) verifyOptions {
	o := verifyOptions{tolerance: DefaultToleranceSeconds * time.Second}
	for _, opt := range opts {
		opt(&o)
	}
	if o.now == nil {
		o.now = func() int64 { return time.Now().Unix() }
	}
	return o
}

type parsedSignature struct {
	timestamp  int64
	signatures []string
}

// parseSignatureHeader parses the Transcodely-Signature header. The header is
// a comma-separated list of key=value pairs: t is the unix timestamp
// (seconds) and each v1 entry is a hex-encoded HMAC-SHA-256. Unknown keys are
// ignored so future scheme versions don't break older receivers.
func parseSignatureHeader(header string) (*parsedSignature, error) {
	var timestamp int64
	haveTimestamp := false
	var signatures []string

	for _, part := range strings.Split(header, ",") {
		eq := strings.IndexByte(part, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(part[:eq])
		value := strings.TrimSpace(part[eq+1:])
		switch key {
		case "t":
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				timestamp = n
				haveTimestamp = true
			}
		case "v1":
			signatures = append(signatures, value)
		}
	}

	if !haveTimestamp {
		return nil, newWebhookSignatureError("signature header is missing the timestamp (t=) component")
	}
	if len(signatures) == 0 {
		return nil, newWebhookSignatureError("signature header has no v1 entries")
	}
	return &parsedSignature{timestamp: timestamp, signatures: signatures}, nil
}

// safeHexEqual reports whether two hex strings decode to the same bytes,
// comparing in constant time. It returns false on a length mismatch (rather
// than panicking inside the constant-time compare), so a malformed candidate
// signature can never crash verification.
func safeHexEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	ab, err := hex.DecodeString(a)
	if err != nil {
		return false
	}
	bb, err := hex.DecodeString(b)
	if err != nil {
		return false
	}
	if len(ab) == 0 || len(ab) != len(bb) {
		return false
	}
	return subtle.ConstantTimeCompare(ab, bb) == 1
}

// VerifySignature verifies a webhook signature header against the raw request
// body. It returns nil on success, a [*WebhookSignatureError] if the header is
// malformed or no v1 entry matches the computed HMAC, or a
// [*WebhookTimestampError] if the timestamp is outside the tolerance window.
//
// Pass more than one secret to accept deliveries signed under either key
// during a secret-rotation overlap window.
func VerifySignature(payload []byte, sigHeader string, secrets []string, opts ...VerifyOption) error {
	o := resolveVerifyOptions(opts)

	parsed, err := parseSignatureHeader(sigHeader)
	if err != nil {
		return err
	}

	toleranceSeconds := int64(o.tolerance / time.Second)
	diff := o.now() - parsed.timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > toleranceSeconds {
		return newWebhookTimestampError(
			fmt.Sprintf("signature timestamp is outside the tolerance window (%ds)", toleranceSeconds),
		)
	}

	// Signed payload is "<unix-timestamp>.<raw-body>".
	signedPayload := make([]byte, 0, len(payload)+20)
	signedPayload = strconv.AppendInt(signedPayload, parsed.timestamp, 10)
	signedPayload = append(signedPayload, '.')
	signedPayload = append(signedPayload, payload...)

	for _, secret := range secrets {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(signedPayload)
		expected := hex.EncodeToString(mac.Sum(nil))
		for _, candidate := range parsed.signatures {
			if safeHexEqual(expected, candidate) {
				return nil
			}
		}
	}

	return newWebhookSignatureError("no signatures matched the expected value")
}
