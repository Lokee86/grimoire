# Multi-repository selection calibration — 2026-07-24

## Purpose

This calibration tests whether Grimoire's production selection constants generalize beyond the original Grimoire corpus. It freezes retrieval implementation commit `c7cb6ee321ef9ccc630a054bb4315872137bf3d8`, evaluates repositories at pinned revisions, and keeps the final test repositories outside the tuning loop.

## Corpus

| Split | Repositories | Cases |
| --- | --- | ---: |
| Calibration | Grimoire, Lexicon, Gum, HTTPie, fd | 39 |
| Validation | Space Rocks, RuboCop, Actual `packages/loot-core` | 42 |
| Held-out test | GDQuest 2D Space Game, Trilium | 10 |
| **Total** | **10 repositories** | **91** |

Each newly added external corpus contains one direct-location, mechanism-explanation, architecture-ownership, call-chain-investigation, and long-mixed-query case. Aggregate metrics are unweighted macro-averages across repositories so larger corpora do not dominate the selection decision.

## Frozen baseline

The previous production configuration was:

- same-file repeat penalty: `10`
- same-subsystem repeat penalty: `18`
- adjacent-primary limit: `3`

On the five-repository calibration split it produced:

- pass rate: 11.33%
- required evidence recall: 21.85%
- R@10: 36.71%
- R@20: 47.78%
- MRR: 0.346
- irrelevant selection: 81.54%
- median latency: 1076.6 ms

The large difference between repositories confirms that the earlier Grimoire-only result did not generalize uniformly. Gum reached 42.9% final recall, while Lexicon reached 9.5% and fd reached 11.1%.

## Bounded calibration grid

Only one selection family was changed at a time. Ranking metrics remained constant because these controls alter final curation and neighbor promotion after candidate ranking.

| Candidate | Pass | Required recall | Irrelevant | Median latency | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `10/18/3` baseline | 11.33% | 21.85% | 81.54% | 1076.6 ms | Baseline |
| subsystem `8` | 9.67% | 22.86% | 81.42% | 1123.6 ms | Reject: pass regression |
| subsystem `12` | 11.33% | 22.83% | 81.70% | 1140.7 ms | Reject: worse noise and latency |
| subsystem `24` | 11.33% | 21.85% | 83.66% | 1196.2 ms | Reject |
| adjacent primaries `2` | 9.67% | 21.18% | 82.18% | 1139.5 ms | Reject |
| adjacent primaries `4` | 11.33% | 22.77% | 81.26% | 1123.0 ms | Advance |

The `10/18/4` candidate preserved calibration pass rate, increased required recall by 0.92 percentage points, and reduced irrelevant selection by 0.28 percentage points.

## Validation

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `10/18/3` | 1.04% | 11.19% | 24.65% | 39.09% | 0.245 | 86.51% | 1280.6 ms |
| `10/18/4` | 1.04% | 11.47% | 24.65% | 39.09% | 0.245 | 86.48% | 1224.8 ms |

The candidate preserved pass and ranking metrics, increased final recall by 0.28 percentage points, slightly reduced irrelevant selection, and lowered median latency by 55.8 ms. It therefore advanced to the sealed test split without further changes.

## Held-out test

The held-out repositories were evaluated only after the candidate was frozen. No constants or algorithms were changed from their results.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `10/18/3` | 0.00% | 32.13% | 45.31% | 53.76% | 0.500 | 80.06% | 1208.9 ms |
| `10/18/4` | 0.00% | 33.98% | 45.31% | 53.76% | 0.500 | 78.38% | 1166.1 ms |

Per repository:

| Repository | Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| GDQuest 2D Space Game | `10/18/3` | 0.0% | 59.3% | 72.8% | 79.0% | 0.767 | 63.6% | 571.3 ms |
| GDQuest 2D Space Game | `10/18/4` | 0.0% | 63.0% | 72.8% | 79.0% | 0.767 | 61.8% | 560.8 ms |
| Trilium | `10/18/3` | 0.0% | 5.0% | 17.9% | 28.6% | 0.233 | 96.5% | 1846.5 ms |
| Trilium | `10/18/4` | 0.0% | 5.0% | 17.9% | 28.6% | 0.233 | 94.9% | 1771.4 ms |

The candidate improved held-out macro recall by 1.85 percentage points, reduced irrelevant selection by 1.68 percentage points, and reduced median latency by 42.8 ms. The gain came from better final-package survival around one additional primary candidate; initial ranking did not change.

## Decision

Adopt `10/18/4` as the production selection configuration.

This is a small, consistently positive curation change across calibration, validation, and held-out test repositories. It is not a retrieval-quality milestone:

- no held-out case contained every required item;
- Trilium final recall remained 5.0%;
- held-out irrelevance remained 78.38%;
- R@10, R@20, and MRR did not improve; and
- repository-to-repository variance remains extreme.

The next calibration cycle should target ranking and evidence-to-package survival defects across the calibration and validation repositories. The consumed GDQuest and Trilium cases must not be used to choose further changes; a new untouched test split is required for another final generalization claim.

## Reproducibility

The repository split, revisions, baseline constants, and test gate are recorded in `evaluation/retrieval/suite.json`. `evaluation/run_retrieval_suite.py` verifies the selected checkout revisions, writes per-repository and macro summaries, and refuses the test split unless `--allow-test` is supplied.
