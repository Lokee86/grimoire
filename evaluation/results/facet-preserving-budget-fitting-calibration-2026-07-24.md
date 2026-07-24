# Facet-preserving budget fitting calibration — 2026-07-24

## Purpose

This cycle targets the final exact-token fitting stage. Ranking, query decomposition, provider fusion, candidate curation, and coverage-aware assembly remain unchanged.

Every external checkout was evaluated from rebuilt `.grimoire/` state. Repositories were not recloned. The second half of each paired comparison reused the exact prepared state created by its fresh baseline.

GDQuest and Trilium were already consumed in earlier cycles. Their results remain regression and generalization evidence rather than statistically untouched test evidence.

## Measured defect

Failure attribution showed that final budget fitting remained the largest source of missing required evidence:

- calibration: 54 of 100 missing requirements;
- validation: 55 of 117 missing requirements; and
- consumed test repositories: 15 of 26 missing requirements.

Assembled-stage candidate diagnostics were added because the earlier evaluator stopped at curation. The new stage showed that several requirements were not missing files: an early chunk from the required file survived, while later chunks carrying additional required symbols were discarded.

Representative examples included `internal/compiler/compiler.go`, `internal/app/query_batch.go`, and `internal/app/vector_manifest.go`. Each file had an included owner chunk and omitted same-file chunks containing additional required symbols.

## Implementation

Coverage-aware adaptive packages now protect one source candidate for each available query facet before spending the remaining source budget. Mechanism, call-chain, and direct-location owners may then protect one same-file companion chunk when that chunk contributes lexical evidence not already represented by the owner.

Companion completion is deliberately bounded:

- one facet owner;
- one companion round;
- same source file;
- same facet claim;
- at least one new BM25, declaration-alias, or leading-line term;
- no mixed-intent owners; and
- no architecture-only owners.

The package schema advances to version 6 and records facet identities, protected claims, available/protected/omitted facet counts, and the configured companion depth. Evaluation reports also retain assembled candidate rank, score, and reasons.

## Rejected variants

The following variants were rejected:

- one protected owner per facet without companions: quality-neutral;
- two or three protected candidates per facet: no recall gain and worse package efficiency;
- two or three distinct files per facet: fd recall regression and higher latency;
- unfiltered same-file completion: Lexicon and fd regressions;
- companions for mixed-intent owners: displaced required Lexicon consumer-state evidence; and
- two companion rounds: no additional recall and worse latency/irrelevance.

## Calibration

Five repositories and 39 cases were evaluated: Grimoire, Lexicon, Gum, HTTPie, and fd.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency | Budget-fit losses |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh legacy fitting | 11.33% | 31.03% | 37.17% | 49.79% | 0.3701 | 74.13% | 1108.3 ms | 54 |
| **Facet owner + one pure companion** | **11.33%** | **31.47%** | 37.17% | 49.79% | 0.3701 | **74.10%** | **1092.6 ms** | **53** |

No calibration repository lost pass rate, required recall, R@10, R@20, or MRR. Grimoire required recall increased from 26.67% to 28.89%; the other four repositories preserved required recall.

The recovered requirement was the `Compile`, `stabilizeTokenCount`, and `Marshal` evidence group in `internal/compiler/compiler.go`. Its owner chunk was already assembled near the front of the package; the new policy preserved the same-file chunk that completed the required symbol group.

## Validation

Space Rocks, RuboCop, and Actual `loot-core` were rebuilt from source with zero reused files.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency | Budget-fit losses |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh legacy fitting | 7.71% | 19.10% | 27.98% | 39.61% | 0.2522 | 81.73% | 1340.8 ms | 55 |
| **Facet owner + one pure companion** | **7.71%** | **19.37%** | 27.98% | 39.61% | 0.2522 | **81.70%** | **1297.0 ms** | **54** |

No validation repository lost a primary quality metric.

- Space Rocks required recall increased from 24.79% to 25.62% and its budget-fitting losses fell from 40 to 39.
- RuboCop preserved all primary quality metrics.
- Actual `loot-core` preserved all primary quality metrics.

## Consumed test repositories, fresh state

GDQuest and Trilium were rebuilt from source with zero reused files.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant | Median latency |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh legacy fitting | 20.00% | 40.19% | 46.74% | 53.76% | 0.4999 | 70.88% | 1357.7 ms |
| Facet owner + one pure companion | 20.00% | 40.19% | 46.74% | 53.76% | 0.4999 | 71.26% | 1406.1 ms |

Quality metrics were neutral. Irrelevant selection increased by 0.38 percentage points and median latency increased by 48.5 ms. Trilium remains at 10.0% required recall; this compiler change does not address its provider and ranking failures.

## Decision

Adopt one protected facet owner and one novel same-file companion for pure mechanism, call-chain, and direct-location evidence groups.

The gain is deliberately bounded: two budget-fitting failures were converted into included evidence across calibration and validation, with no repository-level primary-metric regression. Most remaining budget losses are attached to evidence that is poorly ranked within a facet or spread across files, so the next work should return to mechanism and call-chain ranking rather than increasing compiler reservation depth.
