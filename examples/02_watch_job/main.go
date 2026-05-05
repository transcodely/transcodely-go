// Watch a job to completion.
package main

import (
	"context"
	"log"
	"os"

	"github.com/transcodely/transcodely-go"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: 02_watch_job <job_id>")
	}
	jobID := os.Args[1]

	client, err := transcodely.New(os.Getenv("TRANSCODELY_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	stream := client.Jobs.Watch(context.Background(), jobID)
	defer stream.Close()
	for stream.Next() {
		event := stream.Current()
		j := event.GetJob()
		log.Printf("[%s] progress=%d%%", j.GetStatus(), j.GetProgress())
		if j.GetStatus() == transcodely.JobStatusCompleted ||
			j.GetStatus() == transcodely.JobStatusFailed ||
			j.GetStatus() == transcodely.JobStatusCanceled {
			log.Printf("terminal: %s", j.GetStatus())
			return
		}
	}
	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
}
