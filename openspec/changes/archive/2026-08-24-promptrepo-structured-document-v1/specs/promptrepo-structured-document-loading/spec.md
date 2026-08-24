## ADDED Requirements

### Requirement: Structured document API 必须 additive
Promptrepo MUST 新增独立 structured document DTO/interfaces，不得删除、重命名或改变 TemplateRole、TemplateContent、ReadTemplate、Render 和 Preview 的字段、默认值、digest 或错误语义。

#### Scenario: 旧 consumer 编译和运行
- **WHEN** 旧 consumer 仅使用 v0.3 public surface
- **THEN** 它无需 structured descriptor 仍能通过原有 conformance tests 并读取旧 Markdown 模板

### Requirement: Descriptor 必须绑定精确 source
Loader MUST 校验 descriptor 的 package、solution、version、role、locale、path 和 source digest 与已解析模板完全一致，任何 mismatch MUST 在解析正文前失败关闭。

#### Scenario: Descriptor digest stale
- **WHEN** 模板正文已变化但 descriptor 仍引用旧 source digest
- **THEN** loader 返回 stale/mismatch error，不计算新的 trusted canonical digest

### Requirement: Parser 必须严格且有界
Loader MUST 对 JSON/YAML/JSONL 执行 UTF-8、size、depth、duplicate-key 和 format-specific safety checks，并拒绝 YAML aliases/tags 和 JSONL duplicate record IDs。

#### Scenario: YAML anchor bomb
- **WHEN** YAML 使用多层 alias 扩展结构
- **THEN** loader 在 expansion 前拒绝文档，内存和 CPU 使用保持在配置上限内

#### Scenario: JSONL line 超限
- **WHEN** 单条 JSONL record 超过 descriptor 的 max_record_bytes
- **THEN** streaming loader 停止并返回 record-too-large，不继续读取或缓存剩余内容

### Requirement: Canonical digest 必须跨 JSON/YAML 稳定
对等价 JSON-compatible value，JSON 和允许的 YAML 表达 MUST 产生相同 canonical digest；source digest MUST 继续区分原始 bytes。

#### Scenario: 仅 YAML 注释变化
- **WHEN** YAML 只改变注释和缩进且解析值不变
- **THEN** source digest 改变而 canonical digest 保持不变

### Requirement: Selector 必须确定性执行
Selector MUST 仅支持已声明且 format-compatible 的 heading、JSON/YAML pointer 和 JSONL record-id；不存在、歧义、重复或不兼容 selector MUST 返回稳定错误。

#### Scenario: YAML pointer
- **WHEN** consumer 对 YAML 文档请求 `/task/layout/views/0`
- **THEN** selector 在 canonical JSON node 上返回唯一节点和完整 lineage

### Requirement: 安全投影不得包含正文
Structured document 的 machine projection MUST 排除 raw body、parsed node、selected node、rendered body 和 supplied values。

#### Scenario: JSON serialization
- **WHEN** LoadedDocument 或 inspect result 被编码为 JSON/YAML
- **THEN** 输出只包含 refs、digests、format、Schema/compiler metadata、size、selector、readiness 和 findings
