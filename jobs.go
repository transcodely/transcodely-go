package transcodely

import (
	"context"
	"io"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Jobs is the Stripe-style namespace for transcoding jobs. Reach it via
// client.Jobs.
type Jobs struct {
	client       transcodelyv1connect.JobServiceClient
	streamClient transcodelyv1connect.JobServiceClient
	cfg          *config
}

func newJobs(c, stream transcodelyv1connect.JobServiceClient, cfg *config) *Jobs {
	return &Jobs{client: c, streamClient: stream, cfg: cfg}
}

// Create submits a new transcoding job. If params.IdempotencyKey is empty, the
// SDK auto-generates one so retried calls are server-side deduped.
func (j *Jobs) Create(ctx context.Context, params *JobCreateParams) (*Job, error) {
	if params == nil {
		params = &JobCreateParams{}
	}
	resp, err := j.client.Create(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetJob(), nil
}

// Get fetches a single job by ID.
func (j *Jobs) Get(ctx context.Context, id string) (*Job, error) {
	resp, err := j.client.Get(ctx, connect.NewRequest(&v1.GetJobRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetJob(), nil
}

// List returns an auto-paging iterator over jobs. Mutate params.Pagination to
// override the page size or starting cursor.
func (j *Jobs) List(ctx context.Context, params *JobListParams) *Iter[*Job] {
	if params == nil {
		params = &JobListParams{}
	}
	limit := int32(0)
	if params.GetPagination() != nil {
		limit = params.GetPagination().GetLimit()
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Job, string, error) {
		req := proto.Clone(params).(*JobListParams)
		if req.Pagination == nil {
			req.Pagination = &PaginationRequest{}
		}
		if limit > 0 {
			req.Pagination.Limit = limit
		}
		req.Pagination.Cursor = cursor
		resp, err := j.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetJobs(), resp.Msg.GetPagination().GetNextCursor(), nil
	})
}

// Cancel stops a pending or running job.
func (j *Jobs) Cancel(ctx context.Context, id string) (*Job, error) {
	resp, err := j.client.Cancel(ctx, connect.NewRequest(&v1.CancelJobRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetJob(), nil
}

// Confirm starts a job currently in AWAITING_CONFIRMATION (created with
// delay_start = true). Returns the job in PROCESSING state.
func (j *Jobs) Confirm(ctx context.Context, id string) (*Job, error) {
	resp, err := j.client.Confirm(ctx, connect.NewRequest(&v1.ConfirmJobRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetJob(), nil
}

// Watch opens a server-stream that pushes job events until the job reaches a
// terminal state. Heartbeats are filtered transparently. Always defer Close().
func (j *Jobs) Watch(ctx context.Context, id string) *Stream[*WatchJobResponse] {
	return newStream(ctx, j.cfg.maxRetries,
		func(msg *WatchJobResponse) bool {
			return msg.GetEvent() == WatchEventHeartbeat
		},
		func(ctx context.Context) (streamConn[*WatchJobResponse], error) {
			conn, err := j.streamClient.Watch(ctx, connect.NewRequest(&v1.WatchJobRequest{Id: id}))
			if err != nil {
				return nil, fromConnectError(err)
			}
			return &jobWatchAdapter{stream: conn}, nil
		})
}

type jobWatchAdapter struct {
	stream *connect.ServerStreamForClient[v1.WatchJobResponse]
}

func (a *jobWatchAdapter) Recv() (*WatchJobResponse, error) {
	if !a.stream.Receive() {
		if err := a.stream.Err(); err != nil {
			return nil, fromConnectError(err)
		}
		return nil, io.EOF
	}
	return a.stream.Msg(), nil
}

func (a *jobWatchAdapter) Close() error {
	if a.stream == nil {
		return nil
	}
	return a.stream.Close()
}
