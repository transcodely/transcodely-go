// Verify and handle incoming webhook deliveries with a net/http receiver.
//
// Transcodely signs every delivery. transcodely.ConstructEvent verifies the
// signature, enforces the timestamp tolerance, and decodes the flat envelope
// into a typed *transcodely.Event — the same Event type client.Events.Retrieve
// returns, so this handler works identically against live and replayed events.
//
// Set WEBHOOK_SIGNING_SECRET to the endpoint secret returned by
// client.WebhookEndpoints.Create (whsec_...).
package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/transcodely/transcodely-go"
)

func main() {
	secret := os.Getenv("WEBHOOK_SIGNING_SECRET")
	if secret == "" {
		log.Fatal("set WEBHOOK_SIGNING_SECRET")
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		event, err := transcodely.ConstructEvent(body, r.Header.Get(transcodely.SignatureHeader), secret)
		if err != nil {
			// Bad signature, stale timestamp, or malformed body. Reply non-2xx
			// so the platform marks this attempt failed and retries per its curve.
			log.Printf("invalid webhook: %v", err)
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}

		// During a secret rotation, accept both keys for the 24h overlap:
		//
		//	event, err := transcodely.ConstructEventWithSecrets(
		//	    body, r.Header.Get(transcodely.SignatureHeader),
		//	    []string{newSecret, previousSecret})

		// Acknowledge fast with a 2xx, then do heavy work asynchronously.
		// Deduplicate on event.ID — a retry carries the same evt_ id.
		switch event.Type {
		case transcodely.EventTypeJobSucceeded:
			if job, ok := event.Job(); ok {
				log.Printf("job %s succeeded (status %s)", job.GetId(), job.GetStatus())
			}
		case transcodely.EventTypeOutputReady:
			if out, ok := event.JobOutput(); ok {
				log.Printf("output %s ready at %s", out.GetId(), out.GetOutputUrl())
			}
		case transcodely.EventTypeVideoUploaded:
			if vid, ok := event.Video(); ok {
				log.Printf("video %s uploaded", vid.GetId())
			}
		default:
			log.Printf("unhandled event %s (%s)", event.Type, event.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	const addr = ":8080"
	log.Printf("listening on %s — POST deliveries to /webhooks", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
