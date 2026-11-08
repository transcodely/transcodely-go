package transcodely

import (
	"context"
	"errors"
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

	gotSeo *v1.GetSeoMarkupRequest
	seo    *v1.GetSeoMarkupResponse

	gotInsights *v1.GenerateInsightsRequest
	insights    *v1.GenerateInsightsResponse

	gotHighlights *v1.GenerateHighlightsRequest
	highlights    *v1.GenerateHighlightsResponse

	gotTranslate *v1.TranslateCaptionsRequest
	translate    *v1.TranslateCaptionsResponse

	gotRestyle *v1.RestyleCaptionsRequest
	restyle    *v1.RestyleCaptionsResponse

	gotSetDefaultStyle *v1.SetDefaultCaptionStyleRequest
	setDefaultStyle    *v1.SetDefaultCaptionStyleResponse

	gotDeleteStyle *v1.DeleteCaptionStyleRequest
	deleteStyle    *v1.DeleteCaptionStyleResponse
	deleteStyleErr error
}

func (f *fakeVideoClient) SetDefaultCaptionStyle(_ context.Context, req *connect.Request[v1.SetDefaultCaptionStyleRequest]) (*connect.Response[v1.SetDefaultCaptionStyleResponse], error) {
	f.gotSetDefaultStyle = req.Msg
	return connect.NewResponse(f.setDefaultStyle), nil
}

func (f *fakeVideoClient) DeleteCaptionStyle(_ context.Context, req *connect.Request[v1.DeleteCaptionStyleRequest]) (*connect.Response[v1.DeleteCaptionStyleResponse], error) {
	f.gotDeleteStyle = req.Msg
	if f.deleteStyleErr != nil {
		return nil, f.deleteStyleErr
	}
	return connect.NewResponse(f.deleteStyle), nil
}

func (f *fakeVideoClient) RestyleCaptions(_ context.Context, req *connect.Request[v1.RestyleCaptionsRequest]) (*connect.Response[v1.RestyleCaptionsResponse], error) {
	f.gotRestyle = req.Msg
	return connect.NewResponse(f.restyle), nil
}

func (f *fakeVideoClient) TranslateCaptions(_ context.Context, req *connect.Request[v1.TranslateCaptionsRequest]) (*connect.Response[v1.TranslateCaptionsResponse], error) {
	f.gotTranslate = req.Msg
	return connect.NewResponse(f.translate), nil
}

func (f *fakeVideoClient) GenerateHighlights(_ context.Context, req *connect.Request[v1.GenerateHighlightsRequest]) (*connect.Response[v1.GenerateHighlightsResponse], error) {
	f.gotHighlights = req.Msg
	return connect.NewResponse(f.highlights), nil
}

func (f *fakeVideoClient) GenerateInsights(_ context.Context, req *connect.Request[v1.GenerateInsightsRequest]) (*connect.Response[v1.GenerateInsightsResponse], error) {
	f.gotInsights = req.Msg
	return connect.NewResponse(f.insights), nil
}

