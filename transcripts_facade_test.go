package transcodely

import (
	"strings"
	"testing"

	"github.com/transcodely/transcodely-go/internal/codec"
	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestTranscriptFacadeEnumConstants pins every facade constant for the
// transcripts/insights enums to its generated value and to the lowercase
// string it carries on the wire, so a constant bound to the wrong value fails.
func TestTranscriptFacadeEnumConstants(t *testing.T) {
	c := codec.NewProtoJSONCodec()
	marshal := func(t *testing.T, msg any) string {
		t.Helper()
		b, err := c.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return strings.ReplaceAll(string(b), " ", "")
	}

	styles := []struct {
		got  TranscriptStyle
		want v1.TranscriptStyle
		wire string
	}{
		{TranscriptStyleClean, v1.TranscriptStyle_TRANSCRIPT_STYLE_CLEAN, "clean"},
		{TranscriptStyleLiteral, v1.TranscriptStyle_TRANSCRIPT_STYLE_LITERAL, "literal"},
	}
	for _, s := range styles {
		if s.got != s.want {
			t.Errorf("style constant = %v, want %v", s.got, s.want)
		}
		if out := marshal(t, &SubtitleResult{TranscriptStyle: s.got}); !strings.Contains(out, `"transcript_style":"`+s.wire+`"`) {
			t.Errorf("style %v marshals as %s, want %q", s.got, out, s.wire)
		}
	}

	formats := []struct {
		got  TranscriptFormat
		want v1.TranscriptFormat
		wire string
	}{
		{TranscriptFormatJSON, v1.TranscriptFormat_TRANSCRIPT_FORMAT_JSON, "json"},
		{TranscriptFormatTXT, v1.TranscriptFormat_TRANSCRIPT_FORMAT_TXT, "txt"},
		{TranscriptFormatMD, v1.TranscriptFormat_TRANSCRIPT_FORMAT_MD, "md"},
	}
	for _, f := range formats {
		if f.got != f.want {
			t.Errorf("format constant = %v, want %v", f.got, f.want)
		}
		if out := marshal(t, &TranscriptRendition{Format: f.got}); !strings.Contains(out, `"format":"`+f.wire+`"`) {
			t.Errorf("format %v marshals as %s, want %q", f.got, out, f.wire)
		}
	}

	statuses := []struct {
		got  InsightsStatus
		want v1.InsightsStatus
		wire string
	}{
		{InsightsStatusPending, v1.InsightsStatus_INSIGHTS_STATUS_PENDING, "pending"},
		{InsightsStatusReady, v1.InsightsStatus_INSIGHTS_STATUS_READY, "ready"},
		{InsightsStatusFailed, v1.InsightsStatus_INSIGHTS_STATUS_FAILED, "failed"},
	}
	for _, s := range statuses {
		if s.got != s.want {
			t.Errorf("status constant = %v, want %v", s.got, s.want)
		}
		if out := marshal(t, &TranscriptInsights{Status: s.got}); !strings.Contains(out, `"status":"`+s.wire+`"`) {
			t.Errorf("status %v marshals as %s, want %q", s.got, out, s.wire)
		}
	}

	if TranscriptStyleUnspecified != v1.TranscriptStyle_TRANSCRIPT_STYLE_UNSPECIFIED ||
		TranscriptFormatUnspecified != v1.TranscriptFormat_TRANSCRIPT_FORMAT_UNSPECIFIED ||
		InsightsStatusUnspecified != v1.InsightsStatus_INSIGHTS_STATUS_UNSPECIFIED {
		t.Error("an Unspecified constant is not bound to its zero value")
	}
}
