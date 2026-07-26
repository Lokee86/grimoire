# Knowledge retrieval evaluation: grimoire

Generated: 2026-07-26 02:37:15-07:00  
Variant: `bm25-vector`  
Vectors requested: `true`  
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
| Vector usage | 8/8 (100.0%) |
| Vector errors | 0 |
| Median latency | 36.9 ms |
| p95 latency | 342.4 ms |

## Cases

| Case | Category | Pass | Required recall | MRR | Irrelevant | Vector | Latency |
| --- | --- | ---: | ---: | ---: | ---: | --- | ---: |
| `architecture-knowledge-separate-lane` | architecture-rationale | true | 100.0% | 1.000 | 3 (60.0%) | yes | 342.4 ms |
| `architecture-vector-supplement` | architecture-rationale | true | 100.0% | 1.000 | 3 (60.0%) | yes | 59.1 ms |
| `command-knowledge-index` | command-behavior | true | 100.0% | 1.000 | 1 (20.0%) | yes | 30.7 ms |
| `command-bm25-only` | command-behavior | true | 100.0% | 1.000 | 3 (60.0%) | yes | 32.7 ms |
| `failure-stale-vector` | failure-fallback | true | 100.0% | 1.000 | 2 (40.0%) | yes | 37.0 ms |
| `failure-source-preserved` | failure-fallback | true | 100.0% | 1.000 | 2 (40.0%) | yes | 27.7 ms |
| `ownership-knowledge-bm25` | ownership-boundary | true | 100.0% | 1.000 | 2 (40.0%) | yes | 36.9 ms |
| `ownership-knowledge-vectors` | ownership-boundary | true | 100.0% | 0.500 | 3 (60.0%) | yes | 39.4 ms |

## Per-case rankings

### `architecture-knowledge-separate-lane`

why does Grimoire index repository rationale separately from production source chunks

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/knowledge.md` | Knowledge retrieval | 25.8881 | true |
| 2 | `docs/reference/vector-store.md` | Vector store | 12.8957 | false |
| 3 | `docs/reference/indexing.md` | Documentation indexing and vectors | 12.7815 | true |
| 4 | `docs/reference/indexing.md` | Prepared source index | 12.4980 | false |
| 5 | `docs/development/testing-and-benchmarks.md` | Prepared source and knowledge-vector smoke test | 11.6173 | false |

### `architecture-vector-supplement`

why are documentation vectors supplemental to deterministic BM25

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `internal/app/README.md` | Knowledge pipeline | 23.1003 | true |
| 2 | `docs/reference/knowledge.md` | Judged documentation evaluation | 21.1654 | false |
| 3 | `docs/development/retrieval-quality.md` | Judged documentation retrieval | 20.5504 | false |
| 4 | `docs/reference/cli.md` | `grimoire knowledge` | 20.3464 | false |
| 5 | `docs/reference/knowledge.md` | Search | 19.8468 | true |

### `command-knowledge-index`

which command builds the independent documentation rationale index

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/cli.md` | `grimoire knowledge` | 17.2415 | true |
| 2 | `internal/app/README.md` | Commands | 12.3879 | false |
| 3 | `README.md` | Quick start | 11.1838 | true |
| 4 | `docs/reference/indexing.md` | Documentation indexing and vectors | 10.6786 | true |
| 5 | `docs/reference/knowledge.md` | Knowledge retrieval | 10.4152 | true |

### `command-bm25-only`

how can knowledge search be forced to use BM25 only

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/knowledge.md` | Search | 19.9617 | true |
| 2 | `docs/reference/indexing.md` | Documentation indexing and vectors | 19.5741 | false |
| 3 | `internal/app/README.md` | Knowledge pipeline | 18.8285 | false |
| 4 | `docs/limits/current-limitations.md` | The Go native loader is Windows-only | 17.2340 | false |
| 5 | `docs/reference/cli.md` | `grimoire knowledge` | 16.4975 | true |

### `failure-stale-vector`

what happens when documentation vectors are stale unavailable or incompatible

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/reference/vector-store.md` | Compatibility and discovery | 17.0252 | true |
| 2 | `internal/app/README.md` | Knowledge pipeline | 16.3268 | true |
| 3 | `docs/reference/indexing.md` | Documentation indexing and vectors | 14.5335 | false |
| 4 | `docs/reference/knowledge.md` | Search | 14.4556 | true |
| 5 | `docs/development/retrieval-quality.md` | Judged documentation retrieval | 13.6398 | false |

### `failure-source-preserved`

does missing documentation vector state fail knowledge retrieval

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/architecture/system-overview.md` | Failure and fallback boundaries | 14.8057 | true |
| 2 | `docs/limits/current-limitations.md` | State maintenance is explicit | 14.0071 | true |
| 3 | `docs/reference/knowledge.md` | Search | 13.7005 | true |
| 4 | `internal/app/README.md` | Commands | 13.4521 | false |
| 5 | `docs/development/testing-and-benchmarks.md` | Judged retrieval evaluation | 13.3450 | false |

### `ownership-knowledge-bm25`

how does internal knowledge rank every query with deterministic BM25 and preserve results when an optional vector ranker fails

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `internal/knowledge/README.md` | Responsibilities | 39.1971 | true |
| 2 | `internal/app/README.md` | Knowledge pipeline | 30.7641 | true |
| 3 | `docs/reference/knowledge.md` | Search | 28.6295 | true |
| 4 | `docs/development/retrieval-quality.md` | Judged documentation retrieval | 28.0798 | false |
| 5 | `internal/knowledgeevaluation/README.md` | Responsibilities | 27.5362 | false |

### `ownership-knowledge-vectors`

which package owns optional documentation semantic ranking and what remains BM25

| Rank | Path | Heading | Score | Relevant |
| ---: | --- | --- | ---: | ---: |
| 1 | `docs/architecture/system-overview.md` | Query-time construction | 20.6216 | false |
| 2 | `docs/architecture/system-overview.md` | Knowledge state | 19.1913 | true |
| 3 | `docs/architecture/system-overview.md` | Retrieval and ranking | 18.5472 | false |
| 4 | `docs/reference/vector-store.md` | Ownership boundary | 18.0974 | true |
| 5 | `README.md` | System flow | 17.9145 | false |

