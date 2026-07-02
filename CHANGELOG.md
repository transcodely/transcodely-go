# Changelog

All notable changes to the Transcodely Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Breaking changes are allowed on minor bumps until v1.0.0.

## [0.1.2](https://github.com/transcodely/transcodely-go/compare/v0.1.1...v0.1.2) (2026-07-02)


### Features

* **origins:** add Cloudflare R2 as a first-class provider ([#4](https://github.com/transcodely/transcodely-go/issues/4)) ([73eed4e](https://github.com/transcodely/transcodely-go/commit/73eed4e5047ebe4c0ca9e0620a5c7094e302c181))
* sync proto — thumbnail path_template + accumulated drift ([#5](https://github.com/transcodely/transcodely-go/issues/5)) ([80594d2](https://github.com/transcodely/transcodely-go/commit/80594d26f29fc341bd8266270e86cbc3771da2dc))

## [0.1.1](https://github.com/transcodely/transcodely-go/compare/v0.1.0...v0.1.1) (2026-05-05)


### Features

* initial v0.1.0 alpha release ([39454e1](https://github.com/transcodely/transcodely-go/commit/39454e160ad9e5cbd55d61c82d05e45ea09a3eb1))

## [v0.1.0] — Alpha

Initial public alpha. Covers 100% of the public RPC surface (56 RPCs across 10 services). Stripe-style facade: lazy resource namespaces, auto-pagination via `*Iter[T]`, auto-idempotency on `Create` mutations, typed error hierarchy (1 base + 8 concrete) usable with `errors.As`, Watch streams with auto-reconnect, calendar-versioned API.
