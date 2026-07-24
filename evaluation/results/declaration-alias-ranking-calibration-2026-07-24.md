# Declaration-alias ranking calibration — 2026-07-24

## Purpose

This cycle targets within-facet lexical ranking without changing query decomposition, candidate reservation, curation, coverage-aware assembly, or token fitting. The comparison baseline is the production `10/18/4` selection configuration with coverage-aware assembly at facet depth `3` and exact-token BM25 ranking.

All external suite checkouts were evaluated from rebuilt `.grimoire/` state. Gum, HTTPie, fd, Space Rocks, RuboCop, Actual `loot-core`, GDQuest, and Trilium were not recloned; only their prepared Grimoire state was removed. Each fresh baseline index reported zero reused files.

GDQuest and Trilium had been inspected in an earlier cycle. Their results here are fresh-state regression and generalization evidence, not statistically untouched test evidence.

## Implementation

The accepted change keeps exact-token BM25 as the primary lexical ranker and adds one bounded repository-derived declaration alias for a query term when that term is absent from the repository's declaration vocabulary.

The declaration vocabulary is built from:

- file names and paths;
- up to six non-comment declaration-header lines per chunk; and
- identifier components split before lowercasing, so `ValidateSnapshot` contributes `validate` and `snapshot`.

An alias candidate must be alphabetic, 5–32 characters long, share at least a four-character prefix, differ in length by no more than five characters, and pass a normalized Levenshtein threshold. At most one alias is selected per absent query term, preferring higher similarity, then lower declaration frequency, then lexical order. A matching declaration receives a score of `1 × similarity`.

The mechanism is deterministic, repository-local, and does not use an LLM, embedding model, global synonym list, stemming library, or fuzzy matching across arbitrary body text.

## Rejected approaches

The calibration cycle rejected broader whole-chunk term coverage, generic prefix morphology, suffix morphology in BM25, term-level reciprocal-rank fusion, and IDF-scaled fixed field bonuses. Those variants either promoted verbose noise, regressed MRR or recall on at least one repository, or produced no ranking improvement.

## Calibration

Five repositories and 39 cases were rebuilt and evaluated: Grimoire, Lexicon, Gum, HTTPie, and fd.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Exact-token baseline | 11.33% | 32.64% | 35.78% | 47.98% | 0.3510 | 74.30% | 1054.2 ms |
| **Declaration alias 1** | **11.33%** | **32.64%** | **35.78%** | **47.98%** | **0.3718** | **74.21%** | **1078.6 ms** |

The candidate increased macro MRR by `0.0209` without changing R@10, R@20, required recall, or pass rate. Ranking-cutoff misses fell from 18 to 17. No calibration repository lost R@10, R@20, MRR, or final required recall.

Per-repository MRR:

| Repository | Baseline | Candidate | Change |
| --- | ---: | ---: | ---: |
| Grimoire | 0.3152 | 0.3152 | unchanged |
| Lexicon | 0.2785 | 0.2798 | +0.0013 |
| Gum | 0.4725 | 0.5725 | +0.1000 |
| HTTPie | 0.5456 | 0.5483 | +0.0028 |
| fd | 0.1429 | 0.1431 | +0.0002 |

## Validation

Space Rocks, RuboCop, and Actual `loot-core` were rebuilt from source: 1,885, 2,107, and 390 files respectively, with zero reused files.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh-state baseline | 7.71% | 17.01% | 24.65% | 39.09% | 0.2450 | 82.39% | 1263.8 ms |
| Declaration alias 1 | 7.71% | 17.01% | 24.65% | 39.09% | 0.2450 | 82.39% | 1267.2 ms |

The candidate was neutral on all validation quality metrics. It neither improved nor regressed any validation repository.

## Reused test repositories, fresh state

GDQuest and Trilium were rebuilt from source: 96 and 3,621 files respectively, with zero reused files.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh-state baseline | 20.00% | 40.19% | 45.31% | 53.76% | 0.4999 | 72.18% | 1301.7 ms |
| **Declaration alias 1** | **20.00%** | **40.19%** | **46.74%** | **53.76%** | **0.4999** | 72.55% | 1350.3 ms |

The candidate improved macro R@10 by `1.43` percentage points. The gain came from GDQuest, whose R@10 increased from 72.76% to 75.62%. Trilium was unchanged. Final recall, R@20, MRR, and pass rate were unchanged; irrelevant selection increased by 0.37 percentage points.

## Prepared-state finding

The earlier reused-state held-out run reported 33.98% required recall. The current fresh-state baseline reports 40.19% on the same pinned GDQuest and Trilium revisions. Because the repository revisions and evaluation cases are fixed, prepared-state freshness is a material benchmark variable. Future claims must record whether `.grimoire/` was rebuilt and must not compare reused and fresh prepared states as paired algorithm measurements.

## Decision

Adopt repository-derived declaration aliases with bonus `1` as the production default.

This is a small first-hit ranking improvement, not a solution to the remaining retrieval problem. Validation required recall remains 17.01%, Trilium required recall remains 10.0%, and neither R@20 nor final package recall improved. The next ranking cycle should target query-facet quality and repository-specific identifier expansion beyond a single near-morphological alias, while preserving the fresh-state benchmark invariant.
