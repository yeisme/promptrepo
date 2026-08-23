# Release and migration

## Provenance and source of truth

The public module was extracted from private tag `sdk/go/promptrepo/v0.1.0`,
peeled commit `e59e5a060b98125e55197cb9f5ed9179cdacc46a`. That private source
is immutable. Public `github.com/yeisme/promptrepo` is canonical beginning at
`v0.2.0-rc.1` and hosts all new SDK fixes and releases.

## RC and stable gates

Before tagging `v0.2.0-rc.1` and again before `v0.2.0`, require:

1. `go mod verify`, `go test ./...`, `go vet ./...`, and CGO-disabled tests and
   builds pass on Go 1.24.
2. `compatibility_v010_test.go` passes without private network access.
3. Strict OpenSpec validation, source diff review, and a secret scan are clean.
4. No Template Registry server/internal dependency appears in `go.mod` or Go
   imports.
5. Release notes name the immutable v0.1.0 provenance and the migration path.

## Tag policy

Only the repository maintainer may create annotated public tags. The intended
sequence is `v0.2.0-rc.1`, consumer canary verification, then `v0.2.0` from the
same reviewed release line. Tags are immutable; a faulty release is corrected
by a new version, never retagging. This repository contains no publish or
release automation.

## Consumer migration and rollback

Consumers replace the old module/import prefix with
`github.com/yeisme/promptrepo`, run their existing tests, and can keep their
current state/config/cache because the schema and paths are unchanged. During
the migration window, consumers can revert their dependency/import change to
the immutable private v0.1.0 tag. If the public RC fails compatibility gates,
do not tag stable; fix forward in a new RC or revert consumer migration. The
private baseline is deprecated only after public stable cutover and remains
read-only for rollback reference.
