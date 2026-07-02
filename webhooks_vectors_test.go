package transcodely

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// webhookVector is one entry of the shared, cross-SDK conformance corpus.
// The file (testdata/webhook_vectors.json) is a verbatim copy of the corpus
// every Transcodely SDK must pass — do not edit it here; resync from source.
type webhookVector struct {
	Name          string   `json:"name"`
	Secret        string   `json:"secret"`
	Secrets       []string `json:"secrets"`
	SigningSecret string   `json:"signing_secret"`
	SigningBody   string   `json:"signing_body"`
	TS            int64    `json:"ts"`
	Body          string   `json:"body"`
	Tolerance     int64    `json:"tolerance"`
	Now           int64    `json:"now"`
	Expect        struct {
		Result         string `json:"result"`
		EventType      string `json:"event_type"`
		EventID        string `json:"event_id"`
		DataID         string `json:"data_id"`
		IdempotencyKey string `json:"idempotency_key"`
	} `json:"expect"`
}

func TestWebhookConformanceCorpus(t *testing.T) {
	raw, err := os.ReadFile("testdata/webhook_vectors.json")
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	var file struct {
		Vectors []webhookVector `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse corpus: %v", err)
	}
	if len(file.Vectors) == 0 {
		t.Fatal("corpus is empty")
	}

	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			// Derive the signature header at runtime, exactly as the corpus
			// spec dictates: t=<ts>,v1=HMAC-SHA-256(signing_secret, <ts>.<signing_body>).
			signingSecret := firstNonEmpty(v.SigningSecret, v.Secret, first(v.Secrets))
			signingBody := v.SigningBody
			if signingBody == "" {
				signingBody = v.Body
			}
			sigHeader := fmt.Sprintf("t=%d,v1=%s", v.TS, signHMAC(v.TS, signingBody, signingSecret))

			verifySecrets := v.Secrets
			if verifySecrets == nil {
				verifySecrets = []string{v.Secret}
			}
			opts := []VerifyOption{
				withNow(func() int64 { return v.Now }),
				WithTolerance(time.Duration(v.Tolerance) * time.Second),
			}

			event, err := ConstructEventWithSecrets([]byte(v.Body), sigHeader, verifySecrets, opts...)

			switch v.Expect.Result {
			case "ok":
				if err != nil {
					t.Fatalf("expected ok, got error: %v", err)
				}
				if v.Expect.EventType != "" && string(event.Type) != v.Expect.EventType {
					t.Errorf("Type = %q, want %q", event.Type, v.Expect.EventType)
				}
				if v.Expect.EventID != "" && event.ID != v.Expect.EventID {
					t.Errorf("ID = %q, want %q", event.ID, v.Expect.EventID)
				}
				if v.Expect.DataID != "" {
					getter, ok := event.Data.(interface{ GetId() string })
					if !ok {
						t.Fatalf("data does not expose GetId(): %T", event.Data)
					}
					if getter.GetId() != v.Expect.DataID {
						t.Errorf("data id = %q, want %q", getter.GetId(), v.Expect.DataID)
					}
				}
				if v.Expect.IdempotencyKey != "" && event.Request.IdempotencyKey != v.Expect.IdempotencyKey {
					t.Errorf("IdempotencyKey = %q, want %q", event.Request.IdempotencyKey, v.Expect.IdempotencyKey)
				}
			case "signature_error":
				assertSignatureError(t, err)
			case "timestamp_error":
				assertTimestampError(t, err)
			case "payload_error":
				assertPayloadError(t, err)
			default:
				t.Fatalf("unknown expected result %q", v.Expect.Result)
			}
		})
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func first(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}
