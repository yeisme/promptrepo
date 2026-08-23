## 1. Extraction and module boundary

- [x] 1.1 Extract the exact private v0.1.0 SDK subtree into the public module.
- [x] 1.2 Rewrite only the Go module/import prefix and verify no server/internal dependency remains.
- [x] 1.3 Preserve the copied production, test, and conformance package layout.

## 2. Public release foundations

- [x] 2.1 Add MIT licensing and Chinese-first bilingual public documentation.
- [x] 2.2 Document source ownership, immutable provenance, migration, deprecation, rollback, RC/stable gates, and tag policy.
- [x] 2.3 Add repository guidance, security reporting, contribution guidance, and pinned Go 1.24 CI.

## 3. Compatibility and verification

- [x] 3.1 Add a local v0.1.0 fixture compatibility test with no private-network dependency.
- [x] 3.2 Run module, test, vet, and CGO-disabled validation.
- [x] 3.3 Run strict OpenSpec validation, normalized source diff review, and secret scan.
