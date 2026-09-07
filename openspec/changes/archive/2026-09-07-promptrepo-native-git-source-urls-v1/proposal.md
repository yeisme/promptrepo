## Why

Repository profiles currently require Promptrepo-specific wrappers such as `github://` or `git+https://`. Users expect the same clone URLs accepted by Git, especially copied GitHub HTTPS and SSH addresses, and bare `github.com/owner/repository` shorthand.

## What Changes

- Accept native `http://`, `https://`, `ssh://`, SCP-style SSH, and bare GitHub repository sources.
- Normalize GitHub web and shorthand addresses to a deterministic HTTPS Git remote.
- Preserve existing `github://`, `git+file://`, `git+https://`, and `git+ssh://` inputs.
- Reject embedded passwords, HTTP user information, URL queries, fragments, malformed hosts, and non-repository GitHub page paths.

## Capabilities

### New Capabilities

- `native-git-source-urls`: Additive source detection, normalization, validation, and compatibility requirements for Git repository profiles.

### Modified Capabilities

None.

## Impact

The public `source` package gains additional accepted input forms without changing exported Go APIs, stored profile fields, error codes, or Git command execution. Template Registry and other consumers can pass user-supplied native clone URLs through the existing `RepositoryProfile.Source` field.
