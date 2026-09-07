## Why

外部 Agent 需要可复现的多步骤提示包。现有 SDK 只拥有单模板渲染；新增合同必须避免改变已发布引用、字段与输出行为。

## What Changes

- 增加实验性 RecipeV1、步骤输入绑定、PromptPackageV1 和内容校验函数。
- DAG 拓扑顺序、跨平台路径与逐文件 digest 使用公共纯函数验证。
- 模型调用、会话、资料导入和文件落盘留在 Registry owner。

## Capabilities

### New Capabilities
- `prompt-package-contract`: 确定性多步骤与可搬运提示包合同。

### Modified Capabilities

无；已有合同保持原形状。

## Impact

分类为 split-owner。公开 SDK 是合同唯一 owner，Registry 是消费实现 owner。新接口处于 v0.1 实验状态，不修改历史 catalog、TemplateContract 或 RenderTemplate。
