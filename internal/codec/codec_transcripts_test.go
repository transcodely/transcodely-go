package codec

import (
	"strings"
	"testing"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
)

// TestUnmarshalTranscriptsAndInsights decodes a job payload the way the API
// emits it: snake_case fields and lowercase enums for the transcript renditions,
// transcript style and insights status.
func TestUnmarshalTranscriptsAndInsights(t *testing.T) {
	c := NewProtoJSONCodec()
	body := `{
	  "id": "job_a1b2c3d4e5f6",
	  "insights": [{
	    "output_id": "out_a1b2c3d4e5f6g7",
	    "language": "eng",
	    "status": "ready",
	    "summary": "A short talk.",
	    "key_points": ["one", "two"],
	    "questions": [{"question": "Why?", "answer": "Because.", "start_seconds": 12.5}],
	    "terms": [{"term": "GOP", "definition": "Group of pictures.", "start_seconds": 3}],
	    "highlights": [{"text": "verbatim", "start_seconds": 1, "end_seconds": 2.5}],
	    "chapters": [{"start_seconds": 0, "title": "Intro"}],
	    "generated_at": "2026-10-24T10:00:00Z"
	  }]
	}`
	var job v1.Job
	if err := c.Unmarshal([]byte(body), &job); err != nil {
		t.Fatalf("unmarshal job: %v", err)
	}
	if len(job.GetInsights()) != 1 {
		t.Fatalf("insights len = %d, want 1", len(job.GetInsights()))
	}
	ins := job.GetInsights()[0]
	if ins.GetStatus() != v1.InsightsStatus_INSIGHTS_STATUS_READY {
		t.Errorf("status = %v, want READY", ins.GetStatus())
	}
	if ins.GetQuestions()[0].GetStartSeconds() != 12.5 || ins.GetTerms()[0].GetTerm() != "GOP" {
		t.Errorf("questions/terms not decoded: %+v", ins)
	}
	if ins.GetHighlights()[0].GetEndSeconds() != 2.5 || len(ins.GetKeyPoints()) != 2 {
		t.Errorf("highlights/key_points not decoded: %+v", ins)
	}
	if ins.GetChapters()[0].GetTitle() != "Intro" || ins.GetGeneratedAt().GetSeconds() == 0 {
		t.Errorf("chapters/generated_at not decoded: %+v", ins)
	}

	res := `{
	  "transcript_style": "literal",
	  "transcript_renditions": [
	    {"format": "json", "storage_key": "a/t.json", "url": "gs://b/a/t.json"},
	    {"format": "txt", "storage_key": "a/t.txt", "url": "gs://b/a/t.txt", "signed_url": "https://x/t.txt"},
	    {"format": "md", "storage_key": "a/t.md", "url": "gs://b/a/t.md"}
	  ]
	}`
	var sr v1.SubtitleResult
	if err := c.Unmarshal([]byte(res), &sr); err != nil {
		t.Fatalf("unmarshal subtitle result: %v", err)
	}
	if sr.GetTranscriptStyle() != v1.TranscriptStyle_TRANSCRIPT_STYLE_LITERAL {
		t.Errorf("transcript_style = %v, want LITERAL", sr.GetTranscriptStyle())
	}
	want := []v1.TranscriptFormat{
		v1.TranscriptFormat_TRANSCRIPT_FORMAT_JSON,
		v1.TranscriptFormat_TRANSCRIPT_FORMAT_TXT,
		v1.TranscriptFormat_TRANSCRIPT_FORMAT_MD,
	}
	if len(sr.GetTranscriptRenditions()) != len(want) {
		t.Fatalf("renditions len = %d, want %d", len(sr.GetTranscriptRenditions()), len(want))
	}
	for i, r := range sr.GetTranscriptRenditions() {
		if r.GetFormat() != want[i] {
			t.Errorf("renditions[%d].format = %v, want %v", i, r.GetFormat(), want[i])
		}
	}
	if sr.GetTranscriptRenditions()[1].GetSignedUrl() != "https://x/t.txt" {
		t.Errorf("signed_url not decoded")
	}
}

// TestMarshalTranscriptAndInsightsOptions pins the request side: the options on
// a generate track serialize as snake_case with a lowercase style.
func TestMarshalTranscriptAndInsightsOptions(t *testing.T) {
	c := NewProtoJSONCodec()
	style := v1.TranscriptStyle_TRANSCRIPT_STYLE_CLEAN
	hint := "lecture for first-year students"
	track := &v1.SubtitleTrack{
		Operation:  v1.SubtitleOperation_SUBTITLE_OPERATION_GENERATE,
		Transcript: &v1.TranscriptOptions{Style: &style},
		Insights:   &v1.InsightsOptions{Enabled: true, Hint: &hint},
	}
	data, err := c.Marshal(track)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)
	for _, frag := range []string{
		`"operation":"generate"`,
		`"transcript":{"style":"clean"}`,
		`"insights":{"enabled":true,"hint":"lecture for first-year students"}`,
	} {
		if !strings.Contains(strings.ReplaceAll(got, " ", ""), strings.ReplaceAll(frag, " ", "")) {
			t.Errorf("marshal output %s missing %s", got, frag)
		}
	}
}
