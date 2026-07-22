# Context Compiler

`internal/compiler` owns final whole-chunk selection and the versioned JSON context-package model.

## Owns

- package versioning;
- package and selection field definitions;
- ranked whole-chunk budget fitting;
- exact serialized-package token accounting;
- final package-byte verification;
- budget-omission counting; and
- retrieval-source metadata supplied by the current composition path.

## Does not own

- candidate discovery or relevance scoring;
- repository or prepared-state access;
- chunk construction;
- tokenizer selection or vocabulary ownership; or
- provider execution and deadlines.

## Main files

- `compiler.go` - package model and selection algorithm.
- `compiler_test.go` - budget and package behavior coverage.

## Selection rule

Candidates are considered in ranked order. For each candidate, the compiler serializes and counts the complete tentative JSON package. A complete chunk is retained only when that package fits the budget. A rejected candidate is counted, but later smaller candidates may still be selected. The final emitted bytes are counted again and must match the package-level `token_count`.

## Related documentation

- [Context package](../../docs/reference/context-package.md)
- [System overview](../../docs/architecture/system-overview.md)
- [Current limitations](../../docs/limits/current-limitations.md)
