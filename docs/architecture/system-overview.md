# System overview

Grimoire contains three independently runnable components arranged as one repository-intelligence pipeline: Lexicon language analysis, Arcana graph analysis, and Grimoire retrieval and package construction.

Source co-location does not merge their runtime state or domain ownership. See [Component architecture](components.md).

## Repository intelligence pipeline

```text
Repository source
  -> Grimoire ignore and eligibility rules
  -> Lexicon-aligned source chunks plus deterministic fallback gaps
  -> immutable prepared source snapshot and BM25 sidecar

Repository source
  -> Lexicon language adapters
  -> immutable Lexicon facts and snapshot
  -> Arcana graph compilation
  -> immutable Arcana graph snapshot
  -> optional Arcana semantic graph index

Repository documentation
  -> independent Grimoire knowledge index
  -> optional documentation embedding batches
  -> content-addressed vector objects
  -> packed documentation vector snapshot
```

Lexicon and Arcana remain independently useful and optional at query time. Documentation vectors are also optional. Source retrieval does not depend on an embedding service or a repository-wide code-vector snapshot.

## Query-time construction

```text
Source query
  -> deterministic BM25 retrieval
  -> concrete exact recovery
  -> Lexicon symbol evidence
  -> Arcana graph traversal and optional semantic graph seeds
  -> candidate merge and deterministic ranking
  -> query-shape analysis
  -> selection and neighbour expansion
  -> automatic evidence assembly or explicit fixed budget
  -> context package compilation

Knowledge query
  -> documentation BM25
  -> optional documentation-vector score
  -> cited knowledge sections
```

The query profile and retrieval policy are computed after source and structural candidates are available. When the caller omits a positive budget, Grimoire activates scope-specific evidence assembly. A positive budget retains fixed fit-to-budget assembly.

## Ownership boundaries

### Lexicon

`lexicon/` owns language extraction, normalized source identities, fact contracts, immutable analysis objects, snapshot publication, adapter execution, and deterministic consumer hooks.

### Arcana

`arcana/` owns Lexicon snapshot ingestion, repository graph construction, packed graph storage, overlays, compaction, graph-derived semantic documents and vector indexes, semantic graph search, traversal, impact analysis, path queries, and the graph protocol. Arcana calls the shared embedding endpoint but stores and invalidates its index inside `.arcana/`.

### Context application orchestration

`internal/app` parses commands, resolves state, schedules independent providers, applies timeout and fallback rules, and passes typed results between packages. Normal source context uses deterministic source retrieval; application-level vector commands own documentation vectors.

### Source state

`internal/index` owns repository traversal, chunking, exact token counts, lexical postings, immutable object reuse, and prepared snapshot publication. `.git/`, `.grimoire/`, nested worktree containers, and generated state are excluded from traversal.

### Knowledge state

`internal/knowledge` owns documentation discovery, section extraction, stable citation handles, BM25 ranking, code links, and graceful vector-score supplementation. `internal/knowledgevector` owns documentation-vector build, freshness validation, and the `knowledge.VectorRanker` implementation.

### Embeddings

`internal/embedding` owns the fixed Qwen3 model identity, managed `llama.cpp` runtime, request contracts, Matryoshka truncation to 512 dimensions, and normalization. It does not decide source relevance or persist vectors.

### Vector storage

`internal/vectorstore` is the Go boundary to `native/vector-engine`. The Rust engine owns immutable vector objects, deterministic snapshot materialization, memory-mapped reads, and exact inner-product search. One Go snapshot handle serializes ABI operations because the native handle is not re-entrant.

### Retrieval and ranking

`internal/retrieve` owns deterministic BM25 source retrieval, concrete exact recovery, and shared candidate provenance. `internal/app` merges exact, lexical, Lexicon-derived, and Arcana-derived source candidates. Documentation vector scores never enter source-candidate ranking.

### Structural integration

`internal/structure` defines common evidence and provider-state contracts. `internal/lexiconfacts` matches immutable Lexicon exports. `internal/arcanagraph` synchronizes and queries Arcana using Lexicon matches and, when available, Arcana-owned semantic graph matches as bounded graph seeds. Source-bearing graph results are localized to prepared chunks without moving graph ownership into Grimoire.

Structural failures are non-fatal to source retrieval.

### Selection and policy

`internal/selection` deduplicates, diversifies, and expands prepared neighbours. `internal/queryshape` classifies intent, specificity, breadth, ambiguity, cross-system scope, and evidence needs. `internal/assembly` preserves a scope-appropriate candidate pool and stops on deterministic evidence coverage.

### Compilation

`internal/compiler` owns exact `o200k_base` accounting, package versioning, omission counts, and final JSON serialization.

### Evaluation

`internal/evaluation` owns deterministic source and structural corpus validation, pipeline-loss attribution, aggregate metrics, and report publication. `internal/knowledgeevaluation` independently owns judged documentation retrieval. The retired repository-wide source-vector evaluator and implementation have been removed.

## Failure and fallback boundaries

- Missing or stale documentation vectors cause knowledge search to remain BM25-only; source context is unaffected.
- A missing Arcana vector index falls back to Lexicon-seeded graph traversal; a failed query against an existing Arcana index warns without discarding other evidence.
- A failed Lexicon or Arcana provider warns without discarding deterministic source evidence.
- Arcana state remains explicitly bound to the Lexicon snapshot it consumed.
- Explicit backend or runtime errors fail setup or service startup rather than silently changing the requested backend.
- Automatic assembly losses and final budget-fitting losses are recorded separately.
- Package compilation remains deterministic for identical prepared state, provider evidence, query, and options.

## State directories

- `.grimoire/` — prepared source state, investigations, and independent knowledge state.
- `.grimoire/knowledge/vectors/<model>/` — optional documentation vector objects and packed snapshot.
- `.lexicon/` — Lexicon immutable analysis state.
- `.arcana/` — Arcana graph state and optional graph-vector indexes.

These formats remain independently versioned and owned. Integration occurs through manifests, exports, and protocols rather than direct cross-component mutation.
