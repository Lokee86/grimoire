# Testing and benchmarks

Grimoire uses component-owned unit and integration suites for deterministic contracts and repository-owned judged corpora for context-retrieval behavior.

## Required checks

The repository contains multiple build roots. A root Go or Cargo command does not verify Lexicon or Arcana automatically.

Grimoire Context and native vector engine, from the repository root:

```bash
cargo fmt --manifest-path native/vector-engine/Cargo.toml --all --check
cargo test --manifest-path native/vector-engine/Cargo.toml
cargo clippy --manifest-path native/vector-engine/Cargo.toml --workspace --all-targets -- -D warnings
cargo build --manifest-path native/vector-engine/Cargo.toml -p grimoire-vector-ffi --release
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

Lexicon:

```bash
cd lexicon
python evaluation/run_tests.py
```

Arcana:

```bash
cd arcana
cargo fmt -- --check
cargo check --all-targets
cargo test --all-targets
```

Formatting should produce no diff after the final run.

## Test ownership

Lexicon adapter and snapshot coverage is documented in [`lexicon/docs/DEVELOPMENT.md`](../../lexicon/docs/DEVELOPMENT.md). Arcana graph coverage is documented in [`arcana/README.md`](../../arcana/README.md). The table below covers Grimoire Context.

| Area | Primary coverage |
| --- | --- |
| CLI dispatch and flags | `internal/app/run_test.go`, `model_test.go`, evaluation tests |
| Prepared traversal and state | `internal/index/*_test.go`, app exclusion tests |
| Embedding contract and query batching | `internal/embedding/*_test.go` |
| Runtime backend selection | `internal/embedding/setup_backend_test.go`, app model tests |
| Native object, snapshot, and search behavior | `native/vector-engine/crates/*` tests |
| Go-to-Rust ABI | `internal/vectorstore/integration_windows_test.go` |
| Vector build, reuse, and concurrency | `internal/app/vector*_test.go` |
| Exact, lexical, and merged retrieval | `internal/retrieve/*_test.go`, app context tests |
| Curation and neighbour expansion | `internal/selection/*_test.go` |
| Query profiling and assembly | `internal/queryshape/*_test.go`, `internal/assembly/*_test.go` |
| Structural providers | `internal/lexiconfacts`, `internal/arcanagraph`, and app structure tests |
| Package fitting and exact tokens | `internal/compiler/*_test.go` |
| Corpus scoring and reports | `internal/evaluation/*_test.go` |

Most tests use temporary repositories, local HTTP servers, and synthetic vectors. Native integration requires a built DLL and skips when unavailable.

## Runtime verification

After managed setup:

```bash
grimoire model info
grimoire model serve
```

In another terminal:

```bash
grimoire model probe
grimoire vector info --root .
```

`model info` verifies discovery only. `model probe` sends a real query/document pair to the running endpoint.

## Prepared and vector smoke test

```bash
grimoire index --root .
grimoire vector build --root .
grimoire vector search --root . --query "where is context compilation implemented"
grimoire context --root . --query "explain context compilation"
```

The last command exercises automatic policy. Add a positive `--budget` to exercise fixed fitting.

## Warm algorithm benchmarks

```bash
go test ./internal/retrieve ./internal/selection \
  -bench 'Benchmark(Search|Exact|Curate)' \
  -benchmem
```

These isolate lexical fallback, conditional exact recovery, and bounded candidate curation. They exclude repository loading, embedding inference, native vector scanning, and package serialization unless the benchmark explicitly includes them.

## Live query-embedding benchmarks

With the local model service running:

```bash
GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT=http://127.0.0.1:9876/v1 \
  go test ./internal/embedding -run '^$' \
  -bench 'BenchmarkLiveQuery(EmbeddingModes|RequestBatching)' \
  -benchtime=3x -count=3
```

Compare modes within the same run. Hardware backend, prompt cache, system load, and runtime version materially affect absolute latency.

## Judged retrieval evaluation

Run the repository-owned corpus:

```bash
grimoire eval retrieval \
  --root . \
  --cases evaluation/retrieval/grimoire.json \
  --modes lexical,fast,full,quality
```

Provider comparisons:

```bash
grimoire eval retrieval --root . --cases evaluation/retrieval/grimoire.json \
  --modes lexical --structural-providers none --variant standalone

grimoire eval retrieval --root . --cases evaluation/retrieval/grimoire.json \
  --modes lexical --structural-providers lexicon --variant lexicon

grimoire eval retrieval --root . --cases evaluation/retrieval/grimoire.json \
  --modes lexical --structural-providers lexicon,arcana --variant lexicon-arcana
```

Automatic policy and assembly:

```bash
grimoire eval retrieval \
  --root . \
  --cases evaluation/retrieval/grimoire.json \
  --modes lexical \
  --adaptive \
  --variant adaptive
```

`--adaptive` cannot be combined with a fixed `--budget` override.

## Report outputs

The evaluator writes JSON and Markdown under `evaluation/results/`. Reports include source and structural recall, irrelevant-selection rates, ranking recall and MRR, query-profile agreement, latency, package size, budget utilization, provider warnings, and loss attribution through retrieval, merge, curation, adaptive assembly, and final fitting.

Important report families include ranking calibration baselines/current runs, query-profile reports, fixed/adaptive query-shape comparisons, and standalone/Lexicon/Lexicon-plus-Arcana comparisons.

Do not compare reports from different repository contents, prepared snapshots, corpora, modes, provider sets, or hardware as though they were paired experiments.

## Multi-repository retrieval suite

The suite runner builds Grimoire once, verifies pinned repository revisions, prepares each selected repository, runs adaptive lexical evaluation, and writes per-repository plus macro-averaged reports under the ignored `evaluation/validation/` directory.

Calibration run:

```bash
python evaluation/run_retrieval_suite.py \
  --workspace-root C:/!bin/workspace \
  --grimoire-root . \
  --split calibration \
  --variant frozen-baseline
```

Validation run:

```bash
python evaluation/run_retrieval_suite.py \
  --workspace-root C:/!bin/workspace \
  --grimoire-root . \
  --split validation \
  --variant candidate-name
```

The test split is deliberately sealed. Run it only after the implementation and constants are frozen:

```bash
python evaluation/run_retrieval_suite.py \
  --workspace-root C:/!bin/workspace \
  --grimoire-root . \
  --split test \
  --variant final \
  --allow-test
```

Use `--skip-index` only for the second half of a paired comparison after the first half rebuilt the same pinned checkout state. A normal `grimoire index` may reuse prepared objects, so a benchmark described as fresh must remove the checkout's `.grimoire/` directory before the first run and record the resulting `scanned`, `reused`, and `updated` counts. Reused-state and fresh-state reports are not paired measurements.

The runner refuses revision drift and changes outside the declared calibration seams. Bounded calibration runs may override the frozen selection values with `--selection-file-penalty`, `--selection-subsystem-penalty`, and `--selection-adjacent-primaries`; compare `--assembly-strategy legacy|coverage` and `--assembly-facet-depth`; or vary `--lexical-declaration-alias-bonus`. Every report records the effective ranking, selection, and assembly values plus aggregate required-evidence failure stages.

## Candidate-selection calibration

The evaluator can vary the production curation configuration without changing the normal context CLI:

```bash
grimoire eval retrieval \
  --root . \
  --cases evaluation/retrieval/grimoire.json \
  --modes lexical \
  --adaptive \
  --selection-file-penalty 10 \
  --selection-subsystem-penalty 18 \
  --selection-adjacent-primaries 4 \
  --assembly-strategy coverage \
  --assembly-facet-depth 3 \
  --variant selection-calibrated
```

The current defaults are file penalty 10, subsystem penalty 18, four neighbor anchors, coverage-aware assembly with three distinct candidates reserved per query facet, and repository-derived declaration aliases with bonus `1`. The alias ranker selects at most one high-similarity code-facing identifier for an absent query term and scores only declaration/path evidence; exact-token BM25 remains primary.

Standalone explicit direct-location questions also use bounded facet-specificity ranking. Each query term contributes only its strongest observed evidence signal, with body and declaration evidence weighted above filename and path matches. The correction applies only to full-weight location questions such as `where`, `find`, `locate`, and `which function`; decomposed direct-location sub-facets retain the established ranking behavior.

Selection values were established in the earlier bounded grid; facet depth was selected on the multi-repository calibration split and validated against Space Rocks, RuboCop, and Actual `loot-core`. Declaration aliases improved calibration MRR and fresh-state test R@10 without changing validation quality metrics. Standalone location specificity then improved calibration and validation ranking plus validation package recall without a repository-level primary-metric regression. These values are measured defaults, not a universal optimum. See [the selection comparison](../../evaluation/results/selection-calibration-comparison-2026-07-23.md), [the coverage-aware comparison](../../evaluation/results/coverage-aware-retrieval-calibration-2026-07-24.md), [the declaration-alias ranking comparison](../../evaluation/results/declaration-alias-ranking-calibration-2026-07-24.md), and [the standalone location-specificity comparison](../../evaluation/results/standalone-location-specificity-calibration-2026-07-24.md).

## Calibration discipline

1. Remove `.grimoire/` and rebuild prepared and vector state for a fresh baseline after implementation changes that affect indexed content, chunking, or embedding identity.
2. Record index reuse counts and run compared variants against that same immutable rebuilt state.
3. Preserve corpus, command parameters, and provider set with the report.
4. Inspect per-case failure stages before acting on aggregate recall.
5. Correct invalid expectations instead of treating them as implementation failures.
6. Add a deterministic regression fixture for confirmed defects.
7. Commit only reports that document a meaningful baseline, comparison, or gate.
