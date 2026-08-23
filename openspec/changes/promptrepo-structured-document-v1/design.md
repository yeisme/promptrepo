## Context

`TemplateRole` 是已发布兼容面，只含 `Role`、`Locale`、`Path`、`Digest`。`ReadTemplate` 由 source adapter 读取 bytes 后返回 `TemplateContent.Body string`。地址语法已经为 heading、json-pointer 和 yaml-pointer 预留 selector，但当前 preview 对 selector 仍失败关闭。结构化能力必须扩展新表面，不能重定义旧结构。

## Goals / Non-Goals

**Goals:**

- 以纯 Go、CGO-free 方式确定性读取和校验 JSON/YAML/JSONL。
- 保持旧 template read/render/preview byte-for-byte 兼容。
- 向 consumer 提供 parsed document 的安全内存对象和不含正文的 machine projection。
- 让 source/canonical/schema/compiler lineage 可被 Registry、Scaena 和 Eikona锁定。

**Non-Goals:**

- 不实现领域 Schema、角色连续性 lint 或 Prompt Bundle compiler。
- 不持久化用户 supplied values、parsed nodes 或 rendered body。
- 不执行 descriptor 引用的任何代码。
- 不实现 JSONL 任意过滤、排序或 join 查询语言。

## Decisions

### 1. Companion descriptor 通过约定路径绑定旧模板

descriptor schema 为 `promptrepo.template-document.v0.1`，content repo 约定路径为 `contracts/documents/<role>.<locale>.document.json`。绑定键为 package/solution/version/role/locale/template path/source digest；任一项不一致即视为 stale 或 tampered。

### 2. 新公共接口 additive 发布

新增独立请求/响应，不向 `TemplateRole` 或 `TemplateContent` 添加改变调用假设的字段。建议的接口职责为：

```text
DocumentResolver.ResolveDescriptor
DocumentLoader.LoadDocument
DocumentSelector.SelectDocument
```

`LoadedDocument` 的 body/node 字段不得参与 JSON/YAML serialization；安全 projection 只包含 format、digests、schema/compiler refs、selector、size、readiness 和 findings。

### 3. 严格解析和 canonicalization

- JSON：duplicate-key rejection、bounded depth/size、UTF-8 only。
- YAML：1.2 JSON-compatible subset；拒绝 alias、anchor、merge key、tag、non-string map key 和非 JSON 数值。
- JSONL：UTF-8、LF、每行独立 object、最终 newline、stable record ID、bounded line/segment。
- JSON/YAML 使用 RFC 8785 JCS 计算 canonical digest。
- JSONL 每条记录 JCS 后加单个 LF；segment digest 对完整 canonical bytes 计算。
- Markdown/Text 保持 raw UTF-8 digest，不对正文做语义规范化。

### 4. Selector 只实现确定性定位

首版支持：

```text
heading:<text>
json-pointer:/path
yaml-pointer:/path
jsonl-id:<record-id>
```

YAML pointer 在归一化 JSON node 上按 RFC 6901 执行。JSONL record ID 必须唯一；首版不支持 `jsonl-id:<id>#/pointer` 或查询表达式。

### 5. Descriptor 只声明 compiler profile，不执行 compiler

Promptrepo 校验 compiler ref/digest 和 declared capability，但只把它交给 consumer。可执行代码不允许来自 repository source；consumer 只能选择内置/allowlisted compiler 实现。

### 6. Error code additive

新增错误分类建议：`DOCUMENT_DESCRIPTOR_MISSING`、`DOCUMENT_FORMAT_MISMATCH`、`DOCUMENT_DUPLICATE_KEY`、`DOCUMENT_PARSE_FAILED`、`DOCUMENT_SCHEMA_INVALID`、`DOCUMENT_CANONICALIZE_FAILED`、`DOCUMENT_TOO_LARGE`、`SELECTOR_NOT_FOUND`、`JSONL_RECORD_DUPLICATE`、`JSONL_RECORD_TOO_LARGE`。旧错误码及 retryable 语义保持不变。

## Risks / Trade-offs

- [YAML library 默认功能过宽] → AST 预扫描和 JSON-compatible conversion，不直接解码到 loose map。
- [Canonical digest 与 source digest 混用] → 类型和 projection 中明确命名，TemplateRole digest 继续是 source digest。
- [新增 interface 让 consumer 实现负担变大] → 不修改现有 Client interface；通过可选 additive interface/type assertion 或新 client surface 暴露。
- [JSONL 大文件耗内存] → streaming reader、line/segment bounds、record-id index callback，不整文件 materialize。

## Migration Plan

1. 固定 v0.3 兼容 baseline 和旧 conformance fixtures。
2. 新增 descriptor DTO/parser 与 file/local fixtures。
3. 新增 JSON/YAML/JSONL strict parsing和 selector。
4. 新增 safe projection 与 body sentinel tests。
5. Registry/content repo canary 通过后发布 additive public minor。
6. 回滚时 consumer 停用 structured document feature；旧 ReadTemplate/Render/Preview 无数据迁移。
