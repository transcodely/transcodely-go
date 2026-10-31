package codec

import (
	"testing"

	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalGenerateHighlightsRequest pins the GenerateHighlights request as
// the snake_case shape the server expects, with unset config fields off the
// wire.
func TestMarshalGenerateHighlightsRequest(t *testing.T) {
	c := NewProtoJSONCodec()

	m := marshalToMap(t, c, &v1.GenerateHighlightsRequest{
		Id:       "vid_abc",
		Language: proto.String("eng"),
		Config: &v1.HighlightsConfig{
			MaxCount:       proto.Int32(3),
			PaddingSeconds: proto.Float64(2),
		},
	})
	if got := string(m["id"]); got != `"vid_abc"` {
		t.Errorf("id = %s; payload: %v", got, m)
	}
	cfg := marshalToMap(t, c, &v1.HighlightsConfig{MaxCount: proto.Int32(3), PaddingSeconds: proto.Float64(2)})
	if got := string(cfg["max_count"]); got != "3" {
		t.Errorf("max_count = %s; payload: %v", got, cfg)
	}
	if got := string(cfg["padding_seconds"]); got != "2" {
		t.Errorf("padding_seconds = %s; payload: %v", got, cfg)
	}
	for _, k := range []string{"min_seconds", "max_seconds", "maxCount"} {
		if _, ok := cfg[k]; ok {
			t.Errorf("%s present when unset: %v", k, cfg)
		}
	}
	if _, ok := m["config"]; !ok {
		t.Errorf("config missing: %v", m)
	}
}

// TestUnmarshalVideoHighlights verifies Video.highlights (the
// video.highlights.* webhook payload and Videos.Get) decodes, with the status
// enum on its lowercase wire value.
func TestUnmarshalVideoHighlights(t *testing.T) {
	c := NewProtoJSONCodec()
	var v v1.Video
	body := `{"id":"vid_abc","highlights":[` +
		`{"id":"hl_a1b2c3d4e5f6g7","start_seconds":18.5,"end_seconds":31.5,"title":"Patience.","status":"ready",` +
		`"playback_url":"https://cdn.example/x.mp4","poster_url":"https://cdn.example/x.jpg","duration_seconds":13,"job_id":"job_abc","language":"eng"},` +
		`{"id":"hl_b1b2c3d4e5f6g7","status":"failed","error_code":"queue_limit_exceeded","language":"eng"}]}`
	if err := c.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	hl := v.GetHighlights()
	if len(hl) != 2 {
		t.Fatalf("highlights = %d, want 2", len(hl))
	}
	if hl[0].GetStatus() != v1.VideoHighlightStatus_VIDEO_HIGHLIGHT_STATUS_READY ||
		hl[0].GetStartSeconds() != 18.5 || hl[0].GetEndSeconds() != 31.5 ||
		hl[0].GetPlaybackUrl() == "" || hl[0].GetDurationSeconds() != 13 || hl[0].GetJobId() != "job_abc" {
		t.Errorf("highlights[0] = %+v", hl[0])
	}
	if hl[1].GetStatus() != v1.VideoHighlightStatus_VIDEO_HIGHLIGHT_STATUS_FAILED ||
		hl[1].GetErrorCode() != "queue_limit_exceeded" || hl[1].PlaybackUrl != nil {
		t.Errorf("highlights[1] = %+v", hl[1])
	}

	m := marshalToMap(t, c, &v1.VideoHighlight{Status: v1.VideoHighlightStatus_VIDEO_HIGHLIGHT_STATUS_PENDING})
	if got := string(m["status"]); got != `"pending"` {
		t.Errorf("status = %s, want \"pending\"; payload: %v", got, m)
	}
}
