## ADDED Requirements

### Requirement: 显式跳过初始化

Manager MUST 在 SkipInitialization 为 true 时不创建配置或缓存目录；默认构造行为保持兼容。

#### Scenario: 读取不存在的目录
- **WHEN** consumer 显式选择 SkipInitialization 并查询仓库列表
- **THEN** 构造不产生本地目录或状态文件
