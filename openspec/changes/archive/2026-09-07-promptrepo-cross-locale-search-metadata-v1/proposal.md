## Why

官方模板只注册英文可编译正文，同时保留中文标题、摘要和 aliases 供人类发现。当前搜索在 `locale=en` 时只匹配英文显示文本，导致中文用户无法找到仍应返回英文 exact ref 的模板。

## What Changes

- 搜索评分使用 solution 的全部 locale 显示 metadata。
- 返回 title、summary、locale 和 exact ref 仍使用请求 locale 的现有选择逻辑。
- 现有 tags、capabilities、trust、兼容过滤和排序字段不变。

## Capabilities

### New Capabilities

- `cross-locale-display-search`: 允许人类使用中文显示词搜索英文可编译模板。

### Modified Capabilities

None.

## Impact

这是增量检索修复：公共 Go API、JSON 字段、exact ref、catalog schema 和请求默认值均不变化。旧查询的当前-locale 分数保持不变；其他 locale 有更强匹配时可能新增结果或提高排序。
