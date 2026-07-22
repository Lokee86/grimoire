# Retrieval

`internal/retrieve` owns the deterministic lexical fallback and the shared candidate provenance shape used by lexical and vector retrieval.

## Owns

- query normalization and term extraction;
- fixed phrase, filename, path, leading-line, and content scoring;
- inspectable score reasons;
- provider source and rank provenance;
- positive-score candidate filtering;
- deterministic score/path/source-range ordering; and
- candidate limiting.

## Does not own

- repository traversal or prepared-state loading;
- chunk construction;
- token-cost calculation;
- budget fitting or output package structure; or
- Lexicon, Arcana, Demon Docs, or semantic provider execution.

## Main files

- `search.go` - current linear lexical search and ordering.
- `search_test.go` - ranking and tie-break coverage.
- `search_benchmark_test.go` - warm search benchmark over 10,000 prepared chunks.

## Current complexity

The fallback scans all prepared chunks for each query and then sorts positive-score candidates. It is not used on the normal vector-backed context path. A larger lexical engine should only be added if measured retrieval failures justify its runtime and maintenance cost.

## Related documentation

- [System overview](../../docs/architecture/system-overview.md)
- [Context package](../../docs/reference/context-package.md)
- [Testing and benchmarks](../../docs/development/testing-and-benchmarks.md)
