# Application Layer

`internal/app` owns Grimoire's command-line contract and operation orchestration.

## Owns

- top-level dispatch for `index`, `context`, `model`, `vector`, and `version`;
- `model setup`, `info`, `serve`, and `probe` dispatch;
- `vector build`, `search`, and `info` dispatch;
- flag definitions and validation;
- repository and prepared-state path resolution;
- composition of indexing, retrieval, compilation, embedding, and native vector operations; and
- JSON command output.

## Does not own

- repository traversal or prepared-state formats;
- ignore-pattern semantics;
- ranking rules;
- model downloading, runtime discovery, or vector processing details;
- token-cost calculation; or
- budget selection internals.

## Main files

- `run.go` - top-level commands, source index/context operations, and shared JSON output.
- `model.go` - embedding setup, runtime, information, and probe commands.
- `vector.go` - vector search and information commands.
- `vector_build.go` - incremental embedding and packed snapshot publication.
- `vector_paths.go` - vector-state layout and source-content identities.
- `run_test.go` - index-to-context integration coverage.
- `model_test.go` - embedding command wiring coverage.
- `vector_test.go` - index, embed, reuse, native search, and result mapping coverage.

## Dependencies

```text
app
 ├── index
 ├── retrieve
 ├── compiler
 ├── embedding
 └── vectorstore
```

## Related documentation

- [CLI reference](../../docs/reference/cli.md)
- [Embedding model](../../docs/reference/embedding-model.md)
- [Vector store](../../docs/reference/vector-store.md)
- [System overview](../../docs/architecture/system-overview.md)
