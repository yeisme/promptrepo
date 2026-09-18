# Documentation

- [Architecture](architecture.md): ownership, public API boundary, state, and
  source adapters.
- [Release](release.md): provenance, RC/stable gates, tag policy, rollback,
  and consumer migration.
- [OpenSpec change（已归档）](../openspec/changes/archive/2026-08-25-promptrepo-public-sdk-extraction-v1/):
  extraction and compatibility record.
- [Template Address architecture](architecture.md#模板-address-与安全-projection):
  source-URI boundary and inspect/validate/render/preview contract.
- [UI template architecture](architecture.md#ui-template-安全与-owner-边界):
  independent address family, bounded validation, digest/snapshot, redaction,
  and Promptrepo/Registry/Scaena ownership.
- [OpenSpec change（已归档）](../openspec/changes/archive/2026-08-25-promptrepo-template-address-inspect-preview-v1/):
  v0.3.0 template Address, input schema, inspect, and preview contract.
- [UI template OpenSpec（已归档）](../openspec/changes/archive/2026-09-03-promptrepo-ui-template-contract-v1/):
  additive `kind=ui-template` public contract and verification record.

## CI/CD

- [模块化、分级 CI/CD](delivery/ci-cd.md)：quick、full、integration、release 的触发场景、真实命令和权限边界。

## Agent Skills

本项目会话的 active skills 由仓库根 `.skills/profiles/targets/shared/promptrepo.txt` 分配，用 `scripts/skills.sh sync-target shared/promptrepo` 生成 `.agents/skills/` 与 `.claude/skills/`。不要手改运行副本。

完整对照、缺口与下一波优化见 [子项目 Skill Profile](../../../docs/skills/subproject-skill-profiles.md)。

