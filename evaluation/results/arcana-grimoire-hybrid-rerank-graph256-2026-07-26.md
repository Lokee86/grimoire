# Arcana semantic graph retrieval evaluation: Grimoire

Generated: 2026-07-26 23:28:51-07:00  
Variant: `hybrid-rerank-graph256`  
Arcana snapshot: `sha256:64a2907de8b3c09a3c37820f3b2249effe2bddc836c391b043dbaea9bbbb3124`  
Embedding identity: `qwen3-embedding-0.6b-q8_0-512d`  
Corpus cases: `5`

The paired modes use the same prepared source, Lexicon export, and Arcana graph snapshot. `lexicon-seeds` bypasses semantic lookup; `lexicon-plus-vector` adds the existing Arcana vector index before the same deterministic graph expansion.

## Aggregate comparison

| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | p95 latency | Median payload | p95 payload | Provider calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.000 | 0.0% | 7217.9 ms | 7875.4 ms | 23835 B | 38373 B | 2.00 |
| lexicon-plus-vector | 0.0% | 10.0% | 0.200 | 20.0% | 8869.0 ms | 10957.0 ms | 59456 B | 61297 B | 3.00 |

## Seed recall at k

| Mode | R@1 | R@3 | R@6 |
| --- | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.0% |
| lexicon-plus-vector | 10.0% | 10.0% | 10.0% |

## Vector-minus-baseline deltas

| Metric | Delta |
| --- | ---: |
| Pass rate | +0.0 pp |
| Required seed recall | +10.0 pp |
| Seed recall@1 | +10.0 pp |
| Seed recall@3 | +10.0 pp |
| Seed recall@6 | +10.0 pp |
| MRR | +0.200 |
| Required structural recall | +20.0 pp |
| Median latency | +1651.2 ms |
| Median payload | +35621 B |
| Mean provider calls | +1.00 |

## Cases

| Case | Mode | Pass | Seed recall | MRR | Structural recall | Latency | Payload | Calls | Error |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `arcana-semantic-seeds` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 7875.4 ms | 38373 B | 2 |  |
| `arcana-semantic-seeds` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 8998.3 ms | 59456 B | 3 |  |
| `arcana-seed-composition` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 7531.5 ms | 23835 B | 2 |  |
| `arcana-seed-composition` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 8750.1 ms | 49468 B | 3 |  |
| `arcana-deterministic-expansion` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 7217.9 ms | 25340 B | 2 |  |
| `arcana-deterministic-expansion` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 8423.3 ms | 61297 B | 3 |  |
| `arcana-graph-documents` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 7108.8 ms | 22011 B | 2 |  |
| `arcana-graph-documents` | lexicon-plus-vector | false | 50.0% | 1.000 | 100.0% | 8869.0 ms | 60600 B | 3 |  |
| `arcana-vector-search` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 6662.5 ms | 17510 B | 2 |  |
| `arcana-vector-search` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 10957.0 ms | 49976 B | 3 |  |

## Per-case seed rankings

### `arcana-semantic-seeds` / `lexicon-seeds`

How does an optional concept search turn ranked graph matches into bounded structural entry points without building new state?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `new` | `arcana/src/storage/oracle.rs` | false | false |
| 2 | lexicon | `new` | `arcana/src/cli_sync_state.rs` | false | false |
| 3 | lexicon | `Search` | `internal/arcanagraph/client.go` | false | false |
| 4 | lexicon | `new` | `arcana/src/benchmark/common.rs` | false | false |
| 5 | lexicon | `new` | `arcana/src/snapshot/graph_tests.rs` | false | false |
| 6 | lexicon | `Entry` | `internal/knowledgevector/state.go` | false | false |

### `arcana-semantic-seeds` / `lexicon-plus-vector`

How does an optional concept search turn ranked graph matches into bounded structural entry points without building new state?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `collectStructuralContext` | `internal/app/context_structure.go` | false | false |
| 2 | vector | `SearchManyWithConfig` | `internal/retrieve/search.go` | false | false |
| 3 | vector | `Search` | `internal/arcanagraph/client.go` | false | false |
| 4 | vector | `Search` | `internal/knowledge/search.go` | false | false |
| 5 | vector | `structuralMatchesAny` | `internal/evaluation/structural_score.go` | false | false |
| 6 | vector | `SearchWithConfig` | `internal/retrieve/search.go` | false | false |

### `arcana-seed-composition` / `lexicon-seeds`

Where are vocabulary-derived symbols and conceptual graph matches balanced before deterministic relationship expansion?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `matches` | `evaluation/agent_discovery/score.go` | false | false |
| 2 | lexicon | `relationship` | `internal/arcanagraph/evidence.go` | false | false |
| 3 | lexicon | `matches` | `internal/knowledgeevaluation/score.go` | false | false |
| 4 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 5 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |
| 6 | lexicon | `graph` | `arcana/src/snapshot/mod.rs` | false | false |

