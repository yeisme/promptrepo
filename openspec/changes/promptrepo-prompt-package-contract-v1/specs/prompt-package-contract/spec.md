## ADDED Requirements

### Requirement: Stable recipe graph
The SDK SHALL validate bounded recipe DAGs and produce a deterministic dependency order without executing steps.

#### Scenario: Deferred output binding
- **WHEN** a step binds an input to another step output
- **THEN** the upstream step precedes it and missing or cyclic dependencies are rejected

### Requirement: Portable package verification
The SDK SHALL validate package schema, portable paths, inventory and content digests without filesystem or network side effects.

#### Scenario: Tampered package
- **WHEN** a file changes, a path escapes or two paths collide on a supported platform
- **THEN** verification fails with a stable error category
