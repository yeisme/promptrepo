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

## Stable release evidence

The `v0.2.0-rc.1` candidate passed public Go 1.24 CI, Template Registry tests,
Sonora's full local quality gate, a credential-free Sonora GitHub Actions run,
and the official prompt catalog deterministic rebuild. Stable `v0.2.0` keeps
the RC contract and persisted-state behavior unchanged.

## Planned v0.3.0 gate

`v0.3.0` is an additive public-contract release for Template Address,
caller-supplied TemplateContract, inspect, validate, in-memory render, and
provider-free preview. Before creating the tag, require:

1. `GOWORK=off go test ./...`, `go vet ./...`, `go mod verify`, and
   CGO-disabled test/build pass from this repository on the Go 1.24
   compatibility floor and a currently supported Go release.
2. `Client`, `Ref`, catalog/state/receipt DTOs, source profiles, persisted state
   and existing deterministic digests remain compatible with `v0.2.0`.
3. Rendered body remains excluded from JSON/YAML, logs, evidence and receipts;
   selector preview fails closed with `SELECTOR_UNSUPPORTED`.
4. Invalid UTF-8 is made valid before Unicode search normalization so the
   Go-1.24-compatible `x/text` line cannot receive the input shape affected by
   GO-2026-5970. Release security evidence must use a patched supported Go
   toolchain for standard-library checks. `govulncheck` remains expected to
   report the module-level symbol because it does not model this input guard;
   that single finding is accepted only while the focused invalid-UTF-8
   regression passes and the scan reports no additional reachable issue.
5. Template Registry generated companion sidecars validate without changing
   the official catalog digest.
6. The maintainer creates one immutable annotated `v0.3.0` tag. Anonymous
   `GOWORK=off go mod download github.com/yeisme/promptrepo@v0.3.0` must pass
   before consumer dependency updates.
7. Sonora is the first consumer canary; Eikona and Scaena follow their own
   owner OpenSpec changes. No consumer may use a local `replace` as release
   evidence.

If a canary fails, fix forward as `v0.3.1` or a new pre-release. Never move or
retag `v0.3.0`; consumers may remain on `v0.2.0` because no state migration is
required.

When the public module raises its compatibility floor to Go 1.25 or newer, it
must upgrade `golang.org/x/text` to v0.39.0 or later and remove this temporary
guard-based vulnerability exception after re-running the compatibility suite.
