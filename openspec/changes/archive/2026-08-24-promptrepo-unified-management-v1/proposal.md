## Why

Promptrepo v0.3 已公开 exact address、inspect、validate、render 和 preview，但多个消费项目仍需各自拼装用户仓库、组织默认、项目 pin 与 session override。若没有统一的 additive 合同，偏好排序容易被误当成授权，exact ref 也可能绕过组织 deny、rights、capability 或 trust 约束。

## What Changes

- 新增 `RepositoryScopeBinding`、`RepositorySet`、`PolicyConstraint`、`PolicyDecision` 与安全 `ManagementProjection` DTO。
- 新增可选 `RepositorySetReader` 与 `PolicyEvaluator`，不向既有 `Client` 增方法。
- embedded engine 只读取既有 user state；organization、project、session 输入由调用方提供，本次计算不持久化。
- 明确 `session exact > project pin > user preference > organization default > official fallback` 只用于确定性排序。
- 明确 source health、policy、operation、rights、capability 与 minimum trust 采用 deny-wins admission。
- 新增 stable English `operation_id` 和 body-free、credential-safe management projection。

## Capabilities

### New Capabilities

- `promptrepo-repository-set`: 四层 scope 的确定性有效仓库集合。
- `promptrepo-policy-decision`: 偏好与准入分离的纯策略判定。
- `promptrepo-management-projection`: 跨 renderer 复用的安全管理投影。

### Modified Capabilities

- 无。v0.1-v0.3 public DTO、`Client`、state schema/path、digest 和错误语义保持不变。

## Impact

- 新增纯 Go DTO、纯函数、optional interfaces、embedded engine reader 和 conformance tests。
- Template Registry 后续负责 organization policy/service mode；领域 owner 负责 project binding 和 session input。
- 本 change 不发布版本、不新增 provider 调用、不保存 credential value、template body 或组织/项目状态。
