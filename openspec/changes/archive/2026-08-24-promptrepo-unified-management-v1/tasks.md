## 1. 公共合同

- [x] 1.1 新增 schema、scope、mode、selection source、readiness、reason code 与 operation ID 常量；验收：机器值稳定英文；验证：`go test ./... -run 'ManagementContract|OperationID' -count=1`。
- [x] 1.2 新增 additive RepositoryScopeBinding、RepositorySet、PolicyConstraint、PolicyDecision 与 ManagementProjection DTO；验收：既有 `Client` 和 state struct 不变；验证：`go test ./...`。
- [x] 1.3 新增 RepositorySetReader 与 PolicyEvaluator optional interfaces；验收：旧 consumer 无需实现；验证：compile-time interface tests。

## 2. 组合与策略

- [x] 2.1 实现 deterministic RepositorySet composition；验收：session/project/user/organization/official 顺序固定、同 ID 去重、policy digest 稳定；验证：`go test ./... -run 'RepositorySet' -count=1`。
- [x] 2.2 实现 deny-wins policy evaluator；验收：exact ref 不绕过 deny，project allow 不扩宽 organization deny，health/trust/rights/capability/operation blockers 有稳定 reason；验证：`go test ./... -run 'Policy' -count=1`。
- [x] 2.3 实现 body-free management projection；验收：JSON/YAML sentinel scan 不含敏感字段和值；验证：`go test ./... -run 'ManagementProjection' -count=1`。

## 3. Embedded engine 与兼容性

- [x] 3.1 embedded Manager 注入 canonical user state 并实现 EffectiveRepositorySet；验收：不接受 caller 伪造 user candidate、不写 state；验证：`go test ./engine -run 'EffectiveRepositorySet' -count=1`。
- [x] 3.2 embedded Manager 实现纯 EvaluatePolicy adapter；验收：无 provider、无 durable write；验证：`go test ./engine -run 'EvaluatePolicy' -count=1`。
- [x] 3.3 补齐 conformance fixture、v0.1-v0.3 regression 和 leak tests；验证：`go test ./...`。

## 4. 文档与验证

- [x] 4.1 更新 README、architecture 和 CHANGELOG 的 unreleased public surface；验收：不声称 v0.4 已发布。
- [x] 4.2 strict OpenSpec 与纯 Go gates；验证：`openspec validate promptrepo-unified-management-v1 --strict --no-interactive && go mod verify && go test ./... && go vet ./... && CGO_ENABLED=0 go test ./... && CGO_ENABLED=0 go build ./...`。
