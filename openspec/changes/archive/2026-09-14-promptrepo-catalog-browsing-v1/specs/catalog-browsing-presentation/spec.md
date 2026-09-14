## ADDED Requirements

### Requirement: 一致的目录浏览与展示

系统 SHALL 支持本 owner 的已批准范围：新增 BrowseRequest/BrowseResult、多选元数据筛选、跨语言关键词、分组排序、双重计数和共用 conformance 样例。

#### Scenario: 分组筛选并比较统计

- **WHEN** 用户组合元数据筛选和分组
- **THEN** 同维度使用 OR、跨维度使用 AND，统计独立于分页且区分条目和去重内容

#### Scenario: 兼容已有消费者

- **WHEN** 旧客户端继续使用单值参数和既有 JSON 字段
- **THEN** 其类型和内容引用保持不变

#### Scenario: 元数据输出

- **WHEN** 用户浏览或导出目录
- **THEN** 展示元数据与来源，不输出提示词正文或凭据
