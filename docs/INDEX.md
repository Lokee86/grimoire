# Grimoire documentation

This directory is the map for Grimoire's implemented system, operating contracts, verification, limitations, and planned work.

## Sections

- [Architecture](architecture/INDEX.md) — system flow, ownership, state boundaries, and prepared-index design.
- [Reference](reference/INDEX.md) — CLI commands, embedding runtime, indexing, vector storage, query policy, and package schema.
- [Development](development/INDEX.md) — tests, judged corpora, benchmark procedure, and retrieval-quality interpretation.
- [Limits](limits/INDEX.md) — constraints and failure modes that exist now.
- [Planning](planning/INDEX.md) — work that is not yet implemented.

## Core code ownership

| Package | Owns |
| --- | --- |
| [`internal/app`](../internal/app/README.md) | CLI orchestration and cross-package workflows |
| [`internal/index`](../internal/index/README.md) | Prepared source state, chunk identities, and incremental rebuilds |
| [`internal/ignore`](../internal/ignore/README.md) | Repository traversal exclusions and Git-ignore semantics |
| [`internal/embedding`](../internal/embedding/README.md) | Fixed model contract, managed runtime, HTTP client, and query batching |
| [`internal/vectorstore`](../internal/vectorstore/README.md) | Native vector-engine binding, object ingestion, and snapshot access |
| [`internal/retrieve`](../internal/retrieve/README.md) | Lexical fallback, exact recovery, and shared candidate provenance |
| [`internal/selection`](../internal/selection/README.md) | Deterministic deduplication, diversification, and neighbour expansion |
| [`internal/queryshape`](../internal/queryshape/README.md) | Prompt profile and retrieval-policy selection |
| [`internal/assembly`](../internal/assembly/README.md) | Scope-specific evidence coverage and automatic candidate limits |
| [`internal/structure`](../internal/structure/README.md) | Common structural-provider contracts and evidence composition |
| [`internal/lexiconfacts`](../internal/lexiconfacts/README.md) | Immutable Lexicon export matching |
| [`internal/arcanagraph`](../internal/arcanagraph/README.md) | Arcana synchronization and graph protocol queries |
| [`internal/compiler`](../internal/compiler/README.md) | Token accounting and versioned package serialization |
| [`internal/evaluation`](../internal/evaluation/README.md) | Judged corpus model, scoring, aggregation, and reports |
| [`native/vector-engine`](../native/vector-engine/README.md) | Content-addressed vector objects and packed exact-search snapshots |

Package-level README files define the narrower control boundaries.

## Documentation rules

1. Reference pages describe current code, defaults, and failure behavior.
2. Architecture pages identify ownership and data flow, not aspirations.
3. Development pages state how claims are measured and name the report artifacts.
4. Limitations record unresolved current constraints without disguising them as plans.
5. Planning pages contain unimplemented work and must not be cited as current behavior.
6. Version numbers, defaults, commands, and schemas must match code and tests.
7. New top-level documentation must be linked from the nearest `INDEX.md`.

When behavior changes, update the owning package README, its public reference page, and any affected limitations or roadmap entry in the same change.
