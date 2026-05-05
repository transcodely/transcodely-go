package transcodely

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
)

// connErrWithMeta builds a *connect.Error with the given code, message, and
// optional metadata (e.g. X-Request-Id, Retry-After). The message is wrapped
// in our JSON envelope so fromConnectError exercises the parsing path too.
func connErrWithMeta(code connect.Code, jsonMsg string, meta map[string]string) *connect.Error {
	e := connect.NewError(code, errors.New(jsonMsg))
	for k, v := range meta {
		e.Meta().Set(k, v)
	}
	return e
}

func TestFromConnectError_StatusMapping(t *testing.T) {
	cases := []struct {
		name       string
		code       connect.Code
		assert     func(error) bool
		wantHTTP   int
	}{
		{"Unauthenticated→Auth", connect.CodeUnauthenticated, func(e error) bool {
			var t *AuthenticationError
			return errors.As(e, &t)
		}, http.StatusUnauthorized},
		{"PermissionDenied→Permission", connect.CodePermissionDenied, func(e error) bool {
			var t *PermissionError
			return errors.As(e, &t)
		}, http.StatusForbidden},
		{"NotFound→NotFound", connect.CodeNotFound, func(e error) bool {
			var t *NotFoundError
			return errors.As(e, &t)
		}, http.StatusNotFound},
		{"AlreadyExists→Conflict", connect.CodeAlreadyExists, func(e error) bool {
			var t *ConflictError
			return errors.As(e, &t)
		}, http.StatusConflict},
		{"InvalidArgument→InvalidRequest", connect.CodeInvalidArgument, func(e error) bool {
			var t *InvalidRequestError
			return errors.As(e, &t)
		}, http.StatusBadRequest},
		{"FailedPrecondition→Precondition", connect.CodeFailedPrecondition, func(e error) bool {
			var t *PreconditionError
			return errors.As(e, &t)
		}, http.StatusPreconditionFailed},
		{"Internal→APIError", connect.CodeInternal, func(e error) bool {
			var t *APIError
			return errors.As(e, &t)
		}, http.StatusInternalServerError},
		{"Unavailable→Connection", connect.CodeUnavailable, func(e error) bool {
			var t *APIConnectionError
			return errors.As(e, &t)
		}, http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := fromConnectError(connErrWithMeta(c.code, "boom", nil))
			if !c.assert(err) {
				t.Errorf("got %T (%v), want a different typed error", err, err)
			}
			var sdkErr Error
			if !errors.As(err, &sdkErr) {
				t.Fatalf("error %T does not implement transcodely.Error", err)
			}
			// HTTP status is exposed via the (un-exported) base; assert via concrete.
			if be, ok := unwrapBase(err); ok && be.HTTPStatus() != c.wantHTTP {
				t.Errorf("HTTPStatus = %d, want %d", be.HTTPStatus(), c.wantHTTP)
			}
		})
	}
}

func TestFromConnectError_NilReturnsNil(t *testing.T) {
	if got := fromConnectError(nil); got != nil {
		t.Errorf("fromConnectError(nil) = %v, want nil", got)
	}
}

func TestFromConnectError_NonConnectErrorWraps(t *testing.T) {
	cause := errors.New("dial tcp: i/o timeout")
	err := fromConnectError(cause)
	var conn *APIConnectionError
	if !errors.As(err, &conn) {
		t.Fatalf("expected *APIConnectionError, got %T", err)
	}
	if !errors.Is(err, cause) {
		t.Errorf("wrapped error should chain back to original cause")
	}
}

func TestFromConnectError_RateLimitReadsRetryAfter(t *testing.T) {
	ce := connErrWithMeta(connect.CodeResourceExhausted, "rate limited", map[string]string{
		"Retry-After": "5",
	})
	err := fromConnectError(ce)
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rl.RetryAfter != 5*time.Second {
		t.Errorf("RetryAfter = %v, want 5s", rl.RetryAfter)
	}
}

