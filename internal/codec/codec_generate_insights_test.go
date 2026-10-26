package codec

import (
	"testing"

	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalGenerateInsightsRequest pins the GenerateInsights request as the
// snake_case shape the server expects, with the optional fields off the wire
// when unset.
func TestMarshalGenerateInsightsRequest(t *testing.T) {
	c := NewProtoJSONCodec()

	m := marshalToMap(t, c, &v1.GenerateInsightsRequest{
		Id:       "vid_abc",
		Language: proto.String("eng"),
		Hint:     proto.String("for beginners"),
		AppId:    proto.String("app_abc"),
	})
	want := map[string]string{
		"id":       `"vid_abc"`,
		"language": `"eng"`,
		"hint":     `"for beginners"`,
		"app_id":   `"app_abc"`,
	}
	for k, v := range want {
		if got, ok := m[k]; !ok || string(got) != v {
			t.Errorf("%s = %s (present=%v), want %s; payload: %v", k, got, ok, v, m)
		}
	}
	if _, ok := m["appId"]; ok {
		t.Errorf("payload unexpectedly contains camelCase key appId: %v", m)
	}

	m = marshalToMap(t, c, &v1.GenerateInsightsRequest{Id: "vid_abc"})
	for _, k := range []string{"language", "hint", "app_id"} {
		if _, ok := m[k]; ok {
			t.Errorf("%s present when unset: %v", k, m)
		}
	}
}

// TestUnmarshalVideoSourcedInsights verifies a video-sourced insights entry
// (the GenerateInsights response and the video.insights.* webhook payload)
// decodes its new fields: the ins_ id, the lowercase source enum, video_id and
// fee_eur.
func TestUnmarshalVideoSourcedInsights(t *testing.T) {
	c := NewProtoJSONCodec()
	var resp v1.GenerateInsightsResponse
	body := `{"insights":{"id":"ins_abc123","source":"video","video_id":"vid_abc","language":"eng","status":"pending","fee_eur":0.07}}`
	if err := c.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ins := resp.GetInsights()
	if ins.GetId() != "ins_abc123" {
		t.Errorf("id = %q", ins.GetId())
	}
	if ins.GetSource() != v1.InsightsSource_INSIGHTS_SOURCE_VIDEO {
		t.Errorf("source = %v, want video", ins.GetSource())
	}
	if ins.GetVideoId() != "vid_abc" || ins.GetLanguage() != "eng" {
		t.Errorf("video_id/language = %q/%q", ins.GetVideoId(), ins.GetLanguage())
	}
	if ins.GetStatus() != v1.InsightsStatus_INSIGHTS_STATUS_PENDING {
		t.Errorf("status = %v, want pending", ins.GetStatus())
	}
	if ins.FeeEur == nil || ins.GetFeeEur() != 0.07 {
		t.Errorf("fee_eur = %v, want 0.07", ins.FeeEur)
	}

	// The source enum marshals back as the lowercase wire value, and a
	// job-sourced entry omits fee_eur.
	m := marshalToMap(t, c, &v1.TranscriptInsights{Source: v1.InsightsSource_INSIGHTS_SOURCE_JOB})
	if got := string(m["source"]); got != `"job"` {
		t.Errorf("source = %s, want \"job\"; payload: %v", got, m)
	}
	if _, ok := m["fee_eur"]; ok {
		t.Errorf("fee_eur present when unset: %v", m)
	}
}
