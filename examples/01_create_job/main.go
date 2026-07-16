// Submit a transcoding job and print its ID.
package main

import (
	"context"
	"log"
	"os"

	"github.com/transcodely/transcodely-go"
	"google.golang.org/protobuf/proto"
)

func main() {
	client, err := transcodely.New(os.Getenv("TRANSCODELY_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	job, err := client.Jobs.Create(context.Background(), &transcodely.JobCreateParams{
		InputUrl: "https://download.samplelib.com/mp4/sample-30s.mp4",
		// Write outputs to Transcodely-managed storage. Drop Managed and set
		// OutputOriginId to write to your own configured origin.
		Managed: proto.Bool(true),
		Outputs: []*transcodely.OutputSpec{{
			Type: transcodely.OutputFormatHLS,
			Video: []*transcodely.VideoVariant{
				{Codec: transcodely.VideoCodecH264, Resolution: transcodely.Resolution1080P},
				{Codec: transcodely.VideoCodecH264, Resolution: transcodely.Resolution720P},
				{Codec: transcodely.VideoCodecH264, Resolution: transcodely.Resolution480P},
			},
		}},
		Metadata: map[string]string{"source": "01_create_job"},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created %s in status %s", job.GetId(), job.GetStatus())
}
