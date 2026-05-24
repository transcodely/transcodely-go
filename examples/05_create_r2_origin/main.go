// Create a Cloudflare R2 origin and print its ID.
//
// R2 exposes an S3-compatible API, so it reuses S3Credentials. Identify the
// bucket either by account ID (shown below, with an optional data-residency
// jurisdiction) or by an explicit endpoint URL (see the commented variant).
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

	origin, err := client.Origins.Create(context.Background(), &transcodely.OriginCreateParams{
		Name:        "R2 source",
		Permissions: []transcodely.OriginPermission{transcodely.OriginPermissionRead},
		R2: &transcodely.R2OriginConfig{
			Bucket:       "my-r2-bucket",
			AccountId:    os.Getenv("R2_ACCOUNT_ID"),   // 32 lowercase hex chars
			Jurisdiction: transcodely.R2JurisdictionEU, // optional: Default, EU, FedRAMP
			Credentials: &transcodely.S3Credentials{
				AccessKeyId:     os.Getenv("R2_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
			},
		},

		// Endpoint escape hatch — set this *instead of* AccountId/Jurisdiction.
		// The server requires exactly one of the two location forms.
		//
		// endpoint := "https://<account>.r2.cloudflarestorage.com"
		// R2: &transcodely.R2OriginConfig{
		//     Bucket:   "my-r2-bucket",
		//     Endpoint: &endpoint,
		//     Credentials: &transcodely.S3Credentials{
		//         AccessKeyId:     os.Getenv("R2_ACCESS_KEY_ID"),
		//         SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		//     },
		// },
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created %s (provider %s) in status %s",
		origin.GetId(), origin.GetProvider(), origin.GetStatus())
}
