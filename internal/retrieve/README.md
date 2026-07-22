# Retrieval

`internal/retrieve` owns deterministic exact and lexical fallback retrieval plus the shared candidate provenance shape used by lexical and vector retrieval.

## Owns

- query normalization and term extraction;
- fixed phrase, filename, path, leading-line, and content scoring;
- inspectable score reasons;
- provider source and rank provenance;
- positive-score candidate filtering;
- deterministic score/path/source-range ordering;
- candidate limiting;
- exact-signal extraction for literal repository queries; and
- exact path/content matching, reason aggregation, and ordering.

## Does not own

- repository traversal or prepared-state loading;
- chunk construction;
- token-cost calculation;
- budget fitting or output package structure; or
- Lexicon, Arcana, Demon Docs, or semantic provider execution.

## Main files

- `search.go` - current linear lexical search and ordering.
- `search_test.go` - ranking and tie-break coverage.
- `exact.go` - targeted candidate scanning, aggregation, limiting, and ranking.
- `exact_signals.go` - concrete signal extraction/classification and literal matching.
- `exact_test.go` - exact signal, aggregation, limiting, and tie-break coverage.
- `search_benchmark_test.go` - warm lexical benchmark over 10,000 prepared chunks.
- `exact_benchmark_test.go` - warm conditional exact-recovery benchmark over 10,000 prepared chunks.

## Current complexity

The fallback scans all prepared chunks for each query and then sorts positive-score candidates. It is not used on the normal vector-backed context path. A larger lexical engine should only be added if measured retrieval failures justify its runtime and maintenance cost.

`Exact` only activates for concrete signals: quoted phrases, paths or filenames, identifier-like tokens, configuration keys, error codes, and version strings. Lowercase natural-language words alone return no candidates. Dotted configuration keys also emit their terminal key, so `damage.max_per_hit` recovers `max_per_hit` in sectioned TOML/YAML content while retaining configuration-key provenance. Exact candidates use source `exact`, preserve one reason per matched path/content signal, and use deterministic score/path/range ordering.

## Related documentation

- [System overview](../../docs/architecture/system-overview.md)
- [Context package](../../docs/reference/context-package.md)
- [Testing and benchmarks](../../docs/development/testing-and-benchmarks.md)
