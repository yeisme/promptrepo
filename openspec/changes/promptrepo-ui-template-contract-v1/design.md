## Context

Promptrepo 的稳定公共面已经提供 `ParseRef`、`TemplateAddress`、`ParseTemplateAddress`、`TemplateRole`、structured document loading 与 digest-safe machine projections。`TemplateAddress` 的 canonical query 顺序为 `kind,locale,role,path,selector,digest,snapshot`，并明确只接受 `kind=template`。直接扩展现有 parser 接受另一种 kind 会改变已发布输入边界，可能让旧调用方把 UI artifact 当作 prompt template 处理。

Scaena 本地审阅页需要引用一个不可执行的 HTML fragment + CSS bundle，Template Registry 需要在发行前复用同一地址、DTO 和校验语义。因此采用并行的 additive `ui-template` 合同，而不是修改现有 template/solution/document 合同。

## Goals / Non-Goals

**Goals:**

- 提供可精确解析和格式化的 `UITemplateAddress`，保留现有地址/API 二进制与行为兼容。
- 提供纯 Go `UITemplateBundleV1`、security profile、slot 和 limits DTO。
- 提供 deterministic、fail-closed、无执行副作用的 validate/inspect/load 合同。
- 让 Template Registry 与 Scaena 对 digest、snapshot、禁止语法和安全投影达成同一理解。
- 保证 JSON/agent projection 不泄露模板正文或 consumer 注入值。

**Non-Goals:**

- 不修改 `ParseRef`、`FormatRef`、`ParseTemplateAddress`、`FormatTemplateAddress` 或 `TemplateRole`。
- 不执行 HTML/CSS、不渲染页面、不运行 JavaScript、不提供 sanitizer rewrite。
- 不管理 artifact release、CAS、mirror、install、browser session、CSP 或 Scaena action。
- 不访问网络、provider、credential、consumer state 或文件系统之外的隐式 source。
- 不定义完整通用 Web component/plugin 系统。

## Decisions

### 1. 新建独立地址类型，不扩宽 TemplateAddress

新增：

```go
type UITemplateAddress struct {
    RepositoryID string
    PackageID    string
    SolutionID   string
    Version      string
    Locale       string
    Role         string
    Path         string
    Digest       string
    Snapshot     string
}

func ParseUITemplateAddress(raw string) (UITemplateAddress, error)
func FormatUITemplateAddress(address UITemplateAddress) string
```

canonical URI 复用 solution identity 和 query 顺序：

```text
promptrepo://official/scaena/storyboard-review@1.0.0?kind=ui-template&locale=zh-CN&role=review&path=ui/review.zh-CN.html&digest=sha256%3A...&snapshot=sha256%3A...
```

`kind=ui-template` 与 locale 必需；role/path/digest/snapshot 可选但 exact production load 要求 digest + snapshot。UI template 不支持 content selector，因为 bundle 是固定的 fragment/CSS/metadata 单元，不允许通过 selector 绕开整体 digest/validation。

parser 拒绝 userinfo、fragment、unknown/duplicate query、absolute/traversal path、非法 locale、非 sha256 digest 和 selector。旧 `ParseTemplateAddress` 继续拒绝 `kind=ui-template`。

### 2. Bundle 是声明式展示合同

```go
type UITemplateBundleV1 struct {
    SchemaVersion string
    Address       UITemplateAddress
    HTMLFragment  []byte
    CSS           []byte
    Slots         []UITemplateSlotV1
    Security      UITemplateSecurityProfileV1
    Limits        UITemplateLimitsV1
    ContentDigest string
    Snapshot      string
}
```

slot 只声明稳定 name、kind、required 和 cardinality；不包含 executable callback、HTTP method、endpoint、arbitrary payload schema 或 consumer state mutation。Scaena 根据 slot name 将自己的 safe component/projection 注入。

V1 默认 bounds：HTML fragment ≤ 256 KiB、CSS ≤ 256 KiB、bundle body 总计 ≤ 512 KiB、slot ≤ 64、slot name ≤ 64 UTF-8 bytes，且全部正文必须是有效 UTF-8。Limits 是 bundle 中显式、可降但不可超过 V1 ceiling 的字段，使 Registry 与 consumer 使用相同边界。

### 3. Validator 拒绝危险内容，不重写内容

HTML validation 使用有界 lexer、展示元素/静态属性 allowlist 与显式结构栈，而非正则
sanitizer 或容错 DOM rewrite，至少拒绝：

- `script`、`style`（CSS 必须独立）、`form`、`input`、`button`、`textarea`、`select`、`iframe`、`object`、`embed`、`svg`、`math`、`meta`、`base`、`link`；
- 所有 `on*` attribute、URL-bearing attributes、`srcdoc`、`contenteditable`；
- `hx-*`、`x-*` 等 framework directive、Mustache/EJS/EL template delimiter，及除
  `data-promptrepo-slot` 外的任意 `data-*`；
