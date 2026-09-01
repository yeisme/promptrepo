## ADDED Requirements

### Requirement: Graph Kit closure 必须通过现有 structured document 接口加载
Promptrepo MUST 允许 consumer 使用 `DocumentResolver`、`DocumentLoader` 和 `DocumentSelector` 对一个 exact Graph Kit manifest 及其 required child documents 完成 provider-free closure load，不要求 graph-specific Client method。

#### Scenario: 加载完整 exact closure
- **WHEN** Git source 已同步到 exact commit，manifest 和 required children 的 descriptors/digests 均有效
- **THEN** consumer 可解析 manifest、加载每个 child、执行允许的 selector，并获得一致 snapshot metadata

### Requirement: Closure 漂移必须失败关闭
Promptrepo conformance MUST 拒绝浮动 snapshot、missing child、descriptor/source/canonical digest mismatch 和 selector identity mismatch，且错误不得包含 document body。

#### Scenario: Child digest 漂移
- **WHEN** required child bytes 与 descriptor 或 manifest digest 不匹配
- **THEN** load 返回 stable typed error、无 partial closure、无 latest fallback、输出不含 child body

### Requirement: Git/GitHub source 必须保持 exact commit 与 bounded path
Git/GitHub adapter MUST 将成功 sync 解析为 exact commit，并仅通过受限参数和 repository-contained path 读取 catalog、descriptor 和 child documents。

#### Scenario: Local Git conformance
- **WHEN** conformance repository 通过 Git adapter 同步
- **THEN** snapshot revision 为 exact commit，重复 sync/resolve 产生相同 refs/digests，且不执行 shell interpolation

