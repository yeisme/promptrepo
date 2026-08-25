## 为什么

`promptrepo` 已在 v0.2.0 提供精确 solution ref、catalog snapshot 与受校验的模板读取，但消费者在执行前仍需自行猜测模板角色、输入字段与可用性。Sonora、Eikona 等拥有 CLI 的消费者需要一套可安全投影为 inspect/preview 输出的 SDK 合同，且不能让预览变成隐式执行、写入 state 或记录用量。

## 能力归属与账本

结论为 **fit**：`promptrepo` 是公开 Go SDK、catalog 与本地模板读取的规范拥有者；各领域 CLI 是 inspect/preview 命令、JSON/YAML/agent 美化渲染及显式 `--output` 文件输出的消费者。此变更不新增 `cmd/promptrepo`。

| 能力 | 状态 | 规范拥有者 | 本次证据 |
| --- | --- | --- | --- |
| 模板不可变地址与 companion 输入合同 | required / deliver-now | promptrepo | parser、contract resolver、SDK 测试 |
| inspect 安全详情 | required / deliver-now | promptrepo engine | 不读取正文的测试 |
| 内存 preview | required / deliver-now | promptrepo engine | 无 provider/state/usage 副作用测试 |
| Sonora/Eikona CLI 渲染 | retain-next | 对应 CLI | 本次只给出稳定 SDK 载荷与示例 |
| user/project source alias | retain-next | source/profile 路由设计 | 现有 file/git/github/s3 不变 |

## 变更内容

- 保持 `Client`、`ParseRef`、`FormatRef`、`TemplateRole`、`Solution`、catalog/state schema/path、既有错误语义与 catalog digest 不变。
- 添加独立的 canonical template `Address` 解析/格式化层。它以既有 solution identity 为底座，查询参数规范顺序为 `kind,locale,role,path,selector,digest,snapshot`；它不是 repository source URI 语法。`kind=template`、相对 path、selector 类型和 digest/snapshot 会被安全校验。
- 以独立 caller-supplied `TemplateContract` 携带输入定义、license、permissions 和可选 contract digest；新增独立 `ContractResolver`，从同一 source 读取 Template Registry 生成的 companion sidecar，严格校验 schema、identity、template path/digest 与 contract digest。Git/GitHub exact commit 标记为 `snapshot_pinned`；file/S3 当前对象标记为 `content_bound`，不虚构 catalog 未提供的历史 contract pin。添加 `Inspector`、`Previewer`、`Renderer`、`Validator` 等独立可选能力接口，`engine.Manager` 实现它们而不修改 `Client`。
- 添加不读取正文的 inspect、只处理调用方内存正文的 Render，及只在内存严格替换 `{{name}}` 的 preview。preview 不调用 provider、不创建 run、不写 state、不记录 usage；正文即使在 Go 内存结果中存在，也不得出现在 JSON/YAML 序列化中。
- 预计作为未来 **v0.3.0** 的加性发布。没有任何既有符号删除或 state migration。

## 影响与回滚

影响仅限本模块的新增公开 DTO、engine 与文档；既有 catalog/state DTO 没有新增字段或 tag。既有 v0.1 fixture 与 v0.2 Client 调用继续工作。若消费者发现回归，回滚方式为还原本变更并继续使用 v0.2.0；state 格式未变化，因此无需数据回滚或迁移。
