# Architecture

## Ownership

`github.com/yeisme/promptrepo` is the canonical public owner of the Go SDK.
The private Template Registry nested SDK tagged `sdk/go/promptrepo/v0.1.0` is
an immutable compatibility baseline, not a second active owner. Template
Registry remains a separate server/control-plane implementation and is not a
dependency of this module.

## Package boundary

```text
consumer -> promptrepo.Client <- engine.Manager -> source.Registry
                                      |                 |-- file
                                      |                 |-- Git/GitHub
                                      |                 `-- anonymous S3
                                      `-> user config/cache state
```

The root package exposes DTOs, error codes, exact-ref parsing, digest helpers,
and the `Client` interface. `engine` is the embedded local implementation;
`source` supplies bounded read-only source adapters. No package exposes server
internals, ORM types, or credentials.

## Compatibility invariants

- `promptrepo.state.v0.1`, `promptrepo.catalog.v0.1`, and
  `promptrepo.stage_receipt.v0.1` remain unchanged through the extraction.
- Config and cache defaults remain `yeisme/promptrepo` below the OS user config
  and cache directories.
- Digest algorithms, error-code strings, JSON tags, and exact-ref format remain
  compatible with private v0.1.0.
- State uses a cross-process lock and atomic rename; unsupported future state
  schemas fail closed.

The `compatibility_v010_test.go` test executes a frozen v0.1.0-shaped fixture
without private-network access and pins the observable result.
