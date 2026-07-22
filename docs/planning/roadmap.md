# Grimoire Roadmap

This roadmap describes implementation order, not release commitments.

## Current foundation

Implemented:

- incremental file records with unchanged-file reuse;
- Git-ignore traversal and protected tool-state exclusions;
- content-addressed prepared storage with atomic publication;
- deterministic fallback chunks;
- exact `o200k_base` chunk and package accounting;
- a deterministic linear lexical baseline;
- the fixed Qwen3 embedding-model contract;
- verified managed model/runtime setup on Windows x64;
- an OpenAI-compatible local embedding client;
- query instruction formatting, 512-dimensional reduction, and L2 normalization; and
- live model serving and probing commands;
- content-addressed immutable chunk-vector objects;
- unchanged-vector reuse by embedding and source-content identity;
- versioned aligned float32 snapshot materialization;
- memory-mapped validation;
- a narrow caller-owned-buffer C ABI; and
- exact serial/concurrent semantic search.

The embedding provider is connected through `vector build` and `vector search`. The context compiler remains lexical-only until hybrid fusion is implemented.

## 1. Persistent chunk vectors — implemented

The initial persistent vector layer is complete. See [Vector store](../reference/vector-store.md).

Unresolved follow-on work:

- benchmark float32 against float16 and int8 encodings;
- add immutable-object garbage collection;
- add non-Windows Go dynamic-library loaders;
- automate vector refresh; and
- consider approximate indexing only if exact-scan measurements require it.

## 2. Prepared lexical retrieval

Replace the current full chunk scan and custom substring score with standard lexical retrieval.

Required behavior:

- incremental inverted postings;
- BM25 corpus scoring;
- exact phrase, filename, path, symbol-like term, and heading boosts as separate inspectable signals;
- deterministic tie-breaking; and
- benchmark and retrieval-quality comparison against the current baseline.

This work is independent of Lexicon. Fallback chunks remain valid searchable documents.

## 3. Hybrid retrieval

Retrieve lexical and semantic candidates independently, then combine them.

Initial direction:

- bounded BM25 and vector candidate pools;
- deterministic reciprocal-rank fusion;
- provider-specific ranks and scores retained as provenance;
- lexical-only fallback when the model is unavailable; and
- explicit package metadata describing which retrieval paths contributed.

Raw BM25 and cosine values should not be treated as directly comparable scales.

## 4. Selection quality

Improve final context construction after hybrid retrieval is measurable.

Candidate work:

- overlap removal;
- file and subsystem diversity;
- adjacent-chunk expansion;
- evidence-class reservations;
- stable package fingerprints; and
- explicit omission reasons beyond budget pressure.

Exact `o200k_base` package enforcement remains the final boundary.

## 5. Incremental maintenance runtime

Keep lexical and vector state current without requiring a manual indexing command.

Standalone Grimoire should own its own behavior. When hosted by Warlock, it should consume shared repository change events and supervision rather than duplicate the umbrella runtime.

One-shot CLI indexing must remain supported.

## 6. Optional structural enrichment

Consume Lexicon structural ranges when available while retaining the fallback chunker.

Lexicon may improve chunk boundaries, symbol metadata, and replacement identity. It is not a prerequisite for lexical search, embeddings, vector search, or hybrid retrieval.

## 7. Optional evidence providers

Add bounded provider interfaces for Arcana graph evidence, Demon Docs documentation evidence, Git-change evidence, and other Warlock facts.

Grimoire remains responsible for retrieval fusion, context selection, budgeting, and the final package.

## 8. Stable external contracts

Before a stable release, define:

- CLI compatibility and exit behavior;
- machine-readable diagnostics;
- prepared-index and vector-index migration policy;
- context-package compatibility policy;
- model/runtime compatibility policy; and
- benchmark gates for latency, memory, and retrieval quality.

## Graduation rule

When a roadmap item becomes implemented:

1. Update the owning package README.
2. Update current architecture documentation.
3. Add or update exact reference documentation.
4. Remove or narrow the corresponding limitation.
5. Replace roadmap detail with links and unresolved follow-on work.
