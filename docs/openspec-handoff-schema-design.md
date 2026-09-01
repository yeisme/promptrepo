# Schema 设计

本文定义跨项目字段语义，不是可直接发布的 JSON Schema 文件。正式 machine schema 必须由各 owner 的 CLI/生成器创建并经过 OpenSpec、conformance 和 digest 校验。

## Schema 集合

| Schema ID | Owner | 用途 |
| --- | --- | --- |
| `promptrepo.template-document.v0.1` | promptrepo/Registry | 描述模板格式、Schema、canonicalization、selector 和 compiler profile |
| `scaena.character-spec.v1` | Scaena | 角色永久身份、脸、发型、身体、左右 marker |
| `scaena.wardrobe-spec.v1` | Scaena | 服装层级、材料 zone、coverage、配饰和鞋 |
| `scaena.character-asset-task.v1` | Scaena | 单次资产生成 Task、layout、render、references、constraints、QC、output |
| `scaena.reference-binding.v1` | Scaena/Eikona handoff | 参考资产职责、禁止影响范围、版本和审批状态 |
| `scaena.prompt-bundle.v1` | Scaena compiler | 编译 Prompt、typed controls、bindings、QC 和 provenance |
| `scaena.asset-event.v1` | Scaena | JSONL append-only lifecycle event |
| `scaena.character-asset-view.v1` | Scaena | 单角色资产衍生视图 |
| `scaena.cast-composition-view.v1` | Scaena | 多人构图与交互视图 |
| `scaena.shot-continuity-view.v1` | Scaena | 镜头连续性、伤痕、污渍、道具位置和状态 Delta |

## CharacterAssetTaskSpec 顶层

```json
{
  "schema_version": "scaena.character-asset-task.v1",
  "spec_version": "1.0",
  "task": {},
  "subject": {},
  "layout": {},
  "render_intent": {},
  "reference_bindings": [],
  "constraints": [],
  "prompt_overrides": {},
  "qc_policy": {},
  "output_intent": {},
  "provenance": {}
}
```

`additionalProperties` 默认 `false`。扩展字段只能进入 `extensions`，key 必须使用反向域名或 owner namespace；production compiler 不得自动解释未知 extension。

