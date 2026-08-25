## 上下文与边界

v0.2.0 的 `promptrepo://repository/package/solution@version?locale=…` 是已发布的 solution ref。它只能标识 solution，不能安全、规范地描述某个模板 role、路径、选择器或已读快照。本变更保留该 ref 原样，并在旁路增加 `Address`；它绝不复用 file/git/github/s3 repository source URI 的语法或权限含义。

```mermaid
flowchart LR
  C[owning CLI consumer] --> I[Inspector: catalog/snapshot only]
  C --> P[Previewer]
  P --> R[ReadTemplate: bounded source read]
  R --> V[declared input validation]
  V --> M[Renderer: in-memory {{name}} render]
  I --> O[safe inspect projection]
  M --> O2[safe preview projection\nbody JSON/YAML omitted]
  O2 -. no provider / no state write / no usage .-> X[no external execution]
```

## 决策

1. **地址是独立的加性语法。** `ParseRef` 和 `FormatRef` 不改动。`ParseTemplateAddress` 仅接受无 userinfo、无 fragment 的 `promptrepo://` identity，query key 仅为 `kind,locale,role,path,selector,digest,snapshot`，canonical formatter 按该顺序输出。`kind=template` 与 `locale` 必需；role/path/selector/digest/snapshot 可选。path 必须是无 traversal 的相对路径；selector 必须以 `heading:`、`json-pointer:` 或 `yaml-pointer:` 开头。digest/snapshot 仅接受 sha256 digest。
2. **contract 不进入 catalog/state。** 已发布 `TemplateRole`、`Solution` 和 state DTO 的字段、tag、unkeyed-composite 形状均保持。新增 `TemplateContract` 由调用方随 inspect/validate/preview 显式提供并在结果返回；非空 digest 必须是小写 `sha256:<64 hex>`，空 digest 允许草稿合同。独立 `ContractResolver` 按 `<solution>/prompts/...` → `<solution>/contracts/<role>.<locale>.json` 约定读取 Template Registry 生成的 companion，统一限制为 4 MiB valid UTF-8 JSON，并校验 sidecar 自身 digest 与 catalog template binding；它不写 catalog/state。Git/GitHub exact commit 返回 `snapshot_pinned`；file/S3 当前对象返回 `content_bound`，让 policy/consumer 能区分一致性保证。
3. **输入值从不回显。** inspect 返回 caller-supplied contract 的字段定义和 `supplied/default/missing` 状态，不返回值。敏感字段禁止 default/example/enum；preview 的渲染正文标记为 `json:"-" yaml:"-"`。错误/issue 仅使用安全 code、field 与静态 message。输入未就绪时 next action 为 `supply_inputs`；rights 为 blocked/prohibited 时必须返回 `blocked`，不得误导 consumer 继续补输入或预览。
4. **Render 与 Preview 分离。** `Renderer.Render` 仅对调用方提供的内存正文执行严格替换，既不读 source 也不写 state；`Preview` 可经既有受 digest 校验的 adapter 读取模板后调用 Render。两者均不调用 provider、不创建 run、不记录 usage；Preview 显式报告三个 false flags。selector grammar 仅预留；Inspect 作为 metadata 返回，Preview 对非空 selector 以 `SELECTOR_UNSUPPORTED` fail closed，直到 selector engine 有 conformance 测试。
5. **新接口不扩大旧 Client。** 以 `ContractResolver`、`Inspector`、`Previewer`、`Renderer`、`Validator` 的独立接口表达 capability；`engine.Manager` 编译期声明实现。消费者可通过类型断言选择升级。source `Adapter` 也保持原方法集合，built-in file/Git/S3 adapter 通过可选 `source.CompanionReader` 提供 bounded sidecar read，避免破坏第三方 adapter。

## 输入规则

名称使用 `[A-Za-z][A-Za-z0-9_]*`。类型为 `string`、`integer`、`number`、`boolean`、`enum`。numeric 的 min/max、文本的 min_length/max_length、string/enum 的 regex、enum 的允许值和 locale labels/descriptions 在 caller-supplied TemplateContract 验证。供给值同时校验未知字段、必填、类型、enum 和约束。严格 scanner 只接受 `{{name}}`，拒绝 triple/stray closing brace run；未声明 placeholder、非法 brace 及有问题的输入使 preview `ready=false`。

## 兼容性、发布与回滚

| 表面 | 分类 | 兼容措施 |
| --- | --- | --- |
| public Go API | additive | 新 symbols/interfaces；`Client` 不动 |
| catalog JSON | unchanged | 不增加字段或 tag；保留 schema `promptrepo.catalog.v0.1` |
| state | unchanged | 不读写新字段，不变更 schema/path |
| source URI/ref | additive | source URI 不变，旧 Ref parser/formatter 不变 |

计划在未来 v0.3.0 发布；不存在弃用窗口，因为无删除或改名。发现回归时回退到 v0.2.0/还原此变更，不需要 state 回滚。

## 风险与未决项

- 模板正文可包含未声明或不严格的 brace：preview 明确报安全 issue，不会猜测替换。
- Go 1.24 兼容线无法直接采用要求 Go 1.25 的 `x/text` v0.39.0；所有进入 NFKC 的文本先经 `strings.ToValidUTF8`，并由 invalid-UTF8 回归覆盖。`govulncheck` 无法建模该 guard，发布时只允许这一项已解释 finding，且支持工具链标准库必须无其它可达漏洞；提高 Go floor 后立即升级依赖并移除例外。
- 自定义 source adapter 若未实现 `source.CompanionReader`，contract resolution 明确返回 unsupported，不回退到任意路径读取或正文猜测。
- CLI 的 JSON/YAML envelope、美化输出和本地 `--output` 文件属于 Sonora/Eikona 消费方，SDK 只提供安全 payload 与 tags。
- user/project scoped source alias 将在后续 profile/source 变更中设计，phase 1 继续使用 file/git/github/s3。
