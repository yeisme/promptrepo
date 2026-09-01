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

## v0.3.0 release evidence

Before tagging `v0.3.0`, the maintainer ran `GOWORK=off go mod verify`,
`GOWORK=off go test ./...`, `GOWORK=off go vet ./...`, CGO-disabled tests and
builds on the Go 1.24 floor, strict OpenSpec validation across all changes,
and a secret scan over the release tree; all passed. The release is an
additive public-contract change: `Client`, `Ref`, catalog/state/receipt DTOs,
source profiles, persisted state, and existing deterministic digests stay
compatible with `v0.2.0`. The tag is published only after anonymous
`GOWORK=off go mod download github.com/yeisme/promptrepo@v0.3.0` succeeds; the
first consumer canary is Pinax, which pins the public tag without a filesystem
replace or private module credential.

## Planned v0.4.0 gate

`v0.4.0` is an additive public-contract release for structured documents
(Markdown, text, JSON, YAML, JSONL) and unified repository-set/policy
management. Before creating the tag, require:

1. `GOWORK=off go mod verify`, `GOWORK=off go test ./...`,
   `GOWORK=off go vet ./...`, and CGO-disabled test/build pass from this
   repository on the Go 1.24 compatibility floor and a currently supported
   Go release.
2. `Client`, `Ref`, catalog/state/receipt DTOs, source profiles, persisted
   state, and existing deterministic digests remain compatible with `v0.3.0`.
3. Structured document bodies and values stay excluded from JSON/YAML
   projections, logs, evidence, and receipts; only the additive loader,
   resolver, and selector contracts expose selection results, and every
   renderer keeps its body-free sentinel behavior.
4. `RepositorySetReader` and `PolicyEvaluator` remain optional interfaces;
   deny-wins policy admission never reads or stores credentials, and the
   management projection stays body-free.
5. The maintainer creates one immutable annotated `v0.4.0` tag. Anonymous
   `GOWORK=off go mod download github.com/yeisme/promptrepo@v0.4.0` must pass
   before consumer dependency updates.
6. Eikona is the first consumer canary for this release through its owner
   OpenSpec change; no consumer may use a local `replace` as release
   evidence.

If a canary fails, fix forward as `v0.4.1` or a new pre-release. Never move or
retag `v0.4.0`; consumers may remain on `v0.3.0` because no state migration is
required.

## v0.4.0 release evidence

Before tagging `v0.4.0`, the maintainer ran `GOWORK=off go mod verify`,
`GOWORK=off go test ./...`, `GOWORK=off go vet ./...`, CGO-disabled tests and
builds on the Go 1.24 floor, strict OpenSpec validation across all changes,
and a secret scan over the release tree; all passed. The release is an
additive public-contract change: `Client`, `Ref`, catalog/state/receipt DTOs,
source profiles, persisted state, and existing deterministic digests stay
compatible with `v0.3.0`. The tag is published only after anonymous
`GOWORK=off go mod download github.com/yeisme/promptrepo@v0.4.0` succeeds; the
first consumer canary is Eikona, which pins the public tag without a
filesystem replace or private module credential.

## Unreleased ui-template gate

受限 UI template 合同是 additive minor candidate，不授权在本 change 中创建 tag、push
或发布。进入 consumer canary 前要求：

1. `GOWORK=off go mod verify`、`GOWORK=off go test ./...`、
   `GOWORK=off go test -race ./...`、`GOWORK=off go vet ./...`、
   `CGO_ENABLED=0 GOWORK=off go test ./...` 与
   `CGO_ENABLED=0 GOWORK=off go build ./...` 全部通过。
2. `ParseRef`、`TemplateAddress`、`TemplateRole`、structured document、catalog digest、
   `Client` 和 durable state compatibility tests 保持原结果。
3. HTML/CSS dangerous corpus、comment/escape bypass、size/UTF-8、slot、digest/snapshot、
   explicit tag balance、duplicate attribute、namespace、symlink/path containment 和
   body-redaction tests 全部通过。
4. machine projection 与错误不得包含 HTML/CSS、consumer values、credential、provider
   payload 或私有绝对路径。
5. Template Registry 是首个 contract consumer，随后由 Scaena exact-ref fixture/canary
   复验；两者不得复制或放宽 Promptrepo validator。
6. bundle ceiling 只表达单个 artifact 的安全边界，不得被解释为全局资产数量 quota。
7. `openspec validate promptrepo-ui-template-contract-v1 --strict --no-interactive`
   通过后，仍需 maintainer 另行决定版本号、tag 与发布。
8. 使用已修复标准库漏洞的 supported Go toolchain 执行 `GOWORK=off govulncheck ./...`；
   UI template 路径不得产生额外 reachable finding。Go 1.24 兼容线允许的唯一已知例外仍是
   `GO-2026-5970`：只有 focused invalid-UTF-8 guard regression 通过且扫描无其他 reachable
   module issue 时才可接受。UI template validator 不得重新引入 `golang.org/x/net/html`。

回滚不需要 state/catalog migration：consumer 停止解析和加载 `kind=ui-template` 即可；
既有 `kind=template` 与 structured document 路径保持可用。
