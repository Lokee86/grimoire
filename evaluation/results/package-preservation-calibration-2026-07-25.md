# Package-preservation calibration

Date: 2026-07-25

## Objective

Calibrate the final source-package funnel rather than adding more retrieval or span-extraction machinery. The historical failure was not a WriteFreely metric: it came from the Grimoire judged corpus, where lexical retrieval found 34 of 45 required items, curation retained 33, and final fitting included only 1.

The integrated pre-calibration pipeline had already changed that corpus and produced a smaller but still material fitting loss:

| Stage | Required items |
| --- | ---: |
| Retrieved and curated | 25 |
| Assembled | 22 |
| Included | 14 |

## Calibrated policy

The retained changes are deliberately bounded:

1. Prefer primary implementation owners over supporting tests, documentation, or configuration when protecting a query facet.
2. For single-facet or exploratory mechanism requests, permit a second distinct implementation file for the same facet. Call-chain requests retain one protected file because held-out validation showed that extra call-chain breadth displaced correct route and connection evidence. Bounded multi-facet requests retain one protected file per facet.
3. Keep lexical plans in their calibrated ranked order. For semantic-only plans, preserve file ranking but prefer declaration-bearing chunks within the selected file, with one declaration companion fallback when semantic evidence has no lexical details.
4. Expand focused call-chain lifecycle language deterministically: creation adds `new`, `create`, and `add`; persistence adds `persist`, `store`, `save`, `insert`, `write`, and `database`.
5. Keep language-aware span extraction opt-in. It added latency without quality gain in the earlier calibration.

## Five-repository calibration

The paired comparison used the same lexical rankings and changed only final protected-file depth.

| Configuration | Macro pass | Macro required recall | R@10 | R@20 | MRR | Irrelevant | Fitting losses |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Protected file depth 1 | 13.0% | 31.9% | 37.5% | 50.1% | 0.375 | 74.5% | 51 |
| Mechanism-only protected file depth 2 | 14.7% | 34.0% | 37.5% | 50.1% | 0.375 | 73.9% | 48 |

Mechanism-only depth 2 improves package recall without changing retrieval rankings. It is the default. A broader mechanism-and-call-chain depth-2 policy reached 34.4% calibration recall but reduced held-out recall, so it was rejected.

On the Grimoire corpus alone, the final funnel becomes:

| Stage | Required items |
| --- | ---: |
| Retrieved and curated | 25 |
| Assembled | 22 |
| Included | 16 |

That is two additional required items included under the same retrieval result and token-budget policy.

## WriteFreely full-stack fixture

Fixture: WriteFreely `8f942b2aed5951aba717268f0f7a597bd487e8e5` with Grimoire, Lexicon, and Arcana.

The initial rerun exposed the exact failure:

| Stage | Required items |
| --- | ---: |
| Retrieved | 5/5 |
| Curated | 5/5 |
| Assembled | 4/5 |
| Included | 0/5 |

After role-aware fitting, semantic declaration ownership, and lifecycle vocabulary expansion:

| Metric | Final |
| --- | ---: |
| Retrieved | 5/5 |
| Curated | 5/5 |
| Assembled | 5/5 |
| Included | 3/5 |
| Required recall | 60.0% |
| Required R@10 | 40.0% |
| Required R@20 | 60.0% |
| Irrelevant package selections | 62.5% |

The two remaining misses are budget-fitting losses. They are no longer retrieval, curation, or assembly failures.

## Hermes verification

Three independent Hermes trials used the exact final 12,000-token Grimoire + Lexicon + Arcana package and the existing WriteFreely rubric.

| Trial | Score | Discovery calls | Time |
| --- | ---: | ---: | ---: |
| 1 | 10/10 | 45 | 271.7 s |
| 2 | 10/10 | 52 | 362.8 s |
| 3 | 10/10 | 52 | 337.0 s |
| Average | 10/10 | 49.7 | 323.8 s |

All trials identified the route entry points, handler/authentication and validation flow, `CreateOwnedPost` versus `CreatePost`, exact SQL boundary and retry behavior, response construction, and post-persistence federation/email work with usable source citations.

The current Hermes model was `gpt-5.6-luna`, so elapsed time is not directly comparable to the earlier Hermes campaign. Correctness remained 3/3 and 10/10 on the exact retained mechanism-only policy.

## Held-out validation

The paired validation used Space Rocks, RuboCop, and Actual loot-core. The frozen Space Rocks revision was unavailable locally, so the suite used the current indexed Space Rocks revision `2e056203acc6f0da4ef7749760f19f2126a425e7`; RuboCop and Actual remained pinned.

| Configuration | Macro pass | Macro required recall | R@10 | R@20 | MRR | Irrelevant | Fitting losses |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Protected file depth 1 | 14.4% | 21.3% | 28.2% | 39.6% | 0.253 | 80.9% | 52 |
| Mechanism-only protected file depth 2 | 14.4% | 21.6% | 28.2% | 39.6% | 0.253 | 80.0% | 51 |

The broader mechanism-and-call-chain depth-2 policy was rejected after it reduced held-out macro recall to 19.5% by displacing a required Actual loot-core server-connection file. Restricting the second file to mechanism evidence restores that requirement and retains the calibration gain.

## Conclusion

The span-extraction work was not the improvement. The useful calibration was final-package ownership plus a small query-language correction. The selected policy measurably reduces fitting loss, preserves lexical ranking behavior, materially repairs the WriteFreely package, and retains perfect agent correctness on the existing rubric.
