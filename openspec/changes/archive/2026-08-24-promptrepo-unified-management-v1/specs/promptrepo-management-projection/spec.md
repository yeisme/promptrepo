## ADDED Requirements

### Requirement: Unified operations MUST have stable identity
Management projection MUST use stable English `operation_id` values while preserving a separate existing command identity.

#### Scenario: Two consumers synchronize repositories
- **WHEN** two domain CLIs render repository sync
- **THEN** both use `promptrepo.repository.sync`
- **AND** each may retain its own command path

### Requirement: Projection MUST be renderer-neutral and safe
Human, agent, JSON, YAML and events MUST derive from the same management projection. The projection MUST NOT contain prompt body, rendered body, supplied values, credentials, Authorization, private source URL/path, provider payload, hidden system prompt, private tool arguments or full chain-of-thought.

#### Scenario: Projection is serialized
- **WHEN** the same projection is encoded as JSON and YAML
- **THEN** operation, readiness, reason codes, digests and next actions are equivalent
- **AND** sensitive sentinel values are absent
