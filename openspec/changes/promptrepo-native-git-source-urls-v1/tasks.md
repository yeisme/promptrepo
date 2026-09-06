# 实施与验收任务

由 scripts/openspec-tasks.py 维护状态。

- [x] contract Document additive native Git source URL contract and rollback | evidence: OpenSpec change validates strictly; legacy schemes and rollback are documented
- [x] implementation Add shared source detection and normalization with legacy compatibility | evidence: Native HTTP/HTTPS/SSH/SCP and bare GitHub sources share add-time and sync-time validation; Promptrepo tests pass
- [x] consumer Update Template Registry examples, integration canary, and operator Skill | evidence: Registry public docs and public canary use native HTTPS; operator Skill source and generated runtimes include repository guidance
- [x] verification Run Promptrepo checks and a real Template Registry GitHub canary | evidence: Promptrepo mod/test/vet/CGO checks passed; Registry workspace tests passed; live HTTPS and bare GitHub sync resolved 32 solutions
