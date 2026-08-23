## Why

Promptrepo 已能解析 exact template address、读取受限正文并提供 inspect/validate/render/preview，但公共模板模型仍以文本正文为中心。Scaena 和 Eikona 需要读取 YAML/JSON 角色资产 Spec、Schema、声明式 compiler profile 和 JSONL fixtures，同时必须保持 `TemplateRole`、`ReadTemplate`、digest 和错误语义兼容。

## What Changes

- 新增 additive structured document descriptor、resolver、loader、selector 和 safe projection。
- 支持 `markdown`、`text`、`json`、`yaml`、`jsonl` 五种 format；format 由 descriptor 声明，文件扩展名只校验一致性。
- JSON/YAML 规范化为 JSON-compatible node，支持 RFC 6901 pointer；JSONL 支持 record-id selector。
- 同时记录 source digest 和 canonical digest，不改变旧 template digest 含义。
- 严格限制大小、深度、重复键、YAML aliases/tags、JSONL record size 和 selector 范围。
- 不在 Promptrepo 实现 Scaena 角色语义、Prompt compiler、provider call、用户数据库或生产审核。

## Capabilities

### New Capabilities

- `promptrepo-structured-document-loading`: 多格式文档描述、加载、规范化、选择器、安全投影和 conformance。

### Modified Capabilities

- 无。旧 public DTO、TemplateRole、ReadTemplate、Render 和 Preview requirement 保持不变。

## Impact

- 新增公共 Go DTO/interfaces、descriptor reader、strict parsers、canonicalization 和 conformance fixtures。
- Template Registry/content repo 生成 descriptor；Promptrepo 只读取和验证。
- Scaena/Eikona 作为 consumer 使用 parsed document 与 exact lineage，不 import Registry internal。
- 该变化必须在现有 v0.3 release/canary 完成后以 additive v0.4 或后续 minor 发布。
