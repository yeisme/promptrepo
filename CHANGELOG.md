# Changelog

All notable changes to this module are documented here.

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
