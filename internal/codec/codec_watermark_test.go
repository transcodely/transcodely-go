package codec

import (
	"encoding/json"
	"strings"
	"testing"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalWatermarkRelative verifies that a create-job request carrying a
// relative-mode watermark (anchor + width_pct + margin_pct) serializes with
// snake_case field names and a simplified lowercase anchor enum ("bottom_right")
// — the exact wire shape the server expects.
func TestMarshalWatermarkRelative(t *testing.T) {
	c := NewProtoJSONCodec()
	anchor := v1.WatermarkAnchor_WATERMARK_ANCHOR_BOTTOM_RIGHT
	widthPct := 15.0
	marginPct := 2.0
	opacity := 0.8
	req := &v1.CreateJobRequest{
		InputUrl: "https://example.com/in.mp4",
		Outputs: []*v1.OutputSpec{{
			Type: v1.OutputFormat_OUTPUT_FORMAT_HLS,
			Watermark: &v1.WatermarkConfig{
				ImageUrl:  "https://cdn.example.com/logo.png",
				Anchor:    &anchor,
				WidthPct:  &widthPct,
				MarginPct: &marginPct,
				Opacity:   &opacity,
			},
		}},
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"watermark":`,
		`"image_url":"https://cdn.example.com/logo.png"`,
		`"anchor":"bottom_right"`,
		`"width_pct":15`,
		`"margin_pct":2`,
		`"opacity":0.8`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %q\nfull payload: %s", want, got)
		}
	}
	// No camelCase field names must leak from the default protojson encoder.
	for _, bad := range []string{`"imageUrl"`, `"widthPct"`, `"marginPct"`} {
		if strings.Contains(got, bad) {
			t.Errorf("payload unexpectedly contains camelCase key %s\nfull payload: %s", bad, got)
		}
	}
	// The verbose enum name must never reach the wire.
	if strings.Contains(got, "WATERMARK_ANCHOR_BOTTOM_RIGHT") {
		t.Errorf("payload contains verbose enum name; want simplified \"bottom_right\"\nfull payload: %s", got)
	}
}

// TestMarshalWatermarkPixel verifies the advanced pixel-placement mode: the
// `pixel` submessage serializes under a snake_case key with snake_case x/y/width
// fields, and image_url/opacity still apply.
func TestMarshalWatermarkPixel(t *testing.T) {
	c := NewProtoJSONCodec()
	opacity := 1.0
	req := &v1.CreateJobRequest{
		InputUrl: "https://example.com/in.mp4",
		Outputs: []*v1.OutputSpec{{
			Type: v1.OutputFormat_OUTPUT_FORMAT_MP4,
			Watermark: &v1.WatermarkConfig{
				ImageUrl: "https://cdn.example.com/logo.png",
				Opacity:  &opacity,
				Pixel: &v1.WatermarkPixelPlacement{
					X:     40,
					Y:     40,
					Width: 240,
				},
			},
		}},
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Outputs []struct {
			Watermark struct {
				ImageURL string  `json:"image_url"`
				Opacity  float64 `json:"opacity"`
				Pixel    struct {
					X     int32 `json:"x"`
					Y     int32 `json:"y"`
					Width int32 `json:"width"`
				} `json:"pixel"`
			} `json:"watermark"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(out.Outputs) != 1 {
		t.Fatalf("outputs len = %d, want 1", len(out.Outputs))
	}
	wm := out.Outputs[0].Watermark
	if wm.ImageURL != "https://cdn.example.com/logo.png" {
		t.Errorf("image_url = %q, want the logo url", wm.ImageURL)
	}
	if wm.Pixel.X != 40 || wm.Pixel.Y != 40 || wm.Pixel.Width != 240 {
		t.Errorf("pixel placement = %d/%d/%d, want 40/40/240", wm.Pixel.X, wm.Pixel.Y, wm.Pixel.Width)
	}
	// In pixel mode the relative anchor must be absent from the wire.
	if strings.Contains(string(data), `"anchor"`) {
		t.Errorf("pixel-mode payload unexpectedly contains an anchor field\nfull payload: %s", data)
	}
}

// TestRoundTripWatermark ensures a watermark config survives a Marshal→Unmarshal
// cycle with its enum and scalar values intact.
func TestRoundTripWatermark(t *testing.T) {
	c := NewProtoJSONCodec()
	anchor := v1.WatermarkAnchor_WATERMARK_ANCHOR_TOP_LEFT
	widthPct := 12.5
	original := &v1.CreateJobRequest{
		InputUrl: "https://example.com/in.mp4",
		Outputs: []*v1.OutputSpec{{
			Type: v1.OutputFormat_OUTPUT_FORMAT_HLS,
			Watermark: &v1.WatermarkConfig{
				ImageUrl: "https://cdn.example.com/logo.webp",
				Anchor:   &anchor,
				WidthPct: &widthPct,
			},
		}},
	}
	data, err := c.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded v1.CreateJobRequest
	if err := c.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wm := decoded.GetOutputs()[0].GetWatermark()
	if wm == nil {
		t.Fatalf("decoded watermark is nil")
	}
	if wm.GetImageUrl() != "https://cdn.example.com/logo.webp" {
		t.Errorf("image_url = %q", wm.GetImageUrl())
	}
	if wm.GetAnchor() != v1.WatermarkAnchor_WATERMARK_ANCHOR_TOP_LEFT {
		t.Errorf("anchor = %v, want TOP_LEFT", wm.GetAnchor())
	}
	if wm.GetWidthPct() != 12.5 {
		t.Errorf("width_pct = %v, want 12.5", wm.GetWidthPct())
	}
}
