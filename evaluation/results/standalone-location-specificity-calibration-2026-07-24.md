# Standalone location specificity calibration — 2026-07-24

## Purpose

This cycle targets within-facet ranking for direct source-location questions. It does not change lexical retrieval, declaration aliases, query decomposition, provider fusion, curation, coverage-aware assembly, or token fitting.

Every external checkout was evaluated from rebuilt `.grimoire/` state. The baseline and candidate in each split used the same pinned repository state. GDQuest and Trilium were already consumed in earlier cycles, so their results remain regression and generalization evidence rather than statistically untouched test evidence.

## Measured defect

Direct-location intent ranking re-added large bonuses for every filename, path, and leading-line match after lexical retrieval had already scored those fields. Generic path nouns could therefore outweigh broader implementation evidence.

In `grimoire-dl-01`, a native `snapshot_api.rs` chunk received 52 additional intent-ranking points from repeated field matches. The required `internal/app/vector_manifest.go` implementation first appeared at rank 21 even though it matched more of the substantive query.

## Implementation

Standalone, explicit location questions now receive a bounded facet-specificity score. Each query term contributes only its strongest observed evidence:

| Evidence | Coverage weight |
| --- | ---: |
| Declaration alias | 1.25 |
| BM25 body match | 1.00 |
| Leading-line match | 0.75 |
| Filename match | 0.50 |
| Path match | 0.25 |

Weighted coverage is multiplied by four and capped at 20 points. Implementation declarations receive 18 points and non-declaration implementation chunks receive 8 points.

The new score is deliberately restricted to standalone direct-location intents with full intent weight and an explicit location prefix such as `where`, `find`, `locate`, or `which function`. Direct-location sub-facets produced while decomposing mechanism or mixed queries retain the previous ranking behavior.

This restriction is material. Applying the score to every generated direct-location facet regressed Grimoire mechanism recall and fd MRR. Applying similar coverage scores to call-chain, mechanism, and architecture intents also regressed recall. Repository-derived declaration co-occurrence expansion was tested and rejected because it added latency and cross-repository regressions without a stable gain.

## Calibration

Five repositories and 39 cases were evaluated: Grimoire, Lexicon, Gum, HTTPie, and fd.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh baseline | 11.33% | 32.64% | 35.78% | 47.98% | 0.3718 | 74.07% |
| **Standalone location specificity** | **13.00%** | **33.09%** | **37.45%** | **50.48%** | **0.3759** | **73.77%** |

No calibration repository lost pass rate, required recall, R@10, R@20, or MRR.

Key case changes:

- Grimoire direct-location R@10 increased from 33.3% to 66.7%, R@20 from 33.3% to 66.7%, and MRR from 0.183 to 0.250.
- In `grimoire-dl-01`, the first required source moved from rank 21 to rank 4, and `internal/app/vector_manifest.go` replaced a native API distractor at the front of the selected package.
- Lexicon required recall increased from 20.45% to 22.73%, R@20 from 44.10% to 48.26%, and MRR from 0.2798 to 0.2835.
- Gum, HTTPie, and fd preserved all primary ranking and package metrics.

## Validation

Space Rocks, RuboCop, and Actual `loot-core` were rebuilt from source with zero reused files: 1,885, 2,107, and 390 files respectively.

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh baseline | 7.71% | 17.01% | 24.65% | 39.09% | 0.2450 | 82.39% |
| **Standalone location specificity** | **7.71%** | **19.10%** | **27.98%** | **39.61%** | **0.2522** | **81.74%** |

No validation repository lost a primary quality metric.

- Space Rocks preserved pass rate and final recall while R@20 increased from 32.27% to 33.84% and MRR from 0.3265 to 0.3281.
- RuboCop was unchanged on all quality metrics.
- Actual `loot-core` required recall doubled from 6.25% to 12.50%, R@10 increased from 6.67% to 16.67%, MRR increased from 0.0701 to 0.0901, and irrelevance fell from 77.46% to 75.36%.

## Reused test repositories, fresh state

GDQuest and Trilium were rebuilt from source with zero reused files. The candidate was neutral on every quality metric:

| Configuration | Pass | Required recall | R@10 | R@20 | MRR | Irrelevant |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh baseline | 20.00% | 40.19% | 46.74% | 53.76% | 0.4999 | 72.55% |
| Standalone location specificity | 20.00% | 40.19% | 46.74% | 53.76% | 0.4999 | 72.55% |

Trilium remains the major unresolved generalization failure at 10.0% required recall. The accepted change does not claim to address broad mechanism, architecture, or large-repository retrieval.

## Decision

Adopt bounded facet-specificity ranking for standalone explicit direct-location questions.

The next ranking work should target mechanism and call-chain facets independently. Broadly applying the location formula is explicitly rejected; each intent requires its own evidence model and repository-level calibration.
