## 1. 兼容基线与合同

- [x] 1.1 固定 TemplateRole、TemplateContent、ReadTemplate、Render、Preview 和错误码兼容 fixtures；验收：旧测试无修改通过；验证：`go test ./...`。
- [x] 1.2 新增 additive descriptor、request/result 和可选 document client interfaces；依赖：1.1；验收：旧 Client interface 和 unkeyed composite consumer 不受影响；验证：`go test ./...`。

## 2. 严格解析与选择器

- [x] 2.1 实现 descriptor binding、bounded source read 和 format mismatch；依赖：1.2；验证：`go test ./... -run 'DocumentDescriptor|FormatMismatch' -count=1`。
- [x] 2.2 实现 strict JSON 和 YAML JSON-compatible conversion；依赖：2.1；验收：duplicate/alias/tag/non-finite fixtures fail closed；验证：`go test ./... -run 'StructuredDocument|YAML|Duplicate' -count=1`。
- [x] 2.3 实现 streaming JSONL、record-id uniqueness 和 segment bounds；依赖：2.1；验证：`go test ./... -run 'JSONL|RecordID|Segment' -count=1`。
- [x] 2.4 实现 heading、JSON/YAML pointer 和 JSONL ID selector；依赖：2.2、2.3；验证：`go test ./... -run 'DocumentSelector' -count=1`。

## 3. Digest、安全投影与 conformance

- [x] 3.1 实现 source/canonical digest 和 JSON/YAML 等价 fixtures；依赖：2.2；验证：`go test ./... -run 'CanonicalDigest' -count=1`。
- [x] 3.2 实现 safe projection 与 body/value sentinel leak tests；依赖：2.4；验证：`go test ./... -run 'DocumentProjection|BodyLeak' -count=1`。
- [x] 3.3 更新 public docs、architecture 和 release compatibility；依赖：3.1、3.2；验收：只记录真实 public surface 和实际验证命令；验证：`openspec validate promptrepo-structured-document-v1 --strict --no-interactive`。

## 4. 最终验证

- [x] 4.1 运行纯 Go gates；依赖：3.3；验证：`go mod verify && go test ./... && go vet ./... && CGO_ENABLED=0 go test ./... && CGO_ENABLED=0 go build ./...`。
