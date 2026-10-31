package transcodely

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestInvalidRequestError_ReadsErrorDetailsFromWire pins the envelope every
// API validator sends (protovalidate schema, codec and feature checks alike):
// Connect invalid_argument with an `error-code` header and a
// transcodely.v1.ErrorDetails detail whose field violations read
// "[<code>] <message>". The violations must reach InvalidRequestError.Errors().
func TestInvalidRequestError_ReadsErrorDetailsFromWire(t *testing.T) {
	detail, err := proto.Marshal(&v1.ErrorDetails{
		Code:    "validation_error",
		Message: "validation failed: outputs[0].content_aware.auto_abr.min_variants: value must be greater than or equal to 1 and less than or equal to 20",
		FieldViolations: []*v1.FieldViolation{{
			Field:       "outputs[0].content_aware.auto_abr.min_variants",
			Description: "[parameter_out_of_range] value must be greater than or equal to 1 and less than or equal to 20",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(map[string]any{
		"code":    "invalid_argument",
		"message": "validation failed: outputs[0].content_aware.auto_abr.min_variants: value must be greater than or equal to 1 and less than or equal to 20",
		"details": []map[string]any{{
			"type":  "transcodely.v1.ErrorDetails",
			"value": base64.RawStdEncoding.EncodeToString(detail),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Error-Code", "invalid_argument")
		w.Header().Set("X-Request-Id", "req_test")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client, err := New("ak_test", WithBaseURL(srv.URL), WithMaxRetries(0))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Jobs.Get(context.Background(), "job_abc123def456")

	var invalid *InvalidRequestError
	if !errors.As(err, &invalid) {
		t.Fatalf("want *InvalidRequestError, got %T: %v", err, err)
	}
	got := invalid.Errors()
	if len(got) != 1 {
		t.Fatalf("want 1 field violation, got %d (%+v)", len(got), got)
	}
	if got[0].Field != "outputs[0].content_aware.auto_abr.min_variants" {
		t.Errorf("field = %q", got[0].Field)
	}
	if got[0].Description != "[parameter_out_of_range] value must be greater than or equal to 1 and less than or equal to 20" {
		t.Errorf("description = %q", got[0].Description)
	}
	if invalid.ErrorCode() != "validation_error" {
		t.Errorf("ErrorCode() = %q, want validation_error", invalid.ErrorCode())
	}
	if invalid.Type() != "invalid_argument" {
		t.Errorf("Type() = %q, want invalid_argument (the Connect code, as the JS and Python SDKs report it)", invalid.Type())
	}
	if invalid.RequestID() != "req_test" {
		t.Errorf("RequestID() = %q", invalid.RequestID())
	}
}
