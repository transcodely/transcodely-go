package transcodely

// Version is the semantic version of the Transcodely Go SDK.
const Version = "0.3.0" // x-release-please-version

// APIVersion is the calendar version of the Transcodely API the SDK targets.
// Sent as the `Transcodely-Version` header on every request.
const APIVersion = "2026-05-03"

// DefaultBaseURL is the production API endpoint.
const DefaultBaseURL = "https://api.transcodely.com"

// userAgent is sent on every request as `User-Agent`.
const userAgent = "transcodely-go/" + Version
