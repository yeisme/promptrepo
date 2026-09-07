## Context

`source.DetectKind` currently recognizes only Promptrepo-specific Git schemes, while `GitAdapter` separately normalizes the selected remote. Template Registry therefore rejects native Git URLs before a sync can begin. The source string is persisted and used as a cache identity, so changing stored values or cache hashing would create an unnecessary migration.

## Goals / Non-Goals

**Goals:**

- Accept common Git clone address forms at repository-add time.
- Validate the same input again at sync time through one shared normalizer.
- Keep legacy source forms and persisted profiles valid.
- Prevent credentials and browser-only GitHub paths from being treated as repository identities.

**Non-Goals:**

- Detect whether an arbitrary HTTP endpoint is a Git repository before sync.
- Add credential storage, GitHub API calls, archive downloads, local path inference, or a second Registry parser.
- Deduplicate cache directories across equivalent source spellings.

## Decisions

`DetectKind` delegates Git candidates to the existing package-private normalizer. This keeps add-time and sync-time validation identical without adding public API.

Native HTTP, HTTPS, and SSH URLs are passed to Git after bounded structural validation. GitHub HTTPS and bare `github.com/owner/repository` forms are canonicalized to `https://github.com/owner/repository.git`; SCP-style SSH remains unchanged because it is already Git's canonical transport syntax.

The persisted `RepositoryProfile.Source` remains the user-supplied string. Cache hashing also remains unchanged so existing snapshots and reads keep their current location. Equivalent spellings may use separate caches, which is safer than an implicit state migration.

URL queries and fragments are rejected. HTTP user information and SSH passwords are rejected; an SSH username such as `git@` remains valid. Credential values continue to belong outside repository profiles.

## Risks / Trade-offs

- A normal web page may pass structural HTTP validation and fail during `git fetch`. The adapter returns the existing source-fetch error, which preserves the current responsibility boundary.
- Equivalent addresses can create duplicate cache entries. The implementation preserves state compatibility; canonical cache migration can be designed separately if evidence shows a material cost.
- Older Promptrepo versions still reject native URLs. Consumers must upgrade the SDK before documenting the syntax as released behavior.

## Migration Plan

This is additive. Existing profiles require no migration and all legacy source forms remain accepted. Rollback consists of reverting the source parser and consumer documentation; profiles created with native syntax can be changed back to the equivalent legacy wrapper if an older SDK must be restored.

## Open Questions

None for this version. Native `git://` and implicit local filesystem paths remain intentionally unsupported.
