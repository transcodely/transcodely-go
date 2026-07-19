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

// TestMarshalGenerateCaptionsJob verifies the F5 AI-captions wire shape: a
// captions-only output (no video variants, no output type) carrying a single
// generate track with language "auto", sourced from a hosted video via
// input_video_id. All three must serialize with snake_case keys and the
// simplified lowercase enum ("generate").
func TestMarshalGenerateCaptionsJob(t *testing.T) {
	c := NewProtoJSONCodec()
	videoID := "vid_a1b2c3d4e5f6g7"
	lang := "auto"
	req := &v1.CreateJobRequest{
		InputVideoId: &videoID,
		Outputs: []*v1.OutputSpec{{
			// Captions-only: no video[] and no type — allowed by the relaxed
			// CEL because every subtitle track is a generate operation.
			SubtitleTracks: []*v1.SubtitleTrack{{
				Operation: v1.SubtitleOperation_SUBTITLE_OPERATION_GENERATE,
				Language:  &lang,
			}},
		}},
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"input_video_id":"vid_a1b2c3d4e5f6g7"`,
		`"subtitle_tracks":`,
		`"operation":"generate"`,
		`"language":"auto"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %q\nfull payload: %s", want, got)
		}
	}
	// input_video_id is the retro-caption source; input_url stays empty (the
	// codec emits unpopulated fields, so it appears as "").
	if strings.Contains(got, `"input_url":"http`) {
		t.Errorf("captions-only retro-caption request should not set input_url\nfull payload: %s", got)
	}
}

