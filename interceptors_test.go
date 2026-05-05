package transcodely

import (
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
)

func TestIsMutating(t *testing.T) {
	cases := []struct {
		procedure string
		want      bool
	}{
		{"/transcodely.v1.JobService/Create", true},
		{"/transcodely.v1.PresetService/Create", true},
		{"/transcodely.v1.JobService/Get", false},
		{"/transcodely.v1.JobService/List", false},
		{"/transcodely.v1.JobService/Update", false},
		{"/transcodely.v1.JobService/Cancel", false},
		{"/transcodely.v1.JobService/Watch", false},
		// "/CreatePolicy" must NOT match — the suffix has to be exactly "/Create".
		{"/transcodely.v1.PolicyService/CreatePolicy", false},
	}
	for _, c := range cases {
		if got := isMutating(c.procedure); got != c.want {
			t.Errorf("isMutating(%q) = %v, want %v", c.procedure, got, c.want)
		}
	}
}

func TestEndsWith(t *testing.T) {
	if !endsWith("/foo/Create", "/Create") {
		t.Errorf("endsWith should match /Create at the tail")
	}
	if endsWith("/foo/CreateThing", "/Create") {
		t.Errorf("endsWith must not match a partial tail")
	}
	if endsWith("Create", "/Create") {
		t.Errorf("endsWith must not match when source is shorter than suffix")
	}
}

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"network error (no Connect wrapper) → retry", errors.New("dial tcp"), true},
		{"Unavailable → retry", connect.NewError(connect.CodeUnavailable, errors.New("x")), true},
		{"DeadlineExceeded → retry", connect.NewError(connect.CodeDeadlineExceeded, errors.New("x")), true},
		{"Aborted → retry", connect.NewError(connect.CodeAborted, errors.New("x")), true},
		{"ResourceExhausted → retry", connect.NewError(connect.CodeResourceExhausted, errors.New("x")), true},
		{"NotFound → no retry", connect.NewError(connect.CodeNotFound, errors.New("x")), false},
		{"InvalidArgument → no retry", connect.NewError(connect.CodeInvalidArgument, errors.New("x")), false},
		{"Unauthenticated → no retry", connect.NewError(connect.CodeUnauthenticated, errors.New("x")), false},
		{"PermissionDenied → no retry", connect.NewError(connect.CodePermissionDenied, errors.New("x")), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isRetryable(c.err); got != c.want {
				t.Errorf("isRetryable(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestBackoff_HonorsRetryAfter(t *testing.T) {
	ce := connect.NewError(connect.CodeResourceExhausted, errors.New("rate limited"))
	ce.Meta().Set("Retry-After", "3")
	got := backoff(0, ce)
	if got != 3*time.Second {
		t.Errorf("backoff with Retry-After=3: got %v, want 3s", got)
	}
}

func TestBackoff_FallsBackToExponentialJitter(t *testing.T) {
	// No Retry-After header → result must be in [0, 2^attempt * 100ms).
	plain := errors.New("transient")
	for attempt := 0; attempt < 6; attempt++ {
		max := time.Duration(1<<attempt) * 100 * time.Millisecond
		if max > 4*time.Second {
			max = 4 * time.Second
		}
		got := backoff(attempt, plain)
		if got < 0 || got >= max {
			t.Errorf("attempt=%d: got %v, want [0, %v)", attempt, got, max)
		}
	}
}

func TestBackoff_IgnoresUnparseableRetryAfter(t *testing.T) {
	ce := connect.NewError(connect.CodeResourceExhausted, errors.New("rate limited"))
	ce.Meta().Set("Retry-After", "later")
	got := backoff(0, ce)
	// Falls through to the jitter path: <100ms.
	if got >= 100*time.Millisecond {
		t.Errorf("got %v, expected jitter (<100ms)", got)
	}
}

func TestInterceptorFunc_NilHandlersFallThrough(t *testing.T) {
	// Verify the adapter never panics when the unary/streaming hook is nil.
	f := interceptorFunc{}
	if got := f.WrapUnary(nil); got != nil {
		t.Errorf("WrapUnary with nil 'unary' should return next (nil), got %T", got)
	}
	if got := f.WrapStreamingClient(nil); got != nil {
		t.Errorf("WrapStreamingClient with nil 'streamingClient' should return next (nil), got %T", got)
	}
}
