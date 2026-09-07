# native-git-source-urls Specification

## Purpose
TBD - created by archiving change promptrepo-native-git-source-urls-v1. Update Purpose after archive.
## Requirements
### Requirement: Native Git repository sources
The SDK SHALL recognize native HTTP, HTTPS, SSH, SCP-style SSH, and bare GitHub repository addresses as Git sources without requiring a Promptrepo-specific scheme.

#### Scenario: GitHub HTTPS address
- **WHEN** a repository profile uses `https://github.com/owner/repository`
- **THEN** source detection selects the Git adapter and synchronization uses `https://github.com/owner/repository.git`

#### Scenario: SSH clone address
- **WHEN** a repository profile uses `ssh://git@example.com/group/repository.git` or `git@example.com:group/repository.git`
- **THEN** source detection selects the Git adapter and preserves the validated Git remote

#### Scenario: Bare GitHub shorthand
- **WHEN** a repository profile uses `github.com/owner/repository`
- **THEN** synchronization uses the deterministic GitHub HTTPS remote

### Requirement: Legacy Git source compatibility
The SDK SHALL continue to accept `github://`, `git+file://`, `git+https://`, and `git+ssh://` repository sources with their existing adapter selection.

#### Scenario: Existing profile after upgrade
- **WHEN** an existing repository profile uses a legacy Git source form
- **THEN** source detection and synchronization continue without requiring a profile migration

### Requirement: Git source trust boundary
The SDK SHALL reject malformed Git source addresses, embedded HTTP credentials, SSH passwords, URL queries, URL fragments, control characters, and browser-only GitHub subpaths before invoking Git.

#### Scenario: Credential-bearing URL
- **WHEN** a repository source contains HTTP user information or an SSH password
- **THEN** validation fails with an invalid-request error without exposing the source value

#### Scenario: GitHub branch page
- **WHEN** a source points to a GitHub page below `owner/repository`, such as `/tree/main`
- **THEN** validation fails instead of cloning a different repository identity

