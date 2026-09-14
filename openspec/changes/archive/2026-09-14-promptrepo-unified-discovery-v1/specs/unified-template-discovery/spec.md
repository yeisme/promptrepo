## ADDED Requirements

### Requirement: 可追踪的统一模板发现

系统 SHALL 提供本 owner 对统一模板发现的支持：共享发现请求与结果、独立角色标注、TTL 缓存、并发刷新、来源配置 patch、Eikona 元数据适配器。

#### Scenario: 枚举已登记来源

- **WHEN** 用户请求模板目录
- **THEN** 结果提供元数据、来源与可用状态，且不输出提示词正文

#### Scenario: 兼容已有消费者

- **WHEN** 旧消费者读取已有引用、结构或内容哈希
- **THEN** 既有合同保持可用，新增发现标注不改变执行授权
