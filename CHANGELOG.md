# Changelog

All notable changes to this module are documented here.

## [0.2.0] - 2026-08-23

### Changed

- Promoted the public SDK from `v0.2.0-rc.1` to stable without API, persisted
  state, schema, digest, or runtime behavior changes.
- Verified the release candidate in Template Registry and Sonora consumers;
  Sonora GitHub Actions run `32617428198` completed successfully without a
  private module credential.
- Rebuilt the official prompt catalog with the public SDK while preserving
  catalog digest
  `sha256:d252148bb4e4f6be95f99848dd49bad43b6414f4b7ccd520fe71b3616ad37b2f`.

### Migration

- Consumers should pin `github.com/yeisme/promptrepo v0.2.0` and remove any
  private module credential setup used only for the legacy module path.
- The private `sdk/go/promptrepo/v0.1.0` tag remains immutable as a rollback
  reference; new development and releases occur only in this repository.

## [0.2.0-rc.1] - 2026-08-23

### Added

- Canonical public module `github.com/yeisme/promptrepo` for the prompt
  repository contract, embedded engine, source adapters, tests, and
  conformance helpers.
- Public release foundations: MIT license, security policy, contribution
  guidance, release/architecture documentation, OpenSpec change, compatibility
  test, and Go 1.24 CI.

### Changed

- Rehomed the private SDK tagged `sdk/go/promptrepo/v0.1.0` at
  `e59e5a060b98125e55197cb9f5ed9179cdacc46a` without behavioral or state
  contract changes. Consumer imports now use `github.com/yeisme/promptrepo`.

### Migration

- Replace the old module/import prefix
  `github.com/yeisme/backend-server-template-registry/sdk/go/promptrepo` with
  `github.com/yeisme/promptrepo`.
- Private v0.1.0 remains immutable and becomes deprecated after public cutover.
  It is not modified, republished, or retagged.
