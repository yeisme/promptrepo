# 验证记录

验证日期：2026-08-31

## 结论

`promptrepo-ui-template-contract-v1` 已完成本地非生产实现和最终质量门。新增能力是独立、
additive 的 `kind=ui-template` 地址、bundle、校验、摘要与 inspect/load 合同；既有
`ParseRef`、`TemplateAddress`、`TemplateRole`、`Client`、catalog/state schema、structured
document 和冻结 digest 未改变。

## 实现证据

- 地址：`UITemplateAddress` parser/formatter 固定
  `kind,locale,role,path,digest,snapshot` 顺序，拒绝 userinfo、fragment、selector、未知或
  重复 query、absolute/traversal path 与非法 digest/locale；旧 parser 继续拒绝
  `kind=ui-template`。
- 安全合同：`UITemplateBundleV1`、slots、`static-review-fragment-v1`、显式 per-bundle
  limits 和稳定错误码已实现；这些 ceiling 不构成全项目资产数量上限。
- HTML/CSS：HTML canonical element/attribute allowlist + bounded template-aware lexer +
  explicit tag stack；CSS tokenizer + parser + delimiter/comment/escape 检查。Framework/template
  directive、未授权`data-*`、重复属性、namespace、非 void 自闭合、未闭合和错序闭合均
  fail closed；allowlist防漂移测试覆盖全部v1元素/属性，不 sanitizer rewrite，不执行或请求网络。
- 摘要：长度分隔 SHA-256 覆盖 canonical address identity、normalized metadata 与原始
  HTML/CSS bytes；content digest 与 repository snapshot 分离并在 exact load 同时验证。
- 投影：bundle JSON/YAML、inspection 和错误消息不包含 HTML/CSS、consumer values 或私有
  absolute path。
- 本地 loader：`uitemplatefs` 只读 bounded regular files，检查 root containment、
  symlink、tamper、digest 和 snapshot，无网络/provider/state side effect。

## 兼容与依赖分类

- 公共面：新增 exported types/functions/interfaces/error codes；无既有 exported identifier
  删除、重命名、retype 或语义扩宽。
- 输入边界：`ParseTemplateAddress` 保持只接受 `kind=template`；新 parser 与旧 parser
  互斥。
- 持久化：无 state/catalog migration，无新增 durable SDK state。
- 依赖：新增 pure-Go `github.com/tdewolff/parse/v2 v2.8.16`，复用其 HTML/CSS lexer/parser；
  不引入 `golang.org/x/net/html`。`golang.org/x/text` 保持 Go 1.24 兼容线的 `v0.34.0`；
  `CGO_ENABLED=0` 测试与构建通过。
- 回滚：consumer 停止消费 optional UI template interfaces 即可；旧 template/document
  路径和 persisted state 不需要回退迁移。

## 执行命令

以下命令均从 `shared/promptrepo` 执行并通过：

```bash
GOTOOLCHAIN=go1.24.13 GOWORK=off go mod verify
GOTOOLCHAIN=go1.24.13 GOWORK=off go test ./... -count=1
GOTOOLCHAIN=go1.26.6 GOWORK=off go test -race ./... -count=1
GOTOOLCHAIN=go1.26.6 GOWORK=off go vet ./...
CGO_ENABLED=0 GOTOOLCHAIN=go1.24.13 GOWORK=off go test ./... -count=1
CGO_ENABLED=0 GOTOOLCHAIN=go1.24.13 GOWORK=off go build ./...
openspec validate promptrepo-ui-template-contract-v1 --strict --no-interactive
git diff --check
```

定向 dangerous-corpus、compatibility、digest/snapshot、redaction 与 loader tests 同样通过。

安全扫描使用修复标准库漏洞的 Go 1.26.6：

```bash
GOTOOLCHAIN=go1.26.6 GOWORK=off govulncheck ./...
```

扫描无标准库 finding，也没有 UI template 或 `golang.org/x/net/html` finding；module graph 已
移除 `golang.org/x/net`。命令仅报告既有 `GO-2026-5970`，这是 Go 1.24 compatibility line
上的 `golang.org/x/text v0.34.0` 静态可达结果；`TestNormalizeSearchHandlesInvalidUTF8`
验证所有输入在进入 `norm.Form.String` 前经过 `strings.ToValidUTF8`。该单项已按 release
policy 记录为临时例外，升级 Go compatibility floor 后必须改用 `x/text v0.39.0+` 并移除例外。

## 脏工作区归因

验证前已存在且未修改/回退的用户工作包括 generated OpenSpec skill copies、
`credentialctl-usage`、`.understand-anything`、`docs/openspec-handoff-schema-design.md` 与
`source/credential.go`/`source/credential_test.go`。本 change 只拥有 UI template 新文件、
对应公共错误码/依赖、文档和本 OpenSpec 目录。

## 未授权动作

未 commit、push、tag、publish、deploy，也未访问 live Registry、Scaena 或 provider。
