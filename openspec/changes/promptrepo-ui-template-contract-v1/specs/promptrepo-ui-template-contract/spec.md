## ADDED Requirements

### Requirement: UI template 地址必须是 additive 独立类型
SDK MUST 提供 `UITemplateAddress`、`ParseUITemplateAddress` 与 `FormatUITemplateAddress`，要求 `kind=ui-template` 并按 `kind,locale,role,path,digest,snapshot` canonical order 格式化；现有 `ParseRef`、`TemplateAddress` 和 `ParseTemplateAddress` MUST 保持不变。

#### Scenario: 解析 canonical UI template address
- **WHEN** consumer 解析 `promptrepo://official/scaena/storyboard-review@1.0.0?kind=ui-template&locale=zh-CN&role=review&path=ui%2Freview.zh-CN.html&digest=sha256%3A0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef&snapshot=sha256%3Afedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210`
- **THEN** parser 必须返回结构化 identity 并由 formatter 生成 canonical 等价地址

#### Scenario: 旧 parser 收到 ui-template
- **WHEN** 调用方把 `kind=ui-template` 传给 `ParseTemplateAddress`
- **THEN** 旧 parser 必须继续拒绝该输入

#### Scenario: 地址含 traversal 或未知 query
- **WHEN** UI template address 含绝对/遍历 path、userinfo、fragment、selector、unknown 或 duplicate query
- **THEN** parser 必须以 `UI_TEMPLATE_ADDRESS_INVALID` fail closed

### Requirement: UITemplateBundleV1 必须是有界声明式合同
`UITemplateBundleV1` MUST 包含 schema version、address、HTML fragment、CSS、slots、security profile、limits、content digest 与 snapshot；V1 MUST 限制 HTML ≤ 256 KiB、CSS ≤ 256 KiB、body 总计 ≤ 512 KiB、slots ≤ 64，并要求有效 UTF-8。

#### Scenario: bundle 超过 ceiling
- **WHEN** HTML、CSS、总 body 或 slot count 超过 V1 ceiling
- **THEN** validation/load 必须返回 `UI_TEMPLATE_LIMIT_EXCEEDED`

### Requirement: HTML validation 必须拒绝执行和外部内容
validator MUST 只接受 static-review profile 明确列出的展示元素与静态属性，并拒绝脚本、document-level nodes、form controls、iframe/object/embed、SVG/MathML、事件属性、URL-bearing attributes、任意 framework/template directive、除 `data-promptrepo-slot` 外的 `data-*`、`srcdoc`、`contenteditable` 与未声明/重复 slot；validator MUST NOT 通过重写或 sanitizer 自动接受原内容。

`static-review-fragment-v1` 的 canonical element allowlist 固定为：

```text
abbr address article aside b bdi bdo blockquote br caption code col colgroup
dd del details dfn div dl dt em figcaption figure footer h1 h2 h3 h4 h5 h6
header hr i ins kbd li main mark menu meter nav ol p pre progress q rp rt ruby
s samp search section small span strong sub summary sup table tbody td tfoot th
thead time tr u ul var wbr
```

canonical exact attribute allowlist 固定为：

```text
abbr alt class colspan datetime dir draggable headers height hidden high id inert
lang low max min open optimum reversed role rowspan scope span spellcheck start
tabindex title translate type value width
```

此外只允许格式合法的 `aria-*` 和唯一注入标记 `data-promptrepo-slot`。Consumer MUST 直接
复用 Promptrepo validator，不得复制或放宽列表；若未来需要增加 element/attribute，MUST
引入 successor security profile，而不是扩宽 `static-review-fragment-v1`。

#### Scenario: HTML 含事件处理器
- **WHEN** fragment 含任意大小写或转义形式的 `on*` event attribute
- **THEN** validator 必须返回 `UI_TEMPLATE_HTML_FORBIDDEN`

#### Scenario: HTML 含外部图片
- **WHEN** fragment 含 `src`、`href` 或等价 URL-bearing attribute
- **THEN** validator 必须 fail closed

#### Scenario: HTML 含 consumer framework directive
- **WHEN** fragment 含 `hx-*`、`x-*`、任意未列入 static-review profile 的 attribute，或除 `data-promptrepo-slot` 外的 `data-*`
- **THEN** validator 必须以 `UI_TEMPLATE_HTML_FORBIDDEN` fail closed

#### Scenario: HTML 含 template expression
- **WHEN** fragment 的 text 或 attribute 含 Mustache、EJS 或 `${...}` template delimiter
- **THEN** validator 必须以 `UI_TEMPLATE_HTML_FORBIDDEN` fail closed

### Requirement: CSS validation 必须拒绝外部或可执行语法
validator MUST 拒绝 `@import`、`url()`、`expression()`、`behavior`、`-moz-binding`、外部 font/source、HTML breaking sequence 与 CSS parser error。

#### Scenario: CSS 使用注释或转义绕过 url
- **WHEN** CSS 以大小写、注释或 escape 形式表达 `url()`/`@import`
- **THEN** tokenizer/parser 必须识别并返回 `UI_TEMPLATE_CSS_FORBIDDEN`

### Requirement: Digest 和 snapshot 必须精确验证
SDK MUST 对 canonical bundle byte stream 计算 sha256 content digest，并在 exact load 时同时校验 address digest、bundle digest 与 repository snapshot。

#### Scenario: body 被修改但 metadata 未变
- **WHEN** HTML/CSS 任一 byte 变化而 address digest 保持旧值
- **THEN** load 必须返回 `UI_TEMPLATE_DIGEST_MISMATCH`

### Requirement: Inspect 和机器投影不得泄露正文
`InspectUITemplate` 与 UI template JSON/agent projection MUST 只输出 metadata、slots、limits、digest/snapshot、body sizes 与 validation result，MUST NOT 输出 HTML/CSS body、consumer supplied values 或私有绝对路径。

#### Scenario: JSON 序列化 bundle inspection
- **WHEN** consumer 将 inspection 编码为 JSON
- **THEN** 输出必须包含 digest/size/validation
- **THEN** 输出不得包含 fragment、CSS 或注入值

### Requirement: Promptrepo UI template API 必须无执行与网络副作用
parse、format、validate、inspect 和 load MUST NOT 执行 HTML/CSS/JavaScript、发起 provider/network 请求、改变 Registry/consumer state 或读取 credential。

#### Scenario: 校验恶意 bundle
- **WHEN** caller 校验包含危险内容的 bundle
- **THEN** SDK 必须仅返回有界 validation error
- **THEN** 不得执行、请求或持久化 bundle 内容
