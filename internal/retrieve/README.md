# Retrieval

`internal/retrieve` owns deterministic exact and BM25 source retrieval plus the candidate provenance shape consumed by the unified discovery layer.

## Owns

- query normalization and term extraction;
- code-aware query and source tokenization with prompt-stopword suppression;
- use of the prepared lexical postings index to localize candidate chunks;
- BM25 term-frequency, rarity, and document-length scoring;
- fixed phrase, filename, path, leading-line, and declaration-alias boosts;
- scoped search over bounded path sets;
- file-level discovery scopes;
- inspectable score reasons;
- provider source and rank provenance;
- positive-score candidate filtering;
- deterministic score/path/source-range ordering;
- candidate limiting;
- exact-signal extraction for literal repository queries; and
- exact path/content matching, reason aggregation, and ordering.

## Does not own

- repository traversal or prepared-state loading;
- chunk construction or lexical-index persistence;
- discovery lane assembly or response schemas;
- documentation retrieval;
- Lexicon, Arcana, or semantic provider execution; or
- agent stopping policy.

## Main files

- `search.go` — shared-corpus BM25 orchestration, postings-based candidate localization, declaration aliases, field boosts, and ordering.
- `scoped_search.go` — BM25 restricted to an allowed path set.
- `file_search.go` — file-level discovery scopes used before chunk retrieval.
- `bm25.go` — corpus statistics and BM25 scoring over the prepared lexical index.
- `exact.go` — targeted exact candidate scanning, aggregation, limiting, and ranking.
- `exact_signals.go` — concrete signal extraction, classification, and literal matching.
- `*_test.go` — exact, file, scoped, multi-query, ranking, and tie-break coverage.
- `search_benchmark_test.go` — warm lexical benchmark over 10,000 prepared chunks.
- `exact_benchmark_test.go` — warm conditional exact-recovery benchmark over 10,000 prepared chunks.

## Current behavior

`SearchMany` compiles a bounded set of queries against one prepared lexical corpus. Candidate documents are localized through postings for query terms and declaration aliases, then scored independently for each query. Repository content is not rescanned once per intent.

Scoped retrieval accepts an explicit path set so a structural or lexical discovery phase can narrow later source ranking without creating another source index.

`Exact` activates only for concrete signals such as quoted phrases, paths, filenames, identifier-like tokens, configuration keys, error codes, and version strings. Lowercase natural-language words alone return no exact candidates. Dotted configuration keys also emit their terminal key, so `damage.max_per_hit` can recover `max_per_hit` while retaining configuration-key provenance.

Exact, source, document, symbol, and relationship limits are assembled independently above this package.

## Related documentation

- [System overview](../../docs/architecture/system-overview.md)
- [Unified discovery contract](../../docs/reference/agent-query.md)
- [Prepared-index architecture](../../docs/architecture/prepared-index.md)
- [Testing and benchmarks](../../docs/development/testing-and-benchmarks.md)

