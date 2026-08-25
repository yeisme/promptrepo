# template-address-inspect-preview Specification

## Purpose
TBD - created by archiving change promptrepo-template-address-inspect-preview-v1. Update Purpose after archive.
## Requirements
### Requirement: Template address is a separately canonical immutable reference

The SDK SHALL preserve `ParseRef` and `FormatRef` unchanged. It SHALL provide a distinct template-address parser and formatter whose `promptrepo://` solution identity uses canonical query order `kind,locale,role,path,selector,digest,snapshot`. The address parser SHALL require `kind=template` and a valid locale, reject user info, fragments, unknown/duplicate query fields, traversal/absolute paths, invalid typed selectors, and invalid sha256 digests. Repository source URIs SHALL remain a distinct grammar.

#### Scenario: Existing solution ref remains unchanged
- **WHEN** a consumer parses and formats `promptrepo://official/audio/podcast@1.0.0?locale=zh-CN`
- **THEN** it SHALL receive the exact same ref
- **AND** template-address additions SHALL not alter its behavior.

### Requirement: Template contracts are additive and catalog/state remain unchanged

The SDK SHALL preserve the released `TemplateRole`, `Solution`, catalog, snapshot, and state DTO field/tag shape. It SHALL provide caller-supplied `TemplateContract` for input definitions, license, permissions, and an optional contract digest through inspect/validate/preview requests and results. A non-empty contract digest SHALL be lowercase `sha256:<64 hex>`; an empty digest SHALL remain valid for draft contracts. The additive `ContractResolver` capability SHALL load a Registry-authored, bounded, valid-UTF-8 `promptrepo.template-contract.v0.1` companion from the same resolved source without adding contract fields to catalog or state. Exact Git/GitHub commits SHALL report `snapshot_pinned`; file and S3 current-object reads SHALL report `content_bound` while still rejecting oversized/invalid-text sidecars, identity, template-path, template-digest, or contract-digest drift.

#### Scenario: A v0.1 catalog has no input definitions
- **WHEN** its canonical digest is calculated by the new SDK
- **THEN** the digest SHALL equal the frozen prior digest.

#### Scenario: A companion is resolved from an exact Git snapshot
- **WHEN** the selected catalog template is `solutions/audio/podcast/prompts/main.zh-CN.md`
- **THEN** the SDK SHALL read `solutions/audio/podcast/contracts/main.zh-CN.json` from the same exact commit
- **AND** it SHALL reject schema, identity, template-path, template-digest, or contract-digest drift.

### Requirement: Inspect returns only safe metadata and readiness

The optional Inspector capability SHALL resolve catalog/snapshot data without reading template body and SHALL return origin/ref/version/digest/locale/role/rights/maturity/tags/capabilities/localized display/caller-supplied contract/input status/readiness/issues/next action. Supplied input values SHALL not be returned.

#### Scenario: Required input is absent
- **WHEN** inspect is called without a required declared input
- **THEN** the result SHALL report `missing` and `ready=false`
- **AND** it SHALL include a safe `INPUT_REQUIRED` issue.

#### Scenario: Rights block use
- **WHEN** inspect resolves a template whose rights are blocked or prohibited
- **THEN** the result SHALL report `ready=false` with `RIGHTS_BLOCKED`
- **AND** `next_action.kind` SHALL be `blocked`, not `supply_inputs` or `preview`.

### Requirement: Preview is non-executing and serialization-safe

The optional Renderer capability SHALL strictly render declared `{{name}}` placeholders in caller-provided memory without reading a source, calling a provider, writing state, creating a run, or recording usage. The optional Previewer capability MAY read a digest-verified template body and invoke Renderer. Both SHALL validate declared, unknown, missing, type, enum and constraints; report rendered digest and byte/rune counts; and omit rendered body from JSON and YAML serialization. Preview SHALL report false provider/state/usage flags.

#### Scenario: Preview receives a selector
- **WHEN** an Address contains a non-empty selector
- **THEN** Inspect SHALL return it as Address metadata
- **AND** Preview SHALL fail closed with `SELECTOR_UNSUPPORTED` until a selector engine has conformance tests.

#### Scenario: A valid preview has defaults
- **WHEN** an optional input is omitted and its declared default is valid
- **THEN** preview SHALL substitute that default in memory and report ready
- **AND** it SHALL not create a run, call a provider, write state, or record usage.

