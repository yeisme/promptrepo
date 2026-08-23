## Why

Domain consumers need the prompt repository SDK without importing a private
Template Registry server module. The existing private v0.1.0 SDK is complete
but cannot be the public canonical package; extracting it now gives consumers a
stable Go 1.24 module while preserving its state and behavior.

## What Changes

- Create public canonical module `github.com/yeisme/promptrepo` by rehoming
  the exact private `sdk/go/promptrepo/v0.1.0` source.
- Preserve the exported API, source behavior, state paths/schema, digests,
  errors, tests, and conformance helpers; alter only module/import paths and
  publication foundations.
- Add provenance, compatibility, security, CI, RC/stable, tag, migration, and
  rollback records for the public release line.
- Deprecate the private baseline after public cutover without changing or
  retagging it.

## Capabilities

### New Capabilities

- `public-sdk-extraction`: public source-of-truth ownership and compatibility
  contract for the extracted Go SDK.
- `public-sdk-release-governance`: public release gates, tag policy, consumer
  migration, rollback, and secret-scan requirements.

### Modified Capabilities

- None.

## Impact

Affected code is this module's root package, `engine`, `source`, conformance
helpers, tests, Go module path, public documentation, CI, and local OpenSpec.
There is no Template Registry server/internal dependency and no private network
requirement in public CI. The private tag
`sdk/go/promptrepo/v0.1.0` at
`e59e5a060b98125e55197cb9f5ed9179cdacc46a` is the immutable source baseline.
