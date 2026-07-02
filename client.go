// Package transcodely is the official Go SDK for the Transcodely video
// transcoding API.
//
// Get an API key at https://transcodely.com and start transcoding:
//
//	import (
//	    "context"
//	    "log"
//	    "os"
//
//	    "github.com/transcodely/transcodely-go"
//	)
//
//	func main() {
//	    client, err := transcodely.New(os.Getenv("TRANSCODELY_API_KEY"))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    job, err := client.Jobs.Create(context.Background(), &transcodely.JobCreateParams{
//	        InputURL: "https://example.com/in.mp4",
//	        Outputs: []*transcodely.OutputSpec{
//	            {Type: transcodely.OutputFormatHLS, Video: []*transcodely.VideoVariant{
//	                {Codec: transcodely.VideoCodecH264, Resolution: transcodely.Resolution1080P},
//	            }},
//	        },
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    log.Printf("created %s in %s", job.GetId(), job.GetStatus())
//	}
//
// All resources hang off the root [Client]: Jobs, Videos, Presets, Origins,
// Apps, APIKeys, Organizations, Memberships, Users, Health, Events,
// WebhookEndpoints. Verify incoming webhook deliveries with the package-level
// [ConstructEvent] (also reachable as client.Webhooks.ConstructEvent).
//
// Errors are typed; switch on them with errors.As. See the [Error] interface
// for the common surface and [APIConnectionError], [APIError],
// [AuthenticationError], [PermissionError], [NotFoundError], [ConflictError],
// [RateLimitError], [InvalidRequestError], [PreconditionError] for the
// concrete classes.
package transcodely

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/transcodely/transcodely-go/internal/codec"
	"github.com/transcodely/transcodely-go/internal/gen/transcodely/v1/transcodelyv1connect"
)

// ErrMissingAPIKey is returned by New when no API key is supplied.
var ErrMissingAPIKey = errors.New("transcodely: api key is required")

// Client is the root Transcodely API client. Construct one with [New], then
// reach the resource you need via the typed namespace fields.
//
// A *Client is safe for concurrent use by multiple goroutines and is meant to
// live for the lifetime of your process. The underlying *http.Client is
// shared across every RPC and connection-pools transparently.
type Client struct {
	cfg *config

	Jobs             *Jobs
	Videos           *Videos
	Presets          *Presets
	Origins          *Origins
	Apps             *Apps
	APIKeys          *APIKeys
	Organizations    *Organizations
	Memberships      *Memberships
	Users            *Users
	Health           *Health
	Events           *Events
	WebhookEndpoints *WebhookEndpoints

	// Webhooks groups the stateless webhook helpers (ConstructEvent,
	// VerifySignature). It needs no client — the package-level functions are
	// equivalent — but is exposed here for discoverability.
	Webhooks Webhooks
}

// New constructs a Client. apiKey is required and should be a value like
// `ak_live_…` or `ak_test_…`. Pass any number of [Option]s to override
// defaults.
func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	cfg := defaultConfig()
	cfg.apiKey = apiKey
	for _, o := range opts {
		o(cfg)
	}

	jsonCodec := codec.NewProtoJSONCodec()
	unaryOpts := []connect.ClientOption{
		connect.WithProtoJSON(),
		connect.WithCodec(jsonCodec),
		connect.WithInterceptors(
			authInterceptor(cfg),
			idempotencyInterceptor(),
			retryInterceptor(cfg),
		),
	}
	streamOpts := []connect.ClientOption{
		connect.WithProtoJSON(),
		connect.WithCodec(jsonCodec),
		connect.WithInterceptors(streamingAuthInterceptor(cfg)),
	}

	c := &Client{cfg: cfg}

	c.Jobs = newJobs(transcodelyv1connect.NewJobServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...),
		transcodelyv1connect.NewJobServiceClient(cfg.httpClient, cfg.baseURL, streamOpts...),
		cfg)
	c.Videos = newVideos(transcodelyv1connect.NewVideoServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...),
		transcodelyv1connect.NewVideoServiceClient(cfg.httpClient, cfg.baseURL, streamOpts...),
		cfg)
	c.Presets = newPresets(transcodelyv1connect.NewPresetServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.Origins = newOrigins(transcodelyv1connect.NewOriginServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.Apps = newApps(transcodelyv1connect.NewAppServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.APIKeys = newAPIKeys(transcodelyv1connect.NewAPIKeyServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.Organizations = newOrganizations(transcodelyv1connect.NewOrganizationServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.Memberships = newMemberships(transcodelyv1connect.NewMembershipServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.Users = newUsers(transcodelyv1connect.NewUserServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))
	c.Health = newHealth(transcodelyv1connect.NewHealthServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...))

	webhookSvc := transcodelyv1connect.NewWebhookServiceClient(cfg.httpClient, cfg.baseURL, unaryOpts...)
	c.Events = newEvents(webhookSvc)
	c.WebhookEndpoints = newWebhookEndpoints(webhookSvc)

	return c, nil
}
