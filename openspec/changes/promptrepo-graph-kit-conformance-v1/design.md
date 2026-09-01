## Context

`engine.Manager` 已分别实现 `DocumentResolver`、`DocumentLoader`、`DocumentSelector`，source registry 已支持 `file`、Git/GitHub 和 S3 适配。Graph Kit 本质是一个 structured manifest 加多个 exact structured child，并不需要新的 graph-specific network protocol。

## Goals / Non-Goals

**Goals:**

- 证明同一 exact Git snapshot 内可以解析、加载和选择完整 Graph Kit closure。
- 为 Auctra/Registry 提供不依赖私有内容的公共 conformance fixture。
- 保持 exported API 向后兼容和 pure Go。

**Non-Goals:**

- 不定义 Auctra projection、workspace、review、cache 或 maturity 规则。
- 不执行 compiler、Prompt、provider 或 Registry release state machine。

## Decisions

### D1：复用三个小接口，不扩展 Client

Consumer 通过本地组合接口依赖 Resolver/Loader/Selector。Promptrepo 不给 `Client` 添加 document methods，也不合并三个接口，避免破坏 consumer mocks。

### D2：Conformance fixture 使用通用 graph manifest

Fixture 只包含 source slots、lens/view/schema/validator refs 与 synthetic rows，所有 child 使用 exact descriptor/digest。它验证 closure，不携带小说人物/剧情或 Auctra private Overlay。

### D3：Snapshot consistency 是 closure 不变量

所有 child 必须属于 manifest resolved snapshot，或由 manifest 明确锁定另一个 exact snapshot；浮动 revision、latest fallback、missing child 或 digest drift 均失败关闭。

## Risks / Trade-offs

- [把 Graph Kit 变成 Promptrepo domain type] → fixture/helper 只操作 structured document refs，不新增 graph business model。
- [Git 测试依赖网络] → conformance 使用 local bare/file Git；GitHub URI 只做 parser/normalization test，真实网络留给显式 canary。

## Migration Plan

1. 添加 conformance fixture 和 Git snapshot tests。
2. 只在测试证明 API 缺口时新增最小 additive helper。
3. 运行 pure-Go 全门禁并发布后提供 exact version handoff。

回滚：删除新 fixture/helper，现有公共 API 和状态不变。

## Open Questions

无。
