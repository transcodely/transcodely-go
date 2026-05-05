# Changelog

All notable changes to the Transcodely Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Breaking changes are allowed on minor bumps until v1.0.0.

## [0.1.1](https://github.com/transcodely/transcodely-go/compare/v0.1.0...v0.1.1) (2026-05-05)


### Features

* initial v0.1.0 alpha release ([39454e1](https://github.com/transcodely/transcodely-go/commit/39454e160ad9e5cbd55d61c82d05e45ea09a3eb1))

## [v0.1.0] — Alpha

Initial public alpha. Covers 100% of the public RPC surface (56 RPCs across 10 services). Stripe-style facade: lazy resource namespaces, auto-pagination via `*Iter[T]`, auto-idempotency on `Create` mutations, typed error hierarchy (1 base + 8 concrete) usable with `errors.As`, Watch streams with auto-reconnect, calendar-versioned API.
