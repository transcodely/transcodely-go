package codec

import (
	"encoding/json"
	"testing"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"google.golang.org/protobuf/proto"
)

// TestMarshalAnimatedThumbnailSpec verifies the F2 animated-preview fields on
// ThumbnailSpec serialize to the wire form the server expects: snake_case field
// names, the mode enum simplified to the lowercase string "animated", and the
// repeated start_offsets emitted as a JSON array.
func TestMarshalAnimatedThumbnailSpec(t *testing.T) {
	c := NewProtoJSONCodec()
	spec := &v1.ThumbnailSpec{
		Mode:            v1.ThumbnailMode_THUMBNAIL_MODE_ANIMATED,
		DurationSeconds: proto.Float64(6),
		Fps:             proto.Int32(10),
		StartOffsets:    []float64{1.5, 4, 9.25},
	}
	data, err := c.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	// Mode simplifies to the lowercase wire string.
	if mode := string(m["mode"]); mode != `"animated"` {
		t.Errorf(`mode = %s, want "animated"\nfull payload: %s`, mode, got)
	}
	// Snake_case keys are present (not the protojson camelCase defaults).
	for _, want := range []string{"duration_seconds", "fps", "start_offsets"} {
		if _, ok := m[want]; !ok {
			t.Errorf("payload missing snake_case key %q\nfull payload: %s", want, got)
		}
	}
	for _, bad := range []string{"durationSeconds", "startOffsets"} {
		if _, ok := m[bad]; ok {
			t.Errorf("payload unexpectedly contains camelCase key %q\nfull payload: %s", bad, got)
		}
	}
	// start_offsets round-trips as a numeric array.
	var offsets []float64
	if err := json.Unmarshal(m["start_offsets"], &offsets); err != nil {
		t.Fatalf("start_offsets re-parse: %v", err)
	}
	if len(offsets) != 3 || offsets[0] != 1.5 || offsets[2] != 9.25 {
		t.Errorf("start_offsets = %v, want [1.5 4 9.25]", offsets)
	}
}

// TestRoundTripAnimatedThumbnailSpec ensures the animated-mode ThumbnailSpec
// survives a Marshal→Unmarshal cycle with its enum and scalar values intact,
// including the lowercase "animated" enum string being expanded back.
func TestRoundTripAnimatedThumbnailSpec(t *testing.T) {
	c := NewProtoJSONCodec()
	original := &v1.ThumbnailSpec{
		Mode:            v1.ThumbnailMode_THUMBNAIL_MODE_ANIMATED,
		DurationSeconds: proto.Float64(4),
		Fps:             proto.Int32(12),
		StartOffsets:    []float64{0, 2.5},
	}
	data, err := c.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded v1.ThumbnailSpec
	if err := c.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.GetMode() != v1.ThumbnailMode_THUMBNAIL_MODE_ANIMATED {
		t.Errorf("mode: got %v, want ANIMATED", decoded.GetMode())
	}
	if decoded.GetDurationSeconds() != 4 {
		t.Errorf("duration_seconds: got %v, want 4", decoded.GetDurationSeconds())
	}
	if decoded.GetFps() != 12 {
		t.Errorf("fps: got %v, want 12", decoded.GetFps())
	}
	if len(decoded.GetStartOffsets()) != 2 || decoded.GetStartOffsets()[1] != 2.5 {
		t.Errorf("start_offsets: got %v, want [0 2.5]", decoded.GetStartOffsets())
	}
}

// TestUnmarshalVideoHoverPreviewURLs verifies the F2 Video hover-preview URL
// fields parse from the snake_case wire form into their typed accessors.
func TestUnmarshalVideoHoverPreviewURLs(t *testing.T) {
	c := NewProtoJSONCodec()
	payload := []byte(`{"video":{"id":"vid_abc","hover_preview_url":"https://cdn.example.com/p.webp","hover_preview_mp4_url":"https://cdn.example.com/p.mp4"}}`)
	var resp v1.GetVideoResponse
	if err := c.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	v := resp.GetVideo()
	if got := v.GetHoverPreviewUrl(); got != "https://cdn.example.com/p.webp" {
		t.Errorf("hover_preview_url: got %q", got)
	}
	if got := v.GetHoverPreviewMp4Url(); got != "https://cdn.example.com/p.mp4" {
		t.Errorf("hover_preview_mp4_url: got %q", got)
	}
}

// TestMarshalCreateUploadHoverPreviews verifies the CreateUploadRequest opt-in
// toggle serializes as the snake_case boolean the server expects.
func TestMarshalCreateUploadHoverPreviews(t *testing.T) {
	c := NewProtoJSONCodec()
	req := &v1.CreateUploadRequest{
		Filename:      "clip.mp4",
		HoverPreviews: true,
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if v, ok := m["hover_previews"]; !ok || string(v) != "true" {
		t.Errorf(`hover_previews = %s (present=%v), want true; full payload: %s`, v, ok, data)
	}
	if _, ok := m["hoverPreviews"]; ok {
		t.Errorf("payload unexpectedly contains camelCase key hoverPreviews: %s", data)
	}
}
