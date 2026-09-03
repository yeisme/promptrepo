## Why

Promptrepo 已发布 structured document resolver/loader/selector 与 Git/GitHub source adapter，但缺少 Graph Kit 完整 closure 的跨 source conformance，Auctra 因而只能在内部 CAS fixture 上证明更新流程。需要用公共现有 API 固定 exact Graph Kit manifest/child load、snapshot consistency 和 body-free failure contract。

## What Changes

- 新增通用 Graph Kit conformance fixture/profile，不引入 Auctra/Nishi 私有类型或内容。
- 验证 Git/GitHub source sync 到 exact commit 后，manifest 与 required child document 均由现有 DocumentResolver/Loader/Selector 读取。
- 验证 exact address、descriptor/source/canonical digest、selector 和 snapshot consistency；child missing/drift 必须 fail closed。
- 提供可供 consumer 组合的 conformance helper/fixture；不向既有 `Client` 或 consumer-implemented interface 加方法。
- 保持 pure Go、`CGO_ENABLED=0`、无 provider execution、无 server 和无 credential 持久化。

## Capabilities

### New Capabilities

- `promptrepo-graph-kit-conformance`: Graph Kit structured closure 在 Git/GitHub source 上的公共 conformance 行为与 fixture。

### Modified Capabilities

- 无。

## Impact

- 主要修改 `engine`、`source` conformance tests/fixtures 和 public docs；预计无需新增公共 DTO。
- 若 conformance 暴露公共 API 缺口，只允许 additive symbol/interface，必须保留现有 `DocumentResolver`、`DocumentLoader`、`DocumentSelector` 语义。
- 后续发布不可变 minor/tag 后，Auctra 和 Registry 才升级依赖；不使用本地 replace 作为 release evidence。
