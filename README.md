# Transcodely Go SDK

The official Go SDK for the [Transcodely](https://transcodely.com) video
transcoding API.

```bash
go get github.com/transcodely/transcodely-go
```

> Requires Go 1.23+.

## Quickstart

```go
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

    job, err := client.Jobs.Create(context.Background(), &transcodely.JobCreateParams{
        InputUrl: "https://download.samplelib.com/mp4/sample-30s.mp4",
        Outputs: []*transcodely.OutputSpec{{
            Type: transcodely.OutputFormatHLS,
            Video: []*transcodely.VideoVariant{
                {Codec: transcodely.VideoCodecH264, Resolution: transcodely.Resolution1080P},
                {Codec: transcodely.VideoCodecH264, Resolution: transcodely.Resolution720P},
            },
        }},
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("created %s in status %s", job.GetId(), job.GetStatus())
}
```

## Design

The SDK mirrors Stripe's Go conventions:

- **Functional options** on `transcodely.New` — `WithBaseURL`, `WithHTTPClient`, `WithMaxRetries`, `WithUserAgent`, `WithAPIVersion`, `WithAutoIdempotency`.
- **Resource namespaces** off the root client: `client.Jobs`, `client.Videos`, `client.Presets`, `client.Origins`, `client.Apps`, `client.APIKeys`, `client.Organizations`, `client.Memberships`, `client.Users`, `client.Health`, `client.Events`, `client.WebhookEndpoints`.
- **Typed errors** via `errors.As`. Concrete types implement the
  [`Error`](https://pkg.go.dev/github.com/transcodely/transcodely-go#Error) interface
  and expose `ErrorCode()` and `RequestID()`. Switch on `*NotFoundError`,
  `*RateLimitError`, `*InvalidRequestError`, etc.
- **Auto-pagination** via `*Iter[T]` — `Next()` / `Current()` / `Err()`.
- **Auto-idempotency** — `Create` mutations get a UUIDv4 `Idempotency-Key`
  unless you set one yourself.
- **Streaming watches** via `*Stream[T]` — heartbeats are filtered, network
  blips reconnect transparently.
- **Wire format** is the same custom snake_case + lowercase-enum JSON the
  TypeScript and Python SDKs use, ported verbatim from the api repo's
  `internal/connect/codec.go`.

## Watch a job

```go
stream := client.Jobs.Watch(ctx, job.GetId())
defer stream.Close()
for stream.Next() {
    event := stream.Current()
    j := event.GetJob()
    log.Printf("[%s] progress=%d%%", j.GetStatus(), j.GetProgress())
    if j.GetStatus() == transcodely.JobStatusCompleted ||
        j.GetStatus() == transcodely.JobStatusFailed ||
        j.GetStatus() == transcodely.JobStatusCanceled {
        break
    }
}
if err := stream.Err(); err != nil {
    log.Fatal(err)
}
```

## Iterate every job

```go
iter := client.Jobs.List(ctx, &transcodely.JobListParams{
    Pagination: &transcodely.PaginationRequest{Limit: 50},
})
for iter.Next() {
    job := iter.Current()
    log.Printf("%s %s", job.GetId(), job.GetStatus())
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

## Typed error handling

```go
job, err := client.Jobs.Get(ctx, "job_does_not_exist")
if err != nil {
    var notFound *transcodely.NotFoundError
    var invalid  *transcodely.InvalidRequestError
    var rate     *transcodely.RateLimitError
    switch {
    case errors.As(err, &notFound):
        log.Printf("not found, request id: %s", notFound.RequestID())
    case errors.As(err, &invalid):
        for _, v := range invalid.Errors() {
            log.Printf("%s: %s", v.Field, v.Description)
        }
    case errors.As(err, &rate):
        time.Sleep(rate.RetryAfter)
    default:
        log.Fatal(err)
    }
}
```

## Origins

An **origin** tells Transcodely where to read source media from. Create one with
`client.Origins.Create` — the provider is inferred from which config block you
set, so there is no `provider` field to pass.

### Amazon S3

```go
origin, err := client.Origins.Create(ctx, &transcodely.OriginCreateParams{
    Name:        "Marketing assets",
    Permissions: []transcodely.OriginPermission{transcodely.OriginPermissionRead},
    S3: &transcodely.S3OriginConfig{
        Bucket: "my-bucket",
        Region: "us-east-1",
        Credentials: &transcodely.S3Credentials{
            AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
            SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
        },
    },
})
```

### Google Cloud Storage

```go
origin, err := client.Origins.Create(ctx, &transcodely.OriginCreateParams{
    Name: "GCS source",
    Gcs: &transcodely.GcsOriginConfig{
        Bucket: "my-gcs-bucket",
        Credentials: &transcodely.GcsCredentials{
            ServiceAccountJson: os.Getenv("GCP_SERVICE_ACCOUNT_JSON"),
        },
    },
})
```

### Public HTTP

```go
origin, err := client.Origins.Create(ctx, &transcodely.OriginCreateParams{
    Name: "CDN mirror",
    Http: &transcodely.HttpOriginConfig{
        BaseUrl: "https://media.example.com",
    },
})
```

### Cloudflare R2

R2 issues S3-compatible access keys, so it reuses `S3Credentials`. Point at the
bucket either by **account ID** (Transcodely derives the endpoint, with an
optional data-residency `Jurisdiction`) or by an explicit **endpoint** URL:

```go
origin, err := client.Origins.Create(ctx, &transcodely.OriginCreateParams{
    Name:        "R2 source",
    Permissions: []transcodely.OriginPermission{transcodely.OriginPermissionRead},
    R2: &transcodely.R2OriginConfig{
        Bucket:       "my-r2-bucket",
        AccountId:    os.Getenv("R2_ACCOUNT_ID"),    // 32 lowercase hex chars
        Jurisdiction: transcodely.R2JurisdictionEU,  // optional: Default, EU, FedRAMP
        Credentials: &transcodely.S3Credentials{
            AccessKeyId:     os.Getenv("R2_ACCESS_KEY_ID"),
            SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
        },
    },

    // Endpoint escape hatch — set this *instead of* AccountId/Jurisdiction:
    //
    // endpoint := "https://<account>.r2.cloudflarestorage.com"
    // R2: &transcodely.R2OriginConfig{
    //     Bucket:      "my-r2-bucket",
    //     Endpoint:    &endpoint,
    //     Credentials: &transcodely.S3Credentials{ /* ... */ },
    // },
})
```

> Provide **exactly one** of `AccountId` or `Endpoint`, and only set
> `Jurisdiction` alongside `AccountId`. Both rules are enforced server-side and
> surface as an `*transcodely.InvalidRequestError`.

## Webhooks

A **webhook endpoint** is an HTTPS URL that receives signed POST requests when
events happen in your app — a job succeeds, a video is uploaded, and so on.

### Verify and handle a delivery

`transcodely.ConstructEvent` verifies the signature, enforces the timestamp
tolerance (5 minutes, Stripe parity), and decodes the body into a typed
`*transcodely.Event`. It returns a `*WebhookSignatureError`,
`*WebhookTimestampError`, or `*WebhookPayloadError` on failure — all reachable
together via the `WebhookError` interface. It needs no client, so a receiver
process never has to hold an API key.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    secret := os.Getenv("WEBHOOK_SIGNING_SECRET") // whsec_... from Create

    event, err := transcodely.ConstructEvent(body, r.Header.Get(transcodely.SignatureHeader), secret)
    if err != nil {
        http.Error(w, "invalid signature", http.StatusBadRequest) // non-2xx → retried
        return
    }

    switch event.Type {
    case transcodely.EventTypeJobSucceeded:
        job, _ := event.Job() // typed accessor: *transcodely.Job
        log.Printf("job %s %s", job.GetId(), job.GetStatus())
    case transcodely.EventTypeVideoUploaded:
        vid, _ := event.Video()
        log.Printf("video %s uploaded", vid.GetId())
    default:
        log.Printf("unhandled %s (%s)", event.Type, event.ID)
    }
    w.WriteHeader(http.StatusOK) // ack fast, then work asynchronously
}
```

`event.Data` holds the decoded resource — `*Job`, `*JobOutput`, `*Video`, or
`*App` — pulled out with the matching accessor (`event.Job()`,
`event.JobOutput()`, `event.Video()`, `event.App()`). For an event type a
newer API version added that this SDK doesn't yet know, `event.RawData()`
returns the raw JSON. Deduplicate on `event.ID`; a retried delivery carries the
same `evt_` id.

During a secret rotation, accept both keys for the 24h overlap window:

```go
event, err := transcodely.ConstructEventWithSecrets(
    body, r.Header.Get(transcodely.SignatureHeader),
    []string{newSecret, previousSecret})
```

### Manage endpoints

```go
endpoint, err := client.WebhookEndpoints.Create(ctx, &transcodely.WebhookEndpointCreateParams{
    AppId:         "app_...",
    Url:           "https://example.com/webhooks",
    EnabledEvents: []string{"job.succeeded", "video.uploaded"}, // or []string{"*"}
})
// endpoint.GetSecret() is populated ONLY here and on RotateSecret — store it now.

iter := client.WebhookEndpoints.List(ctx, &transcodely.WebhookEndpointListParams{AppId: "app_..."})
for iter.Next() {
    ep := iter.Current()
    log.Printf("%s %s", ep.GetId(), ep.GetStatus())
}

rotated, _ := client.WebhookEndpoints.RotateSecret(ctx, endpoint.GetId())
_ = rotated.GetSecret() // the new secret, shown once
```

### Query and replay events

Every event the API hands back is the **same** `*transcodely.Event` shape
`ConstructEvent` produces, so a handler tested against a retrieved event
behaves identically to one driven by a live delivery.

```go
event, _ := client.Events.Retrieve(ctx, "evt_...")

iter := client.Events.List(ctx, &transcodely.EventListParams{
    AppId:      "app_...",
    Pagination: &transcodely.PaginationRequest{Limit: 50},
})
for iter.Next() { /* iter.Current() is a *transcodely.Event */ }

// Re-queue delivery (all subscribed endpoints, or a subset).
deliveries, _ := client.Events.Resend(ctx, "evt_...")
```

## Configuration

| Option | Default | Notes |
|---|---|---|
| `WithBaseURL(url)` | `https://api.transcodely.com` | Override for staging or self-hosted |
| `WithHTTPClient(c)` | `&http.Client{Timeout: 60s}` | Inject a custom transport |
| `WithMaxRetries(n)` | `2` | Network errors, 5xx, 429, 503 are retried with jitter |
| `WithUserAgent(ua)` | — | Appended to the default `transcodely-go/<version>` |
| `WithAPIVersion(v)` | calendar version baked at SDK build time | Sent as `Transcodely-Version` header |
| `WithAutoIdempotency(b)` | `true` | Auto-generate UUIDv4 `Idempotency-Key` on `Create` mutations |

## Environment variables

The SDK does not read any environment variables itself. Pass `os.Getenv(...)`
into `New(...)` so your callers stay in control.

## Examples

See [`examples/`](../../examples/go) for ready-to-run programs:

- `01_create_job` — submit a job
- `02_watch_job` — stream a job to completion
- `03_pagination` — iterate every job
- `04_error_handling` — typed error matching
- `05_create_r2_origin` — register a Cloudflare R2 origin
- `06_verify_webhook` — verify and handle webhook deliveries (net/http)

## Generated code

`internal/gen/` holds the protobuf-generated message and Connect-RPC client
types. Treat it as opaque — the public API surface is everything in the root
package; lift types you need via the re-exports in [`types.go`](types.go).

## License

MIT — see [LICENSE](LICENSE).
