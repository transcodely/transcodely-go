package transcodely

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestPreconditionError_BillingPastDueFromWire replays a billing_past_due
// refusal exactly as the API (connect-go v1.18.1) sends it: HTTP 400 with
// Connect code failed_precondition. It must surface as *PreconditionError
// and report the status the wire actually carried.
func TestPreconditionError_BillingPastDueFromWire(t *testing.T) {
	const msg = "billing past due: new jobs are paused until the overdue invoice is paid"
	detail, err := proto.Marshal(&v1.ErrorDetails{Code: "billing_past_due", Message: msg})
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(map[string]any{
		"code":    "failed_precondition",
		"message": msg,
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

	var pre *PreconditionError
	if !errors.As(err, &pre) {
		t.Fatalf("want *PreconditionError, got %T: %v", err, err)
	}
	if pre.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("HTTPStatus() = %d, want 400 (what the wire carried)", pre.HTTPStatus())
	}
	if pre.ErrorCode() != "billing_past_due" {
		t.Errorf("ErrorCode() = %q, want billing_past_due", pre.ErrorCode())
	}
	if pre.Type() != "failed_precondition" {
		t.Errorf("Type() = %q, want failed_precondition", pre.Type())
	}
	if pre.RequestID() != "req_test" {
		t.Errorf("RequestID() = %q", pre.RequestID())
	}
}

// TestHTTPStatusForConnectCode_MatchesConnectGo checks every Connect code
// against the status connect-go itself writes for it (its unexported
// connectCodeToHTTP, reached through ErrorWriter on a Connect-protocol
// request), so the SDK's synthesized HTTPStatus() can never drift from the
// wire again.
func TestHTTPStatusForConnectCode_MatchesConnectGo(t *testing.T) {
	ew := connect.NewErrorWriter()
	for code := connect.CodeCanceled; code <= connect.CodeUnauthenticated; code++ {
		req := httptest.NewRequest(http.MethodPost, "/transcodely.v1.JobService/GetJob", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		if !ew.IsSupported(req) {
			t.Fatal("ErrorWriter does not recognise a Connect unary JSON request")
		}
		rec := httptest.NewRecorder()
		if err := ew.Write(rec, req, connect.NewError(code, errors.New("boom"))); err != nil {
			t.Fatalf("code=%v: ErrorWriter.Write: %v", code, err)
		}
		if got := httpStatusForConnectCode(code); got != rec.Code {
			t.Errorf("code=%v: httpStatusForConnectCode = %d, connect-go writes %d", code, got, rec.Code)
		}
	}
}
