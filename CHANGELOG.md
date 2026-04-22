# Changelog

All notable changes to this project will be documented in this file.

## v0.4.0 - 2026-04-22

### Added
- Added a new `Client` API for reusable synthesis configuration.
- Added package-level helper functions for one-off use cases:
  - `Bytes`
  - `BytesSSML`
  - `Save`
  - `SaveSSML`
  - `WriteTo`
  - `WriteSSMLTo`
  - `Stream`
  - `StreamSSML`
- Added first-class `Request` support with symmetric `Text(...)` and `SSML(...)` builders.
- Added structured batch APIs:
  - `Batch`
  - `SaveBatch`
  - `WriteZIP`
- Added voice filtering helpers:
  - `FilterVoices`
  - `FindVoice`
- Added runnable demo under `cmd/demo`.
- Added example coverage for streaming, SSML, request-based usage, batch output, ZIP output, and HTTP streaming.

### Changed
- Reworked the public API around direct synthesis workflows instead of the old task-queue-first model.
- Improved error propagation in the internal streaming path so websocket and writer failures are surfaced to callers.
- Added context-aware streaming support in the internal communication layer.
- Updated README with new API usage, streaming examples, runnable demo commands, and migration guidance.
- Marked the legacy `Speech` API as deprecated while keeping it as a compatibility wrapper.

### Fixed
- Fixed missing `go.sum` state so the module builds and tests cleanly.
- Fixed potential voice parsing panics when deriving `VoiceLangRegion`.
- Fixed file-save behavior to avoid leaving broken output files behind on synthesis failure.

### Verification
- `go test ./...`
- `go vet ./...`
