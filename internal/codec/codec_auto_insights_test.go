package codec

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalAutoInsights verifies the auto_insights upload opt-in serializes
// as the snake_case boolean the server expects on all three video-creating
// requests, beside auto_captions, and stays off the wire when unset.
func TestMarshalAutoInsights(t *testing.T) {
	cases := map[string]struct {
		set   proto.Message
		unset proto.Message
	}{
		"CreateUploadRequest": {
			set:   &v1.CreateUploadRequest{Filename: "clip.mp4", AutoCaptions: proto.Bool(true), AutoInsights: proto.Bool(true)},
			unset: &v1.CreateUploadRequest{Filename: "clip.mp4", AutoCaptions: proto.Bool(true)},
		},
		"CreateFromUrlRequest": {
			set:   &v1.CreateFromUrlRequest{Url: "https://example.com/a.mp4", AutoCaptions: proto.Bool(true), AutoInsights: proto.Bool(true)},
			unset: &v1.CreateFromUrlRequest{Url: "https://example.com/a.mp4", AutoCaptions: proto.Bool(true)},
		},
		"CreateMultipartUploadRequest": {
			set:   &v1.CreateMultipartUploadRequest{Filename: "big.mp4", AutoCaptions: proto.Bool(true), AutoInsights: proto.Bool(true)},
			unset: &v1.CreateMultipartUploadRequest{Filename: "big.mp4", AutoCaptions: proto.Bool(true)},
		},
	}
	c := NewProtoJSONCodec()
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			m := marshalToMap(t, c, tc.set)
			if v, ok := m["auto_insights"]; !ok || string(v) != "true" {
				t.Errorf("auto_insights = %s (present=%v), want true; payload: %v", v, ok, m)
			}
			if v, ok := m["auto_captions"]; !ok || string(v) != "true" {
				t.Errorf("auto_captions = %s (present=%v), want true; payload: %v", v, ok, m)
			}
			if _, ok := m["autoInsights"]; ok {
				t.Errorf("payload unexpectedly contains camelCase key autoInsights: %v", m)
			}

			m = marshalToMap(t, c, tc.unset)
			if _, ok := m["auto_insights"]; ok {
				t.Errorf("auto_insights present when unset: %v", m)
			}
		})
	}
}

func marshalToMap(t *testing.T, c *ProtoJSONCodec, msg proto.Message) map[string]json.RawMessage {
	t.Helper()
	data, err := c.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	return m
}
