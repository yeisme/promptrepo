## 设计

搜索先用已选择 locale 的 title/summary/aliases 计算原分数，再按 locale key 排序遍历其他显示 metadata，采用最高分。相同分数保留当前 locale 的理由，避免 Go map 顺序改变结果。Tags 仍属于 solution 级，只按原权重参与。

结果卡继续显示请求 locale 的文本并生成同 locale exact ref。其他 locale 只作为检索词，不会注册第二模板、改变编译 locale 或让中文 review 文档进入 digest。

## 兼容与回滚

公共字段和 ranking profile 字符串不变。私有 v0.1.0 兼容测试继续锁定原查询分数；新增测试覆盖 `locale=en` 下中文 alias 命中且返回英文卡片。若跨 locale 匹配造成不可接受噪声，回滚单个评分调用即可，catalog 和存储数据无需迁移。
