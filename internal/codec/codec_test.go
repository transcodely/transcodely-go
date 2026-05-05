package codec

import (
	"encoding/json"
	"strings"
	"testing"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestMarshalSimplifiesEnums verifies the Marshal path strips the proto enum
// prefix and lowercases the value, matching the wire format the server emits.
func TestMarshalSimplifiesEnums(t *testing.T) {
	c := NewProtoJSONCodec()
	req := &v1.CreateJobRequest{
		InputUrl: "https://example.com/in.mp4",
		Outputs: []*v1.OutputSpec{{
			Type: v1.OutputFormat_OUTPUT_FORMAT_HLS,
			Video: []*v1.VideoVariant{{
				Codec:      v1.VideoCodec_VIDEO_CODEC_H264,
				Resolution: v1.Resolution_RESOLUTION_1080P,
			}},
		}},
		Priority: v1.JobPriority_JOB_PRIORITY_STANDARD,
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"input_url":"https://example.com/in.mp4"`,
		`"type":"hls"`,
		`"codec":"h264"`,
		`"resolution":"1080p"`,
		`"priority":"standard"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\nfull payload: %s", want, got)
		}
	}
}

// TestUnmarshalExpandsEnums verifies the Unmarshal path accepts the simplified
// wire form and produces the correct enum integer in the proto message.
func TestUnmarshalExpandsEnums(t *testing.T) {
	c := NewProtoJSONCodec()
	payload := []byte(`{"job":{"id":"job_a","status":"processing","priority":"premium"}}`)
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

// TestUnmarshalAcceptsFullEnumNames verifies backwards-compat: callers that
// happen to send the verbose `JOB_STATUS_PROCESSING` form still parse.
func TestUnmarshalAcceptsFullEnumNames(t *testing.T) {
	c := NewProtoJSONCodec()
	payload := []byte(`{"job":{"id":"job_a","status":"JOB_STATUS_PROCESSING"}}`)
	var resp v1.GetJobResponse
	if err := c.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.GetJob().GetStatus() != v1.JobStatus_JOB_STATUS_PROCESSING {
		t.Errorf("status: got %v, want PROCESSING", resp.GetJob().GetStatus())
	}
}

// TestRoundTrip ensures Marshal then Unmarshal recovers the original message.
func TestRoundTrip(t *testing.T) {
	c := NewProtoJSONCodec()
	original := &v1.CreateJobRequest{
		InputUrl: "https://example.com/in.mp4",
		Outputs: []*v1.OutputSpec{{
			Type: v1.OutputFormat_OUTPUT_FORMAT_DASH,
			Video: []*v1.VideoVariant{{
				Codec:      v1.VideoCodec_VIDEO_CODEC_AV1,
				Resolution: v1.Resolution_RESOLUTION_2160P,
			}},
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
	if decoded.GetInputUrl() != original.GetInputUrl() {
		t.Errorf("input_url mismatch")
	}
	if len(decoded.GetOutputs()) != 1 ||
		decoded.GetOutputs()[0].GetType() != v1.OutputFormat_OUTPUT_FORMAT_DASH ||
		decoded.GetOutputs()[0].GetVideo()[0].GetCodec() != v1.VideoCodec_VIDEO_CODEC_AV1 ||
		decoded.GetOutputs()[0].GetVideo()[0].GetResolution() != v1.Resolution_RESOLUTION_2160P {
		t.Errorf("decoded output mismatch: %+v", decoded.GetOutputs())
	}
}

// TestSnakeCaseFields verifies the Marshal output uses snake_case (not the
// default protojson camelCase).
func TestSnakeCaseFields(t *testing.T) {
	c := NewProtoJSONCodec()
	priority := v1.JobPriority_JOB_PRIORITY_ECONOMY
	req := &v1.CreateJobRequest{InputUrl: "x", Priority: priority}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if _, ok := m["input_url"]; !ok {
		t.Errorf("expected snake_case key 'input_url', got: %v", keys(m))
	}
	if _, ok := m["inputUrl"]; ok {
		t.Errorf("unexpected camelCase 'inputUrl' present in output")
	}
}

func keys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
