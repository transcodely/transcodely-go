// Create an S3-compatible origin (Hetzner Object Storage) and print its ID.
//
// Any S3-compatible object store — Hetzner Object Storage, Wasabi, DigitalOcean
// Spaces, MinIO, Backblaze B2 — uses the standard `s3` provider plus an explicit
// endpoint. Transcodely switches to path-style addressing automatically; the
// region is still required because the AWS SDK uses it to sign requests.
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

	// For Hetzner the region is the location code (fsn1, nbg1, hel1) and the
	// endpoint is https://<region>.your-objectstorage.com. Wasabi, DO Spaces,
	// Backblaze B2 and MinIO follow the same shape with their own endpoint.
	endpoint := "https://fsn1.your-objectstorage.com"

	origin, err := client.Origins.Create(context.Background(), &transcodely.OriginCreateParams{
		Name:        "Hetzner source",
		Permissions: []transcodely.OriginPermission{transcodely.OriginPermissionRead},
		S3: &transcodely.S3OriginConfig{
			Bucket:   "my-video-bucket",
			Region:   "fsn1",
			Endpoint: &endpoint,
			Credentials: &transcodely.S3Credentials{
				AccessKeyId:     os.Getenv("HETZNER_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("HETZNER_SECRET_ACCESS_KEY"),
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created %s (provider %s) in status %s",
		origin.GetId(), origin.GetProvider(), origin.GetStatus())
}
