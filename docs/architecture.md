# Architecture

## Ownership

`github.com/yeisme/promptrepo` is the canonical public owner of the Go SDK.
The private Template Registry nested SDK tagged `sdk/go/promptrepo/v0.1.0` is
an immutable compatibility baseline, not a second active owner. Template
Registry remains a separate server/control-plane implementation and is not a
dependency of this module.

## Package boundary

```text
consumer -> promptrepo.Client <- engine.Manager -> source.Registry
                                      |                 |-- file
                                      |                 |-- Git/GitHub
                                      |                 `-- anonymous S3
                                      `-> user config/cache state
```

The root package exposes DTOs, error codes, exact-ref parsing, digest helpers,
and the `Client` interface. `engine` is the embedded local implementation;
`source` supplies bounded read-only source adapters. No package exposes server
internals, ORM types, or credentials.

## 模板 Address 与安全 projection

`Ref` 和 repository source URI 是两个刻意隔离的层：

```text
RepositoryProfile.Source -> file/git/github/s3 adapter -> synchronized snapshot
promptrepo Ref           -> solution identity in that snapshot
Template Address         -> role/path/selector/digest/snapshot projection
```

已有 `ParseRef`/`FormatRef` 原封不动地保留 solution ref 语法。加性的
`ParseTemplateAddress`/`FormatTemplateAddress` 只接受 `kind=template`，并以
`kind,locale,role,path,selector,digest,snapshot` 固定顺序生成 query。相对 path
禁止 traversal；selector 只能是 `heading:`、`json-pointer:` 或 `yaml-pointer:`；
digest/snapshot 只能是 sha256。source URI 绝不作为 Address 解析，也不从 Address
取得任何网络或凭据能力。

已发布的 `TemplateRole`、`Solution`、catalog 与 state DTO 均不增加字段或 tag。
输入定义、license、permissions 与可选 contract digest 位于新的 caller-supplied
`TemplateContract`，随 inspect/validate/preview 的加性请求和结果传递。独立
`ContractResolver` 从相同 source 读取 Template Registry 生成的 companion。
Git/GitHub 从精确 commit 读取并标记 `snapshot_pinned`；file/S3 校验当前对象的
sidecar digest、identity、template path/digest，并标记 `content_bound`。
sidecar，并核对 schema、identity、template path/digest 与 contract digest；本 SDK 不把
这些数据序列化到 catalog/snapshot/state。

`Client` 不增加方法。想使用新能力的消费者通过独立 `ContractResolver`、`Inspector`、`Validator`、
`Renderer`、`Previewer` 接口进行类型断言。Render 仅对调用方提供的内存正文做严格
`{{name}}` 替换；Inspect 仅读取 state/catalog metadata；Preview 在既有 adapter 完成
digest 校验后读取正文并调用 Render。Preview 不进入 state write path、不创建 run、
不调用 provider、不记录 usage，且 Render/Preview 的 `RenderedBody` 均标记为
`json:"-" yaml:"-"`。selector grammar 是预留语法；Inspect 可显示 selector metadata，
但 Preview 对任何非空 selector 返回 `SELECTOR_UNSUPPORTED`，直到 selector engine 具备
独立 conformance 测试。

## Compatibility invariants

- `promptrepo.state.v0.1`, `promptrepo.catalog.v0.1`, and
  `promptrepo.stage_receipt.v0.1` remain unchanged through the extraction.
- Config and cache defaults remain `yeisme/promptrepo` below the OS user config
  and cache directories.
- Digest algorithms, error-code strings, JSON tags, and exact-ref format remain
  compatible with private v0.1.0.
- State uses a cross-process lock and atomic rename; unsupported future state
  schemas fail closed.
- 模板 Address 和 inspect/validate/render/preview 均为未来 v0.3.0 的加性
  surface；没有 state migration。发现回归时还原该 v0.3.0 change 并继续使用
  v0.2.0，state 无需回滚。

The `compatibility_v010_test.go` test executes a frozen v0.1.0-shaped fixture
without private-network access and pins the observable result.
