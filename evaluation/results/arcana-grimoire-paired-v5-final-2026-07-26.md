# Arcana semantic graph retrieval evaluation: Grimoire

Generated: 2026-07-26 20:56:11-07:00  
Variant: `paired-v5-final`  
Arcana snapshot: `sha256:4a9cefb03363fb367355195aa56212caa7ee0babb5cad795010fe5f799945b75`  
Embedding identity: `qwen3-embedding-0.6b-q8_0-512d`  
Corpus cases: `5`

The paired modes use the same prepared source, Lexicon export, and Arcana graph snapshot. `lexicon-seeds` bypasses semantic lookup; `lexicon-plus-vector` adds the existing Arcana vector index before the same deterministic graph expansion.

## Aggregate comparison

| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | p95 latency | Median payload | p95 payload | Provider calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.000 | 0.0% | 4892.8 ms | 6043.7 ms | 23835 B | 38342 B | 2.00 |
| lexicon-plus-vector | 0.0% | 0.0% | 0.000 | 0.0% | 5894.9 ms | 6697.5 ms | 28611 B | 57601 B | 3.00 |

## Seed recall at k

| Mode | R@1 | R@3 | R@6 |
| --- | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.0% |
| lexicon-plus-vector | 0.0% | 0.0% | 0.0% |

## Vector-minus-baseline deltas

| Metric | Delta |
| --- | ---: |
| Pass rate | +0.0 pp |
| Required seed recall | +0.0 pp |
| Seed recall@1 | +0.0 pp |
| Seed recall@3 | +0.0 pp |
| Seed recall@6 | +0.0 pp |
| MRR | +0.000 |
| Required structural recall | +0.0 pp |
| Median latency | +1002.1 ms |
| Median payload | +4776 B |
| Mean provider calls | +1.00 |

## Cases

| Case | Mode | Pass | Seed recall | MRR | Structural recall | Latency | Payload | Calls | Error |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `arcana-semantic-seeds` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 6043.7 ms | 38342 B | 2 |  |
| `arcana-semantic-seeds` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 6631.2 ms | 51203 B | 3 |  |
| `arcana-seed-composition` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 5127.7 ms | 23835 B | 2 |  |
| `arcana-seed-composition` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 5854.3 ms | 25037 B | 3 |  |
| `arcana-deterministic-expansion` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 4478.3 ms | 25653 B | 2 |  |
| `arcana-deterministic-expansion` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 6697.5 ms | 57601 B | 3 |  |
| `arcana-graph-documents` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 4892.8 ms | 23055 B | 2 |  |
| `arcana-graph-documents` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 5894.9 ms | 7906 B | 3 |  |
| `arcana-vector-search` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 4650.4 ms | 17510 B | 2 |  |
| `arcana-vector-search` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 5682.2 ms | 28611 B | 3 |  |

## Per-case seed rankings

### `arcana-semantic-seeds` / `lexicon-seeds`

How does an optional concept search turn ranked graph matches into bounded structural entry points without building new state?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `new` | `arcana/src/cli_sync_state.rs` | false | false |
| 2 | lexicon | `new` | `arcana/src/snapshot/graph_tests.rs` | false | false |
| 3 | lexicon | `new` | `arcana/src/storage/oracle.rs` | false | false |
| 4 | lexicon | `Search` | `internal/arcanagraph/client.go` | false | false |
| 5 | lexicon | `Entry` | `internal/knowledgevector/state.go` | false | false |
| 6 | lexicon | `new` | `arcana/src/benchmark/common.rs` | false | false |

### `arcana-semantic-seeds` / `lexicon-plus-vector`

How does an optional concept search turn ranked graph matches into bounded structural entry points without building new state?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `Search` | `internal/knowledge/search.go` | false | false |
| 2 | lexicon | `new` | `arcana/src/cli_sync_state.rs` | false | false |
| 3 | vector | `bm25Score` | `internal/knowledge/search.go` | false | false |
| 4 | lexicon | `new` | `arcana/src/snapshot/graph_tests.rs` | false | false |
| 5 | vector | `SearchManyWithConfig` | `internal/retrieve/search.go` | false | false |
| 6 | lexicon | `new` | `arcana/src/storage/oracle.rs` | false | false |

### `arcana-seed-composition` / `lexicon-seeds`

Where are vocabulary-derived symbols and conceptual graph matches balanced before deterministic relationship expansion?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `relationship` | `internal/arcanagraph/evidence.go` | false | false |
| 2 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 3 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |
| 4 | lexicon | `graph` | `arcana/src/snapshot/mod.rs` | false | false |
| 5 | lexicon | `matches` | `evaluation/agent_discovery/score.go` | false | false |
| 6 | lexicon | `matches` | `internal/knowledgeevaluation/score.go` | false | false |

