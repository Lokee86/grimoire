# Context Compiler

`internal/compiler` owns final whole-chunk selection and the versioned JSON context-package model.

## Owns

- package versioning;
- package and selection field definitions;
- ranked whole-chunk budget fitting;
- selected estimated-cost totals;
- budget-omission counting; and
- retrieval-source metadata supplied by the current composition path.

## Does not own

- candidate discovery or relevance scoring;
- repository or prepared-state access;
- chunk construction;
- model tokenization; or
- provider execution and deadlines.

## Main files

- `compiler.go` - package model and selection algorithm.
- `compiler_test.go` - budget and package behavior coverage.

## Selection rule

Candidates are considered in ranked order. A complete chunk is selected when it fits the remaining budget. An oversized candidate is skipped and counted, but later smaller candidates may still be selected.

## Related documentation

- [Context package](../../docs/reference/context-package.md)
- [System overview](../../docs/architecture/system-overview.md)
- [Current limitations](../../docs/limits/current-limitations.md)
