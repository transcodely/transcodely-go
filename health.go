package transcodely

import (
	"context"

	"connectrpc.com/connect"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Health is the Stripe-style namespace for liveness checks.
type Health struct {
	client transcodelyv1connect.HealthServiceClient
}

func newHealth(c transcodelyv1connect.HealthServiceClient) *Health {
	return &Health{client: c}
}

// Check pings the API. Pass an empty service string for an overall check.
func (h *Health) Check(ctx context.Context, service string) (*HealthCheckResponse, error) {
	req := &v1.HealthCheckRequest{}
	if service != "" {
		req.Service = &service
	}
	resp, err := h.client.Check(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}
