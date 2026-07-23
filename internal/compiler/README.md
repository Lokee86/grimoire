# Context Compiler

`internal/compiler` owns final whole-item selection and the versioned JSON context-package model.

## Owns

- package versioning;
- package, structural-provider-state, structural-evidence, and source-selection fields;
- deterministic budget fitting across complete structural facts and complete source chunks;
- exact serialized-package token accounting;
- final package-byte verification;
- separate structural and source budget-omission counts; and
- provider-source metadata supplied by the composition path.

## Does not own

- candidate or structural-evidence discovery;
- relevance scoring;
- repository or prepared-state access;
- chunk construction;
- tokenizer selection or vocabulary ownership; or
- provider execution and deadlines.

## Main files

- `compiler.go` - package model and whole-item fitting algorithm.
- `compiler_test.go` - budget, structural evidence, and package behavior coverage.

## Selection rule

The compiler first considers the highest-ranked structural fact, then the highest-ranked source selection. It next considers the remaining structural facts and finally the remaining source candidates. This gives first-class structural data an early reserved opportunity without allowing it to consume every token before implementation source is considered.

For every item, the compiler serializes and counts the complete tentative JSON package. An item is retained only when that package fits the budget. A rejected item is counted, but later smaller items may still be selected. Individual facts and chunks are never truncated. The final emitted bytes are counted again and must match the package-level `token_count`.

## Related documentation

- [Context package](../../docs/reference/context-package.md)
- [System overview](../../docs/architecture/system-overview.md)
- [Current limitations](../../docs/limits/current-limitations.md)
