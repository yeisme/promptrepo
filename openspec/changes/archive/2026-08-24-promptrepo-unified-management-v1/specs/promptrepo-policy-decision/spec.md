## ADDED Requirements

### Requirement: Admission MUST use deny-wins intersection
Policy evaluation MUST independently apply repository health, all repository allow/deny constraints, operation permission, minimum trust, rights and required capabilities. Any blocker MUST make `allowed=false` with stable reason codes.

#### Scenario: Project allows organization-denied repository
- **WHEN** organization policy denies a repository and project policy allows it
- **THEN** the decision remains blocked with `REPOSITORY_DENIED`

#### Scenario: Required capability is missing
- **WHEN** a resolved solution lacks a required consumer capability
- **THEN** the decision is blocked with `CAPABILITY_MISSING`
- **AND** the missing capability is listed deterministically

### Requirement: Exact selection MUST NOT bypass policy
Session exact selection, project pin and user preference MUST affect ordering only. They MUST NOT suppress policy reasons or change a blocked result to allowed.

#### Scenario: Exact ref targets prohibited rights
- **WHEN** an exact ref resolves to rights marked blocked or prohibited
- **THEN** policy returns `RIGHTS_BLOCKED`
- **AND** no preference field changes the result
