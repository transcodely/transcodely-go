package codec

import (
	"encoding/json"
	"testing"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TestToScreamingSnake covers the CamelCase → SCREAMING_SNAKE_CASE conversion,
// including the tricky initialism cases.
func TestToScreamingSnake(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"JobStatus", "JOB_STATUS"},
		{"VideoCodec", "VIDEO_CODEC"},
		{"OutputFormat", "OUTPUT_FORMAT"},
		{"APIKeyEnvironment", "API_KEY_ENVIRONMENT"},
		{"HTTPCredentials", "HTTP_CREDENTIALS"},
		{"DRMSystem", "DRM_SYSTEM"},
		{"HDRFormat", "HDR_FORMAT"},
		{"Resolution", "RESOLUTION"},
	}
	for _, c := range cases {
		if got := toScreamingSnake(c.in); got != c.want {
			t.Errorf("toScreamingSnake(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestEnumPrefix verifies the prefix derived for several real proto enums.
func TestEnumPrefix(t *testing.T) {
	job := (&v1.Job{}).ProtoReflect().Descriptor()
	statusEnum := job.Fields().ByName("status").Enum()
	priorityEnum := job.Fields().ByName("priority").Enum()

	if got := enumPrefix(statusEnum); got != "JOB_STATUS_" {
		t.Errorf("status prefix = %q, want JOB_STATUS_", got)
	}
	if got := enumPrefix(priorityEnum); got != "JOB_PRIORITY_" {
		t.Errorf("priority prefix = %q, want JOB_PRIORITY_", got)
	}
}

// TestSimplifyExpandRoundTrip is the wire-format conformance matrix. One
// failure here means the SDK's enum mapping has drifted from the server's.
func TestSimplifyExpandRoundTrip(t *testing.T) {
	job := (&v1.Job{}).ProtoReflect().Descriptor()
	statusEnum := job.Fields().ByName("status").Enum()
	priorityEnum := job.Fields().ByName("priority").Enum()

	variant := (&v1.VideoVariant{}).ProtoReflect().Descriptor()
	codecEnum := variant.Fields().ByName("codec").Enum()
	resEnum := variant.Fields().ByName("resolution").Enum()

	output := (&v1.OutputSpec{}).ProtoReflect().Descriptor()
	formatEnum := output.Fields().ByName("type").Enum()

	type tc struct {
		canonical, simple string
		enum              protoreflect.EnumDescriptor
	}
	cases := []tc{
		{"JOB_STATUS_PENDING", "pending", statusEnum},
		{"JOB_STATUS_PROCESSING", "processing", statusEnum},
		{"JOB_STATUS_COMPLETED", "completed", statusEnum},
		{"JOB_STATUS_AWAITING_CONFIRMATION", "awaiting_confirmation", statusEnum},
		{"JOB_PRIORITY_STANDARD", "standard", priorityEnum},
		{"JOB_PRIORITY_ECONOMY", "economy", priorityEnum},
		{"JOB_PRIORITY_PREMIUM", "premium", priorityEnum},
		{"VIDEO_CODEC_H264", "h264", codecEnum},
		{"VIDEO_CODEC_AV1", "av1", codecEnum},
		{"RESOLUTION_1080P", "1080p", resEnum},
		{"RESOLUTION_2160P", "2160p", resEnum},
		{"OUTPUT_FORMAT_HLS", "hls", formatEnum},
		{"OUTPUT_FORMAT_DASH", "dash", formatEnum},
	}

	for _, c := range cases {
		if got := simplifyEnumValue(c.canonical, c.enum); got != c.simple {
			t.Errorf("simplify %q: got %q, want %q", c.canonical, got, c.simple)
		}
		if got := expandEnumValue(c.simple, c.enum); got != c.canonical {
			t.Errorf("expand %q: got %q, want %q", c.simple, got, c.canonical)
		}
	}
}

// TestExpandEnumValueUnknown leaves an unrecognised value untouched so
// protojson can produce a meaningful error downstream.
func TestExpandEnumValueUnknown(t *testing.T) {
	job := (&v1.Job{}).ProtoReflect().Descriptor()
	statusEnum := job.Fields().ByName("status").Enum()
	if got := expandEnumValue("totally_unknown", statusEnum); got != "totally_unknown" {
		t.Errorf("expand unknown: got %q, want it returned unchanged", got)
	}
}

// TestUnmarshalMixedFormEnums verifies a payload may freely mix simplified
// and canonical enum names; the codec accepts both.
func TestUnmarshalMixedFormEnums(t *testing.T) {
	c := NewProtoJSONCodec()
	payload := []byte(`{"job":{"id":"job_a","status":"processing","priority":"JOB_PRIORITY_PREMIUM"}}`)
	var resp v1.GetJobResponse
	if err := c.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.GetJob().GetStatus() != v1.JobStatus_JOB_STATUS_PROCESSING {
		t.Errorf("status: got %v, want PROCESSING", resp.GetJob().GetStatus())
	}
	if resp.GetJob().GetPriority() != v1.JobPriority_JOB_PRIORITY_PREMIUM {
		t.Errorf("priority: got %v, want PREMIUM", resp.GetJob().GetPriority())
	}
}

// TestMarshalRepeatedAndNested verifies enum simplification reaches inside
// repeated fields and into deeply nested messages (OutputSpec → VideoVariant).
func TestMarshalRepeatedAndNested(t *testing.T) {
	c := NewProtoJSONCodec()
	req := &v1.CreateJobRequest{
		InputUrl: "https://example.com/in.mp4",
		Outputs: []*v1.OutputSpec{
			{
				Type: v1.OutputFormat_OUTPUT_FORMAT_HLS,
				Video: []*v1.VideoVariant{
					{Codec: v1.VideoCodec_VIDEO_CODEC_H264, Resolution: v1.Resolution_RESOLUTION_720P},
					{Codec: v1.VideoCodec_VIDEO_CODEC_H265, Resolution: v1.Resolution_RESOLUTION_1080P},
				},
			},
			{
				Type: v1.OutputFormat_OUTPUT_FORMAT_DASH,
				Video: []*v1.VideoVariant{
					{Codec: v1.VideoCodec_VIDEO_CODEC_AV1, Resolution: v1.Resolution_RESOLUTION_2160P},
				},
			},
		},
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Outputs []struct {
			Type  string `json:"type"`
			Video []struct {
				Codec      string `json:"codec"`
				Resolution string `json:"resolution"`
			} `json:"video"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(out.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(out.Outputs))
	}
	if out.Outputs[0].Type != "hls" {
		t.Errorf("outputs[0].type = %q, want hls", out.Outputs[0].Type)
	}
	if out.Outputs[1].Type != "dash" {
		t.Errorf("outputs[1].type = %q, want dash", out.Outputs[1].Type)
	}
	if out.Outputs[0].Video[0].Codec != "h264" || out.Outputs[0].Video[1].Codec != "h265" {
		t.Errorf("HLS codecs mismatch: %+v", out.Outputs[0].Video)
	}
	if out.Outputs[1].Video[0].Resolution != "2160p" {
		t.Errorf("DASH resolution = %q, want 2160p", out.Outputs[1].Video[0].Resolution)
	}
}

// TestUnmarshalIgnoresUnknownFields makes sure adding a new server field
// doesn't break older SDKs.
func TestUnmarshalIgnoresUnknownFields(t *testing.T) {
	c := NewProtoJSONCodec()
	payload := []byte(`{"job":{"id":"job_a"},"made_up_field":42}`)
	var resp v1.GetJobResponse
	if err := c.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := resp.GetJob().GetId(); got != "job_a" {
		t.Errorf("job id = %q, want job_a", got)
	}
}

// TestMarshalRejectsNonProtoMessage verifies the type guard surfaces a
// recognisable error rather than a runtime panic.
func TestMarshalRejectsNonProtoMessage(t *testing.T) {
	c := NewProtoJSONCodec()
	_, err := c.Marshal(map[string]any{"not": "a proto"})
	if err == nil {
		t.Fatal("expected error for non-proto message, got nil")
	}
}
