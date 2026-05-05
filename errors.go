package transcodely

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"
)

// FieldViolation describes a single field-level validation failure.
type FieldViolation struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

// Error is the base interface implemented by every typed error the SDK returns.
//
// All concrete error types (*APIConnectionError, *APIError, *AuthenticationError,
// *PermissionError, *NotFoundError, *ConflictError, *RateLimitError,
// *InvalidRequestError, *PreconditionError) embed *baseError, so a single
// errors.As(err, &target) check is enough for typed handling.
type Error interface {
	error
	// Code is the server-side machine-readable code (e.g. "JOB_NOT_FOUND").
	ErrorCode() string
	// RequestID is the `req_*` identifier the server returned in `X-Request-Id`,
	// or the empty string if absent.
	RequestID() string
}

type baseError struct {
	msg       string
	code      string
	typ       string
	httpCode  int
	requestID string
	errors    []FieldViolation
	cause     error
}

func (e *baseError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.code != "" && e.msg != "" {
		return fmt.Sprintf("transcodely: %s: %s", e.code, e.msg)
	}
	if e.msg != "" {
		return "transcodely: " + e.msg
	}
	return "transcodely: unknown error"
}

func (e *baseError) Unwrap() error      { return e.cause }
func (e *baseError) ErrorCode() string  { return e.code }
func (e *baseError) RequestID() string  { return e.requestID }
func (e *baseError) Type() string       { return e.typ }
func (e *baseError) HTTPStatus() int    { return e.httpCode }
func (e *baseError) Errors() []FieldViolation {
	if e == nil {
		return nil
	}
	return e.errors
}

// APIConnectionError indicates a network failure: DNS resolution, TLS handshake,
// connection refused, no HTTP response received.
type APIConnectionError struct{ baseError }

// APIError indicates a server-side internal error (5xx).
type APIError struct{ baseError }

// AuthenticationError indicates a 401: missing, invalid, revoked or expired API key.
type AuthenticationError struct{ baseError }

// PermissionError indicates a 403: authenticated but lacking permission.
type PermissionError struct{ baseError }

// NotFoundError indicates a 404: the requested resource was not found.
type NotFoundError struct{ baseError }

// ConflictError indicates a 409: already exists, idempotency conflict, slug taken.
type ConflictError struct{ baseError }

// RateLimitError indicates a 429: quota or rate limit exceeded. RetryAfter
// reflects the `Retry-After` response header.
type RateLimitError struct {
	baseError
	RetryAfter time.Duration
}

// InvalidRequestError indicates a 400/422: request body or parameters invalid.
// Errors() returns the per-field violations from the server.
type InvalidRequestError struct{ baseError }

// PreconditionError indicates a 412: preconditions not met (e.g. job not
// cancelable in current state).
type PreconditionError struct{ baseError }

// errorPayload mirrors the server JSON error envelope.
type errorPayload struct {
	Type    string           `json:"type,omitempty"`
	Code    string           `json:"code,omitempty"`
	Message string           `json:"message,omitempty"`
	Errors  []FieldViolation `json:"errors,omitempty"`
}

// fromConnectError converts a *connect.Error returned by the generated client
// into a typed SDK error.
func fromConnectError(err error) error {
	if err == nil {
		return nil
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return &APIConnectionError{baseError: baseError{
			msg:   err.Error(),
			cause: err,
		}}
	}

	base := baseError{
		msg:      connectErr.Message(),
		httpCode: httpStatusForConnectCode(connectErr.Code()),
		cause:    err,
	}

	// Try to recover the structured error envelope from the message body.
	// Some servers return JSON in the message; fall back to plain text.
	if payload := parseErrorPayload(connectErr.Message()); payload != nil {
		base.code = payload.Code
		base.typ = payload.Type
		base.errors = payload.Errors
		if payload.Message != "" {
			base.msg = payload.Message
		}
	}

	if connectErr.Meta() != nil {
		if rid := connectErr.Meta().Get("X-Request-Id"); rid != "" {
			base.requestID = rid
		}
	}

	switch connectErr.Code() {
	case connect.CodeUnauthenticated:
		return &AuthenticationError{baseError: base}
	case connect.CodePermissionDenied:
		return &PermissionError{baseError: base}
	case connect.CodeNotFound:
		return &NotFoundError{baseError: base}
	case connect.CodeAlreadyExists:
		return &ConflictError{baseError: base}
	case connect.CodeResourceExhausted:
		retryAfter := time.Duration(0)
		if connectErr.Meta() != nil {
			if v := connectErr.Meta().Get("Retry-After"); v != "" {
				if secs, err := strconv.Atoi(v); err == nil {
					retryAfter = time.Duration(secs) * time.Second
				}
			}
		}
		return &RateLimitError{baseError: base, RetryAfter: retryAfter}
	case connect.CodeInvalidArgument:
		return &InvalidRequestError{baseError: base}
	case connect.CodeFailedPrecondition:
		return &PreconditionError{baseError: base}
	case connect.CodeInternal, connect.CodeUnknown, connect.CodeDataLoss:
		return &APIError{baseError: base}
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded, connect.CodeAborted:
		return &APIConnectionError{baseError: base}
	}
	return &APIError{baseError: base}
}

func parseErrorPayload(message string) *errorPayload {
	if len(message) == 0 || message[0] != '{' {
		return nil
	}
	var p errorPayload
	if err := json.Unmarshal([]byte(message), &p); err != nil {
		return nil
	}
	return &p
}

// httpStatusForConnectCode mirrors the canonical Connect→HTTP status mapping.
// Useful for surfacing HTTPStatus() to callers without forcing a connect-go dep.
func httpStatusForConnectCode(code connect.Code) int {
	switch code {
	case connect.CodeCanceled:
		return 408
	case connect.CodeUnknown, connect.CodeInternal, connect.CodeDataLoss:
		return 500
	case connect.CodeInvalidArgument, connect.CodeOutOfRange:
		return 400
	case connect.CodeDeadlineExceeded:
		return 504
	case connect.CodeNotFound:
		return 404
	case connect.CodeAlreadyExists, connect.CodeAborted:
		return 409
	case connect.CodePermissionDenied:
		return 403
	case connect.CodeResourceExhausted:
		return 429
	case connect.CodeFailedPrecondition:
		return 412
	case connect.CodeUnimplemented:
		return 501
	case connect.CodeUnavailable:
		return 503
	case connect.CodeUnauthenticated:
		return 401
	}
	return 500
}
