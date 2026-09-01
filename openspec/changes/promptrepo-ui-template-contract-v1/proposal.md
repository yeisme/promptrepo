## Why

Scaena 的本地审阅页需要可版本化、可校验的展示模板，但 Promptrepo 当前公共地址只表达普通 `template`，且不能安全描述 HTML fragment、CSS、slot 与浏览器安全策略。需要一个 additive 公共合同，让 consumer 精确引用受限 UI 模板，同时不把渲染、服务端状态或发布职责带入 Promptrepo。

## What Changes

- 新增独立 `UITemplateAddress` 与 `kind=ui-template` parser/formatter，不改变现有 `TemplateAddress`、`ParseTemplateAddress` 或 `TemplateRole` 语义。
- 新增 `UITemplateBundle`、slot/security profile、content digest、snapshot 与 inspect/load 接口，支持受限 HTML fragment + CSS 的精确引用和验证。
- 规定 JSON/agent 安全投影默认省略模板正文，只输出 metadata、digest、limits 与 validation result。
- 固定合同禁止 JavaScript、事件处理器、form、iframe、SVG、外部 URL、CSS `url()`/`@import`，并为长度、文件数、slot 与 UTF-8 提供有界校验。
- Promptrepo 只提供纯 Go、无网络副作用的公共类型和验证；不执行模板、不管理发行、不接触 provider 或 consumer 领域状态。

## Capabilities

### New Capabilities

- `promptrepo-ui-template-contract`：`ui-template` 地址、bundle、验证、安全投影和 consumer-facing load/inspect 合同。

### Modified Capabilities

无。现有 template address、structured document loading、solution/catalog digest 合同保持不变。

## Impact

- 影响 `promptrepo` Go 公共 API、地址解析、DTO、验证器、文档与兼容性测试。
- Template Registry 可实现发行/安装，Scaena 可实现安全渲染；两者只消费公共合同，不反向把服务或产品状态放入本库。
- 这是 additive minor capability；旧调用方无需迁移，旧 parser 不接受 `ui-template` 的行为继续保持。
