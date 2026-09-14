package transcodely

import (
	"testing"

	"github.com/transcodely/transcodely-go/internal/codec"
	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// The facade re-exports every OutputReport message, and each alias is the
// generated type itself rather than a distinct look-alike. These declarations
// fail to compile if an alias is missing or points somewhere else, which is
// the whole guarantee types.go makes.
var (
	_ *v1.OutputReport         = (*OutputReport)(nil)
	_ *v1.OutputReportVideo    = (*OutputReportVideo)(nil)
	_ *v1.OutputReportColor    = (*OutputReportColor)(nil)
	_ *v1.OutputReportAudio    = (*OutputReportAudio)(nil)
	_ *v1.OutputReportVerdict  = (*OutputReportVerdict)(nil)
	_ *v1.OutputReportMismatch = (*OutputReportMismatch)(nil)

	_ *OutputReport = (&JobOutput{}).GetReport()
)

// A GetJob response carrying an output report decodes through the SDK's own
// codec, with the wire's snake_case field names, and the verdict is readable
// off the typed message.
func TestOutputReport_DecodesFromTheWire(t *testing.T) {
	payload := []byte(`{"job":{"id":"job_abc123def456","status":"completed","outputs":[{
		"id":"out_abc123def4567","status":"completed","report":{
			"container":"mp4",
			"duration_seconds":30.5,
			"checked_at":"2026-09-14T10:00:00Z",
			"video":{
				"codec":"hevc","profile":"main10","level":"4.0","pix_fmt":"yuv420p10le",
				"width":1920,"height":1080,"frame_rate":29.97,"bitrate_kbps":4800,
				"hdr_format":"hdr10",
				"color":{"primaries":"bt2020","transfer":"smpte2084","matrix":"bt2020nc","range":"tv"}
			},
			"audio":[{"codec":"aac","channels":2,"sample_rate_hz":48000,"bitrate_kbps":128,"language":"eng"}],
			"verdict":{"matches_request":false,"mismatches":[
				{"field":"video.codec","expected":"h264","actual":"hevc"}
			]}
		}}]}}`)

	var resp v1.GetJobResponse
	if err := codec.NewProtoJSONCodec().Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	outputs := resp.GetJob().GetOutputs()
	if len(outputs) != 1 {
		t.Fatalf("outputs = %d, want 1", len(outputs))
	}
	report := outputs[0].GetReport()
	if report == nil {
		t.Fatal("output carries no report")
	}

	if got := report.GetVerdict().GetMatchesRequest(); got {
		t.Error("verdict.matches_request = true, want false")
	}
	mismatches := report.GetVerdict().GetMismatches()
	if len(mismatches) != 1 {
		t.Fatalf("mismatches = %d, want 1", len(mismatches))
	}
	if got := mismatches[0].GetField(); got != "video.codec" {
		t.Errorf("mismatches[0].field = %q, want %q", got, "video.codec")
	}
	if got, want := mismatches[0].GetExpected(), "h264"; got != want {
		t.Errorf("mismatches[0].expected = %q, want %q", got, want)
	}
	if got, want := mismatches[0].GetActual(), "hevc"; got != want {
		t.Errorf("mismatches[0].actual = %q, want %q", got, want)
	}

	if got, want := report.GetContainer(), "mp4"; got != want {
		t.Errorf("container = %q, want %q", got, want)
	}
	if got, want := report.GetDurationSeconds(), 30.5; got != want {
		t.Errorf("duration_seconds = %v, want %v", got, want)
	}
	video := report.GetVideo()
	if got, want := video.GetCodec(), "hevc"; got != want {
		t.Errorf("video.codec = %q, want %q", got, want)
	}
	if got, want := video.GetPixFmt(), "yuv420p10le"; got != want {
		t.Errorf("video.pix_fmt = %q, want %q", got, want)
	}
	if got, want := video.GetWidth(), int32(1920); got != want {
		t.Errorf("video.width = %d, want %d", got, want)
	}
	if got, want := video.GetHdrFormat(), "hdr10"; got != want {
		t.Errorf("video.hdr_format = %q, want %q", got, want)
	}
	if got, want := video.GetColor().GetTransfer(), "smpte2084"; got != want {
		t.Errorf("video.color.transfer = %q, want %q", got, want)
	}
	if n := len(report.GetAudio()); n != 1 {
		t.Fatalf("audio streams = %d, want 1", n)
	}
	if got, want := report.GetAudio()[0].GetSampleRateHz(), int32(48000); got != want {
		t.Errorf("audio[0].sample_rate_hz = %d, want %d", got, want)
	}
}

// An output with no report reads as nil rather than an empty report, so
// "not measured" stays distinguishable from "measured, nothing wrong".
func TestOutputReport_AbsentIsNil(t *testing.T) {
	payload := []byte(`{"job":{"id":"job_abc123def456","outputs":[{"id":"out_abc123def4567"}]}}`)

	var resp v1.GetJobResponse
	if err := codec.NewProtoJSONCodec().Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := resp.GetJob().GetOutputs()[0].GetReport(); got != nil {
		t.Errorf("report = %v, want nil", got)
	}
}
