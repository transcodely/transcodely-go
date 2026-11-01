package codec

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalCreateRequestsSourcePinned pins retain-at-create on all three
// create requests as snake_case source_pinned, and off the wire when unset.
func TestMarshalCreateRequestsSourcePinned(t *testing.T) {
	c := NewProtoJSONCodec()
	for name, msg := range map[string]proto.Message{
		"CreateUpload":          &v1.CreateUploadRequest{Filename: "a.mp4", SourcePinned: proto.Bool(true)},
		"CreateMultipartUpload": &v1.CreateMultipartUploadRequest{Filename: "a.mp4", SourcePinned: proto.Bool(true)},
		"CreateFromUrl":         &v1.CreateFromUrlRequest{Url: "https://example.com/a.mp4", SourcePinned: proto.Bool(true)},
	} {
		m := marshalToMap(t, c, msg)
		if got := string(m["source_pinned"]); got != "true" {
			t.Errorf("%s: source_pinned = %q; payload: %v", name, got, m)
		}
	}
	m := marshalToMap(t, c, &v1.CreateUploadRequest{Filename: "a.mp4"})
	if _, ok := m["source_pinned"]; ok {
		t.Errorf("source_pinned present when unset: %v", m)
	}
}

// TestUnmarshalSourceRetentionFields verifies Video.source_expires_at and the
// UsageSummary retained-source pair decode from their snake_case wire names.
func TestUnmarshalSourceRetentionFields(t *testing.T) {
	c := NewProtoJSONCodec()

	var v v1.Video
	body := `{"id":"vid_abc","source_pinned":false,"source_expires_at":"2026-11-08T21:23:00Z"}`
	if err := c.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("unmarshal video: %v", err)
	}
	want := time.Date(2026, 11, 8, 21, 23, 0, 0, time.UTC)
	if v.GetSourceExpiresAt() == nil || !v.GetSourceExpiresAt().AsTime().Equal(want) {
		t.Errorf("source_expires_at = %v, want %v", v.GetSourceExpiresAt(), want)
	}

	var u v1.UsageSummary
	body = `{"storage_gb_avg":1.5,"storage_cost":0.03,"retained_source_gb_avg":500,"retained_source_cost":10,"total_cost":10.03}`
	if err := c.Unmarshal([]byte(body), &u); err != nil {
		t.Fatalf("unmarshal usage: %v", err)
	}
	if u.GetRetainedSourceGbAvg() != 500 || u.GetRetainedSourceCost() != 10 || u.GetStorageGbAvg() != 1.5 {
		t.Errorf("usage = %+v", &u)
	}
}