### `arcana-seed-composition` / `lexicon-plus-vector`

Where are vocabulary-derived symbols and conceptual graph matches balanced before deterministic relationship expansion?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `expandRelationships` | `internal/lexiconfacts/rank.go` | false | false |
| 2 | lexicon | `relationship` | `internal/arcanagraph/evidence.go` | false | false |
| 3 | vector | `relationshipAggregate` | `internal/lexiconfacts/relationship_provenance.go` | false | false |
| 4 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 5 | vector | `mergeRelationshipCandidate` | `internal/lexiconfacts/rank.go` | false | false |
| 6 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |

### `arcana-deterministic-expansion` / `lexicon-seeds`

Which boundary resolves a few graph anchors and then asks for roles, dependents, uncertain references, and short execution paths?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `Paths` | `internal/arcanagraph/query.go` | false | false |
| 2 | lexicon | `paths` | `arcana/src/protocol/path_queries.rs` | false | false |
| 3 | lexicon | `Paths` | `arcana/src/protocol/request.rs` | false | false |
| 4 | lexicon | `References` | `arcana/src/repository/model.rs` | false | false |
| 5 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 6 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |

### `arcana-deterministic-expansion` / `lexicon-plus-vector`

Which boundary resolves a few graph anchors and then asks for roles, dependents, uncertain references, and short execution paths?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `operational_role` | `arcana/src/protocol/analysis_queries.rs` | false | false |
| 2 | lexicon | `Paths` | `internal/arcanagraph/query.go` | false | false |
| 3 | vector | `resolveAnchors` | `internal/agentquery/resolve.go` | false | false |
| 4 | lexicon | `paths` | `arcana/src/protocol/path_queries.rs` | false | false |
| 5 | vector | `execute` | `arcana/src/protocol/session.rs` | false | false |
| 6 | lexicon | `Paths` | `arcana/src/protocol/request.rs` | false | false |

### `arcana-graph-documents` / `lexicon-seeds`

Where is each repository graph entity rendered with bounded inbound, outbound, and unresolved neighborhood text for embeddings?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `unresolved` | `arcana/src/repository/mod.rs` | false | false |
| 2 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 3 | lexicon | `unresolved` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 4 | lexicon | `Graph` | `arcana/src/repository/repository_snapshot_error.rs` | false | false |
| 5 | lexicon | `text` | `arcana/src/repository/repository_snapshot_validation.rs` | false | false |
| 6 | lexicon | `Unresolved` | `internal/arcanagraph/query.go` | false | false |

### `arcana-graph-documents` / `lexicon-plus-vector`

Where is each repository graph entity rendered with bounded inbound, outbound, and unresolved neighborhood text for embeddings?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `embeddingsResponse` | `internal/embedding/client.go` | false | false |
| 2 | lexicon | `unresolved` | `arcana/src/repository/mod.rs` | false | false |
| 3 | vector | `query.go` | `internal/embedding/query.go` | false | false |
| 4 | lexicon | `graph` | `arcana/src/repository/repository_snapshot.rs` | false | false |
| 5 | vector | `QueryInput` | `internal/embedding/query.go` | false | false |
| 6 | lexicon | `unresolved` | `arcana/src/repository/repository_snapshot.rs` | false | false |

### `arcana-vector-search` / `lexicon-seeds`

How is a conceptual task embedded, checked against immutable graph identity, scored over packed node vectors, and returned in stable order?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `Packed` | `arcana/src/snapshot/graph_error.rs` | false | false |
| 2 | lexicon | `Packed` | `arcana/src/snapshot/overlay_error.rs` | false | false |
| 3 | lexicon | `Packed` | `arcana/src/benchmark/error.rs` | false | false |
| 4 | lexicon | `Packed` | `arcana/src/cli_commands.rs` | false | false |
| 5 | lexicon | `node` | `arcana/src/cli_update_tests.rs` | false | false |
| 6 | lexicon | `Node` | `arcana/src/lexicon/object.rs` | false | false |

### `arcana-vector-search` / `lexicon-plus-vector`

How is a conceptual task embedded, checked against immutable graph identity, scored over packed node vectors, and returned in stable order?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `main` | `internal/graphrank.test` | false | false |
| 2 | lexicon | `Packed` | `arcana/src/snapshot/graph_error.rs` | false | false |
| 3 | vector | `recognizedTasks` | `internal/queryshape/signals.go` | false | false |
| 4 | lexicon | `Packed` | `arcana/src/snapshot/overlay_error.rs` | false | false |
| 5 | vector | `graph_documents` | `arcana/src/vector/documents.rs` | false | false |
| 6 | lexicon | `Packed` | `arcana/src/benchmark/error.rs` | false | false |

