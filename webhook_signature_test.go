package transcodely

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	testSecret  = "whsec_test_12345678901234567890abcdef"
	testSecretB = "whsec_test_xxxxxxxxxxxxxxxxxxxxabcd"
	testTS      = int64(1716480293)
)

func signHMAC(ts int64, body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.%s", ts, body)
	return hex.EncodeToString(mac.Sum(nil))
}

func fixedNow(ts int64) VerifyOption { return withNow(func() int64 { return ts }) }

func TestVerifySignature(t *testing.T) {
	const body = `{"id":"evt_x","object":"event"}`
	now := fixedNow(testTS)

	t.Run("accepts a well-formed single-signature header", func(t *testing.T) {
		header := fmt.Sprintf("t=%d,v1=%s", testTS, signHMAC(testTS, body, testSecret))
		if err := VerifySignature([]byte(body), header, []string{testSecret}, now); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})

	t.Run("accepts the second v1 entry when the first does not match", func(t *testing.T) {
		wrong := strings.Repeat("0", 64)
		right := signHMAC(testTS, body, testSecret)
		header := fmt.Sprintf("t=%d,v1=%s,v1=%s", testTS, wrong, right)
		if err := VerifySignature([]byte(body), header, []string{testSecret}, now); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})

	t.Run("accepts an array of secrets and matches against any", func(t *testing.T) {
		header := fmt.Sprintf("t=%d,v1=%s", testTS, signHMAC(testTS, body, testSecretB))
		if err := VerifySignature([]byte(body), header, []string{testSecret, testSecretB}, now); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})

	t.Run("rejects a tampered body with WebhookSignatureError", func(t *testing.T) {
		header := fmt.Sprintf("t=%d,v1=%s", testTS, signHMAC(testTS, body, testSecret))
		tampered := strings.Replace(body, "evt_x", "evt_y", 1)
		assertSignatureError(t, VerifySignature([]byte(tampered), header, []string{testSecret}, now))
	})

	t.Run("rejects a body signed with the wrong secret", func(t *testing.T) {
		header := fmt.Sprintf("t=%d,v1=%s", testTS, signHMAC(testTS, body, testSecretB))
		assertSignatureError(t, VerifySignature([]byte(body), header, []string{testSecret}, now))
	})

	t.Run("accepts a timestamp exactly at the tolerance edge (299s)", func(t *testing.T) {
		old := testTS - 299
		header := fmt.Sprintf("t=%d,v1=%s", old, signHMAC(old, body, testSecret))
		if err := VerifySignature([]byte(body), header, []string{testSecret}, now); err != nil {
			t.Fatalf("expected success at 299s, got %v", err)
		}
	})

	t.Run("rejects a timestamp 301s out as WebhookTimestampError", func(t *testing.T) {
		old := testTS - 301
		header := fmt.Sprintf("t=%d,v1=%s", old, signHMAC(old, body, testSecret))
		assertTimestampError(t, VerifySignature([]byte(body), header, []string{testSecret}, now))
	})

	t.Run("honors a custom tolerance window", func(t *testing.T) {
		old := testTS - 60
		header := fmt.Sprintf("t=%d,v1=%s", old, signHMAC(old, body, testSecret))
		err := VerifySignature([]byte(body), header, []string{testSecret}, now, WithTolerance(30*time.Second))
		assertTimestampError(t, err)
	})

	t.Run("throws WebhookSignatureError for header missing t=", func(t *testing.T) {
		header := "v1=" + strings.Repeat("0", 64)
		assertSignatureError(t, VerifySignature([]byte(body), header, []string{testSecret}, now))
	})

	t.Run("throws WebhookSignatureError for header with no v1 entries", func(t *testing.T) {
		header := fmt.Sprintf("t=%d,foo=bar", testTS)
		assertSignatureError(t, VerifySignature([]byte(body), header, []string{testSecret}, now))
	})

	t.Run("ignores unknown scheme keys (forward-compat for v2= etc.)", func(t *testing.T) {
		sig := signHMAC(testTS, body, testSecret)
		header := fmt.Sprintf("t=%d,v0=ignored,v1=%s,v2=ignored", testTS, sig)
		if err := VerifySignature([]byte(body), header, []string{testSecret}, now); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})

	t.Run("rejects mismatched signature lengths without crashing", func(t *testing.T) {
		halfSig := strings.Repeat("abcd", 8) // 32 hex chars, not 64
		header := fmt.Sprintf("t=%d,v1=%s", testTS, halfSig)
		assertSignatureError(t, VerifySignature([]byte(body), header, []string{testSecret}, now))
	})
}

// TestVerifySignature_KnownAnswer is a known-answer test against a golden HMAC
// computed externally with OpenSSL (independent of this code), proving the SDK
// signs over exactly "<ts>.<body>" with the full secret. Reproduce:
//
//	printf '%s' '1700000000.{"id":"job_abc123","status":"succeeded"}' \
//	  | openssl dgst -sha256 -hmac "whsec_known_answer_test_key_here"
func TestVerifySignature_KnownAnswer(t *testing.T) {
	const (
		secret = "whsec_known_answer_test_key_here"
		ts     = int64(1700000000)
		body   = `{"id":"job_abc123","status":"succeeded"}`
		golden = "738628e4926e9ad49a18b13f0e83519f30e3a79650f68528a4b69dfe27abdd93"
	)
	now := fixedNow(ts)

	header := fmt.Sprintf("t=%d,v1=%s", ts, golden)
	if err := VerifySignature([]byte(body), header, []string{secret}, now); err != nil {
		t.Fatalf("golden signature should verify, got %v", err)
	}

	// Flip the first hex nibble (7→8): must be rejected.
	flipped := "8" + golden[1:]
	badHeader := fmt.Sprintf("t=%d,v1=%s", ts, flipped)
	assertSignatureError(t, VerifySignature([]byte(body), badHeader, []string{secret}, now))
}

func assertSignatureError(t *testing.T, err error) {
	t.Helper()
	var target *WebhookSignatureError
	if !errors.As(err, &target) {
		t.Fatalf("expected *WebhookSignatureError, got %T (%v)", err, err)
	}
}

func assertTimestampError(t *testing.T, err error) {
	t.Helper()
	var target *WebhookTimestampError
	if !errors.As(err, &target) {
		t.Fatalf("expected *WebhookTimestampError, got %T (%v)", err, err)
	}
}
