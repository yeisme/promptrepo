# 统一模板发现

## Why

用户需要一条列表命令直接看见全部已登记来源的具体模板。旧仓库列表只打印成功，过期缓存与方案／角色计数混用使模板看似丢失。

## What Changes

共享发现请求与结果、独立角色标注、TTL 缓存、并发刷新、来源配置 patch、Eikona 元数据适配器。

## Capabilities

### New Capabilities

- `unified-template-discovery`: 可追踪来源、数量与可用状态的模板元数据发现。

### Modified Capabilities

无删除、改名或收窄已有合同。

## Impact

关联变化：promptrepo-unified-discovery-v1, template-registry-unified-discovery-v1, eikona-template-discovery-v1, official-template-discovery-metadata-v1。Registry、共享 SDK、Eikona 和官方内容分别维护自己的状态，不迁移私有数据。
