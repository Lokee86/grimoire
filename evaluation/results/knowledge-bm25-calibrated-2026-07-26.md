# Knowledge retrieval evaluation: grimoire

Generated: 2026-07-26 02:35:58-07:00  
Variant: `bm25`  
Vectors requested: `false`  
Corpus cases: `8`  

## Aggregate

| Metric | Value |
| --- | ---: |
| Pass rate | 100.0% |
| Required-section recall | 100.0% |
| Recall@1 | 50.0% |
| Recall@3 | 62.5% |
| Recall@5 | 100.0% |
| MRR | 0.938 |
| Irrelevant selections | 19 (47.5%) |
| Vector usage | 0/8 (0.0%) |
| Vector errors | 0 |
| Median latency | 4.5 ms |
| p95 latency | 10.8 ms |

## Cases

| Case | Category | Pass | Required recall | MRR | Irrelevant | Vector | Latency |
| --- | --- | ---: | ---: | ---: | ---: | --- | ---: |
| `architecture-knowledge-separate-lane` | architecture-rationale | true | 100.0% | 1.000 | 3 (60.0%) | no | 5.5 ms |
| `architecture-vector-supplement` | architecture-rationale | true | 100.0% | 1.000 | 3 (60.0%) | no | 4.1 ms |
| `command-knowledge-index` | command-behavior | true | 100.0% | 1.000 | 1 (20.0%) | no | 5.1 ms |
| `command-bm25-only` | command-behavior | true | 100.0% | 1.000 | 3 (60.0%) | no | 10.8 ms |
| `failure-stale-vector` | failure-fallback | true | 100.0% | 1.000 | 2 (40.0%) | no | 3.5 ms |
| `failure-source-preserved` | failure-fallback | true | 100.0% | 1.000 | 2 (40.0%) | no | 4.0 ms |
| `ownership-knowledge-bm25` | ownership-boundary | true | 100.0% | 1.000 | 2 (40.0%) | no | 5.5 ms |
| `ownership-knowledge-vectors` | ownership-boundary | true | 100.0% | 0.500 | 3 (60.0%) | no | 4.5 ms |

## Per-case rankings

### `architecture-knowledge-separate-lane`

why does Grimoire index repository rationale separately from production source chunks

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/knowledge.md` | Knowledge retrieval | 25.7194 | true |
| 2 | `docs/reference/vector-store.md` | Vector store | 12.7416 | false |
| 3 | `docs/reference/indexing.md` | Documentation indexing and vectors | 12.6000 | true |
| 4 | `docs/reference/indexing.md` | Prepared source index | 12.3059 | false |
| 5 | `docs/development/testing-and-benchmarks.md` | Prepared source and knowledge-vector smoke test | 11.4578 | false |

### `architecture-vector-supplement`

why are documentation vectors supplemental to deterministic BM25

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `internal/app/README.md` | Knowledge pipeline | 22.9019 | true |
| 2 | `docs/reference/knowledge.md` | Judged documentation evaluation | 20.9837 | false |
| 3 | `docs/development/retrieval-quality.md` | Judged documentation retrieval | 20.3598 | false |
| 4 | `docs/reference/cli.md` | `grimoire knowledge` | 20.1816 | false |
| 5 | `docs/reference/knowledge.md` | Search | 19.6761 | true |

### `command-knowledge-index`

which command builds the independent documentation rationale index

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/cli.md` | `grimoire knowledge` | 17.0775 | true |
| 2 | `internal/app/README.md` | Commands | 12.2384 | false |
| 3 | `README.md` | Quick start | 11.0433 | true |
| 4 | `docs/reference/indexing.md` | Documentation indexing and vectors | 10.5100 | true |
| 5 | `docs/reference/knowledge.md` | Knowledge retrieval | 10.2519 | true |

### `command-bm25-only`

how can knowledge search be forced to use BM25 only

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/knowledge.md` | Search | 19.7800 | true |
| 2 | `docs/reference/indexing.md` | Documentation indexing and vectors | 19.4496 | false |
| 3 | `internal/app/README.md` | Knowledge pipeline | 18.6257 | false |
| 4 | `docs/limits/current-limitations.md` | The Go native loader is Windows-only | 17.1703 | false |
| 5 | `docs/reference/cli.md` | `grimoire knowledge` | 16.3296 | true |

### `failure-stale-vector`

what happens when documentation vectors are stale unavailable or incompatible

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/vector-store.md` | Compatibility and discovery | 16.8430 | true |
| 2 | `internal/app/README.md` | Knowledge pipeline | 16.1913 | true |
| 3 | `docs/reference/indexing.md` | Documentation indexing and vectors | 14.3574 | false |
| 4 | `docs/reference/knowledge.md` | Search | 14.3282 | true |
| 5 | `docs/development/retrieval-quality.md` | Judged documentation retrieval | 13.4808 | false |

### `failure-source-preserved`

does missing documentation vector state fail knowledge retrieval

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/architecture/system-overview.md` | Failure and fallback boundaries | 14.6375 | true |
| 2 | `docs/limits/current-limitations.md` | State maintenance is explicit | 13.8529 | true |
| 3 | `docs/reference/knowledge.md` | Search | 13.5369 | true |
| 4 | `internal/app/README.md` | Commands | 13.3043 | false |
| 5 | `docs/development/testing-and-benchmarks.md` | Judged retrieval evaluation | 13.2049 | false |

### `ownership-knowledge-bm25`

how does internal knowledge rank every query with deterministic BM25 and preserve results when an optional vector ranker fails

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `internal/knowledge/README.md` | Responsibilities | 39.0310 | true |
| 2 | `internal/app/README.md` | Knowledge pipeline | 30.5538 | true |
| 3 | `docs/reference/knowledge.md` | Search | 28.4562 | true |
| 4 | `docs/development/retrieval-quality.md` | Judged documentation retrieval | 27.9057 | false |
| 5 | `internal/knowledgeevaluation/README.md` | Responsibilities | 27.3627 | false |

### `ownership-knowledge-vectors`

which package owns optional documentation semantic ranking and what remains BM25

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/architecture/system-overview.md` | Query-time construction | 20.4728 | false |
| 2 | `docs/architecture/system-overview.md` | Knowledge state | 19.0288 | true |
| 3 | `docs/architecture/system-overview.md` | Retrieval and ranking | 18.3726 | false |
| 4 | `docs/reference/vector-store.md` | Ownership boundary | 17.9878 | true |
| 5 | `README.md` | System flow | 17.7810 | false |

