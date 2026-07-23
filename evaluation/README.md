# Semantic corpus evaluation

This directory contains the repeatable real-repository acceptance harness for Lexicon adapters. It complements adapter fixtures: fixtures prove specific language behavior, while this corpus checks that the same relations survive realistic repository size, syntax, ambiguity, and dependency layouts.

## Commands

From the Lexicon repository root:

```text
python evaluation/bootstrap_corpus.py
python evaluation/inventory_workspace.py
python evaluation/run_tests.py
python evaluation/run_validation.py --jobs 3
```

`bootstrap_corpus.py` restores the externally pinned corpus repositories beneath the workspace-level `corpus/` directory. It currently prepares the Python doc-ledger snapshot and Lexicanter at the revisions recorded in `corpus.json`.

`run_tests.py` runs the application suite and every adapter suite.

`run_validation.py` builds the required adapters, scans every selected case twice, validates both JSONL files, compares them byte-for-byte, checks positive and negative relation gates, writes bounded audit samples, and generates a combined semantic report.

Useful selectors:

```text
python evaluation/run_validation.py --adapter gdscript
python evaluation/run_validation.py --case gdscript-space-rocks-client --jobs 1
python evaluation/compare_jsonl.py LEFT.jsonl RIGHT.jsonl
```

## Corpus

| Adapter | Calibration | Validation | Holdout |
| --- | --- | --- | --- |
| Python | pinned doc-ledger snapshot | Space Rocks tools | — |
| Ruby | Lexicon Ruby adapter | Space Rocks API | — |
| TypeScript/Svelte | workspace-mcp | pinned Lexicanter snapshot | Space Rocks Astro site |
| GDScript | Alien Attack | Space Rocks client | Speedy Saucer |
| Rust | Grimoire vector engine | Arcana | — |

The Go adapter has its own existing two-repository validation report in `docs/GO_ADAPTER_VALIDATION.md`.

## Gates and artifacts

A case fails when:

- an adapter command or JSONL validation fails;
- two identical scans produce different bytes;
- a required relation has zero emitted edges;
- a relation declared as an expected negative gate is emitted.

Generated scan files, audit samples, summaries, and semantic reports live under `evaluation/validation/generated/` and are ignored by Git. A successful complete run replaces the tracked `evaluation/validation/baseline.json` without timing data, so changes to semantic output remain reviewable.

The harness measures reproducibility and observable relation coverage. It does not claim that every emitted edge is correct or that every unresolved reference is a defect. Precision and recall still require targeted manual audits and language-specific fixtures.
