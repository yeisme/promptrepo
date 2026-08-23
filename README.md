# promptrepo

`promptrepo` 是一个独立、纯 Go 的提示词仓库 SDK。它提供用户级 repository
profile、不可变 catalog snapshot、确定性搜索、精确 `promptrepo://` 引用、模板
正文的 digest 校验读取，以及 staged installation receipt。它不执行 Prompt、不调用
模型，也不依赖 Template Registry 服务端或其内部包。

`promptrepo` is a standalone, pure-Go prompt repository SDK. It provides
user-scoped repository profiles, immutable catalog snapshots, deterministic
search, exact `promptrepo://` references, digest-verified template reads, and
staged installation receipts. It does not execute prompts, call models, or
depend on Template Registry server/internal packages.

## 安装 / Install

```bash
go get github.com/yeisme/promptrepo@v0.2.0-rc.1
```

模块要求 Go 1.24 或更高版本；常规构建和测试支持 `CGO_ENABLED=0`。
The module requires Go 1.24 or later; normal builds and tests support
`CGO_ENABLED=0`.

## 快速开始 / Quick start

```go
package main

import (
    "context"

    "github.com/yeisme/promptrepo"
    "github.com/yeisme/promptrepo/engine"
)

func main() error {
    client, err := engine.New(engine.Options{})
    if err != nil {
        return err
    }
    _, err = client.AddRepository(context.Background(), promptrepo.AddRepositoryRequest{
        Profile: promptrepo.RepositoryProfile{
            ID: "official", Source: "github://yeisme/prompt-templates", Trust: "official",
        },
    })
    return err
}
```

## 稳定合同 / Stable contract

- State schema: `promptrepo.state.v0.1`
- Catalog schema: `promptrepo.catalog.v0.1`
- Receipt schema: `promptrepo.stage_receipt.v0.1`
- Exact reference: `promptrepo://official/audio/podcast-narration@1.0.0?locale=zh-CN`

默认 state 位于 OS user config/cache 目录的 `yeisme/promptrepo`。State 由 engine
原子写入并使用跨进程锁；未来 schema 会以 `STATE_SCHEMA_TOO_NEW` fail closed。

Built-in sources are `file://`, Git (`git+file`, `git+https`, `git+ssh`,
`github://`), and anonymous read-only `s3://`. Profiles hold credential
references only, never credential values. See [docs/architecture.md](docs/architecture.md).

## 私有 v0.1.0 迁移 / private v0.1.0 migration

This repository is the canonical public home for the SDK beginning with
`v0.2.0-rc.1`. It was extracted without behavioral change from private tag
`sdk/go/promptrepo/v0.1.0` at commit
`e59e5a060b98125e55197cb9f5ed9179cdacc46a`.

Consumers replace only the module/import prefix:

```text
github.com/yeisme/backend-server-template-registry/sdk/go/promptrepo
→ github.com/yeisme/promptrepo
```

The private v0.1.0 source is immutable. It will be deprecated after public
cutover, not modified or retagged. See [CHANGELOG.md](CHANGELOG.md) and
[docs/release.md](docs/release.md) for RC, stable, rollback, and migration
gates.

## 开发 / Development

```bash
go mod verify
go test ./...
go vet ./...
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go build ./...
openspec validate promptrepo-public-sdk-extraction-v1 --strict --no-interactive
```

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and
[docs/README.md](docs/README.md).
