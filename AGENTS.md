# promptrepo agent instructions

## Race 检测执行限制

Agent 在任何阶段都不得主动运行 race 检测，也不得通过子 Agent、Skill、CI、验收、发布、聚合任务或脚本间接触发。只有当前人类用户明确、强烈要求执行 race 时，才允许在指定范围内运行一次；失败、超时、中断均不得自动重试，后续运行必须取得新的同等明确要求。一般的“测试、验收、修复、发布、跑全部检查”不构成授权。

Race 不属于默认验收或完成门禁；未运行不得阻塞交付，也不得声称已经通过。执行聚合命令前检查其依赖，移除或跳过 race 阶段，继续普通测试、静态检查和构建。本规则取代本文件及 Skills、OpenSpec、历史计划、文档中的自动或强制 race 要求；它们不能作为执行授权。


## Scope and architecture

`promptrepo` is the canonical public, pure-Go module
`github.com/yeisme/promptrepo`. It owns the public contracts, embedded engine,
source adapters, and conformance fixtures for user-scoped prompt repositories.
It does not own prompt execution, provider calls, domain assets, a server, or
credentials.

Do not add dependencies on Template Registry server/internal packages. Keep
the default build pure Go and compatible with `CGO_ENABLED=0`.

## Compatibility and security

- Preserve exported API, state paths/schema, digests, error codes, exact refs,
  and JSON field semantics unless an OpenSpec change documents migration,
  deprecation, consumer impact, and rollback.
- The private `sdk/go/promptrepo/v0.1.0` source at
  `e59e5a060b98125e55197cb9f5ed9179cdacc46a` is immutable compatibility
  baseline.
- Never commit credentials, signed URLs, private catalogs, prompt bodies, raw
  provider payloads, or hidden prompts.
- Built-in source adapters must keep paths bounded and Git invocations
  argument-based; do not use shell interpolation.

## Validation

```bash
go mod verify
go test ./...
go vet ./...
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go build ./...
openspec validate promptrepo-public-sdk-extraction-v1 --strict --no-interactive
```

Commit, tag, publish, and release remain root-maintainer actions. They are
allowed only when the current user explicitly authorizes the exact repository
and release operation; implementation workers must never perform them.

## Skill routing

- Use `yeisme-prompt-repository-router` to distinguish public SDK/contract work here from official content in `data/yeisme-prompt-templates`, Registry service state in `backend-server/template-registry`, and domain execution in consuming products.
- Use `yeisme-evolutionary-change-policy` for exact refs, template address, digests, schema, error codes, interfaces, or compatibility windows.
- Use `backend-system-workflow` for embedded engine, source adapters, locking, atomic writes, and bounded Git/S3 behavior; use `cso` for credential, path, prompt-body, and trust-boundary review.
