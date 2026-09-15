package transcodely

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// CreatedIngestRule wraps a freshly created rule with its inbound secret. The
// secret is returned once, here — we keep only enough of it to verify a
// delivery, so it cannot be read back later. Store it wherever the event
// sender will read it from; if you lose it, rotate the rule for a new one.
type CreatedIngestRule struct {
	Rule   *IngestRule
	Secret string
}

// UpdatedIngestRule is the result of an update.
//
// Secret is non-empty only when the update rotated it, and like the create
// secret it is shown once. EventsSkippedWhileDisabled is returned when the
// update switched a paused rule back on: pausing keeps accepting deliveries
// and records them as skipped, but nothing retries them on its own, so this
// is how many objects are waiting for [IngestRules.ReplayEvent].
type UpdatedIngestRule struct {
	Rule                       *IngestRule
	Secret                     string
	EventsSkippedWhileDisabled int64
}

// IngestRules is the Stripe-style namespace for ingest rules — standing
// instructions that turn an object landing in your bucket into a job, with no
// server of yours in the path.
//
// A rule hangs off one readable storage origin. Your provider posts its
// object-created events to the rule's GetEndpointUrl(), authenticated with the
// rule's secret; a delivery that passes the rule's filters becomes a job.
// Duplicate deliveries are absorbed — an object is identified by (rule,
// bucket, key, etag) and produces exactly one job.
//
// The inbound endpoint itself is not part of this SDK: it is called by your
// storage provider, not by you.
type IngestRules struct {
	client transcodelyv1connect.IngestRuleServiceClient
}

func newIngestRules(c transcodelyv1connect.IngestRuleServiceClient) *IngestRules {
	return &IngestRules{client: c}
}

// Create adds an ingest rule to a readable origin. The returned
// CreatedIngestRule carries the inbound secret in full, once.
func (i *IngestRules) Create(ctx context.Context, params *IngestRuleCreateParams) (*CreatedIngestRule, error) {
	if params == nil {
		params = &IngestRuleCreateParams{}
	}
	resp, err := i.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return &CreatedIngestRule{
		Rule:   resp.Msg.GetRule(),
		Secret: resp.Msg.GetSecret(),
	}, nil
}

// Get fetches an ingest rule by ID (`ing_*`). The secret is never returned
// again — the rule carries only GetSecretPrefix() and GetSecretHint().
func (i *IngestRules) Get(ctx context.Context, id string) (*IngestRule, error) {
	resp, err := i.client.Get(ctx, connect.NewRequest(&v1.GetIngestRuleRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetRule(), nil
}

// List returns an iterator over ingest rules, newest first. Narrow it with
// IngestRuleListParams.OriginId or .Enabled.
func (i *IngestRules) List(ctx context.Context, params *IngestRuleListParams) *Iter[*IngestRule] {
	if params == nil {
		params = &IngestRuleListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*IngestRule, string, error) {
		req := proto.Clone(params).(*IngestRuleListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := i.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetRules(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Update mutates a rule's name, enabled state, filters or action, and
// optionally rotates its secret. It MERGES: every field is optional and an
// update applies only what it carries, down to the individual filters and the
// individual parts of the action. Narrowing a rule to a new prefix is
// therefore Filters with only Prefix set, and the suffix, content-type and
// size filters are left alone.
//
// Removing something rather than changing it takes the two clear flags.
// ClearFilters empties the filter set before Filters is applied — on its own
// it widens the rule to every object in the bucket, and together with Filters
// it replaces the set outright. ClearAction throws the stored action away, so
// Action must then be complete: at least one output and exactly one
// destination. That is the only way to drop an action's thumbnails or
// metadata, since a repeated or map field sent empty reads as "not sent".
//
// For the same reason Managed: false does not turn managed storage off — a
// false boolean is indistinguishable from an absent one, so it leaves the
// destination alone. Send OutputOriginId instead.
//
// After a rotation the previous secret keeps working for 24 hours, so the
// sender can be updated without dropping an event.
func (i *IngestRules) Update(ctx context.Context, params *IngestRuleUpdateParams) (*UpdatedIngestRule, error) {
	if params == nil {
		params = &IngestRuleUpdateParams{}
	}
	resp, err := i.client.Update(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return &UpdatedIngestRule{
		Rule:                       resp.Msg.GetRule(),
		Secret:                     resp.Msg.GetSecret(),
		EventsSkippedWhileDisabled: resp.Msg.GetEventsSkippedWhileDisabled(),
	}, nil
}

// Delete removes a rule; its endpoint stops accepting events immediately. The
// storage events it already received are kept, and the rule as it stood at
// deletion is returned.
func (i *IngestRules) Delete(ctx context.Context, id string) (*IngestRule, error) {
	resp, err := i.client.Delete(ctx, connect.NewRequest(&v1.DeleteIngestRuleRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetRule(), nil
}

// ListEvents returns an iterator over the storage events received, newest
// first — every delivery and what came of it. Omit RuleId for every rule in
// scope; set Status to read only the skipped or failed ones.
func (i *IngestRules) ListEvents(ctx context.Context, params *IngestEventListParams) *Iter[*StorageEvent] {
	if params == nil {
		params = &IngestEventListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*StorageEvent, string, error) {
		req := proto.Clone(params).(*IngestEventListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		req.Pagination.Cursor = cursor
		resp, err := i.client.ListEvents(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetEvents(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Test dry-runs an object key against a rule: whether the filters match and,
// when they do, the exact job request the rule would submit. Nothing is
// stored and no job is created.
//
// Supply Etag when you want the preview to name the idempotency key a real
// delivery would carry — the key is derived from it.
func (i *IngestRules) Test(ctx context.Context, params *IngestRuleTestParams) (*IngestRuleTestResult, error) {
	if params == nil {
		params = &IngestRuleTestParams{}
	}
	resp, err := i.client.Test(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// ReplayEvent re-queues a skipped or refused event (`sev_*`), giving the
// object one more pass through the rule.
//
// It exists because deduplication is permanent: re-sending the event, or
// re-uploading the same bytes, is absorbed and produces nothing. The event is
// reset rather than duplicated, so it keeps its ID and its history. Only
// skipped and failed events can be replayed.
func (i *IngestRules) ReplayEvent(ctx context.Context, eventID string) (*StorageEvent, error) {
	resp, err := i.client.ReplayEvent(ctx, connect.NewRequest(&v1.ReplayIngestEventRequest{EventId: eventID}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetEvent(), nil
}
