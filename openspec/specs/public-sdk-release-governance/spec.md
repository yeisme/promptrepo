# public-sdk-release-governance Specification

## Purpose
TBD - created by archiving change promptrepo-public-sdk-extraction-v1. Update Purpose after archive.
## Requirements
### Requirement: Public releases use reproducible RC and stable gates
Before a maintainer creates `v0.2.0-rc.1` or `v0.2.0`, the repository SHALL
pass Go 1.24 module verification, tests, vet, CGO-disabled tests/builds, strict
OpenSpec validation, normalized source diff review, and secret scan. Public CI
MUST NOT require private network access or credentials.

#### Scenario: CI validates the public SDK
- **WHEN** a pull request or main-branch push runs CI
- **THEN** CI SHALL run `go mod verify`, `go test ./...`, `go vet ./...`, and
  meaningful `CGO_ENABLED=0` tests/builds using Go 1.24
- **AND** it SHALL use read-only repository permissions and no secrets.

### Requirement: Tags and private baseline remain immutable
The public release line SHALL use immutable maintainer-created tags in the
order `v0.2.0-rc.1` then `v0.2.0`. The private
`sdk/go/promptrepo/v0.1.0` baseline SHALL not be modified, republished, or
retagged and SHALL be deprecated only after public cutover.

#### Scenario: A release candidate fails a gate
- **WHEN** an RC fails a compatibility or release gate
- **THEN** maintainers SHALL not retag the failed version
- **AND** they SHALL fix forward in a new version or revert consumer migration.

### Requirement: Consumer migration has a rollback path
Consumers SHALL migrate by replacing only the old module/import prefix with
`github.com/yeisme/promptrepo`; current compatible state/config/cache SHALL
remain usable. During the deprecation window, consumers MUST be able to roll
back by restoring the immutable private v0.1.0 dependency.

#### Scenario: Consumer reverts public RC adoption
- **WHEN** a consumer canary finds a public RC regression
- **THEN** the consumer SHALL be able to restore the private v0.1.0 dependency
- **AND** existing promptrepo state SHALL remain intact.

