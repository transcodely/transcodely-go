# Changelog

All notable changes to the Transcodely Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Breaking changes are allowed on minor bumps until v1.0.0.

## [v0.1.0] — Alpha

Initial public alpha. Covers 100% of the public RPC surface (56 RPCs across 10 services). Stripe-style facade: lazy resource namespaces, auto-pagination via `*Iter[T]`, auto-idempotency on `Create` mutations, typed error hierarchy (1 base + 8 concrete) usable with `errors.As`, Watch streams with auto-reconnect, calendar-versioned API.
