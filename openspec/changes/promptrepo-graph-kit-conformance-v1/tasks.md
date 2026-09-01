## 1. Contract freeze

- [x] 1.1 固定现有 DocumentResolver/Loader/Selector、Git/GitHub source 与 exact address compatibility tests。
- [x] 1.2 建立跨题材 synthetic Graph Kit manifest/child conformance fixture，确认无 Auctra/Nishi 私有内容。

## 2. Conformance implementation

- [x] 2.1 增加 local Git exact snapshot closure success test，覆盖 manifest、descriptor、load、selector 和 canonical digest。
- [x] 2.2 增加 missing child、digest drift、floating revision、path escape 和 body sentinel fail-closed tests。
- [x] 2.3 仅在现有接口无法表达 closure 时增加最小 additive helper；不得给现有 consumer-implemented interface 加方法。

## 3. Handoff 与验证

- [x] 3.1 更新 public docs/CHANGELOG，说明 Graph Kit 是 structured-document conformance profile而非新 domain API。
- [x] 3.2 运行 `go mod verify`、`go test ./...`、`go vet ./...`、`CGO_ENABLED=0 go test ./...`、`CGO_ENABLED=0 go build ./...`。
- [x] 3.3 运行 `openspec validate promptrepo-graph-kit-conformance-v1 --strict --no-interactive`，记录供 Auctra/Registry 升级的不可变版本 handoff；不执行 tag/publish。