// TestRoundTripGenerateCaptionsJob ensures the generate track and input_video_id
// survive a Marshal→Unmarshal cycle with their native enum/value intact.
func TestRoundTripGenerateCaptionsJob(t *testing.T) {
	c := NewProtoJSONCodec()
	videoID := "vid_a1b2c3d4e5f6g7"
	lang := "auto"
	original := &v1.CreateJobRequest{
		InputVideoId: &videoID,
		Outputs: []*v1.OutputSpec{{
			SubtitleTracks: []*v1.SubtitleTrack{{
				Operation: v1.SubtitleOperation_SUBTITLE_OPERATION_GENERATE,
				Language:  &lang,
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
	if decoded.GetInputVideoId() != videoID {
		t.Errorf("input_video_id: got %q, want %q", decoded.GetInputVideoId(), videoID)
	}
	tracks := decoded.GetOutputs()[0].GetSubtitleTracks()
	if len(tracks) != 1 {
		t.Fatalf("subtitle_tracks: got %d, want 1", len(tracks))
	}
	if tracks[0].GetOperation() != v1.SubtitleOperation_SUBTITLE_OPERATION_GENERATE {
		t.Errorf("operation: got %v, want GENERATE", tracks[0].GetOperation())
	}
	if tracks[0].GetLanguage() != "auto" {
		t.Errorf("language: got %q, want %q", tracks[0].GetLanguage(), "auto")
	}
}

func r2TestCredentials() *v1.S3Credentials {
	return &v1.S3Credentials{
		AccessKeyId:     "0123456789abcdef0123456789abcdef",
		SecretAccessKey: "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
	}
}

// TestMarshalR2OriginAccountIDForm verifies the account-id form serializes with
// snake_case keys (account_id, access_key_id, secret_access_key) and a lowercase
// jurisdiction enum — the exact wire shape the server expects. R2 reuses the
// S3Credentials message, so the credential keys match S3.
func TestMarshalR2OriginAccountIDForm(t *testing.T) {
	c := NewProtoJSONCodec()
	req := &v1.CreateOriginRequest{
		Name: "My R2 origin",
		R2: &v1.R2OriginConfig{
			Bucket:       "my-r2-bucket",
			Credentials:  r2TestCredentials(),
			AccountId:    "0123456789abcdef0123456789abcdef",
			Jurisdiction: v1.R2Jurisdiction_R2_JURISDICTION_EU,
		},
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		`"account_id":"0123456789abcdef0123456789abcdef"`,
		`"access_key_id":"0123456789abcdef0123456789abcdef"`,
		`"secret_access_key":`,
		`"jurisdiction":"eu"`,
		`"bucket":"my-r2-bucket"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %q\nfull payload: %s", want, got)
		}
	}
	for _, bad := range []string{`"accountId"`, `"accessKeyId"`, `"secretAccessKey"`} {
		if strings.Contains(got, bad) {
			t.Errorf("payload unexpectedly contains camelCase key %s\nfull payload: %s", bad, got)
		}
	}
}

// TestMarshalR2OriginEndpointForm verifies the endpoint escape hatch: the explicit
// endpoint is present while account_id stays empty and jurisdiction stays
// unspecified (the server's CEL rules require exactly one of the two locations).
func TestMarshalR2OriginEndpointForm(t *testing.T) {
	c := NewProtoJSONCodec()
	endpoint := "https://example.r2.cloudflarestorage.com"
	req := &v1.CreateOriginRequest{
		Name: "My R2 origin (endpoint)",
		R2: &v1.R2OriginConfig{
			Bucket:      "my-r2-bucket",
			Credentials: r2TestCredentials(),
			Endpoint:    &endpoint,
		},
	}
	data, err := c.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		R2 struct {
			AccountID    string `json:"account_id"`
			Jurisdiction string `json:"jurisdiction"`
			Endpoint     string `json:"endpoint"`
		} `json:"r2"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if out.R2.Endpoint != endpoint {
		t.Errorf("endpoint = %q, want %q", out.R2.Endpoint, endpoint)
	}
	if out.R2.AccountID != "" {
		t.Errorf("account_id = %q, want empty in endpoint form", out.R2.AccountID)
	}
	// EmitUnpopulated emits the zero enum, which simplifies to "unspecified".
	if out.R2.Jurisdiction != "" && out.R2.Jurisdiction != "unspecified" {
		t.Errorf(`jurisdiction = %q, want empty or "unspecified" in endpoint form`, out.R2.Jurisdiction)
	}
}

// TestRoundTripR2Origin ensures both R2 construction forms survive a
// Marshal→Unmarshal cycle with their native struct/enum values intact.
func TestRoundTripR2Origin(t *testing.T) {
	c := NewProtoJSONCodec()
	endpoint := "https://example.r2.cloudflarestorage.com"
	cases := map[string]*v1.CreateOriginRequest{
		"account-id form": {
			Name: "acct",
			R2: &v1.R2OriginConfig{
				Bucket:       "my-r2-bucket",
				Credentials:  r2TestCredentials(),
				AccountId:    "0123456789abcdef0123456789abcdef",
				Jurisdiction: v1.R2Jurisdiction_R2_JURISDICTION_EU,
			},
		},
		"endpoint form": {
			Name: "endpoint",
			R2: &v1.R2OriginConfig{
				Bucket:      "my-r2-bucket",
				Credentials: r2TestCredentials(),
				Endpoint:    &endpoint,
			},
		},
	}
	for name, original := range cases {
		t.Run(name, func(t *testing.T) {
			data, err := c.Marshal(original)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded v1.CreateOriginRequest
			if err := c.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			r2 := decoded.GetR2()
			if r2 == nil {
				t.Fatalf("decoded R2 config is nil")
			}
			want := original.GetR2()
			if r2.GetBucket() != want.GetBucket() {
				t.Errorf("bucket: got %q, want %q", r2.GetBucket(), want.GetBucket())
			}
			if r2.GetCredentials().GetAccessKeyId() != want.GetCredentials().GetAccessKeyId() {
				t.Errorf("access_key_id mismatch")
			}
			if r2.GetAccountId() != want.GetAccountId() {
				t.Errorf("account_id: got %q, want %q", r2.GetAccountId(), want.GetAccountId())
			}
			if r2.GetJurisdiction() != want.GetJurisdiction() {
				t.Errorf("jurisdiction: got %v, want %v", r2.GetJurisdiction(), want.GetJurisdiction())
			}
			if r2.GetEndpoint() != want.GetEndpoint() {
				t.Errorf("endpoint: got %q, want %q", r2.GetEndpoint(), want.GetEndpoint())
			}
		})
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
