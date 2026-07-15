# Transcodely Go SDK

`github.com/transcodely/transcodely-go` — the official Go SDK, generated from the
`api` repo's public protos (`buf generate` → `internal/gen`, re-exported via
`types.go` so users never import `internal/gen`). Wire format is snake_case JSON +
simplified lowercase enums (a port of the `api` repo's `internal/connect/codec.go`).
Upstream (`../api`) is authoritative for wire/enum/error behavior — see
[`api/docs/reference/api-conventions.md`](https://github.com/transcodely/api/blob/master/docs/reference/api-conventions.md).

## Resyncing protos

```bash
./scripts/sync-protos.sh && buf generate
```

`scripts/sync-protos.sh` copies every proto from `../api/proto/transcodely/v1`
except the three internal-service ones (`admin.proto`, `staff.proto`,
`worker.proto`) — those aren't part of the public SDK surface. There is no CI
check or BSR wiring that does this automatically; a proto change upstream is
not done until it's been synced here by hand (org-wide convention, see the
`.github` repo's `CLAUDE.md`).

- A new field on an already-exported message (e.g. a new `optional` field on
  `OutputSpec`) surfaces automatically — `types.go` re-exports whole structs
  as type aliases (`OutputSpec = v1.OutputSpec`), not field-by-field.
- A brand-new message type needs an explicit alias added to `types.go`.
- After resyncing, run `go build ./... && go test ./...` and skim the new/changed
  proto for anything requiring a hand-written facade addition (a new resource
  namespace, a new typed error code, etc.) — `buf generate` only regenerates
  `internal/gen`, it doesn't wire up ergonomic helpers.

## Docs are the contract (drift)

The public docs (`transcodely/web` → `src/routes/(docs)/docs/**`, especially
`getting-started/sdks/go` and the per-resource SDK method maps such as the one in
`api-reference/webhooks`) document this SDK's exact public surface. Rules:

- Any public-surface change (methods, params structs, enums, webhook event types,
  error accessors) must be mirrored in those web docs pages **in the same release
  window**. Web's mechanical drift gate validates proto-level facts but does NOT
  parse `go` code fences — whoever changes this SDK owns the docs snippets.
- Before renaming/removing anything public, grep the web repo's docs for usages
  (` ```go ` fences calling `client.<Resource>.<Method>`); docs may also reference
  capabilities that shipped here first.
- Vendored proto comments flow into generated code and docs — when resyncing, take
  the api repo's comments verbatim (they are maintained as public documentation there).

## Release automation

[Release Please](https://github.com/googleapis/release-please) (`.github/workflows/release.yml`,
`release-please-config.json`) tracks `version.go`'s `x-release-please-version`
marker as an `extra-files` target and keeps `CHANGELOG.md` + the Git tag in
sync on every merge to `master` — driven by Conventional Commits (`feat:` /
`fix:` bump the version; `chore:`, `refactor:`, `test:`, etc. are
changelog-hidden). Don't hand-edit `Version` in `version.go` or the entries at
the top of `CHANGELOG.md`; let release-please's PR do it. If they ever drift
from the latest tag, suspect the `release.yml` workflow run or the
`extra-files` wiring in `release-please-config.json` before hand-fixing the
files.

## Design

Mirrors Stripe's Go conventions — functional options, resource namespaces off
the root client (`client.Jobs`, `client.Videos`, ...), typed errors via
`errors.As`, `*Iter[T]` auto-pagination, auto-idempotency on `Create` calls.
See `README.md` for usage examples.

## Commands

```bash
go build ./...   # Compile
go test ./...    # Run tests (race detector: go test -race ./...)
go vet ./...     # Static analysis
gofmt -l .        # Formatting check (CI fails on any output)
```

## Commit Messages

Conventional Commits, enforced by `.github/workflows/pr-title.yml`:
`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`, `ci:`.
