# Knowledge package

`internal/knowledge` owns the independent repository-documentation index and deterministic retrieval contract.

## Responsibilities

- Discover eligible documentation and prose-heavy configuration while respecting permanent exclusions, ignore rules, generated-content policy, and explicit exclusions.
- Split documents into stable path/heading-derived sections with exact byte and line spans.
- Persist document and section hashes, BM25 term frequencies, Git metadata, and exact citation handles.
- Extract bounded code-link hints for documented symbols, paths, endpoints, messages, and configuration contract names that also exist in repository source.
- Apply path, kind, heading, commit, and commit-time filters.
- Rank every query with deterministic BM25 and optionally accept supplemental scores through `VectorRanker`.
- Preserve BM25 results when an optional vector ranker fails and expose the failure through `SearchResponse.VectorError`.

## Boundary

This package does not discover or load native vector snapshots, call an embedding endpoint, prepare production source chunks, or perform graph traversal. `internal/knowledgevector` implements the optional vector ranker. Source retrieval belongs to `internal/index`, `internal/retrieve`, Lexicon, and Arcana.
