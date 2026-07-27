# Grimoire documentation

This directory maps the Grimoire platform, its context engine, operating contracts, verification, limitations, and planned work. Lexicon and Arcana retain their own component documentation under their source directories.

## Sections

- [Architecture](architecture/INDEX.md) — component boundaries, system flow, state ownership, and prepared-index design.
- [Reference](reference/INDEX.md) — CLI commands, deterministic source retrieval, independent knowledge retrieval, documentation vectors, query policy, and package schemas.
- [Development](development/INDEX.md) — tests, judged corpora, benchmark procedure, and retrieval-quality interpretation.
- [Limits](limits/INDEX.md) — constraints and failure modes that exist now.
- [Planning](planning/INDEX.md) — work that is not yet implemented.
- [Investigation ledger](../internal/investigation/README.md) — persistent agent-facing discovery state and integration contract.
- [Lexicon documentation](../lexicon/docs/README.md) — language analysis, adapters, snapshots, contracts, and validation.
- [Arcana documentation](../arcana/docs/) — graph ingestion and repository-snapshot contracts.

## Component ownership

| Component | Location | Owns |
| --- | --- | --- |
| Lexicon | [`lexicon/`](../lexicon/) | Language analysis, normalized facts, adapters, immutable analysis snapshots |
| Arcana | [`arcana/`](../arcana/) | Graph ingestion, packed graph state, traversal, impact, paths, and graph protocol |
| Grimoire | repository root | Deterministic source retrieval, independent documentation knowledge, ranking, budgeting, evidence assembly, context packages, and investigation-ledger orchestration |

See [Component architecture](architecture/components.md) for the dependency and independent-use rules.

## Context-engine code ownership

| Package | Owns |
| --- | --- |
| [`internal/app`](../internal/app/README.md) | CLI orchestration and cross-package workflows |
| [`internal/index`](../internal/index/README.md) | Prepared source state, chunk identities, and incremental rebuilds |
| [`internal/ignore`](../internal/ignore/README.md) | Repository traversal exclusions and Git-ignore semantics |
| [`internal/embedding`](../internal/embedding/README.md) | Fixed model contract, managed runtime, HTTP client, and embedding requests |
| [`internal/knowledge`](../internal/knowledge/README.md) | Documentation discovery, section identities, BM25 ranking, filters, citations, and code links |
| [`internal/knowledgevector`](../internal/knowledgevector/README.md) | Optional documentation-vector construction, freshness, and supplemental ranking |
| [`internal/vectorstore`](../internal/vectorstore/README.md) | Native vector-engine binding, object ingestion, serialized snapshot access, and exact scanning |
| [`internal/retrieve`](../internal/retrieve/README.md) | Deterministic source BM25, exact recovery, and shared candidate provenance |
| [`internal/selection`](../internal/selection/README.md) | Deterministic deduplication, diversification, and neighbour expansion |
| [`internal/queryshape`](../internal/queryshape/README.md) | Prompt profile and retrieval-policy selection |
| [`internal/assembly`](../internal/assembly/README.md) | Scope-specific evidence coverage and automatic candidate limits |
| [`internal/structure`](../internal/structure/README.md) | Common structural-provider contracts and evidence composition |
| [`internal/lexiconfacts`](../internal/lexiconfacts/README.md) | Immutable Lexicon export matching |
| [`internal/arcanagraph`](../internal/arcanagraph/README.md) | Arcana synchronization and graph protocol queries |
| [`internal/repostate`](../internal/repostate/README.md) | Repository identity inspection, deterministic state preparation, and Arcana-vector availability |
| [`internal/agentquery`](../internal/agentquery/README.md) | Progressive orient, search, trace, impact, and inspect contracts |
| [`internal/agentruntime`](../internal/agentruntime/README.md) | Unified code/knowledge query orchestration and investigation-handle reuse |
| [`internal/investigation`](../internal/investigation/README.md) | Persistent snapshot-bound evidence ledger |
| [`internal/compiler`](../internal/compiler/README.md) | Token accounting and versioned package serialization |
| [`internal/evaluation`](../internal/evaluation/README.md) | Judged source-corpus model, scoring, aggregation, and reports |
| [`internal/knowledgeevaluation`](../internal/knowledgeevaluation/README.md) | Judged documentation corpus, scoring, aggregation, and reports |
| [Lodestone](../../lodestone/README.md) | External content-addressed vector database and packed exact-search snapshots |

Package-level README files define the narrower control boundaries.

## Documentation rules

1. Reference pages describe current code, defaults, and failure behavior.
2. Architecture pages identify ownership and data flow, not aspirations.
3. Development pages state how claims are measured and name the report artifacts.
4. Limitations record unresolved current constraints without disguising them as plans.
5. Planning pages contain unimplemented work and must not be cited as current behavior.
6. Version numbers, defaults, commands, and schemas must match code and tests.
7. New top-level documentation must be linked from the nearest `INDEX.md`.
8. Component-specific behavior belongs in that component's documentation even though the source shares one repository.

When behavior changes, update the owning package README, its public reference page, and any affected limitations or roadmap entry in the same change.
