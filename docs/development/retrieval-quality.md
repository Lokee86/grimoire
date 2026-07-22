# Retrieval Quality and Latency Baselines

Grimoire keeps two complementary baselines:

1. deterministic checked-in quality fixtures for candidate merging, exact recovery, curation, and package construction; and
2. repository-scale semantic runs against real prepared and vector snapshots.

## Checked-in quality fixtures

Fixtures live under `internal/app/testdata/retrieval-quality/`. `cases.json` defines:

- the query;
- a deterministic simulated semantic ranking;
- paths that must survive context compilation;
- retrieval sources that must appear; and
- the exact package budget.

`TestRetrievalQualityFixtures` builds a real prepared index from the fixture repository, performs targeted exact recovery, merges provider candidates, applies candidate curation, compiles the exact-budget package, and verifies deterministic serialized output.

The initial corpus covers:

- raw identifier recovery;
- quoted error-code recovery;
- configuration-key recovery;
- version-string recovery;
- adjacent-chunk context;
- provider deduplication; and
- exact-budget package construction.

These fixtures are regression gates, not a claim that four queries represent general retrieval quality. New failure cases should be reduced into small deterministic fixtures before changing ranking or selection rules.

## Warm algorithm benchmarks

```bash
go test ./internal/retrieve ./internal/selection \
  -bench 'Benchmark(Search|Exact|Curate)' \
  -benchmem
```

The benchmarks cover:

- lexical fallback over 10,000 chunks;
- conditional exact recovery over 10,000 chunks; and
- curation of 200 ranked candidates with prepared neighbours.

July 22, 2026 Windows amd64 baseline on an Intel i9-11900H, median of three runs:

| Operation | Median | Allocated bytes | Allocations |
| --- | ---: | ---: | ---: |
| Conditional exact recovery, 10,000 chunks | 1.07 ms | 729,807 | 14 |
| Lexical fallback, 10,000 chunks | 13.06 ms | 8,828,138 | 70,025 |
| Candidate curation, 200 primaries | 2.96 ms | 1,148,649 | 20,753 |

They exclude repository loading, model inference, native vector scanning, and JSON package serialization unless the benchmark explicitly includes them. The curation allocation count is acceptable for the current bounded candidate set but is an obvious optimization target if complete context latency shows it matters.

## Live query-embedding modes

Run the local model service, then execute:

```bash
GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT=http://127.0.0.1:9876/v1 \
  go test ./internal/embedding -run '^$' \
  -bench 'BenchmarkLiveQuery(EmbeddingModes|RequestBatching)' \
  -benchtime=3x -count=3
```

The benchmark gives every 16-token window iteration-specific material so the `llama.cpp` prompt cache cannot make repeated later windows appear artificially cheap. Each value below is the median of three independent three-iteration runs on the same Windows amd64 i9-11900H system.

| Query tokens | `fast`: bounded split requests | `full`: one query | `quality`: full plus windows |
| ---: | ---: | ---: | ---: |
| 16 | 69.0 ms | 70.9 ms | 68.3 ms |
| 32 | 113.2 ms | 207.6 ms | 298.4 ms |
| 64 | 214.0 ms | 497.5 ms | 693.1 ms |
| 128 | 303.6 ms | 1,088.5 ms | 1,350.0 ms |

Fast mode retains the complete query, divides it into 16-token windows, groups at most four windows into each request, and runs at most two requests concurrently. At 32, 64, and 128 tokens it was about 1.8x, 2.3x, and 3.6x faster than one full-query embedding. For queries of 16 tokens or fewer, all modes reduce to one input and observed differences are measurement noise. `quality` intentionally spends more time to preserve both global-query meaning and mechanically isolated concepts.

The request-shape benchmark isolates whether split windows should all be sent in one request or grouped into bounded 64-token requests:

| Query tokens | All windows in one request | Sequential 64-token requests | Bounded 64-token requests, concurrency 2 |
| ---: | ---: | ---: | ---: |
| 128 | 386.1 ms | 317.5 ms | 292.9 ms |
| 256 | 497.1 ms | 507.0 ms | 456.4 ms |
| 512 | 875.9 ms | 917.8 ms | 813.1 ms |

