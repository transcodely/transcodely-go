package transcodely

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// fakeJobClient is a minimal JobServiceClient stub. It embeds the generated
// interface so only the methods a test exercises need implementations; any
// other call would nil-panic, surfacing accidental use. It records the request
// message it received and returns a canned response.
type fakeJobClient struct {
	transcodelyv1connect.JobServiceClient

	gotCreate *v1.CreateJobRequest
	created   *v1.CreateJobResponse
}

func (f *fakeJobClient) Create(_ context.Context, req *connect.Request[v1.CreateJobRequest]) (*connect.Response[v1.CreateJobResponse], error) {
	f.gotCreate = req.Msg
	return connect.NewResponse(f.created), nil
}

// A Create carrying a Clip forwards start/end seconds verbatim, and the job's
// echoed clip is returned unwrapped.
func TestJobs_Create_ClipPassthrough(t *testing.T) {
	fake := &fakeJobClient{
		created: &v1.CreateJobResponse{
			Job: &v1.Job{
				Id:   "job_abc123",
				Clip: &ClipConfig{StartSeconds: 2, EndSeconds: 7},
			},
		},
	}
	j := newJobs(fake, fake, &config{})

	job, err := j.Create(context.Background(), &JobCreateParams{
		InputUrl: "https://example.com/video.mp4",
		Clip:     &ClipConfig{StartSeconds: 2, EndSeconds: 7},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Request was forwarded verbatim, including the clip range.
	if fake.gotCreate.GetClip() == nil {
		t.Fatal("Create() did not forward a clip")
	}
	if got := fake.gotCreate.GetClip().GetStartSeconds(); got != 2 {
		t.Errorf("clip.start_seconds = %v, want 2", got)
	}
	if got := fake.gotCreate.GetClip().GetEndSeconds(); got != 7 {
		t.Errorf("clip.end_seconds = %v, want 7", got)
	}

	// The created job (with its echoed clip) is returned unwrapped.
	if job.GetId() != "job_abc123" {
		t.Errorf("job.id = %q, want job_abc123", job.GetId())
	}
	if got := job.GetClip().GetEndSeconds(); got != 7 {
		t.Errorf("job.clip.end_seconds = %v, want 7", got)
	}
}

// An open-ended clip (end unset = end of input) forwards start_seconds and a
// zero end_seconds.
func TestJobs_Create_ClipOpenEnded(t *testing.T) {
	fake := &fakeJobClient{created: &v1.CreateJobResponse{Job: &v1.Job{Id: "job_open"}}}
	j := newJobs(fake, fake, &config{})

	if _, err := j.Create(context.Background(), &JobCreateParams{
		InputUrl: "https://example.com/video.mp4",
		Clip:     &ClipConfig{StartSeconds: 10},
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if got := fake.gotCreate.GetClip().GetStartSeconds(); got != 10 {
		t.Errorf("clip.start_seconds = %v, want 10", got)
	}
	if got := fake.gotCreate.GetClip().GetEndSeconds(); got != 0 {
		t.Errorf("clip.end_seconds = %v, want 0 (end of input)", got)
	}
}

// Create guards nil params and still forwards a request.
func TestJobs_Create_NilParams(t *testing.T) {
	fake := &fakeJobClient{created: &v1.CreateJobResponse{Job: &v1.Job{}}}
	j := newJobs(fake, fake, &config{})

	if _, err := j.Create(context.Background(), nil); err != nil {
		t.Fatalf("Create(nil) error = %v", err)
	}
	if fake.gotCreate == nil {
		t.Error("Create(nil) did not forward a request")
	}
	if fake.gotCreate.GetClip() != nil {
		t.Error("Create(nil) should not set a clip")
	}
}
