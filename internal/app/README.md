# Application Layer

`internal/app` owns Grimoire's command-line contract and operation orchestration.

## Owns

- top-level dispatch for `index`, `context`, `model`, and `version`;
- `model setup`, `info`, `serve`, and `probe` dispatch;
- flag definitions and validation;
- repository and prepared-state path resolution;
- composition of indexing, retrieval, compilation, and embedding operations; and
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
- `run_test.go` - index-to-context integration coverage.
- `model_test.go` - embedding command wiring coverage.

## Dependencies

```text
app
 ├── index
 ├── retrieve
 ├── compiler
 └── embedding
```

## Related documentation

- [CLI reference](../../docs/reference/cli.md)
- [Embedding model](../../docs/reference/embedding-model.md)
- [System overview](../../docs/architecture/system-overview.md)
