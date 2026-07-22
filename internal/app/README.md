# Application Layer

`internal/app` owns Grimoire's command-line contract and operation orchestration.

## Owns

- top-level dispatch for `index`, `context`, `model`, `vector`, and `version`;
- `model setup`, `info`, `serve`, and `probe` dispatch;
- `vector build`, `search`, and `info` dispatch;
- flag definitions and validation;
- repository and prepared-state path resolution;
- composition of indexing, multi-provider retrieval, candidate curation, compilation, embedding, and native vector operations; and
- JSON command output.

## Does not own

- repository traversal or prepared-state formats;
- ignore-pattern semantics;
- ranking rules;
- model downloading, runtime discovery, or vector processing details;
- token-cost calculation; or
- budget selection internals.

## Main files

- `run.go` - top-level dispatch, source indexing, path resolution, and shared JSON output.
- `context.go` - context CLI flags, semantic fallback, and package compilation.
- `context_semantic.go` - query planning, batched embeddings, concurrent vector searches, and deterministic hit merging.
- `context_candidates.go` - exact/vector/lexical merge, duplicate-provider evidence, curation, and source reporting.
- `model.go` - embedding setup, runtime, information, and probe commands.
- `vector.go` - vector search and information commands.
- `vector_build.go` - incremental embedding orchestration and packed snapshot publication.
- `vector_ingest.go` - durable per-batch immutable vector ingestion for resumable builds.
- `vector_manifest.go` - persistent snapshot provenance and exact prepared-index freshness validation.
- `vector_paths.go` - vector-state layout and source-content identities.
- `run_test.go` - index-to-context integration coverage.
- `retrieval_quality_test.go` - checked-in deterministic retrieval-quality corpus.
- `model_test.go` - embedding command wiring coverage.
- `vector_test.go` - index, embed, reuse, native search, vector-backed context, and result mapping coverage.
- `vector_resume_test.go` - interrupted-build checkpoint and resume coverage.

## Dependencies

```text
app
 ├── index
 ├── retrieve
 ├── selection
 ├── compiler
 ├── embedding
 └── vectorstore
```

## Related documentation

- [CLI reference](../../docs/reference/cli.md)
- [Embedding model](../../docs/reference/embedding-model.md)
- [Vector store](../../docs/reference/vector-store.md)
- [Retrieval quality and latency](../../docs/development/retrieval-quality.md)
- [System overview](../../docs/architecture/system-overview.md)
