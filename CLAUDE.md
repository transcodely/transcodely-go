# Transcodely Go SDK

`github.com/transcodely/transcodely-go` — the official Go SDK, generated from the
api repo's public protos (`buf generate` → `internal/gen`, re-exported via
`types.go` so users never import `internal/gen`). Wire format is snake_case JSON +
simplified lowercase enums (a port of the api repo's `internal/connect/codec.go`).
Upstream (`../api`) is authoritative for wire/enum/error behavior.

- Regenerate from proto: `./scripts/sync-protos.sh && buf generate`
- Build / test: `go build ./... && go test ./...`

---

## TODO: surface `disable_audio` + single-variant streaming presets

Upstream API changes this SDK still needs to expose (api PR #119, worker PR #57).

### What changed in the API
1. **`disable_audio` (video-only output).** New field:
   - `OutputSpec.disable_audio` — `optional bool`, override semantics: when unset
     it inherits the referenced preset's value; `true` drops audio for that output.
   - `Preset.disable_audio` (`bool`) plus `CreatePresetRequest` / `UpdatePresetRequest`.
   - Wire name is snake_case `disable_audio`. The server rejects `disable_audio: true`
     combined with explicit `audio[]` tracks (`parameter_invalid`).
   - Pricing is unchanged (no audio cost component).
2. **Graceful no-audio sources.** A source video with no audio track no longer fails
   packaging — it now produces a valid video-only output automatically (worker PR #57).
   No SDK code change; worth a line in the README / audio docs.
3. **Single-variant streaming presets.** The API dropped the "minimum 2 ABR variants"
   rule, so a single-variant HLS/DASH/CMAF preset is now valid (e.g. one vertical 720p
   HLS stream). No proto change — just relax any client-side mirror and update examples.

### Work items
- [ ] `./scripts/sync-protos.sh && buf generate` — `OutputSpec.DisableAudio *bool` and the
      preset fields then appear in `internal/gen` automatically. Set via
      `&transcodely.OutputSpec{DisableAudio: proto.Bool(true)}`.
- [ ] Add an `examples/` demo creating a video-only output (and/or a single-variant
      streaming preset).
- [ ] Add a round-trip test asserting it serializes as `"disable_audio": true`
      (follow the existing codec / enum-coverage tests).
- [ ] Update `README.md` + `CHANGELOG.md` (release-please) when it lands.
- [ ] Blocked on api PR #119 (proto) merging first.

Refs: transcodely/api#119, transcodely/worker#57.
