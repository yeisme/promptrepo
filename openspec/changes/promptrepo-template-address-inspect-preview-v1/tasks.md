## 1. 合同与 catalog（owner: promptrepo implementer）

- [x] 1.1 添加 Address、输入 schema、inspect/preview DTO 与独立 capability interfaces；不得改变 `Client` 或 state schema。验收：旧 consumer compile test 与 ref round-trip 通过。
- [x] 1.2 新增 caller-supplied `TemplateContract`（输入、license、permissions、digest）并保持 `TemplateRole`/`Solution`/catalog/state 形状不变。验收：v0.2 unkeyed literal 编译、state/canonical digest 不漂移；非法 contract schema 拒绝。
- [x] 1.3 新增独立 `ContractResolver` 与可选 `source.CompanionReader`，从相同 source 读取 Registry sidecar；Git/GitHub exact commit 标记 `snapshot_pinned`，file/S3 当前对象标记 `content_bound`。验收：identity/template/contract digest 漂移 fail closed，旧 `Client`/`Adapter` 方法集不变。

## 2. engine 行为（owner: promptrepo implementer，依赖 1）

- [x] 2.1 实现无正文读取的 inspect readiness 与安全 next action。验收：结果不带 body/value，字段状态和 rights/metadata 正确。
- [x] 2.2 实现 provider-free Render 和零 provider/zero-write 内存 preview、严格 scanner `{{name}}` 替换与 byte/rune/digest 统计。验收：default/type/enum/missing/unknown/placeholder/brace 路径覆盖，selector fail closed，JSON 不泄漏 body，state 字节不变。

## 3. 文档与验证（owner: promptrepo implementer，依赖 1、2）

- [x] 3.1 更新中文优先 README/architecture/docs index，说明地址与 source URI 的两层边界、未来 v0.3.0、caller-supplied contract、selector staged/fail-closed、非执行 preview 和 `--output` 消费者责任；不宣传未发布 CLI 命令。
- [x] 3.2 运行 `GOWORK=off go test ./...`、`GOWORK=off go vet ./...`、`GOWORK=off go mod verify`、CGO-disabled test/build、两项严格 OpenSpec 验证和 `git diff --check`。失败时先归因，且不得通过改变旧合同规避。
- [x] 3.3 使用 Go 1.24.13 与 Go 1.26.7 运行兼容/race/build 门禁，并执行 `govulncheck ./...`。验收：支持工具链无标准库 finding；仅保留 GO-2026-5970 这一项已由 `strings.ToValidUTF8` 输入 guard 和回归测试覆盖的静态 finding，提高 Go floor 后升级 `x/text` 并移除例外。