func (f *fakeVideoClient) GetSeoMarkup(_ context.Context, req *connect.Request[v1.GetSeoMarkupRequest]) (*connect.Response[v1.GetSeoMarkupResponse], error) {
	f.gotSeo = req.Msg
	return connect.NewResponse(f.seo), nil
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

func TestVideos_GetSeoMarkup_RequestAndPassthrough(t *testing.T) {
	fake := &fakeVideoClient{
		seo: &v1.GetSeoMarkupResponse{
			JsonLd:       `{"@type":"VideoObject"}`,
			SitemapEntry: "<url><loc>https://example.com/watch/42</loc></url>",
			HtmlFragment: `<section class="tx-insights"></section>`,
		},
	}
	v := newVideos(fake, fake, &config{})

	resp, err := v.GetSeoMarkup(context.Background(), &VideoGetSeoMarkupParams{
		VideoId:       "vid_abc",
		WatchUrl:      "https://example.com/watch/42",
		DeepLinkParam: strptr("start"),
		Language:      strptr("eng"),
	})
	if err != nil {
		t.Fatalf("GetSeoMarkup() error = %v", err)
	}
	if fake.gotSeo.GetVideoId() != "vid_abc" || fake.gotSeo.GetWatchUrl() != "https://example.com/watch/42" {
		t.Errorf("request = %q/%q, want vid_abc/https://example.com/watch/42",
			fake.gotSeo.GetVideoId(), fake.gotSeo.GetWatchUrl())
	}
	if fake.gotSeo.GetDeepLinkParam() != "start" || fake.gotSeo.GetLanguage() != "eng" {
		t.Errorf("optional fields = %q/%q, want start/eng",
			fake.gotSeo.GetDeepLinkParam(), fake.gotSeo.GetLanguage())
	}
	if resp.GetJsonLd() != `{"@type":"VideoObject"}` ||
		resp.GetSitemapEntry() == "" || resp.GetHtmlFragment() == "" {
		t.Errorf("response not passed through: %+v", resp)
	}

	if _, err := v.GetSeoMarkup(context.Background(), nil); err != nil {
		t.Fatalf("GetSeoMarkup(nil) error = %v", err)
	}
	if fake.gotSeo == nil {
		t.Error("GetSeoMarkup(nil) did not forward a request")
	}
}

func TestVideos_GenerateInsights_RequestAndUnwrap(t *testing.T) {
	fake := &fakeVideoClient{
		insights: &v1.GenerateInsightsResponse{
			Insights: &v1.TranscriptInsights{
				Id:       "ins_abc123",
				Source:   v1.InsightsSource_INSIGHTS_SOURCE_VIDEO,
				VideoId:  "vid_abc",
				Language: "eng",
				Status:   v1.InsightsStatus_INSIGHTS_STATUS_PENDING,
			},
		},
	}
	v := newVideos(fake, fake, &config{})

	ins, err := v.GenerateInsights(context.Background(), &VideoGenerateInsightsParams{
		Id:       "vid_abc",
		Language: strptr("eng"),
		Hint:     strptr("A cooking show for beginners"),
	})
	if err != nil {
		t.Fatalf("GenerateInsights() error = %v", err)
	}
	if fake.gotInsights.GetId() != "vid_abc" || fake.gotInsights.GetLanguage() != "eng" ||
		fake.gotInsights.GetHint() != "A cooking show for beginners" || fake.gotInsights.AppId != nil {
		t.Errorf("forwarded request = %+v", fake.gotInsights)
	}
	if ins.GetId() != "ins_abc123" || ins.GetSource() != InsightsSourceVideo ||
		ins.GetVideoId() != "vid_abc" || ins.GetStatus() != InsightsStatusPending {
		t.Errorf("GenerateInsights() = %+v, want the response's insights entry", ins)
	}

	fake.gotInsights = nil
	if _, err := v.GenerateInsights(context.Background(), nil); err != nil {
		t.Fatalf("GenerateInsights(nil) error = %v", err)
	}
	if fake.gotInsights == nil {
		t.Error("GenerateInsights(nil) did not forward a request")
	}
}

func TestVideos_GenerateHighlights_RequestAndUnwrap(t *testing.T) {
	fake := &fakeVideoClient{
		highlights: &v1.GenerateHighlightsResponse{
			Highlights: []*v1.VideoHighlight{
				{Id: "hl_a1b2c3d4e5f6g7", StartSeconds: 18.5, EndSeconds: 31.5, Status: v1.VideoHighlightStatus_VIDEO_HIGHLIGHT_STATUS_PENDING, Language: "eng"},
				{Id: "hl_b1b2c3d4e5f6g7", StartSeconds: 40, EndSeconds: 52, Status: v1.VideoHighlightStatus_VIDEO_HIGHLIGHT_STATUS_PENDING, Language: "eng"},
			},
		},
	}
	v := newVideos(fake, fake, &config{})

	maxCount := int32(3)
	hls, err := v.GenerateHighlights(context.Background(), &VideoGenerateHighlightsParams{
		Id:       "vid_abc",
		Language: strptr("eng"),
		Config:   &HighlightsConfig{MaxCount: &maxCount},
	})
	if err != nil {
		t.Fatalf("GenerateHighlights() error = %v", err)
	}
	if fake.gotHighlights.GetId() != "vid_abc" || fake.gotHighlights.GetLanguage() != "eng" ||
		fake.gotHighlights.GetConfig().GetMaxCount() != 3 || fake.gotHighlights.AppId != nil {
		t.Errorf("forwarded request = %+v", fake.gotHighlights)
	}
	if len(hls) != 2 || hls[0].GetId() != "hl_a1b2c3d4e5f6g7" || hls[0].GetStatus() != VideoHighlightStatusPending {
		t.Errorf("GenerateHighlights() = %+v, want the response's highlights", hls)
	}

	fake.gotHighlights = nil
	if _, err := v.GenerateHighlights(context.Background(), nil); err != nil {
		t.Fatalf("GenerateHighlights(nil) error = %v", err)
	}
	if fake.gotHighlights == nil {
		t.Error("GenerateHighlights(nil) did not forward a request")
	}
}

func TestVideos_TranslateCaptions_RequestAndResult(t *testing.T) {
	fake := &fakeVideoClient{
		translate: &v1.TranslateCaptionsResponse{
			JobId:            "job_a1b2c3d4e5f6",
			VideoId:          "vid_abc",
			SourceLanguage:   "eng",
			TargetLanguages:  []string{"ger"},
			InFlight:         []*v1.TranslationInFlight{{Language: "spa", JobId: "job_f6e5d4c3b2a1"}},
			EstimatedFeeEur:  0.36,
			InsightsIncluded: true,
		},
	}
	v := newVideos(fake, fake, &config{})

	include := false
	res, err := v.TranslateCaptions(context.Background(), &VideoTranslateCaptionsParams{
		Id:              "vid_abc",
		SourceLanguage:  strptr("eng"),
		TargetLanguages: []string{"deu", "spa"},
		IncludeInsights: &include,
	})
	if err != nil {
		t.Fatalf("TranslateCaptions() error = %v", err)
	}
	got := fake.gotTranslate
	if got.GetId() != "vid_abc" || got.GetSourceLanguage() != "eng" ||
		len(got.GetTargetLanguages()) != 2 || got.GetTargetLanguages()[0] != "deu" ||
		got.IncludeInsights == nil || got.GetIncludeInsights() || got.AppId != nil {
		t.Errorf("forwarded request = %+v", got)
	}
	if res.GetJobId() != "job_a1b2c3d4e5f6" || res.GetSourceLanguage() != "eng" ||
		len(res.GetTargetLanguages()) != 1 || res.GetTargetLanguages()[0] != "ger" ||
		len(res.GetInFlight()) != 1 || res.GetInFlight()[0].GetLanguage() != "spa" ||
		res.GetInFlight()[0].GetJobId() != "job_f6e5d4c3b2a1" ||
		res.GetEstimatedFeeEur() != 0.36 || !res.GetInsightsIncluded() {
		t.Errorf("TranslateCaptions() = %+v, want the response verbatim", res)
	}

	fake.gotTranslate = nil
	if _, err := v.TranslateCaptions(context.Background(), nil); err != nil {
		t.Fatalf("TranslateCaptions(nil) error = %v", err)
	}
	if fake.gotTranslate == nil {
		t.Error("TranslateCaptions(nil) did not forward a request")
	}
}

func TestVideos_RestyleCaptions_RequestAndResult(t *testing.T) {
	compact := CueStylePresetCompact
	fake := &fakeVideoClient{
		restyle: &v1.RestyleCaptionsResponse{
			JobId:    "job_a1b2c3d4e5f6",
			VideoId:  "vid_abc",
			Language: "eng",
			CueStyle: &CueStyle{
				Preset:           &compact,
				MaxCharsPerLine:  ptrU32(42),
				MaxLines:         ptrU32(2),
				MaxCueDurationMs: ptrU32(4000),
			},
			StyleVariant: CueStylePresetCompact,
			AddsVariant:  true,
		},
	}
	v := newVideos(fake, fake, &config{})

	res, err := v.RestyleCaptions(context.Background(), &VideoRestyleCaptionsParams{
		Id:          "vid_abc",
		Language:    strptr("eng"),
		CueStyle:    &CueStyle{Preset: &compact, MaxLines: ptrU32(2)},
		MakeDefault: boolptr(true),
	})
	if err != nil {
		t.Fatalf("RestyleCaptions() error = %v", err)
	}
	got := fake.gotRestyle
	if got.GetId() != "vid_abc" || got.GetLanguage() != "eng" ||
		got.GetCueStyle().GetPreset() != CueStylePresetCompact ||
		got.GetCueStyle().GetMaxLines() != 2 || got.GetCueStyle().MaxCharsPerLine != nil ||
		got.AppId != nil || got.MakeDefault == nil || !got.GetMakeDefault() {
		t.Errorf("forwarded request = %+v", got)
	}
	if res.GetStyleVariant() != CueStylePresetCompact || !res.GetAddsVariant() {
		t.Errorf("RestyleCaptions() variant = %v adds = %v, want compact true", res.GetStyleVariant(), res.GetAddsVariant())
	}
	if res.GetJobId() != "job_a1b2c3d4e5f6" || res.GetVideoId() != "vid_abc" ||
		res.GetLanguage() != "eng" || res.GetCueStyle().GetPreset() != CueStylePresetCompact ||
		res.GetCueStyle().GetMaxCharsPerLine() != 42 || res.GetCueStyle().GetMaxLines() != 2 ||
		res.GetCueStyle().GetMaxCueDurationMs() != 4000 {
		t.Errorf("RestyleCaptions() = %+v, want the response verbatim", res)
	}

	fake.gotRestyle = nil
	if _, err := v.RestyleCaptions(context.Background(), nil); err != nil {
		t.Fatalf("RestyleCaptions(nil) error = %v", err)
	}
	if fake.gotRestyle == nil {
		t.Error("RestyleCaptions(nil) did not forward a request")
	}
}

func ptrU32(v uint32) *uint32 { return &v }

func boolptr(b bool) *bool { return &b }

// captionVariantVideo is a video carrying English AI captions in two styles,
// long as the default and standard as the other, plus a stale translation.
func captionVariantVideo() *Video {
	long, standard := CueStylePresetLong, CueStylePresetStandard
	return &Video{
		Id: "vid_abc",
		TextTracks: []*VideoTextTrack{
			{Id: "trk_long", Language: "eng", Label: "English", AutoGenerated: true, IsDefault: true, StyleVariant: &long},
			{Id: "trk_std", Language: "eng", Label: "English (standard)", AutoGenerated: true, StyleVariant: &standard},
			{Id: "trk_bul", Language: "bul", Label: "Bulgarian", AutoGenerated: true, IsDefault: true, StyleVariant: &long, TranslatedFrom: "eng", Stale: true},
		},
	}
}

func TestVideos_SetDefaultCaptionStyle_RequestAndResult(t *testing.T) {
	fake := &fakeVideoClient{
		setDefaultStyle: &v1.SetDefaultCaptionStyleResponse{Video: captionVariantVideo()},
	}
	v := newVideos(fake, fake, &config{})

	res, err := v.SetDefaultCaptionStyle(context.Background(), &VideoSetDefaultCaptionStyleParams{
		Id:           "vid_abc",
		Language:     "eng",
		StyleVariant: CueStylePresetLong,
		AppId:        strptr("app_k1l2m3n4o5"),
	})
	if err != nil {
		t.Fatalf("SetDefaultCaptionStyle() error = %v", err)
	}
	got := fake.gotSetDefaultStyle
	if got.GetId() != "vid_abc" || got.GetLanguage() != "eng" ||
		got.GetStyleVariant() != CueStylePresetLong || got.GetAppId() != "app_k1l2m3n4o5" {
		t.Errorf("forwarded request = %+v", got)
	}
	tracks := res.GetTextTracks()
	if len(tracks) != 3 {
		t.Fatalf("SetDefaultCaptionStyle() tracks = %d, want 3", len(tracks))
	}
	if !tracks[0].GetIsDefault() || tracks[0].GetStyleVariant() != CueStylePresetLong || tracks[0].GetLabel() != "English" {
		t.Errorf("default track = %+v", tracks[0])
	}
	if tracks[1].GetIsDefault() || tracks[1].GetStyleVariant() != CueStylePresetStandard || tracks[1].GetLabel() != "English (standard)" {
		t.Errorf("other style track = %+v", tracks[1])
	}
	if !tracks[2].GetStale() || tracks[2].GetTranslatedFrom() != "eng" {
		t.Errorf("translation track = %+v, want stale", tracks[2])
	}

	fake.gotSetDefaultStyle = nil
	if _, err := v.SetDefaultCaptionStyle(context.Background(), nil); err != nil {
		t.Fatalf("SetDefaultCaptionStyle(nil) error = %v", err)
	}
	if fake.gotSetDefaultStyle == nil {
		t.Error("SetDefaultCaptionStyle(nil) did not forward a request")
	}
}

func TestVideos_DeleteCaptionStyle_RequestAndResult(t *testing.T) {
	video := captionVariantVideo()
	video.TextTracks = video.TextTracks[:1]
	fake := &fakeVideoClient{deleteStyle: &v1.DeleteCaptionStyleResponse{Video: video}}
	v := newVideos(fake, fake, &config{})

	res, err := v.DeleteCaptionStyle(context.Background(), &VideoDeleteCaptionStyleParams{
		Id:           "vid_abc",
		Language:     "eng",
		StyleVariant: CueStylePresetStandard,
	})
	if err != nil {
		t.Fatalf("DeleteCaptionStyle() error = %v", err)
	}
	got := fake.gotDeleteStyle
	if got.GetId() != "vid_abc" || got.GetLanguage() != "eng" ||
		got.GetStyleVariant() != CueStylePresetStandard || got.AppId != nil {
		t.Errorf("forwarded request = %+v", got)
	}
	if tracks := res.GetTextTracks(); len(tracks) != 1 || tracks[0].GetId() != "trk_long" {
		t.Errorf("DeleteCaptionStyle() tracks = %+v, want only the default", tracks)
	}

	fake.gotDeleteStyle = nil
	if _, err := v.DeleteCaptionStyle(context.Background(), nil); err != nil {
		t.Fatalf("DeleteCaptionStyle(nil) error = %v", err)
	}
	if fake.gotDeleteStyle == nil {
		t.Error("DeleteCaptionStyle(nil) did not forward a request")
	}
}

func TestVideos_DeleteCaptionStyle_DefaultRefusedIsTypedError(t *testing.T) {
	fake := &fakeVideoClient{
		deleteStyleErr: connect.NewError(connect.CodeInvalidArgument, errors.New("the default caption style cannot be removed")),
	}
	v := newVideos(fake, fake, &config{})

	res, err := v.DeleteCaptionStyle(context.Background(), &VideoDeleteCaptionStyleParams{
		Id: "vid_abc", Language: "eng", StyleVariant: CueStylePresetLong,
	})
	if res != nil {
		t.Errorf("DeleteCaptionStyle() result = %+v, want nil on error", res)
	}
	var invalid *InvalidRequestError
	if !errors.As(err, &invalid) {
		t.Fatalf("DeleteCaptionStyle() error = %T %v, want *InvalidRequestError", err, err)
	}
}
