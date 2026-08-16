# Changelog

All notable changes to the Transcodely Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Breaking changes are allowed on minor bumps until v1.0.0.

## [0.3.8](https://github.com/transcodely/transcodely-go/compare/v0.3.7...v0.3.8) (2026-08-16)


### Features

* **billing:** budgets, outstanding balance and mid-cycle settlement ([#46](https://github.com/transcodely/transcodely-go/issues/46)) ([c65d839](https://github.com/transcodely/transcodely-go/commit/c65d83902d7ed59857ef22432fdcccf2aff5af53))
* sync protos from api 5.5.0 — player config + caption styling ([#43](https://github.com/transcodely/transcodely-go/issues/43)) ([3501637](https://github.com/transcodely/transcodely-go/commit/35016375cc29799ab2477c58a8fd97e1c14886c8))

## [0.3.7](https://github.com/transcodely/transcodely-go/compare/v0.3.6...v0.3.7) (2026-08-11)


### Features

* billing profile + portal sessions from api 4.13.0 ([#39](https://github.com/transcodely/transcodely-go/issues/39)) ([ef3a511](https://github.com/transcodely/transcodely-go/commit/ef3a511f00566dbd7a9a41cba29c0def6f540ccb))
* billing standing from api 4.15.0 ([#41](https://github.com/transcodely/transcodely-go/issues/41)) ([27f5b7f](https://github.com/transcodely/transcodely-go/commit/27f5b7f4ae5bb297f84f8ca22dd5893cf626c4a5))

## [0.3.6](https://github.com/transcodely/transcodely-go/compare/v0.3.5...v0.3.6) (2026-08-08)


### Features

* job cost breakdown + video_id from api 4.9.0 ([#36](https://github.com/transcodely/transcodely-go/issues/36)) ([97ea8e9](https://github.com/transcodely/transcodely-go/commit/97ea8e99fc310b27c760529fdccd5c512f0d9d72))

## [0.3.5](https://github.com/transcodely/transcodely-go/compare/v0.3.4...v0.3.5) (2026-08-02)


### Features

* **client:** auto_captions on video uploads + captions_cost on Video ([bd26b5d](https://github.com/transcodely/transcodely-go/commit/bd26b5d855c1eb137246d72a0e3c7cb76e19014f))

## [0.3.4](https://github.com/transcodely/transcodely-go/compare/v0.3.3...v0.3.4) (2026-08-02)


### Features

* **client:** animated preview fps up to 30 + mp4-only format ([517e71a](https://github.com/transcodely/transcodely-go/commit/517e71a713250ab55c872ef2b60221e96a752633))

## [0.3.3](https://github.com/transcodely/transcodely-go/compare/v0.3.2...v0.3.3) (2026-08-02)


### Features

* **client:** surface signed CDN asset URLs on job reads ([bab1277](https://github.com/transcodely/transcodely-go/commit/bab1277b30fb8faac6e6290943ffc13204d0e1c4))

## [0.3.2](https://github.com/transcodely/transcodely-go/compare/v0.3.1...v0.3.2) (2026-07-29)


### Features

* expose daily usage buckets and webhook endpoint health + event object-id filtering ([#21](https://github.com/transcodely/transcodely-go/issues/21)) ([921ec3e](https://github.com/transcodely/transcodely-go/commit/921ec3e2eb4c72451d9a5cd650429d4d9785a1f2))
* post-launch feature program integration (transcodely-go) ([#29](https://github.com/transcodely/transcodely-go/issues/29)) ([18d1ca4](https://github.com/transcodely/transcodely-go/commit/18d1ca4b922f929afb6513017f0fd15c19d4ebd4))
* sync protos with api 3.10 — billing, captions text tracks, app suspension ([#30](https://github.com/transcodely/transcodely-go/issues/30)) ([430bcdc](https://github.com/transcodely/transcodely-go/commit/430bcdcfc7f46274f0204bf7fa581be9e4faf76a))


### Bug Fixes

* **examples:** set managed on the create-job example ([#19](https://github.com/transcodely/transcodely-go/issues/19)) ([b1d680d](https://github.com/transcodely/transcodely-go/commit/b1d680d385198a82fd3583623e166dff3235c9c7))


### Documentation

* note content-aware encoding is rejected at create ([#25](https://github.com/transcodely/transcodely-go/issues/25)) ([29b3e36](https://github.com/transcodely/transcodely-go/commit/29b3e3636e0bd4f9950e651f1982dec6ec44226c))
* set managed on the README quickstart create-job ([#27](https://github.com/transcodely/transcodely-go/issues/27)) ([eff2650](https://github.com/transcodely/transcodely-go/commit/eff2650785226db0330fe7b6d63b174097fd360b))

## [0.3.1](https://github.com/transcodely/transcodely-go/compare/v0.3.0...v0.3.1) (2026-07-15)


### Features

* **webhooks:** wire WebhookService into the public client ([#17](https://github.com/transcodely/transcodely-go/issues/17)) ([0d20c68](https://github.com/transcodely/transcodely-go/commit/0d20c68f3dae77db96bee5fae0ad85c60979a87b))

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
