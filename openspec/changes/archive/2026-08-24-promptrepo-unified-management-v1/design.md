## Context

Prompt Repository 的 canonical owner 已明确分层：embedded engine 持有 user repository profile，Template Registry 持有 organization repository/policy，领域项目持有 project binding，session override 只存在于本次调用。公共 SDK 需要组合这些输入，但不能成为新的中央数据库。

## Goals / Non-Goals

**Goals:**

- 以 additive、纯 Go、CGO-free 合同表达四层 RepositorySet。
- 把 preference ranking 与 admission policy 分成两个确定性阶段。
- 让 embedded 与未来 service mode 使用同一请求、结果、reason code 和安全投影。
- 保持 stdout/events/evidence 所需 DTO 不含正文、凭据和私有 source。

**Non-Goals:**

- 不在 SDK 持久化 organization、project 或 session 状态。
- 不实现 Registry RBAC、组织数据库、credential resolver 或 domain install。
- 不修改既有 state schema、state path、`RepositoryProfile.Scope=user` 语义或 `Client`。
- 不执行 provider、模板 compiler 或领域工作流。

## Decisions

### 1. RepositorySet 是调用时投影

`EffectiveRepositorySetRequest` 接收 caller-supplied bindings 与安全 candidate metadata。embedded engine 注入本机现有 user profiles，并拒绝 caller 伪造 user candidate；其它 scope 仍由对应 owner 提供。

返回的 `RepositorySet` 只包含 repository ID、scope、selection source、trust、health、policy digest 和 scope ref digest。原始 `scope_ref`、source URL、credential ref、模板正文和本地路径不进入结果。

### 2. 排序与准入严格分离

排序优先级固定为：

```text
session exact selection
project pin
user preference
organization default
official fallback
```

同层按较高 `priority`、声明顺序、repository ID 稳定排序；同一 repository 首次出现即锁定位置。排序不会设置 `allowed=true`。

`EvaluateRepositoryPolicy` 独立检查：enabled/health、每层 repository allow/deny、operation allow/deny、minimum trust、rights allow/deny 与 required capabilities。每个 policy 都是约束，后层 allow 不能扩宽前层 deny。

### 3. optional interfaces 保持旧 Client 不变

新增：

```text
RepositorySetReader.EffectiveRepositorySet(context, request)
PolicyEvaluator.EvaluatePolicy(context, request)
```

同时导出纯函数，Registry service 和测试 fixture 可直接复用。旧 consumer 不 type-assert 新接口时行为完全不变。

### 4. 管理投影是安全最小集

`ManagementProjection` 只允许 operation、command identity、scope digest、mode、repository ID、health/trust/readiness、policy/snapshot digest、stable reasons 与注册 action。结构中没有 body、rendered body、input values、credentials、Authorization、source URL、provider payload 或推理字段。

### 5. Schema 与兼容规则

首版 schema 常量：

```text
promptrepo.repository-set.v0.1
promptrepo.policy-decision.v0.1
promptrepo.management-projection.v0.1
```

schema version 与 Go module minor 分离；未来只能 additive 增字段或发布新 schema version。reason code、operation ID、scope、mode、selection source、readiness 枚举均为稳定英文机器合同。

## Risks / Trade-offs

- [调用方伪造 user state] → embedded Manager 忽略 caller-supplied user candidate，并从现有 state 注入 canonical user metadata。
- [project allow 覆盖 organization deny] → 所有 policy 独立求交，任意 deny 都保留。
- [exact selection 绕过 admission] → exact 只改变第一排序位置，仍必须调用相同 evaluator。
- [投影泄漏] → DTO 不接受敏感字段，并以 JSON/YAML sentinel tests 固定字段面。
- [state migration 风险] → EffectiveRepositorySet 只读，不调用 `withWriteState`，state bytes 测试保持不变。

## Migration Plan

1. 固定 v0.3 public compatibility baseline。
2. 新增 DTO、常量、pure composition/evaluation/projection。
3. 新增 embedded Manager optional interface 实现和 conformance fixtures。
4. 完成本地纯 Go gates 与 strict OpenSpec validation。
5. 后续在明确 release gate 下发布 v0.4 RC；消费者 canary 通过后再发布 stable。
