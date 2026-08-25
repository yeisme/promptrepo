## Context

The SDK exists as a finished nested Go module in the private Template Registry
repository at tag `sdk/go/promptrepo/v0.1.0`, peeled commit
`e59e5a060b98125e55197cb9f5ed9179cdacc46a`. Its module path prevents public
consumption and couples resolution to the server repository. The public module
must preserve an already-observable pre-1.0 API and user-state contract while
moving canonical ownership.

## Goals / Non-Goals

**Goals:**

- Make `github.com/yeisme/promptrepo` the canonical public Go 1.24 SDK.
- Retain private v0.1.0 behavior, exported API, schema names/paths, digests,
  error codes, source boundaries, conformance helpers, and tests.
- Provide evidence that public CI validates compatibility without private
  network access.
- Define RC/stable gates, tag immutability, consumer migration, deprecation,
  and rollback.

**Non-Goals:**

- Changing SDK behavior, state schemas, stored data, catalog formats, or
  exported APIs.
- Importing Template Registry server/internal packages, adding a server, or
  publishing/releasing from this repository.
- Moving consumer code or changing the immutable private baseline.

## Decisions

1. **Extract the exact tagged SDK subtree.** The public source begins as the
   17 tracked files from the peeled commit; only Go module/import prefixes and
   publication/foundation files change. This limits behavior drift. Rewriting
   the SDK or copying the current private branch was rejected because neither
   gives a reproducible compatibility baseline.
2. **Keep the public module pure Go and Go 1.24.** The existing
   `golang.org/x/text` dependency is retained; there are no server, ORM, AWS,
   or HTTP-service dependencies. CI validates `CGO_ENABLED=0`. Adding a server
   dependency or cgo would violate the consumer boundary and complicate public
   distribution.
3. **Compatibility test uses a frozen v0.1.0-shaped local fixture.** It runs
   public root/engine/source packages against a local catalog and pins ref,
   digest, receipt, state, and error outcomes. It does not fetch the private
   repository, so public CI is reproducible.
4. **Private v0.1.0 is preserved for a migration window.** Consumers change
   only the module/import prefix. The private tag is never changed or retagged;
   it is deprecated after public stable cutover. A consumer can roll back by
   restoring the old dependency while the migration window remains open.
5. **Tags are maintainer-only and immutable.** `v0.2.0-rc.1` precedes consumer
   canaries; `v0.2.0` requires the same gates again. A defective release is
   fixed in a new version, not by moving a tag.

## Risks / Trade-offs

- **Public import-path change is source-breaking for callers** → document the
  one-prefix migration, retain private v0.1.0 during the window, and pin API
  behavior with a local compatibility test.
- **Extraction may silently drift from the tag** → record the peeled commit,
  retain copied tests, and run a normalized source diff at release review.
- **A public source may contain sensitive material** → only SDK code and
  non-sensitive local fixtures are copied; run secret scans before RC/stable.
- **Future state incompatibility could corrupt user data** → retain the
  fail-closed schema check and atomic state writes; no migration is included.

## Migration Plan

1. Verify the extracted module against the private tag and local compatibility
   fixture; run Go, vet, CGO-disabled, strict OpenSpec, diff, and secret gates.
2. A maintainer creates `v0.2.0-rc.1`; consumers canary by replacing the old
   module/import prefix only.
3. After canaries and all gates pass, a maintainer creates immutable `v0.2.0`.
4. Mark the private v0.1.0 location deprecated while keeping it immutable and
   available for rollback reference.
5. If RC/stable fails, do not retag. Revert the consumer dependency change or
   publish a new RC/fix version; existing state remains compatible.

## Open Questions

- The public release maintainer owns the exact deprecation-window duration and
  consumer-canary roster; this extraction does not require a new SDK behavior
  decision.
