# Changelog

All notable changes to this module are documented here.

## [Unreleased]

### Added

- Canonical `TemplateAddress` parsing and formatting with additive
  `kind,locale,role,path,selector,digest,snapshot` qualifiers.
- Caller-supplied `TemplateContract` and strict input definitions for string,
  integer, number, boolean, enum, defaults, examples, ranges, patterns,
  localized guidance, and sensitivity.
- Optional `Inspector`, `Validator`, `Renderer`, and `Previewer` capabilities;
  the existing `Client` interface remains unchanged.
- Optional `ContractResolver` plus built-in file, Git/GitHub, and S3 companion
  readers for Registry-authored template contracts, with explicit
  `snapshot_pinned` or `content_bound` consistency; the
  existing source `Adapter` method set remains unchanged.
- Provider-free in-memory rendering and preview metadata with rendered digest,
  byte/rune counts, and body-safe JSON/YAML serialization.
- Invalid UTF-8 is sanitized before Unicode search normalization, preserving
  the Go 1.24 compatibility floor while closing GO-2026-5970's input path.
- Contract license projections require canonical safe SPDX-style text, and
  permissions require canonical identifiers, so inspect/preview output cannot
  carry control text or URLs.

### Compatibility

- Planned release version: `v0.3.0`.
- Existing `Ref`, `Client`, catalog, state, receipt, source profile, and digest
  contracts are unchanged; no state migration is required.
- Selectors remain inspect-only and fail closed with `SELECTOR_UNSUPPORTED`
  during preview until a conformance-tested selector engine is available.

### Release gate

- Do not publish or advertise `v0.3.0` as installable until the maintainer has
  created the immutable public tag and anonymous
  `GOWORK=off go mod download github.com/yeisme/promptrepo@v0.3.0` succeeds.
- After publication, rerun the Sonora/Eikona/Scaena consumer canaries without a
  filesystem replace or private module credential.

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