Bounded 64-token requests were approximately 24%, 8%, and 7% faster than one all-window request at 128, 256, and 512 tokens respectively. The complete query is still represented; longer queries produce more bounded requests rather than being truncated. Absolute latency varies with machine load and runtime scheduling, so comparisons should be made within each benchmark row rather than across separate benchmark runs.

## Repository-scale baseline

The current baseline repositories are:

| Repository | Prepared files | Prepared chunks | Baseline embedded | Baseline reused | Result |
| --- | ---: | ---: | ---: | ---: | --- |
| Grimoire | 114 | 282 | 3 | 279 | Measured incremental baseline after the initial 276-chunk build |
| Lexicon | 211 | 730 | 730 | 0 | Initial snapshot published |
| Space Rocks | 1,902 | 8,355 | deferred | — | Prepared large-corpus scaling case; snapshot not published |

The Space Rocks count is retained as the large-corpus scaling case. On the current CPU-only Q8 embedding runtime, fully materializing 8,355 chunks would primarily measure sustained model inference rather than retrieval or vector-store behavior. Vector builds now ingest each completed embedding batch immediately, so an interrupted future build can resume from durable immutable objects instead of restarting completed batches.

## Warm context-command measurements

The Windows development run used the release Rust DLL through an explicit `--engine` path. A packaged build has the equivalent DLL beside `grimoire.exe`. Each query was warmed once and then measured three times with a 2,000-token budget; the table reports the median. No run emitted a semantic-fallback warning. Representative Grimoire and Lexicon queries also completed under the normal two-second timeout with no warning.

| Repository | Query | Median | Sources | Representative result | Assessment |
| --- | --- | ---: | --- | --- | --- |
| Grimoire | Where is vector snapshot freshness validated? | 657 ms | vector | `internal/app/vector_manifest.go` | Direct implementation hit |
| Grimoire | How does context fall back when semantic retrieval fails? | 689 ms | vector | retrieval and architecture documentation | Relevant, but implementation evidence could rank higher |
| Grimoire | Where are exact token budgets enforced? | 637 ms | vector | `docs/reference/context-package.md` and compiler tests | Relevant contract and enforcement evidence |
| Grimoire | `validateVectorSnapshotManifest` | 655 ms | exact, adjacent, vector | `context.go`, `vector.go`, `vector_manifest.go` | Exact recovery and neighbour expansion worked |
| Lexicon | How are repository changes detected and cached analysis reused? | 735 ms | vector | `README.md`, `docs/APPLICATION.md`, `spec/snapshots-v1.md` | Relevant architectural evidence |
| Lexicon | Where is the normalized consumer contract defined? | 647 ms | vector, adjacent | adapter and consumer-runner tests | Quality gap: the defining consumer contract did not survive selection |
| Lexicon | Which Warlock directories are always ignored? | 641 ms | vector | path handling, ignore tests, and `files.go` | Relevant but includes one temporary-file distraction |

The approximately 0.64–0.74 second warm command latency includes query embedding, snapshot validation, exact vector scan, optional exact recovery, candidate curation, complete package serialization, and process startup. At 282 and 730 vectors, the difference between repositories is small; query embedding is the evident dominant cost.

The consumer-contract miss is retained as an observed quality failure rather than reclassified as success. It should become a reduced deterministic fixture before ranking rules change. The ignored-directory query also shows why repository hygiene and permanent temporary-file exclusions matter to retrieval quality.

Do not compare provider scores across exact, vector, lexical, or adjacent candidates. Provider rank, source, reason, selected path, and final package usefulness are the stable interpretation boundaries.

## Change policy

A retrieval or selection change should not be accepted solely because one demonstration query looks better. It should include:

1. a reduced deterministic fixture for the observed failure or intended behavior;
2. benchmark comparison when algorithmic work changes the hot path;
3. repository-scale spot checks; and
4. inspection of the final package, not only the top search result.

## Related documentation

- [Testing and benchmarks](testing-and-benchmarks.md)
- [Context package](../reference/context-package.md)
- [Vector store](../reference/vector-store.md)
- [Roadmap](../planning/roadmap.md)
