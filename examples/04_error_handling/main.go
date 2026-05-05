// Catch typed errors and respond appropriately.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/transcodely/transcodely-go"
)

func main() {
	client, err := transcodely.New(os.Getenv("TRANSCODELY_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Jobs.Get(context.Background(), "job_does_not_exist")
	if err == nil {
		log.Fatal("expected error")
	}

	var notFound *transcodely.NotFoundError
	var auth *transcodely.AuthenticationError
	var invalid *transcodely.InvalidRequestError
	var rate *transcodely.RateLimitError
	var generic transcodely.Error

	switch {
	case errors.As(err, &notFound):
		log.Printf("not found, request id: %s", notFound.RequestID())
	case errors.As(err, &auth):
		log.Printf("auth failed — check TRANSCODELY_API_KEY")
	case errors.As(err, &invalid):
		for _, v := range invalid.Errors() {
			log.Printf("%s: %s", v.Field, v.Description)
		}
	case errors.As(err, &rate):
		time.Sleep(rate.RetryAfter)
	case errors.As(err, &generic):
		log.Printf("transcodely error: %s (req %s)", generic.ErrorCode(), generic.RequestID())
	default:
		log.Fatal(err)
	}
}
