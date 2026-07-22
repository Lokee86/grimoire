# Retrieval

`internal/retrieve` owns candidate discovery and deterministic lexical ranking over a prepared snapshot.

## Owns

- query normalization and term extraction;
- fixed phrase, filename, path, leading-line, and content scoring;
- inspectable score reasons;
- positive-score candidate filtering;
- deterministic score/path/source-range ordering; and
- candidate limiting.

## Does not own

- repository traversal or prepared-state loading;
- chunk construction;
- token-cost calculation;
- budget fitting or output package structure; or
- future Lexicon, Arcana, Demon Docs, or semantic provider execution.

## Main files

- `search.go` - current linear lexical search and ordering.
- `search_test.go` - ranking and tie-break coverage.
- `search_benchmark_test.go` - warm search benchmark over 10,000 prepared chunks.

## Current complexity

The current implementation scans all prepared chunks for each query and then sorts positive-score candidates. A prepared postings index is planned.

## Related documentation

- [System overview](../../docs/architecture/system-overview.md)
- [Context package](../../docs/reference/context-package.md)
- [Testing and benchmarks](../../docs/development/testing-and-benchmarks.md)
