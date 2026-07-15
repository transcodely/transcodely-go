# Changelog

All notable changes to the Transcodely Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Breaking changes are allowed on minor bumps until v1.0.0.

## [0.3.0](https://github.com/transcodely/transcodely-go/compare/v0.2.0...v0.3.0) (2026-07-15)


### ⚠ BREAKING CHANGES

* App.webhook, CreateAppRequest.webhook, and UpdateAppRequest.webhook (WebhookConfig / CreateWebhookConfig / UpdateWebhookConfig) are removed. App-level webhook configuration is superseded by the WebhookService endpoints API (WebhookEndpoint).

### Features

* sync protos — explicit app scoping; remove legacy app webhook config ([#15](https://github.com/transcodely/transcodely-go/issues/15)) ([0e5527f](https://github.com/transcodely/transcodely-go/commit/0e5527f9b2fc042a27ee3c76dda3634b955f0c6b))

## [0.2.0](https://github.com/transcodely/transcodely-go/compare/v0.1.3...v0.2.0) (2026-07-12)


### ⚠ BREAKING CHANGES

* removed the API-key environment field and the APIKeyEnvironment enum from the generated types, and removed livemode from the webhook Event.

### Features

* proto resync — rotation metadata + measured output metrics ([#12](https://github.com/transcodely/transcodely-go/issues/12)) ([6d88c70](https://github.com/transcodely/transcodely-go/commit/6d88c7026d91b776bd7703e7785093f5516a9815))
* resync protos — API-key environment and webhook livemode removed ([#14](https://github.com/transcodely/transcodely-go/issues/14)) ([22e061e](https://github.com/transcodely/transcodely-go/commit/22e061e4153bc9a2cbbec91d22a6c92c72e49bba))


### Documentation

* add CLAUDE.md ([#10](https://github.com/transcodely/transcodely-go/issues/10)) ([7aafc5e](https://github.com/transcodely/transcodely-go/commit/7aafc5ec2e15fa64443c0f7fcd9b4d3c3b607ed2))

## [0.1.3](https://github.com/transcodely/transcodely-go/compare/v0.1.2...v0.1.3) (2026-07-07)


### Documentation

* **examples:** add S3-compatible (custom-endpoint) origin example ([#7](https://github.com/transcodely/transcodely-go/issues/7)) ([e5e23f2](https://github.com/transcodely/transcodely-go/commit/e5e23f2cf2157b845341a083ab1b23de3efe9485))

## [0.1.2](https://github.com/transcodely/transcodely-go/compare/v0.1.1...v0.1.2) (2026-07-02)


### Features

* **origins:** add Cloudflare R2 as a first-class provider ([#4](https://github.com/transcodely/transcodely-go/issues/4)) ([73eed4e](https://github.com/transcodely/transcodely-go/commit/73eed4e5047ebe4c0ca9e0620a5c7094e302c181))
* sync proto — thumbnail path_template + accumulated drift ([#5](https://github.com/transcodely/transcodely-go/issues/5)) ([80594d2](https://github.com/transcodely/transcodely-go/commit/80594d26f29fc341bd8266270e86cbc3771da2dc))

## [0.1.1](https://github.com/transcodely/transcodely-go/compare/v0.1.0...v0.1.1) (2026-05-05)


### Features

* initial v0.1.0 alpha release ([39454e1](https://github.com/transcodely/transcodely-go/commit/39454e160ad9e5cbd55d61c82d05e45ea09a3eb1))

## [v0.1.0] — Alpha

Initial public alpha. Covers 100% of the public RPC surface (56 RPCs across 10 services). Stripe-style facade: lazy resource namespaces, auto-pagination via `*Iter[T]`, auto-idempotency on `Create` mutations, typed error hierarchy (1 base + 8 concrete) usable with `errors.As`, Watch streams with auto-reconnect, calendar-versioned API.
