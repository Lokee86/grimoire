# System overview

Grimoire contains three independently runnable components arranged as one repository-intelligence pipeline: Lexicon language analysis, Arcana graph analysis, and Grimoire Context retrieval and package construction.

Source co-location does not merge their runtime state or domain ownership. See [Component architecture](components.md).

## Repository intelligence pipeline

```text
Repository files
  -> Lexicon language adapters
  -> immutable Lexicon facts and snapshot
  -> Arcana graph compilation
  -> immutable Arcana graph snapshot

Repository files
  -> Grimoire ignore and eligibility rules
  -> normalized source chunks
  -> immutable prepared snapshot
  -> missing-text embedding batches
  -> content-addressed vector objects
  -> packed vector snapshot
```

Lexicon and Arcana are independently useful and remain optional at Grimoire Context query time. Their co-location makes source evolution and release coordination easier; it does not make source retrieval depend on structural state.

## Query-time construction

```text
Query
  -> query embedding plan
  -> vector search or lexical fallback
  -> concrete exact recovery
  -> available Lexicon and Arcana evidence
  -> candidate merge and ranking
  -> query-shape analysis
  -> selection and neighbour expansion
  -> automatic evidence assembly or explicit fixed budget
  -> package compilation
```

The query profile and retrieval policy are computed after provider candidates are available, allowing prompt semantics to be refined by ranking confidence and graph dispersion without hiding those signals inside the rank score.

When the caller omits a positive budget, Grimoire activates the policy and applies scope-specific evidence assembly. When the caller supplies a positive budget, Grimoire emits the profile in shadow form but retains fixed fit-to-budget assembly.

## Ownership boundaries

### Lexicon

`lexicon/` owns language extraction, normalized source identities, fact contracts, immutable analysis objects, snapshot publication, adapter execution, and deterministic consumer hooks.

### Arcana

`arcana/` owns Lexicon snapshot ingestion, repository graph construction, packed graph storage, overlays, compaction, traversal, impact analysis, path queries, and the graph protocol.

### Context application orchestration

`internal/app` parses Grimoire Context commands, resolves state, schedules independent retrieval providers, applies timeout and fallback rules, and passes typed results between packages. It does not own ranking formulas, graph semantics, vector persistence, or token accounting.

### Source state

`internal/index` owns repository traversal, chunking, exact token counts, immutable object reuse, and prepared snapshot publication. `.git/`, `.grimoire/`, and nested state/output paths are excluded from traversal.

### Embeddings

`internal/embedding` owns the fixed Qwen3 model identity, managed `llama.cpp` runtime, query instructions, request batching, Matryoshka truncation to 512 dimensions, and normalization. It does not persist or rank vectors.

### Vector storage

`internal/vectorstore` is the Go boundary to `native/vector-engine`. The Rust engine owns immutable vector objects, deterministic snapshot materialization, memory-mapped reads, and exact inner-product search.

### Retrieval and ranking

`internal/retrieve` owns deterministic lexical fallback, concrete exact recovery, and the shared candidate provenance shape. `internal/app` orchestrates vector search and merges vector, exact, lexical, and structural-provider candidates. Concrete exact signals supplement ranked search rather than replacing it. Missing or incompatible semantic state degrades to lexical retrieval with a warning.

### Structural integration

`internal/structure` defines common evidence and provider-state contracts. `internal/lexiconfacts` matches immutable Lexicon exports. `internal/arcanagraph` synchronizes and queries Arcana using Lexicon matches as bounded graph seeds.

These packages integrate the components; they do not take ownership of Lexicon or Arcana domain logic. Structural failures are non-fatal to source retrieval.

### Selection and policy

`internal/selection` deduplicates, diversifies, and expands prepared neighbours. `internal/queryshape` classifies intent, specificity, breadth, ambiguity, cross-system scope, and evidence needs. `internal/assembly` preserves a scope-appropriate candidate pool and stops on deterministic evidence coverage.

### Compilation

`internal/compiler` owns exact `o200k_base` accounting, package versioning, omission counts, and final JSON serialization. It receives already ranked and, for automatic requests, already assembled evidence.

### Evaluation

`internal/evaluation` owns corpus validation, source and structural scoring, pipeline-loss attribution, aggregate metrics, and Markdown/JSON reporting. `internal/app` runs the production pipeline for each case.

## Failure and fallback boundaries

- A stale or missing vector snapshot prevents semantic search but not lexical context construction.
- A failed Lexicon or Arcana provider emits warnings and does not discard source evidence.
- Arcana state remains explicitly bound to the Lexicon snapshot it consumed.
- Explicit backend or runtime errors fail setup or service startup rather than silently changing the requested backend.
- Automatic assembly losses and final budget-fitting losses are recorded as separate evaluation stages.
- Package compilation remains deterministic for identical prepared state, provider evidence, query, and options.

## State directories

- `.grimoire/` — prepared and vector state.
- `.lexicon/` — Lexicon immutable analysis state.
- `.arcana/` — Arcana graph state.

The source code now shares one repository, but these state formats remain independently versioned and owned. Integration occurs through manifests, exports, and protocols rather than direct cross-component mutation.