- 未声明、重复或非法 slot；
- 重复 attribute、namespace、非 void element 自闭合、未闭合或错序闭合；
- document-level nodes（`html/head/body/doctype`），因为 artifact 只能是 fragment。

`static-review-fragment-v1` 的 element/attribute allowlist 以 capability spec 中的完整集合为
canonical input boundary。Promptrepo validator 是唯一实现；Registry 与 Scaena 必须调用
该 validator，不能复制名单。后续新增元素或属性会扩宽稳定安全输入面，因此必须发布
successor security profile，不能原地修改 v1。

CSS validation 至少拒绝 `@import`、`url()`、`expression()`、`behavior`、`-moz-binding`、外部 font/source、HTML breaking sequences 和 parser errors。V1 不接受网络资源或 data URL。校验只返回 violations，不进行“清洗后继续”，避免 digest 与实际执行内容不一致。

### 4. Digest 覆盖 canonical bundle，而不是文件系统偶然布局

`ContentDigest` 对 schema version、canonical address identity、normalized slot/security/limits metadata、HTML bytes 和 CSS bytes 的长度分隔 canonical byte stream 计算 sha256。`Snapshot` 继续表达 repository/catalog snapshot。loader 必须同时校验 address digest、bundle digest 和 snapshot binding；路径只用于 source lookup，不进入 consumer output 的私有绝对路径。

### 5. Inspect/Load 分离，machine projection 省略正文

公共接口：

```go
type UITemplateInspector interface {
    InspectUITemplate(context.Context, InspectUITemplateRequest) (UITemplateInspectionV1, error)
}

type UITemplateLoader interface {
    LoadUITemplate(context.Context, LoadUITemplateRequest) (UITemplateBundleV1, error)
}
```

Inspect 返回 address、schema/security profile、slots、limits、digest/snapshot、body byte counts、validation result/violations，不返回 HTML/CSS body。Load 只向进程内调用者返回 body；`MarshalJSON`/machine DTO 默认省略 HTML/CSS 和 supplied values。若调用者需要持久化内容，应由其 owning CLI/application service 完成，Promptrepo 不提供手写 metadata 路径。

### 6. Error code 兼容并保持可修复

新增稳定错误码：

| code | 含义 |
| --- | --- |
| `UI_TEMPLATE_ADDRESS_INVALID` | 地址、kind、query 或 path 非法 |
| `UI_TEMPLATE_LIMIT_EXCEEDED` | body/slot/field 超出 ceiling |
| `UI_TEMPLATE_HTML_FORBIDDEN` | HTML 含禁止节点/属性/URL |
| `UI_TEMPLATE_CSS_FORBIDDEN` | CSS 含外部/可执行/非法语法 |
| `UI_TEMPLATE_SLOT_INVALID` | slot 未声明、重复或非法 |
| `UI_TEMPLATE_DIGEST_MISMATCH` | canonical digest 不匹配 |
| `UI_TEMPLATE_SNAPSHOT_MISMATCH` | source snapshot 不匹配 |

错误信息不得回显完整 body；最多返回 violation code、line/column（可用时）和有界 excerpt digest。

## Risks / Trade-offs

- [两套 address API 增加表面积] → 明确命名与 shared internal parser helpers；换取已发布 `TemplateAddress` 不被扩宽。
- [禁止标签过严影响设计自由] → V1 聚焦 review layout；新能力通过 security profile v2/additive change 演进，不以 sanitizer 放宽。
- [CSS parser 实现复杂] → 采用有界 tokenizer/parser 和禁止 token contract，conformance corpus 覆盖大小写、转义与注释绕过。
- [body 未序列化影响调试] → Inspect 提供 size/digest/violation；正文只能在明确 load 的进程内使用。
- [Registry 与 Scaena 验证漂移] → 两者直接复用 promptrepo validator/conformance fixtures，不复制规则。

## Migration Plan

1. 新增 address/DTO/errors/validator，不修改现有 exported identifiers。
2. 增加 parser canonicalization、dangerous corpus、digest 和 JSON redaction tests。
3. 增加 filesystem fixture loader/inspector；默认无网络 source。
4. Template Registry 和 Scaena 分别升级为 consumer；旧 prompt/template consumer 无需迁移。
5. 若新 API 回滚，只移除尚未发布的 ui-template consumer；现有 template/document 行为和 catalog digest 不变。

## Open Questions

- V1 首个内容仓固定只需 `zh-CN`，但地址与 bundle 保持 locale 通用；具体 fallback locale policy 由 consumer/Registry 决定。
- 若未来需要安全 inline icon，必须新增专门的 typed icon asset contract；V1 不开放 SVG/data URL。