### `arcana-seed-composition` / `lexicon-plus-vector`

Where are vocabulary-derived symbols and conceptual graph matches balanced before deterministic relationship expansion?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `relationshipsForSeed` | `internal/lexiconfacts/evidence.go` | false | false |
| 2 | vector | `graphSourceMatches` | `internal/arcanagraph/candidates.go` | false | false |
| 3 | vector | `expandRelationships` | `internal/lexiconfacts/rank.go` | false | false |
| 4 | vector | `relationshipGraph` | `internal/lexiconfacts/rank.go` | false | false |
| 5 | vector | `relationship` | `internal/arcanagraph/evidence.go` | false | false |
| 6 | vector | `relationshipSite` | `internal/lexiconfacts/relationship_provenance.go` | false | false |

### `arcana-deterministic-expansion` / `lexicon-seeds`

Which boundary resolves a few graph anchors and then asks for roles, dependents, uncertain references, and short execution paths?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `paths` | `arcana/src/protocol/path_queries.rs` | false | false |
| 2 | lexicon | `Paths` | `internal/arcanagraph/query.go` | false | false |
| 3 | lexicon | `References` | `arcana/src/repository/model.rs` | false | false |
| 4 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 5 | lexicon | `Paths` | `arcana/src/protocol/request.rs` | false | false |
| 6 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |

### `arcana-deterministic-expansion` / `lexicon-plus-vector`

Which boundary resolves a few graph anchors and then asks for roles, dependents, uncertain references, and short execution paths?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `resolveAnchors` | `internal/agentquery/resolve.go` | false | false |
| 2 | vector | `operational_role` | `arcana/src/protocol/analysis_queries.rs` | false | false |
| 3 | vector | `paths` | `arcana/src/protocol/path_queries.rs` | false | false |
| 4 | vector | `execute` | `arcana/src/protocol/session.rs` | false | false |
| 5 | vector | `Paths` | `internal/arcanagraph/query.go` | false | false |
| 6 | vector | `addArcanaAnchors` | `internal/agentquery/resolve.go` | false | false |

### `arcana-graph-documents` / `lexicon-seeds`

Where is each repository graph entity rendered with bounded inbound, outbound, and unresolved neighborhood text for embeddings?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `unresolved` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 2 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 3 | lexicon | `text` | `arcana/src/repository/repository_snapshot_validation.rs` | false | false |
| 4 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |
| 5 | lexicon | `Unresolved` | `internal/arcanagraph/query.go` | false | false |
| 6 | lexicon | `unresolved` | `arcana/src/repository/mod.rs` | false | false |

### `arcana-graph-documents` / `lexicon-plus-vector`

Where is each repository graph entity rendered with bounded inbound, outbound, and unresolved neighborhood text for embeddings?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `graph_documents` | `arcana/src/vector/documents.rs` | true | false |
| 2 | vector | `graph_neighbors` | `arcana/src/protocol/traversal.rs` | false | false |
| 3 | vector | `seedGraphProximity` | `internal/arcanagraph/rerank_graph.go` | false | false |
| 4 | vector | `EmbedQueries` | `internal/embedding/query.go` | false | false |
| 5 | vector | `EmbedQuery` | `internal/embedding/client.go` | false | false |
| 6 | vector | `neighbors` | `arcana/src/protocol/queries.rs` | false | false |

### `arcana-vector-search` / `lexicon-seeds`

How is a conceptual task embedded, checked against immutable graph identity, scored over packed node vectors, and returned in stable order?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `Packed` | `arcana/src/snapshot/graph_error.rs` | false | false |
| 2 | lexicon | `Packed` | `arcana/src/snapshot/overlay_error.rs` | false | false |
| 3 | lexicon | `Packed` | `arcana/src/benchmark/error.rs` | false | false |
| 4 | lexicon | `Packed` | `arcana/src/cli_commands.rs` | false | false |
| 5 | lexicon | `Node` | `arcana/src/lexicon/object.rs` | false | false |
| 6 | lexicon | `node` | `arcana/src/cli_update_tests.rs` | false | false |

### `arcana-vector-search` / `lexicon-plus-vector`

How is a conceptual task embedded, checked against immutable graph identity, scored over packed node vectors, and returned in stable order?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `graph_documents` | `arcana/src/vector/documents.rs` | false | false |
| 2 | vector | `embed_batch` | `arcana/src/vector/build.rs` | false | false |
| 3 | vector | `PackedGraph` | `arcana/src/storage/reader.rs` | false | false |
| 4 | vector | `StableID` | `internal/evidence/descriptor.go` | false | false |
| 5 | vector | `seedGraphProximity` | `internal/arcanagraph/rerank_graph.go` | false | false |
| 6 | vector | `nodeGroupID` | `internal/arcanagraph/evidence.go` | false | false |

