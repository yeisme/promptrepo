## 1. 地址与公共 DTO

- [x] 1.1 实现 `UITemplateAddress` parser/formatter；owner：Promptrepo implementer；scope：新增 public Go API/internal shared parsing helpers；依赖：无；并行 lane：address；验收：kind/query order/path/digest/snapshot 严格，旧 `ParseTemplateAddress` 仍拒绝 ui-template；验证：`go test ./... -run 'UITemplateAddress|TemplateAddress'`；预期：PASS；失败复查：运行旧 compatibility tests 确认没有扩宽旧 parser。
- [x] 1.2 实现 `UITemplateBundleV1`、slot/security/limits/inspection/load DTO 与稳定错误码；owner：Promptrepo implementer；scope：public contract；依赖：1.1；并行 lane：contract；验收：ceiling、schema version、zero-value/error semantics 与 docs 一致；验证：`go test ./... -run 'UITemplateBundle|UITemplateLimits|UITemplateError'`；预期：PASS；失败复查：不得修改 `TemplateRole` 或 catalog digest shape。

## 2. 校验与摘要

- [x] 2.1 实现有界 HTML fragment validator；owner：Promptrepo implementer；scope：pure-Go validation；依赖：1.2；并行 lane：validator；验收：禁止标签/属性/URL/slot/UTF-8/bounds 全部 fail closed，不 sanitizer rewrite；验证：`go test ./... -run 'UITemplateHTML'`；预期：PASS；失败复查：增加大小写、entity、escape、malformed corpus，复杂逻辑写中文注释说明安全不变量。
- [x] 2.2 实现有界 CSS tokenizer/validator；owner：Promptrepo implementer；scope：pure-Go validation；依赖：1.2；并行 lane：validator（与 2.1 文件 lease 不重叠时并行）；验收：识别 `@import`、`url()`、expression/behavior/binding、注释/转义绕过和 parser errors；验证：`go test ./... -run 'UITemplateCSS'`；预期：PASS；失败复查：不得用简单大小写正则作为唯一安全门。
- [x] 2.3 实现 canonical bundle digest 与 snapshot binding；owner：Promptrepo implementer；scope：digest utility；依赖：1.2、2.1–2.2；并行 lane：digest；验收：metadata/body 任一变化改变 digest，长度分隔防歧义；验证：`go test ./... -run 'UITemplateDigest|UITemplateSnapshot'`；预期：PASS；失败复查：跨平台换行/路径不得造成非规范差异。
- [x] 2.4 实现 inspect/load 与 machine body redaction；owner：Promptrepo implementer；scope：interfaces + filesystem fixture loader；依赖：2.3；并行 lane：loader；验收：inspect/JSON 不含 body/values/absolute path，load 精确验证后才返回 body；验证：`go test ./... -run 'InspectUITemplate|LoadUITemplate|UITemplateRedaction'`；预期：PASS；失败复查：用 unique sentinel 扫 stdout/error/JSON。

## 3. 兼容、文档与验证

- [x] 3.1 增加旧 public API/golden compatibility suite；owner：Promptrepo implementer；scope：`ParseRef`、TemplateAddress、TemplateRole、structured document/catalog digests；依赖：1–2；并行 lane：compatibility；验收：旧 fixtures 输出 byte/semantic compatible；验证：`go test ./... -run 'Compatibility|TemplateAddress|StructuredDocument'`；预期：PASS；失败复查：新能力必须移回独立类型，不修改旧输入边界。
- [x] 3.2 更新 README/architecture/release consumer guide；owner：Promptrepo implementer；scope：human docs；依赖：API 稳定；并行 lane：docs；验收：说明 Promptrepo 不渲染/发行/执行、Scaena/Registry owner、正文省略与 exact ref；验证：`git diff --check` 和示例 parser test；预期：一致；失败复查：CLI/协议字段保持 English，说明文档用中文。
- [x] 3.3 执行最终验证；owner：Promptrepo verifier；scope：稳定 diff；依赖：3.1–3.2；并行 lane：serial-final；验收：test/race/vet/build/OpenSpec strict 全绿；验证：`go test ./... && go test -race ./... && go vet ./... && CGO_ENABLED=0 go build ./... && openspec validate promptrepo-ui-template-contract-v1 --strict --no-interactive`；预期：PASS；失败复查：先分类既有 dirty worktree，不修改无关 contract。
- [x] 3.4 执行发布候选安全复核；owner：Promptrepo verifier；scope：UI template dependency/validator；依赖：3.3；并行 lane：serial-final；验收：使用 patched Go toolchain 扫描时 UI template 不产生 reachable finding，`golang.org/x/net/html` 不在 module graph；验证：`GOTOOLCHAIN=go1.26.6 GOWORK=off govulncheck ./...`、`GOWORK=off go list -m all`、dangerous-corpus tests；预期：仅保留既有且由 invalid-UTF-8 guard 覆盖的 `GO-2026-5970`，无其他 reachable issue；失败复查：不得通过提高 Go 1.24 compatibility floor 隐式规避依赖问题。
