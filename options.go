package transcodely

import (
	"net/http"
	"time"
)

// Option configures a Client. Options are applied in order; later options
// override earlier ones.
type Option func(*config)

type config struct {
	apiKey      string
	orgID       string
	baseURL     string
	httpClient  *http.Client
	maxRetries  int
	userAgent   string
	apiVersion  string
	idempotency bool
}

func defaultConfig() *config {
	return &config{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		maxRetries: 2,
		userAgent:  userAgent,
		apiVersion: APIVersion,
		// Auto-injection of an Idempotency-Key on POST mutations is on by default.
		// Disable with WithAutoIdempotency(false) if your server side does not
		// (yet) understand the header.
		idempotency: true,
	}
}

// WithBaseURL overrides the API base URL. Defaults to DefaultBaseURL.
// Useful for staging, local docker-compose, or self-hosted deployments.
func WithBaseURL(url string) Option {
	return func(c *config) { c.baseURL = url }
}

// WithOrganization sets the organization every request is made against, sent
// as the `X-Organization-ID` header.
//
// API-key clients never need this: a key is already scoped to one app, and
// therefore to one organization. It is required for the org-scoped surfaces a
// dashboard session token reaches — [Client.Billing] and [APIKeys.Create]
// among them — which answer 400 invalid_argument without it.
func WithOrganization(orgID string) Option {
	return func(c *config) { c.orgID = orgID }
}

// WithHTTPClient supplies a custom *http.Client. Useful for injecting a
// shared transport, custom timeouts, proxy configuration, or test stubs.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithMaxRetries sets how many times the SDK retries a transient failure
// (network errors, 5xx, 429, 503) before giving up. Defaults to 2.
// Set to 0 to disable retries.
func WithMaxRetries(n int) Option {
	return func(c *config) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

// WithUserAgent appends a token to the default user-agent string.
// Use this to identify your application: WithUserAgent("acme-encoder/1.4.2").
func WithUserAgent(ua string) Option {
	return func(c *config) {
		if ua != "" {
			c.userAgent = userAgent + " " + ua
		}
	}
}

// WithAPIVersion overrides the calendar API version sent in the
// Transcodely-Version header. You should rarely need this — the SDK pins a
// version known to be compatible with its types.
func WithAPIVersion(v string) Option {
	return func(c *config) {
		if v != "" {
			c.apiVersion = v
		}
	}
}

// WithAutoIdempotency toggles automatic Idempotency-Key generation on POST
// mutations. Defaults to true.
func WithAutoIdempotency(enabled bool) Option {
	return func(c *config) { c.idempotency = enabled }
}
