# Grimoire documentation

This directory documents Grimoire's unified discovery interface, component ownership, repository state, verification, limitations, and planned work. Lexicon and Arcana retain component-specific documentation under their source directories.

## Sections

- [Architecture](architecture/INDEX.md) — component boundaries, the Lexicon–Arcana–Grimoire lifecycle, codemaps, state ownership, and prepared-index design.
- [Reference](reference/INDEX.md) — installation, agent setup, discovery commands, Lexicon and Arcana operations, JSON/MCP contracts, source and document semantics, vectors, and state.
- [Development](development/INDEX.md) — tests, judged corpora, benchmark procedure, implementation coverage, behavioral contracts, and outcome interpretation.
- [Limits](limits/INDEX.md) — current constraints and failure modes.
- [Planning](planning/INDEX.md) — unimplemented work.
- [Documentation policy](documentation-policy.md) — canonical ownership, product/component boundaries, coverage, and status rules.
- [Documentation procedure](documentation-procedure.md) — required update, graduation, indexing, and verification workflow.
- [Investigation ledger](../internal/investigation/README.md) — persistent agent-facing discovery state.
- [Lexicon reference](reference/lexicon.md) and [Lexicon documentation](../lexicon/docs/README.md) — language analysis, adapters, snapshots, contracts, operations, and codemap.
- [Arcana reference](reference/arcana.md) and [Arcana documentation](../arcana/docs/README.md) — graph ingestion, packed storage, snapshots, protocol operations, vectors, and codemap.

## Component ownership

| Component | Location | Owns |
| --- | --- | --- |
| Grimoire | repository root | Unified discovery API, source and document retrieval, stable handles, state preparation, and investigation sessions |
| Lexicon | [`lexicon/`](../lexicon/) | Language analysis, normalized symbols and relationships, adapters, and immutable snapshots |
| Arcana | [`arcana/`](../arcana/) | Graph ingestion, packed graph state, traversal, impact, paths, and graph protocol |

See [Component architecture](architecture/components.md) for dependency and independent-use rules.

## Grimoire package ownership

| Package | Owns |
| --- | --- |
| [`internal/app`](../internal/app/README.md) | CLI/MCP orchestration and repository preparation |
| [`internal/agentruntime`](../internal/agentruntime/README.md) | Flattened source/document/symbol/relationship discovery and sessions |
| [`internal/evidence`](../internal/evidence/README.md) | Provider-neutral intents, roles, links, graph signals, identities, and merge rules |
| [`internal/lexical`](../internal/lexical/README.md) | Shared deterministic tokenization, document analysis, postings, candidates, and lexical persistence |
| [`internal/agentquery`](../internal/agentquery/README.md) | Orient, search, trace, impact, and source inspection |
| [`internal/index`](../internal/index/README.md) | Prepared source state and immutable chunk identities |
| [`internal/retrieve`](../internal/retrieve/README.md) | Exact and BM25 source discovery |
| [`internal/knowledge`](../internal/knowledge/README.md) | Document indexing, sections, freshness, BM25, citations, and code links |
| [`internal/knowledgevector`](../internal/knowledgevector/README.md) | Optional document-vector ranking and freshness |
| [`internal/lexiconfacts`](../internal/lexiconfacts/README.md) | Read-only Lexicon export integration |
| [`internal/arcanagraph`](../internal/arcanagraph/README.md) | Arcana graph protocol integration |
| [`internal/structure`](../internal/structure/README.md) | Provider-neutral symbol and relationship contracts |
| [`internal/repostate`](../internal/repostate/README.md) | Repository identity and aligned state preparation |
| [`internal/investigation`](../internal/investigation/README.md) | Persistent snapshot-bound evidence ledger |
| [`internal/embedding`](../internal/embedding/README.md) | Embedding runtime and request contract |
| [`internal/vectorstore`](../internal/vectorstore/README.md) | Lodestone vector-storage compatibility boundary |
| [`internal/knowledgeevaluation`](../internal/knowledgeevaluation/README.md) | Judged document-retrieval evaluation |
| [`evaluation/agent_discovery`](../evaluation/agent_discovery/README.md) | Progressive discovery and agent-outcome scoring |

Historical package-assembly reports may remain as calibration records. The retired assembly, compiler, curation, query-shape, and source-evaluation implementations are no longer part of the Grimoire codebase.

## Documentation rules

1. Reference pages describe current code, defaults, and failure behavior.
2. Architecture pages identify ownership and data flow, not aspirations.
3. Development pages state how claims are measured and name report artifacts.
4. Limitations record unresolved constraints without disguising them as plans.
5. Planning pages contain unimplemented work and must not be cited as current behavior.
6. Commands, schemas, defaults, and field names must match code and tests.
7. Source and documentation must be described as separate evidence classes.
8. Component-specific behavior belongs in the owning component's documentation.

When behavior changes, update the owning package README, public reference pages, and affected limitations or roadmap entries in the same change.
