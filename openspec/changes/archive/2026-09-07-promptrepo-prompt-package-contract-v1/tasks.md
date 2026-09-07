# 实施与验收任务

由 scripts/openspec-tasks.py 维护状态。

- [x] contract 增量方案、步骤绑定和包合同 | evidence: Experimental additive contracts; existing public APIs preserved
- [x] verify 纯函数、路径、篡改与兼容性验证 | evidence: Standalone CGO_ENABLED=0 go test ./... and go vet ./... passed
