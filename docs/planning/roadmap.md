# Grimoire Roadmap

This roadmap describes implementation order, not release commitments.

## Current foundation

Implemented:

- incremental file records with unchanged-file reuse;
- Git-ignore traversal and protected tool-state exclusions;
- content-addressed prepared storage with atomic publication;
- deterministic fallback chunks;
- exact `o200k_base` chunk and package accounting;
- the fixed Qwen3 embedding-model contract;
- verified managed model/runtime setup on Windows x64;
- an OpenAI-compatible local embedding client;
- query instruction formatting, 512-dimensional reduction, and L2 normalization;
- content-addressed immutable chunk-vector objects;
- unchanged-vector reuse by embedding and source-content identity;
- versioned aligned float32 snapshot materialization;
- memory-mapped validation;
- a narrow caller-owned-buffer C ABI;
- exact serial/concurrent semantic search;
- vector-backed `grimoire context` retrieval;
- a persistent vector-snapshot manifest bound to the exact prepared-index identity; and
- deterministic lexical fallback when semantic retrieval is unavailable.

The normal context path now performs exact full-vector retrieval. BM25 or another general lexical engine is not a prerequisite and should only be added if measured retrieval failures justify its cost.

## 1. Selection quality

Improve final context construction now that semantic retrieval reaches the compiler.

Candidate work:

- overlap and duplicate removal;
- adjacent-chunk expansion;
- file and subsystem diversity;
- evidence-class reservations;
- stable package fingerprints;
- explicit omission reasons beyond budget pressure; and
- retrieval-quality fixtures covering source code and documentation.

Exact `o200k_base` package enforcement remains the final boundary.

## 2. Targeted exact recovery

Add inexpensive exact lookup for cases embeddings commonly under-rank:

- paths and filenames;
- raw identifiers;
- quoted phrases;
- configuration keys;
- error codes; and
- version strings.

These lookups should be conditional and compact, not a mandatory full-text search pass. Their candidates must retain source, rank, and reason provenance.

## 3. Incremental maintenance runtime

Keep prepared and vector state current without requiring separate manual commands.

Standalone Grimoire should own its own behavior. When hosted by Warlock, it should consume shared repository change events and supervision rather than duplicate the umbrella runtime.

One-shot `index` and `vector build` commands must remain supported.

## 4. Optional structural enrichment

Consume Lexicon structural ranges and symbol facts when available while retaining the fallback chunker.

Lexicon may improve chunk boundaries, symbol metadata, exact lookup, and replacement identity. It is not a prerequisite for embeddings, vector search, context compilation, or fallback operation.

## 5. Optional evidence providers

Add bounded provider interfaces for Arcana graph evidence, Demon Docs documentation evidence, Git-change evidence, and other Warlock facts.

Grimoire remains responsible for retrieval, context selection, budgeting, provenance, and the final package.

## 6. Vector-engine follow-on work

Measure before increasing storage or search complexity:

- benchmark float32 against float16 and int8 encodings;
- add immutable-object garbage collection;
- add non-Windows Go dynamic-library loaders;
- optimize exact-scan kernels when measurements justify it; and
- consider approximate indexing only if exact-scan latency becomes material.

## 7. Stable external contracts

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
