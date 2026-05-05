package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// CreatedAPIKey wraps a freshly created key with its one-time-visible secret.
type CreatedAPIKey struct {
	Key       *APIKey
	PlainText string
}

// APIKeys is the Stripe-style namespace for API keys.
type APIKeys struct {
	client transcodelyv1connect.APIKeyServiceClient
}

func newAPIKeys(c transcodelyv1connect.APIKeyServiceClient) *APIKeys {
	return &APIKeys{client: c}
}

// Create issues a new API key. The plaintext secret is only available once,
// in the returned CreatedAPIKey — store it immediately.
func (a *APIKeys) Create(ctx context.Context, params *APIKeyCreateParams) (*CreatedAPIKey, error) {
	if params == nil {
		params = &APIKeyCreateParams{}
	}
	resp, err := a.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return &CreatedAPIKey{
		Key:       resp.Msg.GetApiKey(),
		PlainText: resp.Msg.GetSecret(),
	}, nil
}

// Get fetches an API key by ID. The plaintext is never returned again.
func (a *APIKeys) Get(ctx context.Context, id string) (*APIKey, error) {
	resp, err := a.client.Get(ctx, connect.NewRequest(&v1.GetAPIKeyRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetApiKey(), nil
}

// List returns an iterator over API keys.
func (a *APIKeys) List(ctx context.Context, params *APIKeyListParams) *Iter[*APIKey] {
	if params == nil {
		params = &APIKeyListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*APIKey, string, error) {
		req := proto.Clone(params).(*APIKeyListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := a.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetApiKeys(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Revoke immediately deactivates an API key.
func (a *APIKeys) Revoke(ctx context.Context, id string) error {
	_, err := a.client.Revoke(ctx, connect.NewRequest(&v1.RevokeAPIKeyRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}
