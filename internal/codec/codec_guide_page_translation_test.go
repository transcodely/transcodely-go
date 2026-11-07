package codec

import (
	"testing"

	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestGuidePageWire pins the guide page fields on the wire: guide_page is
// the lowercase GuidePageSetting value both ways, guide_page_enabled is an
// optional bool that stays off the wire when unset (a video.* webhook
// payload never carries it), and UpdateVideoRequest.guide_page serializes as
// "on"/"off"/"default".
func TestGuidePageWire(t *testing.T) {
	c := NewProtoJSONCodec()

	// Webhook-shaped payload: guide_page present, guide_page_enabled absent.
	var v v1.Video
	body := `{"id":"vid_abc","guide_page":"on"}`
	if err := c.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v.GetGuidePage() != v1.GuidePageSetting_GUIDE_PAGE_SETTING_ON {
		t.Errorf("guide_page = %v, want ON", v.GetGuidePage())
	}
	if v.GuidePageEnabled != nil || v.GuideUrl != nil {
		t.Errorf("guide_page_enabled/guide_url should be nil when absent: %+v", &v)
	}

	// API-shaped payload: the effective value and the URL are present.
	var full v1.Video
	body = `{"id":"vid_abc","guide_page":"default","guide_page_enabled":true,"guide_url":"https://play.example/v/vid_abc/guide"}`
	if err := c.Unmarshal([]byte(body), &full); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if full.GetGuidePage() != v1.GuidePageSetting_GUIDE_PAGE_SETTING_DEFAULT ||
		full.GuidePageEnabled == nil || !full.GetGuidePageEnabled() ||
		full.GetGuideUrl() != "https://play.example/v/vid_abc/guide" {
		t.Errorf("video = %+v", &full)
	}

	// Request side.
	m := marshalToMap(t, c, &v1.UpdateVideoRequest{
		Id:        "vid_abc",
		GuidePage: v1.GuidePageSetting_GUIDE_PAGE_SETTING_OFF.Enum(),
	})
	if got := string(m["guide_page"]); got != `"off"` {
		t.Errorf("guide_page = %s, want \"off\"; payload: %v", got, m)
	}
	if _, ok := m["guide_page_enabled"]; ok {
		t.Errorf("guide_page_enabled must not be a request field: %v", m)
	}
	m = marshalToMap(t, c, &v1.UpdateVideoRequest{Id: "vid_abc"})
	if _, ok := m["guide_page"]; ok {
		t.Errorf("guide_page present when unset: %v", m)
	}

	hc := marshalToMap(t, c, &v1.HostingConfig{GuidePagesEnabled: proto.Bool(true)})
	if got := string(hc["guide_pages_enabled"]); got != "true" {
		t.Errorf("guide_pages_enabled = %s; payload: %v", got, hc)
	}
}

// TestTranslateCaptionsWire pins the TranslateCaptions request/response and
// the translation fields on tracks as snake_case.
func TestTranslateCaptionsWire(t *testing.T) {
	c := NewProtoJSONCodec()

	m := marshalToMap(t, c, &v1.TranslateCaptionsRequest{
		Id:              "vid_abc",
		SourceLanguage:  proto.String("eng"),
		TargetLanguages: []string{"ger", "spa"},
		IncludeInsights: proto.Bool(false),
	})
	if got := string(m["source_language"]); got != `"eng"` {
		t.Errorf("source_language = %s; payload: %v", got, m)
	}
	if got := string(m["target_languages"]); got != `["ger","spa"]` {
		t.Errorf("target_languages = %s; payload: %v", got, m)
	}
	if got := string(m["include_insights"]); got != "false" {
		t.Errorf("include_insights = %s; payload: %v", got, m)
	}
	if _, ok := m["app_id"]; ok {
		t.Errorf("app_id present when unset: %v", m)
	}

	var resp v1.TranslateCaptionsResponse
	body := `{"job_id":"job_a1b2c3d4e5f6","video_id":"vid_abc","source_language":"eng","target_languages":["ger"],` +
		`"in_flight":[{"language":"spa","job_id":"job_f6e5d4c3b2a1"}],"estimated_fee_eur":0.36,"insights_included":true}`
	if err := c.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.GetJobId() != "job_a1b2c3d4e5f6" || len(resp.GetInFlight()) != 1 ||
		resp.GetInFlight()[0].GetLanguage() != "spa" || resp.GetEstimatedFeeEur() != 0.36 || !resp.GetInsightsIncluded() {
		t.Errorf("response = %+v", &resp)
	}

	// Track-side fields.
	var track v1.VideoTextTrack
	if err := c.Unmarshal([]byte(`{"id":"trk_abc","language":"ger","translated_from":"eng"}`), &track); err != nil {
		t.Fatalf("unmarshal track: %v", err)
	}
	if track.GetTranslatedFrom() != "eng" {
		t.Errorf("translated_from = %q", track.GetTranslatedFrom())
	}
	st := marshalToMap(t, c, &v1.SubtitleTrack{Translations: []string{"ger"}})
	if got := string(st["translations"]); got != `["ger"]` {
		t.Errorf("translations = %s; payload: %v", got, st)
	}
	up := marshalToMap(t, c, &v1.CreateUploadRequest{AutoTranslations: []string{"ger"}})
	if got := string(up["auto_translations"]); got != `["ger"]` {
		t.Errorf("auto_translations = %s; payload: %v", got, up)
	}
}
