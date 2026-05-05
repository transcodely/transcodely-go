// Iterate every job using auto-pagination.
package main

import (
	"context"
	"log"
	"os"

	"github.com/transcodely/transcodely-go"
)

func main() {
	client, err := transcodely.New(os.Getenv("TRANSCODELY_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	iter := client.Jobs.List(context.Background(), &transcodely.JobListParams{
		Pagination: &transcodely.PaginationRequest{Limit: 50},
	})
	defer iter.Close()

	seen := 0
	for iter.Next() {
		job := iter.Current()
		seen++
		log.Printf("%d. %s %s", seen, job.GetId(), job.GetStatus())
		if seen >= 200 {
			break
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("total seen: %d", seen)
}
