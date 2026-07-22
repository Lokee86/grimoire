# Application Layer

`internal/app` owns Grimoire's command-line contract and orchestration.

## Owns

- command dispatch for `index`, `context`, and `version`;
- flag definitions and validation;
- repository and prepared-state path resolution;
- composition of index, retrieval, and compiler operations; and
- JSON command output.

## Does not own

- repository traversal or index formats;
- ignore-pattern semantics;
- ranking rules;
- token-cost calculation; or
- budget selection internals.

## Main files

- `run.go` - CLI implementation and shared output helpers.
- `run_test.go` - flag wiring and index-to-context integration coverage.

## Dependencies

```text
app
 ├── index
 ├── retrieve
 └── compiler
```

## Related documentation

- [CLI reference](../../docs/reference/cli.md)
- [System overview](../../docs/architecture/system-overview.md)