func TestFromConnectError_RateLimitWithoutRetryAfter(t *testing.T) {
	ce := connErrWithMeta(connect.CodeResourceExhausted, "rate limited", nil)
	err := fromConnectError(ce)
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rl.RetryAfter != 0 {
		t.Errorf("RetryAfter = %v, want zero", rl.RetryAfter)
	}
}

func TestFromConnectError_ParsesJSONErrorEnvelope(t *testing.T) {
	envelope := `{"type":"validation_error","code":"JOB_INPUT_URL_INVALID","message":"validation failed","errors":[{"field":"input_url","description":"must be a valid URL"}]}`
	ce := connErrWithMeta(connect.CodeInvalidArgument, envelope, nil)
	err := fromConnectError(ce)
	var sdkErr Error
	if !errors.As(err, &sdkErr) {
		t.Fatalf("not a typed Error: %T", err)
	}
	if sdkErr.ErrorCode() != "JOB_INPUT_URL_INVALID" {
		t.Errorf("ErrorCode() = %q, want JOB_INPUT_URL_INVALID", sdkErr.ErrorCode())
	}
	be, _ := unwrapBase(err)
	if be == nil {
		t.Fatal("could not unwrap base error")
	}
	if be.Type() != "validation_error" {
		t.Errorf("Type() = %q, want validation_error", be.Type())
	}
	if len(be.Errors()) != 1 || be.Errors()[0].Field != "input_url" {
		t.Errorf("Errors() = %+v, want one input_url violation", be.Errors())
	}
}

func TestFromConnectError_CapturesRequestID(t *testing.T) {
	ce := connErrWithMeta(connect.CodeInternal, "boom", map[string]string{
		"X-Request-Id": "req_abc123",
	})
	err := fromConnectError(ce)
	var sdkErr Error
	if !errors.As(err, &sdkErr) {
		t.Fatalf("not a typed Error: %T", err)
	}
	if sdkErr.RequestID() != "req_abc123" {
		t.Errorf("RequestID() = %q, want req_abc123", sdkErr.RequestID())
	}
}

func TestHTTPStatusForConnectCode(t *testing.T) {
	cases := []struct {
		code connect.Code
		want int
	}{
		{connect.CodeUnauthenticated, http.StatusUnauthorized},
		{connect.CodePermissionDenied, http.StatusForbidden},
		{connect.CodeNotFound, http.StatusNotFound},
		{connect.CodeAlreadyExists, http.StatusConflict},
		{connect.CodeInvalidArgument, http.StatusBadRequest},
		{connect.CodeFailedPrecondition, http.StatusPreconditionFailed},
		{connect.CodeInternal, http.StatusInternalServerError},
		{connect.CodeResourceExhausted, http.StatusTooManyRequests},
		{connect.CodeUnavailable, http.StatusServiceUnavailable},
		{connect.CodeDeadlineExceeded, http.StatusGatewayTimeout},
		{connect.CodeUnimplemented, http.StatusNotImplemented},
	}
	for _, c := range cases {
		if got := httpStatusForConnectCode(c.code); got != c.want {
			t.Errorf("code=%v: got %d, want %d", c.code, got, c.want)
		}
	}
}

// unwrapBase exposes the *baseError sitting inside any of the typed SDK errors,
// so tests can assert on Type / HTTPStatus / Errors without re-implementing
// every concrete-type switch.
func unwrapBase(err error) (*baseError, bool) {
	switch e := err.(type) {
	case *AuthenticationError:
		return &e.baseError, true
	case *PermissionError:
		return &e.baseError, true
	case *NotFoundError:
		return &e.baseError, true
	case *ConflictError:
		return &e.baseError, true
	case *RateLimitError:
		return &e.baseError, true
	case *InvalidRequestError:
		return &e.baseError, true
	case *PreconditionError:
		return &e.baseError, true
	case *APIError:
		return &e.baseError, true
	case *APIConnectionError:
		return &e.baseError, true
	}
	return nil, false
}
