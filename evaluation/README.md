# Lexicon semantic corpus evaluation

This directory owns the repeatable real-repository acceptance harness for Lexicon language adapters.

Fixtures prove specific language behavior. The corpus checks that the same semantic relations survive realistic repository size, syntax, ambiguity, dependency layouts, and repeated execution.

## Does not own

The evaluation harness does not define facts-v1 semantics, adapter behavior, or production application behavior. Normative contracts live under [`spec/`](../spec/README.md), and language-specific guarantees live in the owning adapter README.

Corpus counts are dated evidence. They are not permanent performance guarantees and do not establish perfect precision or recall.

## Commands

From the Lexicon repository root:

```text
python evaluation/bootstrap_corpus.py
python evaluation/inventory_workspace.py
python evaluation/run_tests.py
python evaluation/run_validation.py --jobs 3
```

`bootstrap_corpus.py` restores externally pinned repositories beneath the workspace-level `corpus/` directory at revisions recorded in `corpus.json`.

`inventory_workspace.py` reports which configured local corpus repositories and runtimes are currently available.

`run_tests.py` runs the application suite and every adapter suite with the repository's canonical commands.

`run_validation.py` builds required adapters, scans every selected case twice, validates both JSONL outputs, compares them byte-for-byte, checks positive and expected-negative relation gates, writes bounded audit samples, and generates combined reports.

Useful focused forms:

```text
python evaluation/run_validation.py --adapter gdscript
python evaluation/run_validation.py --case gdscript-space-rocks-client --jobs 1
python evaluation/compare_jsonl.py LEFT.jsonl RIGHT.jsonl
```

## Corpus roles

| Adapter | Calibration | Validation | Holdout |
| --- | --- | --- | --- |
| Python | Pinned doc-ledger snapshot | Space Rocks tools | — |
| Ruby | Lexicon Ruby adapter | Space Rocks API | — |
| JavaScript / TypeScript / Svelte | workspace-mcp | Pinned Lexicanter snapshot | Space Rocks Astro site |
| GDScript | Alien Attack | Space Rocks client | Speedy Saucer |
| Rust | Grimoire vector engine | Arcana | — |

The Go adapter has a separate dated two-repository record in [`docs/GO_ADAPTER_VALIDATION.md`](../docs/GO_ADAPTER_VALIDATION.md). Go application and adapter tests remain part of `run_tests.py`.

Calibration cases are used to understand and tune language-general behavior. Validation cases protect known realistic behavior. Holdouts provide a separate check against overfitting to calibration repositories.

## Gates

A case fails when:

- the adapter command fails;
- facts-v1 validation fails;
- two identical scans produce different bytes;
- a required relation has zero emitted edges;
- a relation declared as an expected negative is emitted;
- required output or summary artifacts are missing.

A passing case establishes reproducible observable relation coverage for that repository revision. It does not prove that every edge is correct or that every unresolved record is a defect.

## Tracked and generated artifacts

| Path | Responsibility |
| --- | --- |
| `corpus.json` | Pinned corpus definitions, repository locations, splits, and gates |
| `validation/baseline.json` | Tracked machine-readable results and output identities from the latest complete accepted run |
| `validation/generated/` | Ignored per-run JSONL, summaries, audit samples, and reports |
| `bin/` | Ignored adapter executables built for validation |
| `corpus_state.json` | Ignored local bootstrap state |

A successful complete validation may replace `validation/baseline.json`. Focused runs must not overwrite the accepted full-corpus baseline with partial results.

## Direct files

| File | Responsibility |
| --- | --- |
| `bootstrap_corpus.py` | Restore pinned external corpus repositories |
| `inventory_workspace.py` | Inspect local corpus and runtime availability |
| `run_tests.py` | Execute the application and adapter test matrix |
| `run_validation.py` | Execute corpus scans, gates, determinism checks, summaries, and baseline publication |
| `compare_jsonl.py` | Compare two adapter outputs and report deterministic differences |
| `corpus.json` | Define cases, repository revisions, splits, and expected relations |

## Review expectations

When semantic output changes:

1. run focused adapter fixtures;
2. run the affected corpus cases twice;
3. inspect relation-count changes and audit samples;
4. determine whether the change is a correction, intended expansion, or regression;
5. run the complete corpus before accepting a new baseline;
6. update dated validation documentation when conclusions materially change.

Do not accept a baseline change solely because the harness is green. The diff must be semantically reviewed.

## Placement rules

Put repeatable cross-repository execution, corpus definitions, baselines, and generated evidence here.

Put language-specific fixtures inside the adapter. Put normative format rules in `spec/`. Put stable conclusions and dated validation summaries in `docs/`.
