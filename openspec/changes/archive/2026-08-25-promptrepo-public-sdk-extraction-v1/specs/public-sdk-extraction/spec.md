## ADDED Requirements

### Requirement: Public SDK source of truth preserves private v0.1.0 compatibility
The system SHALL expose `github.com/yeisme/promptrepo` as the canonical public
Go module. Its initial public implementation SHALL be sourced from private tag
`sdk/go/promptrepo/v0.1.0` at peeled commit
`e59e5a060b98125e55197cb9f5ed9179cdacc46a`; apart from Go module/import paths
and publication/foundation files, it SHALL preserve SDK behavior, exported API,
tests, conformance helpers, state paths/schema, digests, JSON fields, and
stable error codes. It MUST NOT depend on Template Registry server/internal
packages.

#### Scenario: Public v0.1.0 fixture is exercised without private access
- **WHEN** public CI runs the compatibility test
- **THEN** it SHALL exercise the public SDK using a local v0.1.0-shaped fixture
- **AND** it SHALL validate the frozen ref, digest, receipt, state, and error
  outcomes without private network access.

#### Scenario: Consumer imports public module
- **WHEN** a Go 1.24 consumer imports the root, engine, and source packages
- **THEN** it SHALL resolve imports under `github.com/yeisme/promptrepo`
- **AND** it SHALL not require Template Registry server/internal packages.

### Requirement: State compatibility remains stable through rehome
The public SDK SHALL retain `promptrepo.state.v0.1`,
`promptrepo.catalog.v0.1`, and `promptrepo.stage_receipt.v0.1`, along with
their OS user config/cache defaults and fail-closed future-schema behavior.

#### Scenario: Existing state is read by the public SDK
- **WHEN** a consumer uses state written by the private v0.1.0 SDK
- **THEN** the public SDK SHALL retain the same `yeisme/promptrepo` state path
  and schema interpretation
- **AND** it SHALL return `STATE_SCHEMA_TOO_NEW` for an unsupported future
  schema rather than silently resetting it.
