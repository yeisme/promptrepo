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
禁止 traversal；selector 只能是 `heading:`、`json-pointer:`、`yaml-pointer:` 或
`jsonl-id:`；
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
但 Preview 对任何非空 selector 继续返回 `SELECTOR_UNSUPPORTED`。新的结构化定位走
独立 `DocumentSelector`，不改变 v0.3 Preview 的正文渲染语义。

## 结构化文档边界

结构化能力是旧模板读取旁边的一条 additive 路径：

```text
catalog TemplateRole
    + exact snapshot
    + contracts/documents/<role>.<locale>.document.json
    -> descriptor binding
    -> bounded source read
    -> strict JSON/YAML/JSONL parser
    -> source digest + canonical digest
    -> optional deterministic selector
```

`DocumentResolver` 只解析并绑定 descriptor；`DocumentLoader` 读取完整文档；
`DocumentSelector` 才消费 selector。带 selector 的地址传给 `LoadDocument` 会失败关闭，
避免调用者误以为已经完成局部选择。三个接口都由 `engine.Manager` 加性实现，旧
`Client` 和 `source.Adapter` 方法集不变；built-in source 继续通过可选
`source.CompanionReader` 从同一 file、exact Git commit 或 S3 object 读取 descriptor。

JSON/YAML canonical digest 使用 RFC 8785 JCS。YAML 先以 AST 检查并拒绝 alias、
anchor、merge key、显式 tag、非字符串 map key、重复键与非有限数值，再转换到 JSON
数据模型。JSONL 在受限 source bytes 上逐行解析、校验唯一 record ID 并增量计算 segment
digest；加载结果不物化整个 JSONL parsed-node 数组，按 ID 选择时才解析目标记录。

`LoadedDocument.Body`、`LoadedDocument.Value`、`SelectedDocument.Body` 和
`SelectedDocument.Value` 是 body-bearing 内存字段，均排除 JSON/YAML serialization。
descriptor 中的 Schema 与 compiler profile 只作为 exact ref/digest lineage 返回；
Promptrepo 不联网解析 Schema、不执行 repository supplied compiler，也不持久化正文、
parsed node 或选择结果。

## UI template 安全与 owner 边界

UI template 是已有 prompt template/document 路径旁边的 additive artifact family：

```text
UITemplateAddress (kind=ui-template, exact digest + snapshot)
    -> body-free .ui-template.json descriptor
    -> bounded relative HTML/CSS reads
    -> HTML lexer + explicit structure validation + CSS tokenizer/parser validation
    -> canonical content digest + snapshot binding
    -> body-free inspection | in-memory exact load
```

`ParseUITemplateAddress` 与 `ParseTemplateAddress` 是互斥 parser：前者只接受
`kind=ui-template`，后者继续只接受 `kind=template`。地址、descriptor 和 inspection
只能携带相对路径；`uitemplatefs.Loader` 在读取前后检查 root containment、regular file、
symlink 和大小，且错误消息不回显绝对路径或正文。

HTML 的 slot marker 固定为 `data-promptrepo-slot`。验证器通过有界 lexer 与显式 tag stack
只接受结构完整、显式闭合的 fragment；只接受已声明且唯一的 marker，required slot 必须
出现；展示元素与静态属性均使用由 capability spec 固定的 allowlist，禁止 document node、namespace、可执行/外部
节点与属性、自定义 element、重复属性、framework/template directive、除 slot marker 外的
任意 `data-*` 和 inline style。CSS 同时经过 tokenizer 与 parser，并补充 delimiter、
comment 和 escape 检查；校验失败直接拒绝原 bytes，不生成 sanitizer 后的第二份内容。
Registry 与 Scaena 必须直接复用 Promptrepo validator；新增 element/attribute 需要新的
security profile，不能原地放宽 `static-review-fragment-v1`。

摘要包含 schema、去除 digest/snapshot 后的 canonical address identity、按 name 排序的
slot、安全 profile、limits 与原始 HTML/CSS bytes。它不包含 `ContentDigest` 自身，也
不把 snapshot 混入 content identity。这样 Registry 可以独立固定 repository snapshot，
consumer 又能判断内容是否被替换。

Promptrepo 不拥有发行目录、CAS、安装状态、浏览器、CSP、Scaena action 或审阅状态。
Template Registry 消费公共 DTO/validator 生成和发行 machine metadata；Scaena 在运行时
复验后注入自己的 safe components。每 bundle 的 byte/slot ceiling 是输入安全边界，不是
全项目资产数量上限。

## RepositorySet 与 policy 边界

统一管理是调用时投影，不是新的 durable store：

```text
embedded user state -----------+
Registry organization input ---+
domain project binding --------+--> RepositorySet --> PolicyDecision
session exact override --------+
```

`RepositorySetReader` 与 `PolicyEvaluator` 是独立 optional interfaces，不进入旧
`Client`。embedded Manager 只注入现有 user profile/health；caller-supplied user
candidate 不得覆盖相同 ID 的 canonical metadata。organization、project、session
binding 只在请求内存在，计算不会触发 `withWriteState`。

排序固定为 `session exact > project pin > user preference > organization default >
official fallback`，但排序结果不包含授权。准入由 `EvaluateRepositoryPolicy` 对每个
policy constraint 求交；organization deny 不能被 project allow 扩宽，exact/pin 也不能
绕过 health、operation、trust、rights 或 capability blocker。

三个安全 schema 分别为 `promptrepo.repository-set.v0.1`、
`promptrepo.policy-decision.v0.1` 和 `promptrepo.management-projection.v0.1`。
RepositorySet 输出把 raw `scope_ref` 单向摘要为 `scope_ref_digest`；ManagementProjection
没有 source URL/path、credential、body、rendered body、input values 或 provider payload
字段。embedded/service renderer 必须从这一投影生成 human/agent/JSON/YAML/events，不能
各自重算 readiness。

## Compatibility invariants

- `promptrepo.state.v0.1`, `promptrepo.catalog.v0.1`, and
  `promptrepo.stage_receipt.v0.1` remain unchanged through the extraction.
- Config and cache defaults remain `yeisme/promptrepo` below the OS user config
  and cache directories.
- Digest algorithms, error-code strings, JSON tags, and exact-ref format remain
  compatible with private v0.1.0.
- State uses a cross-process lock and atomic rename; unsupported future state
  schemas fail closed.
- 模板 Address 和 inspect/validate/render/preview 已在 v0.3.0 additive 发布；没有
  state migration，v0.2.0 仍可作为模块级回滚版本。
- structured document DTO/interfaces、错误码和 descriptor path 是下一 additive
  minor 的开发中表面；未引入 state migration。回滚时 consumer 停用这些可选接口，
  旧 `ReadTemplate`/`Render`/`Preview` 路径不需要数据迁移。
- `UITemplateAddress`、`UITemplateBundleV1`、validator 与 filesystem fixture loader 是
  独立 additive 表面；旧 `TemplateAddress`、`TemplateRole`、catalog digest、`Client`
  和 durable state 均不改变。回滚只需停止消费 `kind=ui-template`。
- RepositorySet、PolicyDecision 和 ManagementProjection 是同一后续 additive minor
  的开发中表面；只读调用不改变 state bytes，旧 consumer 无需 type-assert 新接口。

The `compatibility_v010_test.go` test executes a frozen v0.1.0-shaped fixture
without private-network access and pins the observable result.
