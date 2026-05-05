package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Presets is the Stripe-style namespace for reusable transcoding presets.
type Presets struct {
	client transcodelyv1connect.PresetServiceClient
}

func newPresets(c transcodelyv1connect.PresetServiceClient) *Presets {
	return &Presets{client: c}
}

// Create defines a new preset.
func (p *Presets) Create(ctx context.Context, params *PresetCreateParams) (*Preset, error) {
	if params == nil {
		params = &PresetCreateParams{}
	}
	resp, err := p.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetPreset(), nil
}

// Get fetches a preset by ID (`pst_*`).
func (p *Presets) Get(ctx context.Context, id string) (*Preset, error) {
	resp, err := p.client.Get(ctx, connect.NewRequest(&v1.GetPresetRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetPreset(), nil
}

// GetBySlug fetches a preset by human-friendly slug.
func (p *Presets) GetBySlug(ctx context.Context, slug string) (*Preset, error) {
	resp, err := p.client.GetBySlug(ctx, connect.NewRequest(&v1.GetPresetBySlugRequest{Slug: slug}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetPreset(), nil
}

// List returns an iterator over presets in the active app.
func (p *Presets) List(ctx context.Context, params *PresetListParams) *Iter[*Preset] {
	if params == nil {
		params = &PresetListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Preset, string, error) {
		req := proto.Clone(params).(*PresetListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := p.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetPresets(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Update mutates a preset.
func (p *Presets) Update(ctx context.Context, params *PresetUpdateParams) (*Preset, error) {
	if params == nil {
		params = &PresetUpdateParams{}
	}
	resp, err := p.client.Update(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetPreset(), nil
}

// Duplicate copies an existing preset under a new slug + name. Both are required.
func (p *Presets) Duplicate(ctx context.Context, sourceID, slug, name string) (*Preset, error) {
	req := &v1.DuplicatePresetRequest{SourceId: sourceID, Slug: slug, Name: name}
	resp, err := p.client.Duplicate(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetPreset(), nil
}

// Archive soft-deletes a preset. Existing jobs that reference it continue
// to run; future job creations are rejected.
func (p *Presets) Archive(ctx context.Context, id string) error {
	_, err := p.client.Archive(ctx, connect.NewRequest(&v1.ArchivePresetRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}
