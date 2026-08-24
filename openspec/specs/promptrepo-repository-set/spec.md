# promptrepo-repository-set Specification

## Purpose
TBD - created by archiving change promptrepo-unified-management-v1. Update Purpose after archive.
## Requirements
### Requirement: RepositorySet MUST compose four scopes without owning them
Promptrepo MUST distinguish `user`, `organization`, `project`, and `session`. Embedded mode MUST read canonical user profiles from existing local state while caller-supplied organization/project/session inputs remain ephemeral. It MUST NOT persist external bindings or change the existing state schema/path.

#### Scenario: Effective set is read twice
- **WHEN** a caller computes an effective set with organization and project bindings
- **THEN** both results are deterministic
- **AND** the local state bytes remain unchanged

### Requirement: Preference order MUST be deterministic and non-authorizing
Composition MUST order session exact selection, project pin, user preference, organization default and official fallback in that order. Duplicate repository IDs MUST keep the first ranked occurrence. Ranking MUST NOT imply admission.

#### Scenario: Exact repository is quarantined
- **WHEN** session exact selection ranks a quarantined repository first
- **THEN** RepositorySet retains the exact first position
- **AND** policy evaluation still blocks it

### Requirement: RepositorySet output MUST be safe
RepositorySet MUST expose only safe repository metadata, digests and health summary. It MUST NOT serialize raw scope refs, source URLs, credential refs, local paths or template bodies.

#### Scenario: Private organization binding
- **WHEN** an organization scope ref contains a private identifier
- **THEN** output contains only `scope_ref_digest`
- **AND** the raw identifier is absent from JSON and YAML projections

