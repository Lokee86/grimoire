# Coverage-aware retrieval calibration — 2026-07-24

## Purpose

This cycle addresses incomplete evidence coverage in multi-part queries. It compares the previous rank-preserving adaptive assembler with facet-aware retrieval and package planning while keeping repository revisions, prepared state, lexical mode, selection values, and token budgets fixed.

The comparison baseline is commit `5eb9e8eacc063b4620a463a59aac2fddfaaaa4d3` with selection configuration `10/18/4`. The already-consumed GDQuest and Trilium test repositories were not used in this cycle.

## Corpus repair and attribution

Aggregate required-evidence failure stages were added to the evaluator and suite summaries. The first attributed run exposed stale Lexicon expectations inherited from its pre-consolidation library design. Those expectations were corrected against the pinned baseline before any candidate was accepted.

The clean calibration baseline contained 108 missing required items:

| Failure stage | Missing items | Share |
| --- | ---: | ---: |
| Budget-fitting loss | 64 | 59.3% |
| Ranking-cutoff miss | 18 | 16.7% |
| Adaptive assembly loss | 12 | 11.1% |
| Provider retrieval miss | 11 | 10.2% |
| Exact recovery miss | 3 | 2.8% |

The dominant measured defect was therefore evidence that survived retrieval and curation but was displaced during final package fitting.

## Implementation

The accepted implementation adds the following deterministic seams:

- decomposed retrieval intents receive stable facet identities;
- exact, lexical, vector, and structural candidates retain facet membership;
- candidates retain their rank within each facet across provider fusion;
- semantic query plans are embedded independently per facet while sharing one validated vector snapshot and a bounded concurrency cap;
- adaptive assembly reserves distinct candidates for uncovered facets before repeated evidence;
- a candidate matching several facets claims only its strongest still-open facet for coverage purposes; and
- assembly decisions record available and represented facets and the configured coverage depth.

Production coverage depth is `3`: up to three distinct candidates are reserved per facet before remaining ranked evidence is considered.

## Calibration

Five repositories and 39 cases were used for calibration: Grimoire, Lexicon, Gum, HTTPie, and fd.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Legacy assembly | 11.33% | 23.58% | 35.78% | 47.98% | 0.351 | 80.92% | 1072.2 ms |
| Coverage depth 1 | 11.33% | 25.45% | 35.78% | 47.98% | 0.351 | 80.38% | 1018.0 ms |
| Coverage depth 2 | 11.33% | 30.77% | 35.78% | 47.98% | 0.351 | 76.93% | 1034.6 ms |
| **Coverage depth 3** | **11.33%** | **32.64%** | **35.78%** | **47.98%** | **0.351** | **74.22%** | **1020.7 ms** |
| Coverage depth 4 | 11.33% | 34.37% | 35.78% | 47.98% | 0.351 | 71.92% | 1042.0 ms |

Depth three increased required recall by 9.06 percentage points and reduced irrelevant selection by 6.70 percentage points. Required failures fell from 108 to 98, and budget-fitting losses fell from 64 to 54.

Per-repository required recall:

| Repository | Legacy | Depth 3 | Change |
| --- | ---: | ---: | ---: |
| Grimoire | 22.22% | 31.11% | +8.89 pp |
| Lexicon | 18.18% | 20.45% | +2.27 pp |
| Gum | 42.86% | 71.43% | +28.57 pp |
| HTTPie | 23.53% | 23.53% | unchanged |
| fd | 11.11% | 16.67% | +5.56 pp |

Depth four was rejected despite its higher macro average because HTTPie required recall regressed from 23.53% to 17.65%. The selection rule remains repository-regression aware rather than choosing the best aggregate blindly.

## Validation

Three repositories and 42 cases were used for validation: Space Rocks, RuboCop, and Actual `loot-core`.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Legacy assembly | 1.04% | 13.69% | 24.65% | 39.09% | 0.245 | 86.39% | 1207.6 ms |
| **Coverage depth 3** | **7.71%** | **17.01%** | **24.65%** | **39.09%** | **0.245** | **82.39%** | **1173.9 ms** |

Per-repository required recall:

| Repository | Legacy | Depth 3 | Change |
| --- | ---: | ---: | ---: |
| Space Rocks | 21.49% | 24.79% | +3.31 pp |
| RuboCop | 13.33% | 20.00% | +6.67 pp |
| Actual `loot-core` | 6.25% | 6.25% | unchanged |

RuboCop also gained one complete case, raising its pass rate from 0% to 20%. No validation repository lost required recall. Validation required failures fell from 123 to 118, and budget-fitting losses fell from 61 to 56.

## Rejected ranking change

Increasing per-facet reservation inside the ranked candidate merger from two candidates to three produced only a small Grimoire R@20 gain and regressed R@10 or MRR on Lexicon, Gum, and fd. It was reverted.

This confirms that assembly coverage and ranking quality are separate seams. The accepted change materially improves final-package survival, but it does not improve R@10, R@20, or MRR. The next retrieval cycle should target within-facet ranking and query-facet quality rather than widening unconditional reservation.

## Decision

Adopt coverage-aware adaptive assembly with facet depth `3` as the production default.

This is a material package-assembly improvement, not a claim that retrieval quality is solved. Validation required recall remains 17.01%, Actual remains at 6.25%, and ranking metrics did not improve. The existing test repositories remain consumed; a future final generalization claim requires a new untouched test split.
