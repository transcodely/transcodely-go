package transcodely

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// fakeVideoClient is a minimal VideoServiceClient stub. It embeds the generated
// interface so only the methods a test exercises need implementations; any
// other call would nil-panic, surfacing accidental use. It records the request
// message it received and returns a canned response.
type fakeVideoClient struct {
	transcodelyv1connect.VideoServiceClient

	gotStats *v1.GetStatsRequest
	stats    *v1.GetStatsResponse

	gotTop *v1.ListTopVideosRequest
	top    *v1.ListTopVideosResponse
}

func (f *fakeVideoClient) GetStats(_ context.Context, req *connect.Request[v1.GetStatsRequest]) (*connect.Response[v1.GetStatsResponse], error) {
	f.gotStats = req.Msg
	return connect.NewResponse(f.stats), nil
}

func (f *fakeVideoClient) ListTopVideos(_ context.Context, req *connect.Request[v1.ListTopVideosRequest]) (*connect.Response[v1.ListTopVideosResponse], error) {
	f.gotTop = req.Msg
	return connect.NewResponse(f.top), nil
}

func strptr(s string) *string { return &s }

func TestVideos_GetStats_RequestAndPassthrough(t *testing.T) {
	fake := &fakeVideoClient{
		stats: &v1.GetStatsResponse{
			Daily: []*v1.VideoStatsDay{
				{Date: "2026-07-01", Plays: 3, WatchSeconds: 120, UniqueViewers: 2},
				{Date: "2026-07-02", Plays: 5, WatchSeconds: 300, UniqueViewers: 4},
			},
			Totals: &v1.VideoStatsTotals{Plays: 8, WatchSeconds: 420, UniqueViewers: 6},
		},
	}
	v := newVideos(fake, fake, &config{})

	resp, err := v.GetStats(context.Background(), &VideoGetStatsParams{
		VideoId:   "vid_abc",
		StartDate: strptr("2026-07-01"),
		EndDate:   strptr("2026-07-07"),
	})
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	// Request was forwarded verbatim.
	if got := fake.gotStats.GetVideoId(); got != "vid_abc" {
		t.Errorf("video_id = %q, want vid_abc", got)
	}
	if fake.gotStats.GetStartDate() != "2026-07-01" || fake.gotStats.GetEndDate() != "2026-07-07" {
		t.Errorf("date range = %q..%q, want 2026-07-01..2026-07-07",
			fake.gotStats.GetStartDate(), fake.gotStats.GetEndDate())
	}

	// The full response is returned unwrapped: both daily rows and totals.
	if got := len(resp.GetDaily()); got != 2 {
		t.Fatalf("daily len = %d, want 2", got)
	}
	if got := resp.GetDaily()[0].GetUniqueViewers(); got != 2 {
		t.Errorf("daily[0].unique_viewers = %d, want 2", got)
	}
	if got := resp.GetTotals().GetPlays(); got != 8 {
		t.Errorf("totals.plays = %d, want 8", got)
	}
	if got := resp.GetTotals().GetWatchSeconds(); got != 420 {
		t.Errorf("totals.watch_seconds = %d, want 420", got)
	}
}

func TestVideos_ListTopVideos_RequestAndPassthrough(t *testing.T) {
	limit := int32(5)
	fake := &fakeVideoClient{
		top: &v1.ListTopVideosResponse{
			Items: []*v1.TopVideo{
				{VideoId: "vid_1", Title: strptr("Intro"), Plays: 42, WatchSeconds: 1000, UniqueViewers: 20},
				{VideoId: "vid_2", Plays: 10, WatchSeconds: 200, UniqueViewers: 5},
			},
		},
	}
	v := newVideos(fake, fake, &config{})

	resp, err := v.ListTopVideos(context.Background(), &VideoListTopVideosParams{
		AppId:     strptr("app_123"),
		StartDate: strptr("2026-07-01"),
		Limit:     &limit,
	})
	if err != nil {
		t.Fatalf("ListTopVideos() error = %v", err)
	}

	// Request was forwarded verbatim.
	if got := fake.gotTop.GetAppId(); got != "app_123" {
		t.Errorf("app_id = %q, want app_123", got)
	}
	if got := fake.gotTop.GetLimit(); got != 5 {
		t.Errorf("limit = %d, want 5", got)
	}
	if got := fake.gotTop.GetStartDate(); got != "2026-07-01" {
		t.Errorf("start_date = %q, want 2026-07-01", got)
	}

	// The full response is returned unwrapped.
	if got := len(resp.GetItems()); got != 2 {
		t.Fatalf("items len = %d, want 2", got)
	}
	top := resp.GetItems()[0]
	if top.GetVideoId() != "vid_1" || top.GetTitle() != "Intro" || top.GetPlays() != 42 {
		t.Errorf("items[0] = %q/%q/%d plays, want vid_1/Intro/42",
			top.GetVideoId(), top.GetTitle(), top.GetPlays())
	}
}

// Both analytics methods guard nil params and still forward a request.
func TestVideos_Analytics_NilParams(t *testing.T) {
	fake := &fakeVideoClient{
		stats: &v1.GetStatsResponse{},
		top:   &v1.ListTopVideosResponse{},
	}
	v := newVideos(fake, fake, &config{})

	if _, err := v.GetStats(context.Background(), nil); err != nil {
		t.Fatalf("GetStats(nil) error = %v", err)
	}
	if fake.gotStats == nil {
		t.Error("GetStats(nil) did not forward a request")
	}

	if _, err := v.ListTopVideos(context.Background(), nil); err != nil {
		t.Fatalf("ListTopVideos(nil) error = %v", err)
	}
	if fake.gotTop == nil {
		t.Error("ListTopVideos(nil) did not forward a request")
	}
}
