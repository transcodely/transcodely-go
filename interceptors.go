package transcodely

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

// authInterceptor injects Authorization, User-Agent, and the calendar API
// version header on every outgoing request. It also captures X-Request-Id
// from responses so downstream interceptors and error helpers can surface it.
func authInterceptor(cfg *config) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+cfg.apiKey)
			req.Header().Set("User-Agent", cfg.userAgent)
			req.Header().Set("Transcodely-Version", cfg.apiVersion)
			return next(ctx, req)
		}
	})
}

// streamingAuthInterceptor mirrors authInterceptor for server-streaming RPCs
// (Watch). Unary and streaming clients each need their own interceptor wiring.
func streamingAuthInterceptor(cfg *config) connect.Interceptor {
	return interceptorFunc{
		unary: func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+cfg.apiKey)
				req.Header().Set("User-Agent", cfg.userAgent)
				req.Header().Set("Transcodely-Version", cfg.apiVersion)
				return next(ctx, req)
			}
		},
		streamingClient: func(next connect.StreamingClientFunc) connect.StreamingClientFunc {
			return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
				conn := next(ctx, spec)
				conn.RequestHeader().Set("Authorization", "Bearer "+cfg.apiKey)
				conn.RequestHeader().Set("User-Agent", cfg.userAgent)
				conn.RequestHeader().Set("Transcodely-Version", cfg.apiVersion)
				return conn
			}
		},
	}
}

// idempotencyInterceptor injects an Idempotency-Key header on the Create RPC of
// every service when the caller did not supply one. Mutations on /Create
// methods are the ones that benefit from server-side idempotency.
func idempotencyInterceptor() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if isMutating(req.Spec().Procedure) && req.Header().Get("Idempotency-Key") == "" {
				req.Header().Set("Idempotency-Key", uuid.New().String())
			}
			return next(ctx, req)
		}
	})
}

// retryInterceptor retries transient failures with exponential backoff and
// full jitter. Only safe-by-construction RPCs are retried automatically;
// non-Create mutations are also retried when the SDK is responsible for the
// Idempotency-Key (so duplicate requests are server-side deduped).
func retryInterceptor(cfg *config) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if cfg.maxRetries == 0 {
				return next(ctx, req)
			}
			var lastErr error
			for attempt := 0; attempt <= cfg.maxRetries; attempt++ {
				resp, err := next(ctx, req)
				if err == nil {
					return resp, nil
				}
				lastErr = err
				if !isRetryable(err) {
					return nil, err
				}
				if attempt == cfg.maxRetries {
					break
				}
				delay := backoff(attempt, err)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
			}
			return nil, lastErr
		}
	})
}

func isMutating(procedure string) bool {
	// Heuristic: only POST-shaped Create RPCs are auto-idempotent. Update,
	// Cancel, Confirm, etc. already have a deterministic side-effect target
	// (a specific entity ID) so a duplicate request is naturally idempotent.
	return endsWith(procedure, "/Create")
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func isRetryable(err error) bool {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		// Network failure (no HTTP response) — always safe to retry.
		return true
	}
	switch connectErr.Code() {
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded,
		connect.CodeAborted, connect.CodeResourceExhausted:
		return true
	}
	return false
}

func backoff(attempt int, err error) time.Duration {
	// Honor server-supplied Retry-After if present.
	var connectErr *connect.Error
	if errors.As(err, &connectErr) && connectErr.Meta() != nil {
		if v := connectErr.Meta().Get("Retry-After"); v != "" {
			if secs, perr := strconv.Atoi(v); perr == nil {
				return time.Duration(secs) * time.Second
			}
		}
	}
	base := 100 * time.Millisecond
	max := 4 * time.Second
	exp := time.Duration(math.Pow(2, float64(attempt))) * base
	if exp > max {
		exp = max
	}
	jitter := time.Duration(rand.Int63n(int64(exp))) //nolint:gosec // jitter only
	return jitter
}

// interceptorFunc adapts ad-hoc functions to the connect.Interceptor interface
// when we need to override both unary and streaming handling.
type interceptorFunc struct {
	unary           func(connect.UnaryFunc) connect.UnaryFunc
	streamingClient func(connect.StreamingClientFunc) connect.StreamingClientFunc
}

func (f interceptorFunc) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	if f.unary == nil {
		return next
	}
	return f.unary(next)
}

func (f interceptorFunc) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	if f.streamingClient == nil {
		return next
	}
	return f.streamingClient(next)
}

func (f interceptorFunc) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
