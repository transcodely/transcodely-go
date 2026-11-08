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

// GetSeoMarkup returns ready-to-paste SEO markup for a hosted video, built for
// YOUR watch page (params.WatchUrl, absolute https): a schema.org VideoObject
// JSON-LD block (with chapter Clip key moments when insights are ready), a
// video-sitemap <url> entry, and a server-renderable HTML fragment with the
// summary, key points, chapter links and the full transcript. Every string is
// escaped for its context; paste it as-is.
func (v *Videos) GetSeoMarkup(ctx context.Context, params *VideoGetSeoMarkupParams) (*VideoSeoMarkup, error) {
	if params == nil {
		params = &VideoGetSeoMarkupParams{}
	}
	resp, err := v.client.GetSeoMarkup(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// GenerateInsights generates AI insights (title suggestions, summary, key
// points, chapters, questions, key terms, highlights, topics) from a hosted
// video's existing AI captions, without transcribing again, and returns the
// entry in status pending. It also appears on Video.Insights, and the
// video.insights.ready or video.insights.failed webhook fires when it
// finishes. Billed EUR 0.02 per video plus EUR 0.005 per minute of source,
// charged only when the insights become ready.
//
// While an entry for the same video and language is pending, the call returns
// that entry and bills nothing new. Set params.Language when the video has
// several AI caption tracks. A video without AI captions is refused with
// failed_precondition (code captions_required).
func (v *Videos) GenerateInsights(ctx context.Context, params *VideoGenerateInsightsParams) (*TranscriptInsights, error) {
	if params == nil {
		params = &VideoGenerateInsightsParams{}
	}
	resp, err := v.client.GenerateInsights(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetInsights(), nil
}

// GenerateHighlights cuts highlight clips from a hosted video: up to
// Config.MaxCount short MP4s (default 5) around the quotes its ready AI
// insights picked, and returns them in status pending. They also appear on
// Video.Highlights, and the video.highlights.ready (or
// video.highlights.failed, when none could be rendered) webhook fires once
// every one has finished.
//
// Each clip is its own clip job on the video, billed as its rendered
// (clipped) minutes at your effective encoding rate, with no processing fee.
// While any highlight of the video is pending, the call returns the current
// set and bills nothing new; once all are ready or failed, a new call
// replaces them. A video without ready insights is refused with
// failed_precondition (code highlights_require_insights).
func (v *Videos) GenerateHighlights(ctx context.Context, params *VideoGenerateHighlightsParams) ([]*VideoHighlight, error) {
	if params == nil {
		params = &VideoGenerateHighlightsParams{}
	}
	resp, err := v.client.GenerateHighlights(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg.GetHighlights(), nil
}

// TranslateCaptions translates a hosted video's existing AI captions into
// other languages without transcribing again. It creates one captions-only
// job on the video that writes, per target language, a caption track, its
// transcript renditions and (on HLS) a caption rendition in the manifest;
// with IncludeInsights (default true) the ready insights in the source
// language are translated too and appear as one more Video.Insights entry
// per language. The result names the job, the normalized source and target
// languages, the estimated fee, and any target already carried by an
// earlier, still-running job (InFlight; those are not requested again).
//
// Billed as the captions_translation fee: EUR 0.12 per source minute per
// language, only for languages actually delivered. Set
// params.SourceLanguage when the video has several AI caption tracks. A
// video without AI captions is refused with failed_precondition (code
// captions_required); deployments without the feature refuse it with
// parameter_unsupported.
func (v *Videos) TranslateCaptions(ctx context.Context, params *VideoTranslateCaptionsParams) (*VideoTranslateCaptionsResult, error) {
	if params == nil {
		params = &VideoTranslateCaptionsParams{}
	}
	resp, err := v.client.TranslateCaptions(ctx, connect.NewRequest(params))
	if err != nil {
		return nil, fromConnectError(err)
	}
	return resp.Msg, nil
}

// RestyleCaptions re-cuts a hosted video's existing AI captions into cues of
// a new CueStyle without transcribing again: the words and their timings
// come from the track's stored transcript, only their grouping into cues
// changes. It creates a captions-only job on the video (JobId in the
// result); when it completes the track's caption file, transcript
// renditions and HLS caption rendition are replaced in place and
// VideoTextTrack.CueStyle shows the new style. Insights are kept, and
// translations of the track keep their old cues until translated again.
//
// Free: no captions, translation or processing fee. Set params.Language when
// the video has several AI caption tracks; leave params.CueStyle unset to
// re-cut into the app's current default. Refused with failed_precondition
// (code captions_required, insights_transcript_unavailable or
// captions_job_in_flight) when there is nothing to restyle or a captions job
// of that language is still running.
func (v *Videos) RestyleCaptions(ctx context.Context, params *VideoRestyleCaptionsParams) (*VideoRestyleCaptionsResult, error) {
	if params == nil {
		params = &VideoRestyleCaptionsParams{}
	}
	resp, err := v.client.RestyleCaptions(ctx, connect.NewRequest(params))
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
