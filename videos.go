package transcodely

import (
	"context"
	"io"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// Videos is the Stripe-style namespace for hosted videos and uploads.
// Reach it via client.Videos.
type Videos struct {
	client       transcodelyv1connect.VideoServiceClient
	streamClient transcodelyv1connect.VideoServiceClient
	cfg          *config
}

func newVideos(c, stream transcodelyv1connect.VideoServiceClient, cfg *config) *Videos {
	return &Videos{client: c, streamClient: stream, cfg: cfg}
}

// CreateUpload returns a presigned URL for uploading a single-part file.
// For files >100 MB, prefer CreateMultipartUpload instead.
func (v *Videos) CreateUpload(ctx context.Context, params *UploadCreateParams) (*v1.CreateUploadResponse, error) {
	if params == nil {
		params = &UploadCreateParams{}
	}
	resp, err := v.client.CreateUpload(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// CompleteUpload finalises a single-part upload. The video transitions to
// `processing` and a transcoding job is created automatically.
func (v *Videos) CompleteUpload(ctx context.Context, params *UploadCompleteParams) (*Video, error) {
	if params == nil {
		params = &UploadCompleteParams{}
	}
	resp, err := v.client.CompleteUpload(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetVideo(), nil
}

// CreateFromUrl creates a hosted video from a publicly-reachable http(s) URL
// in a single call, with no presigned-upload round-trip. The video is
// returned in `processing` status; playback/embed URLs populate once it
// reaches `ready` (subscribe to the video.ready webhook event or use Watch).
func (v *Videos) CreateFromUrl(ctx context.Context, params *VideoCreateFromUrlParams) (*Video, error) {
	if params == nil {
		params = &VideoCreateFromUrlParams{}
	}
	resp, err := v.client.CreateFromUrl(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetVideo(), nil
}

// CreateMultipartUpload begins a multipart upload session.
func (v *Videos) CreateMultipartUpload(ctx context.Context, params *MultipartCreateParams) (*v1.CreateMultipartUploadResponse, error) {
	if params == nil {
		params = &MultipartCreateParams{}
	}
	resp, err := v.client.CreateMultipartUpload(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// GetUploadPartURLs requests presigned URLs for the given part numbers.
func (v *Videos) GetUploadPartURLs(ctx context.Context, params *MultipartPartURLsParams) (*v1.GetUploadPartUrlsResponse, error) {
	if params == nil {
		params = &MultipartPartURLsParams{}
	}
	resp, err := v.client.GetUploadPartUrls(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// CompleteMultipartUpload finalises a multipart upload. Pass the etags
// returned by S3 / GCS for each part.
func (v *Videos) CompleteMultipartUpload(ctx context.Context, params *MultipartCompleteParams) (*Video, error) {
	if params == nil {
		params = &MultipartCompleteParams{}
	}
	resp, err := v.client.CompleteMultipartUpload(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetVideo(), nil
}

// AbortMultipartUpload cancels an in-flight multipart upload and frees
// server-side state.
func (v *Videos) AbortMultipartUpload(ctx context.Context, params *MultipartAbortParams) error {
	if params == nil {
		params = &MultipartAbortParams{}
	}
	_, err := v.client.AbortMultipartUpload(ctx, connect.NewRequest(params))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}

// Get fetches a video by ID.
func (v *Videos) Get(ctx context.Context, id string) (*Video, error) {
	resp, err := v.client.Get(ctx, connect.NewRequest(&v1.GetVideoRequest{Id: id}))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetVideo(), nil
}

// List returns an iterator over videos. The Video service uses page-token
// pagination (PageSize / PageToken) rather than the cursor model.
func (v *Videos) List(ctx context.Context, params *VideoListParams) *Iter[*Video] {
	if params == nil {
		params = &VideoListParams{}
	}
	return newIter(ctx, func(ctx context.Context, cursor string) ([]*Video, string, error) {
		req := proto.Clone(params).(*VideoListParams)
		if cursor != "" {
			req.PageToken = &cursor
		}
		resp, err := v.client.List(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, "", fromConnectError(err)
		}
		return resp.Msg.GetVideos(), resp.Msg.GetNextPageToken(), nil
	})
}

// Update mutates a video's metadata (title, description, tags, …).
func (v *Videos) Update(ctx context.Context, params *VideoUpdateParams) (*Video, error) {
	if params == nil {
		params = &VideoUpdateParams{}
	}
	resp, err := v.client.Update(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetVideo(), nil
}

// Delete removes a video and its renditions. Storage cleanup is async.
func (v *Videos) Delete(ctx context.Context, id string) error {
	_, err := v.client.Delete(ctx, connect.NewRequest(&v1.DeleteVideoRequest{Id: id}))
	if err != nil {
		return fromConnectError(err)
	}
	return nil
}

// Watch streams video lifecycle events until the video is ready or fails.
func (v *Videos) Watch(ctx context.Context, id string) *Stream[*WatchVideoResponse] {
	return newStream(ctx, v.cfg.maxRetries, nil,
		func(ctx context.Context) (streamConn[*WatchVideoResponse], error) {
			conn, err := v.streamClient.Watch(ctx, connect.NewRequest(&v1.WatchVideoRequest{Id: id}))
			if err != nil {
				return nil, fromConnectError(err)
			}
			return &videoWatchAdapter{stream: conn}, nil
		})
}

// GetUsage returns the usage summary for a billing month.
func (v *Videos) GetUsage(ctx context.Context, billingMonth string) (*UsageSummary, error) {
	req := &v1.GetUsageRequest{}
	if billingMonth != "" {
		req.BillingMonth = &billingMonth
	}
	resp, err := v.client.GetUsage(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetUsage(), nil
}

// GetStats returns playback analytics for a single video: plays, watch time,
// and unique viewers aggregated per UTC day, plus totals across the requested
// date range. The full response is returned unwrapped because it carries both
// the per-day rows and the range totals. Stats come from a best-effort
// playback beacon rolled up hourly, so recent activity may lag by up to an
// hour.
func (v *Videos) GetStats(ctx context.Context, params *VideoGetStatsParams) (*v1.GetStatsResponse, error) {
	if params == nil {
		params = &VideoGetStatsParams{}
	}
	resp, err := v.client.GetStats(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// ListTopVideos returns an app's top videos ranked by plays over a date range.
func (v *Videos) ListTopVideos(ctx context.Context, params *VideoListTopVideosParams) (*v1.ListTopVideosResponse, error) {
	if params == nil {
		params = &VideoListTopVideosParams{}
	}
	resp, err := v.client.ListTopVideos(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

type videoWatchAdapter struct {
	stream *connect.ServerStreamForClient[v1.WatchVideoResponse]
}

func (a *videoWatchAdapter) Recv() (*WatchVideoResponse, error) {
	if !a.stream.Receive() {
		if err := a.stream.Err(); err != nil {
			return nil, fromConnectError(err)
		}
		return nil, io.EOF
	}
	return a.stream.Msg(), nil
}

func (a *videoWatchAdapter) Close() error {
	if a.stream == nil {
		return nil
	}
	return a.stream.Close()
}