建议的 Draft 2020-12 骨架：

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://schemas.yeisme.dev/scaena/character-asset-task/v1",
  "title": "Scaena Character Asset Task",
  "type": "object",
  "required": [
    "schema_version",
    "spec_version",
    "task",
    "subject",
    "layout",
    "render_intent",
    "constraints",
    "qc_policy",
    "output_intent"
  ],
  "properties": {
    "schema_version": {"const": "scaena.character-asset-task.v1"},
    "spec_version": {"const": "1.0"},
    "task": {"$ref": "#/$defs/task"},
    "subject": {"$ref": "#/$defs/subject"},
    "layout": {"$ref": "#/$defs/layout"},
    "render_intent": {"$ref": "#/$defs/renderIntent"},
    "reference_bindings": {
      "type": "array",
      "items": {"$ref": "https://schemas.yeisme.dev/scaena/reference-binding/v1"}
    },
    "constraints": {
      "type": "array",
      "items": {"$ref": "#/$defs/constraint"}
    },
    "prompt_overrides": {"$ref": "#/$defs/promptOverrides"},
    "qc_policy": {"$ref": "https://schemas.yeisme.dev/scaena/qc-policy/v1"},
    "output_intent": {"$ref": "#/$defs/outputIntent"},
    "provenance": {"$ref": "#/$defs/provenance"},
    "extensions": {"type": "object"}
  },
  "unevaluatedProperties": false
}
```

`$id` 是设计标识；正式发布地址和 digest 由 Registry release 锁定，不要求 consumer 联网获取该 URL。

## 用户示例到 canonical 字段的映射

| 输入字段 | Canonical 字段 | 处理 |
| --- | --- | --- |
| `task` | `task` | 保留并补 workspace/disposition |
| `character` | `subject.inline_character` | 可选显式 extract 为 character ref |
| `turnaround_layout` | `layout` | 规范化 view/rotation/alignment |
| `render` | `render_intent` | Prompt prose 与 typed controls 分离 |
| `render.composition.character_count` | `layout.subject_count` + `layout.depiction_count` | 同角色三视图规范化为 1 + 3 |
| `positive_prompt` | `prompt_overrides.positive_sections` | compatibility ingress，默认 append/untrusted |
| `negative_prompt[]` | `constraints[]` | 分配 stable ID/category/severity |
| `quality_control` | `qc_policy` | 转 checks/reject refs/evidence type |
| `output` | `output_intent` | asset family、version、format、review state |

Importer 必须输出该映射的 Delta report；它不能静默丢弃原始字段或假装 top-level Prompt 与结构化事实没有重复。

## Schema、Semantic Lint、Compiler 与 QC 的边界

| 层 | 负责 | 不负责 |
| --- | --- | --- |
| JSON Schema | required、type、enum、range、pattern、oneOf、if/then | 艺术意图、左右关系正确性、参考图平均脸风险 |
| Semantic Lint | 跨字段冲突、side/anchor、材料覆盖、view rotation、approval state、stale | 生成结果是否真的符合 |
| Compiler | section order、constraint dedupe、reference instruction、typed controls | 人工接受、provider 权限 |
| Post-generation QC | 图像身份、比例、材料、手脚、文字水印、多人关系 evidence | 自动修改 canonical state |

任何一层都不能用“模型大概能理解”替代上一层的确定性检查。

## Task

```text
task.id                   stable string
task.type                 enum
task.asset_type           enum
task.description          zh-CN prose
task.language             BCP-47 locale
task.workspace_mode       quick | candidate | production
task.artifact_disposition preview | candidate | canonical_proposal | export
```

首版 `task.type`：

```text
face_master
body_master
character_turnaround
face_view
expression
viseme
wardrobe
pose
prop_interaction
lighting
fusion_master
continuity_state
cast_composition
shot
```

## Subject 与角色模块

`subject` 使用 `oneOf`：

```text
inline_character: scaena.character-spec.v1
character_ref: owner/ref/version/digest
```

Schema 形态：

```json
{
  "oneOf": [
    {
      "required": ["inline_character"],
      "properties": {
        "inline_character": {
          "$ref": "https://schemas.yeisme.dev/scaena/character-spec/v1"
        }
      }
    },
    {
      "required": ["character_ref"],
      "properties": {
        "character_ref": {"$ref": "#/$defs/exactOwnerRef"}
      }
    }
  ],
  "unevaluatedProperties": false
}
```

永久字段包括 gender、adult assertion、age appearance、identity、temperament、face、hair、body 和 stable markers。所有可左右定位的 marker 必须使用结构化 anchor：

```json
{
  "id": "FACE_MARK_TEAR_MOLE_01",
  "kind": "mole",
  "side": "left",
  "anchor": "below_eye",
  "description": "左眼下方很小的浅褐色泪痣",
  "mirror_forbidden": true
}
```

身体比例同时提供 number 与显示文本：

```text
body.height_cm: 168
body.height_impression: "约168厘米"
body.head_body_ratio: 7.5
body.proportion_description: "7.5头身"
```

## Wardrobe 与材料安全

服装使用 layer + material zone，不只保存一段长 prose：

```text
wardrobe.layers[].id
wardrobe.layers[].garment_type
wardrobe.layers[].material_ref
wardrobe.layers[].body_zones[]
wardrobe.layers[].opacity_policy
wardrobe.layers[].coverage_policy
wardrobe.layers[].front_structure
wardrobe.layers[].back_structure
```

`opacity_policy`：`opaque`、`semi_transparent`、`transparent`。角色基础服装 production schema 默认禁止敏感 body zone 使用 `semi_transparent|transparent`，除非下面存在 approved opaque coverage layer。该规则用于校验层次关系，不从文字猜测“应该安全”。

鞋类单独登记 `footwear.id/type/material/color/sole/thickness/lacing/brand_policy`，三视图通过同一 footwear ref 锁定。

## Layout 与三视图

```text
layout.kind                 turnaround_sheet
layout.orientation          landscape
layout.aspect_ratio         16:9
layout.subject_count        1
layout.depiction_count      3
layout.identity_policy      same_subject
layout.coordinate_space     normalized_canvas_v1
layout.views[]
layout.alignment
```

每个 view：

```json
{
  "view_id": "front_view",
  "label": {"zh-CN": "正面"},
  "orientation": "front",
  "rotation": {"yaw_deg": 0, "pitch_deg": 0, "roll_deg": 0},
  "face_visibility": "required",
  "full_body": true,
  "requirements": []
}
```

左侧面使用 yaw `-90`，背面使用 yaw `180` 且 `face_visibility=forbidden`。Schema 校验 depiction count 与 views 长度一致；semantic lint 校验 view ID、orientation 和 rotation 不冲突。

JSON Schema 无法直接比较 `depiction_count` 与 `views.length`，因此两者都必须存在：Schema 分别校验范围，semantic lint 负责相等性。`character_turnaround` 条件规则：

```json
{
  "if": {
    "properties": {
      "task": {
        "properties": {"type": {"const": "character_turnaround"}}
      }
    }
  },
  "then": {
    "properties": {
      "layout": {
        "required": ["subject_count", "depiction_count", "views", "alignment"],
        "properties": {
          "subject_count": {"const": 1},
          "depiction_count": {"minimum": 2, "maximum": 8},
          "identity_policy": {"const": "same_subject"}
        }
      }
    }
  }
}
```

## RenderIntent

```text
render_intent.style.description
render_intent.camera.lens_equivalent_mm: 70
render_intent.camera.projection: near_orthographic
render_intent.camera.angle: eye_level
render_intent.lighting
render_intent.canvas.width: 1792
render_intent.canvas.height: 1024
render_intent.generation.seed_policy
render_intent.generation.seed?
render_intent.generation.reference_strength: 0..1
render_intent.generation.stylization_strength: 0..1
render_intent.generation.detail_strength: 0..1
```

`seed_policy=fixed` 时 `seed` 必填；`family_locked` 由 Scaena 在资产族创建时通过 application service 生成并锁定；`random` 不允许进入 production master。

固定 seed 条件：

```json
{
  "if": {
    "properties": {"seed_policy": {"const": "fixed"}},
    "required": ["seed_policy"]
  },
  "then": {
    "required": ["seed"],
    "properties": {
      "seed": {"type": "integer", "minimum": 0}
    }
  }
}
```

production 禁止 random 属于 task/render 跨节点规则，由 semantic lint 处理；不能只靠局部 render schema。

## Constraints

字符串数组可以导入，但 canonical 记录为：

```json
{
  "id": "continuity.marker.side-lock",
  "category": "continuity",
  "severity": "blocking",
  "enforce_at": "both",
  "target_paths": ["/subject/inline_character/stable_markers"],
  "text": "固定识别点和配饰不得镜像或换边"
}
```

category 首版：`identity`、`proportion`、`wardrobe`、`material`、`safety`、`anatomy`、`layout`、`camera`、`background`、`text_overlay`、`continuity`、`reference_scope`。

Schema 可以使用 `uniqueItems` 防止完全相同的对象，但无法仅按 `constraint.id` 保证唯一；stable ID 冲突、同 ID 不同语义和 precedence 由 semantic lint/compiler处理。

## QCPolicy

```text
qc_policy.checks[].id
qc_policy.checks[].category
qc_policy.checks[].severity
qc_policy.checks[].enforce_at
qc_policy.checks[].target_paths[]
qc_policy.checks[].description
qc_policy.checks[].evidence_required
qc_policy.reject_if[] -> blocking check refs
```

`must_pass` 字符串导入后转换为 blocking/manual_visual checks。自动检测结果只能提交 evidence，不得自己把 candidate 标记 accepted。

## PromptBundle

```json
{
  "schema_version": "scaena.prompt-bundle.v1",
  "bundle_id": "PB_...",
  "task_ref": {},
  "source_documents": [],
  "compiler": {},
  "positive_prompt": "body-bearing private field",
  "negative_prompt": "body-bearing private field",
  "reference_bindings": [],
  "edit_scope": {},
  "generation_intent": {},
  "qc_checklist": [],
  "provenance": {},
  "bundle_digest": "sha256:..."
}
```

Digest 输入包含 source/canonical/schema/semantic-rule/compiler-profile/compiler-implementation digests 和 normalized bundle body，排除存储路径与 wall-clock build time。

Bundle 的安全投影使用独立 DTO，不通过给 `positive_prompt`/`negative_prompt` 加 `omitempty` 来防泄漏；body-bearing DTO 字段应显式 `json:"-" yaml:"-"`，持久化只能进入受控 CAS encoder。

## JSONL 公共 Envelope

事件记录：

```text
schema_version
event_id
sequence
project_id
entity_ref
event_type
occurred_at
source_refs[]
payload
record_digest
```

衍生视图记录：

```text
schema_version
view_type
record_id
project_id
as_of_revision
source_digests[]
payload
record_digest
```

衍生记录不包含 build timestamp；generation manifest 单独记录 built_at。`record_digest` 排除自身字段后按 JCS 计算。

## 多人 CastComposition

```text
scene_ref / shot_ref
coordinate_space: normalized_frame_v1
camera
members[]:
  member_id
  character_ref/digest
  identity_anchor_ref/digest
  wardrobe_ref/digest
  continuity_state_ref/digest
  physical_height_cm
  position {x,y}
  screen_scale
  z_order
  facing
  gaze_target_ref?
relations[]:
  relation_id
  type
  from/to
  from_anchor/to_anchor?
```

relation type 首版：`gaze`、`contact`、`holding`、`occludes`、`supports`、`attacks`、`follows`。接触关系必须指定可见 body/prop anchor，避免只写“牵手”而无法验证左右手。
