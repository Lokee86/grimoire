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

## Judged repository-scale evaluation

Repository-owned corpora live under `evaluation/retrieval/`. They are separate from deterministic unit fixtures and contain explicit required, supporting, and optional forbidden evidence for real development questions.

The initial judged set contains:

| Repository | Cases | Purpose |
| --- | ---: | --- |
| Grimoire | 12 | Retrieval implementation and package-construction behavior |
| Lexicon | 12 | Language analysis, snapshots, consumers, and repository ownership |
| Space Rocks | 32 | Large heterogeneous Go, GDScript, Ruby, configuration, and cross-language work |

Each corpus covers direct locations, mechanism explanations, architecture ownership, call-chain investigations, and long mixed implementation requests. Space Rocks includes multi-file and cross-language expectations.

Run all four modes with:

```bash
grimoire eval retrieval \
  --cases evaluation/retrieval/grimoire.json \
  --root .
```

The runner writes JSON and Markdown under `evaluation/results/` and reports complete-case pass rate, separate source and structural recall, separate irrelevant-evidence rates, median and p95 latency, provider warnings, and the stage where required evidence was lost. It also records the rank where each required evidence item first becomes complete and summarizes pre-curation required recall at ranks 10 and 20, mean reciprocal rank, and judged-path relevance at ranks 10 and 20. These ranking metrics isolate retrieval order from later exact merging, curation, and token-budget effects.

Run the three comparable variants against the same prepared and vector snapshot:

```bash
# Source retrieval only
grimoire eval retrieval \
  --cases evaluation/retrieval/grimoire.json \
  --root . \
  --structural-providers none \
  --variant standalone

# Source retrieval plus Lexicon symbols and relationships
grimoire eval retrieval \
  --cases evaluation/retrieval/grimoire.json \
  --root . \
  --structural-providers lexicon \
  --variant lexicon

# Source retrieval plus Lexicon and Arcana graph evidence
grimoire eval retrieval \
  --cases evaluation/retrieval/grimoire.json \
  --root . \
  --structural-providers lexicon,arcana \
  --variant lexicon-arcana
```

A case may declare `required_structural`, `supporting_structural`, and `forbidden_structural`. Each expectation names a provider and evidence kind, then optionally constrains the subject symbol/path, relationship and target, ordered call-chain symbols, or unresolved expression. Structural failure attribution distinguishes a provider miss from loss during cross-provider composition and loss during final package budgeting.

Lexicon and Arcana execute through the same production path used by `grimoire context`. Lexicon exports the immutable current snapshot into Grimoire's cache. Arcana is synchronized to that same snapshot and queried from the Lexicon-matched seeds. Standalone and assisted reports must use the same source, prepared, vector, Lexicon, and Arcana state where applicable.

## Failure reduction policy

Do not tune retrieval while constructing a judged corpus or recording its initial baseline. For each confirmed failure:

1. identify the responsible stage from the report;
2. reduce the behavior to the smallest deterministic repository fixture;
3. confirm the fixture fails before changing retrieval;
4. modify only the responsible stage;
5. confirm the fixture passes; and
6. rerun the full judged corpus to detect category, relevance, or latency regressions.

The first confirmed live regression was curation failing to promote an immediately adjacent chunk when that chunk already existed later in the retrieved list. The deterministic fixture now verifies that the retrieved neighbor moves forward without losing provider rank or provenance.

Do not compare provider scores across exact, vector, lexical, Lexicon, or adjacent candidates. Provider rank, source, reason, selected path, and final package usefulness are the stable interpretation boundaries.

## Default-mode decision

Do not choose the default from latency alone. `fast` remains acceptable only when its required-evidence recall is at least 90% of `quality`, failed cases do not materially increase, median latency is materially lower, long-query p95 remains acceptable, later query windows are not systematically lost, and no category consistently succeeds under `full` while failing under `fast`.

The dated JSON and Markdown files under `evaluation/results/` are the authoritative measured baselines. Historical spot-query tables are not substitutes for the judged corpus.

## Related documentation

- [Testing and benchmarks](testing-and-benchmarks.md)
- [Ranking calibration corpus](ranking-calibration-corpus.md)
- [Context package](../reference/context-package.md)
- [Vector store](../reference/vector-store.md)
- [CLI reference](../reference/cli.md)
- [Roadmap](../planning/roadmap.md)
