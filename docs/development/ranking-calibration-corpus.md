# Ranking calibration corpus

`evaluation/retrieval/grimoire.json` is the repository-owned judged corpus used to calibrate ranking, query profiles, structural evidence, and package assembly. External controls are required because self-retrieval alone can reward repository-specific vocabulary and layout.

## Current status

The Grimoire corpus contains direct-location, mechanism-explanation, architecture-ownership, call-chain, and long mixed-query cases. It also carries judged query-profile expectations used by fixed-versus-adaptive evaluation.

The pinned Gum control at commit `716d8b5d0221558f944b5a078dbbcca8572534fb` supplies five manually judged cases covering command dispatch, filtering, timeout/exit ownership, a process call chain, and file-picker behavior. It exposed both weak lexical term normalization and the distinction between initial ranking quality and final package survival.

Additional clean controls remain useful across Go, Python, TypeScript, Rust, Ruby, and GDScript. Candidate repositories must be pinned before expectations are authored; floating repository heads are not valid calibration dependencies.

## Case design

A useful case should:

- represent a task an agent or developer would actually ask;
- identify the minimum source or structural evidence required to answer it;
- distinguish required evidence from merely helpful evidence;
- identify forbidden evidence only when its presence is materially misleading;
- record why each expectation matters; and
- classify expected query shape when the classification is clear.

Do not judge a file as required merely because it contains query words. Required evidence needs a defensible role in the answer. Reject cases whose ownership is too ambiguous to establish reliable ground truth.

## Coverage categories

- Direct location: exact symbol, path, or definition discovery.
- Mechanism explanation: a bounded implementation flow.
- Architecture ownership: responsibility across packages or subsystems.
- Call-chain investigation: ordered operational relationships.
- Long mixed query: multiple constraints and evidence types in one prompt.

New cases should cover a distinct failure mode, repository shape, language, or task category rather than repeat an existing lexical pattern.

## Query-profile expectations

Profile expectations may constrain intent, specificity, breadth, ambiguity, cross-system scope, evidence needs, and selected scope. They should be semantic and stable under harmless ranking changes.

A mismatch is not automatically a classifier bug. First check whether the prompt is genuinely ambiguous or the expectation overstates its scope.

## Frozen multi-repository suite

`evaluation/retrieval/suite.json` freezes the comparison baseline at commit `5eb9e8eacc063b4620a463a59aac2fddfaaaa4d3`, the production selection configuration `10/18/4`, the coverage-aware assembly configuration with facet depth `3`, and three repository-level splits:

| Split | Repositories | Permitted use |
| --- | --- | --- |
| Calibration | Grimoire, Lexicon, Gum, HTTPie, fd | Diagnose failures and tune deterministic behavior or bounded constants. |
| Validation | Space Rocks, RuboCop, Actual `loot-core` | Choose between candidates and reject cross-repository regressions. Do not tune directly to individual cases. |
| Test | GDQuest 2D Space Game, Trilium | One final run after implementation and constants are frozen. |

The split is repository-level so closely related cases cannot leak between calibration and evaluation. Aggregate reports must include both per-repository results and an unweighted macro-average across repositories; a large corpus must not dominate merely because it contains more cases.

`evaluation/run_retrieval_suite.py` verifies pinned checkout revisions and rejects a run when retrieval implementation paths differ from the recorded baseline outside the declared calibration seam. Selection and assembly values are recorded in the manifest and every generated report. The test split is sealed unless the caller explicitly passes `--allow-test`.

## July 24, 2026 calibration result

The suite contains 91 cases across ten repositories: 39 calibration cases, 42 validation cases, and 10 held-out test cases. The accepted production configuration is `10/18/4`, increasing adjacent-primary coverage from three to four while retaining the existing file and subsystem penalties.

Across the calibration split, required recall increased from 21.85% to 22.77% and irrelevant selection fell from 81.54% to 81.26%, with pass rate unchanged. Validation required recall increased from 11.19% to 11.47%, irrelevant selection fell slightly, and median latency decreased by 55.8 ms. On the once-consumed held-out test split, required recall increased from 32.13% to 33.98%, irrelevant selection fell from 80.06% to 78.38%, and median latency decreased by 42.8 ms.

The change improves final package assembly rather than initial ranking: R@10, R@20, MRR, and held-out pass rate were unchanged. The test split produced no fully passing cases, and Trilium required recall remained 5.0%; the result is a bounded curation improvement, not evidence that retrieval quality is solved. No further tuning was performed against the held-out cases.

## Coverage-aware assembly calibration

The next cycle used only the calibration and validation repositories. GDQuest and Trilium were already consumed and were not rerun. Aggregate failure-stage reporting first showed that 59.3% of missing calibration requirements were present before final fitting but absent from the package. It also exposed stale Lexicon expectations left over from the pre-consolidation library design; those expectations were corrected against the pinned baseline before candidate comparison.

The accepted implementation assigns stable facet identities to decomposed query intents, retains per-facet candidate ranks across exact, lexical, vector, and structural fusion, and reserves three distinct candidates per facet during adaptive assembly. One candidate can claim only one open facet for coverage purposes, preventing a generic multi-match chunk from satisfying the whole query.

Against the repaired legacy-assembly baseline, calibration required recall increased from 23.58% to 32.64%, irrelevant selection fell from 80.92% to 74.22%, and median latency fell from 1072.2 ms to 1020.7 ms. Validation required recall increased from 13.69% to 17.01%, irrelevant selection fell from 86.39% to 82.39%, median latency fell from 1207.6 ms to 1173.9 ms, and macro pass rate increased from 1.04% to 7.71%. No validation repository lost required recall.

Depth four was rejected despite a higher calibration macro average because HTTPie required recall regressed from 23.53% to 17.65%. Increasing the ranking reservation from two to three candidates per facet was also rejected because it regressed R@10 or MRR on Lexicon, Gum, and fd.

## Declaration-alias ranking calibration

The next cycle rebuilt every external suite checkout from source by removing only its `.grimoire/` prepared state; no repository was recloned. The accepted deterministic ranker derives at most one high-similarity identifier alias for a query term that is absent from the repository's declaration vocabulary. Vocabulary comes from paths and non-comment declaration headers, with CamelCase splitting performed before lowercasing. Production alias bonus is `1`.

Calibration macro MRR increased from `0.3510` to `0.3718` without changing R@10, R@20, required recall, or pass rate. Validation quality metrics were unchanged. On fresh-state GDQuest and Trilium, macro R@10 increased from 45.31% to 46.74%; all other primary ranking and package metrics were unchanged, while irrelevant selection increased by 0.37 percentage points.

The fresh-state test baseline reported 40.19% required recall versus 33.98% in the earlier reused-state run. Prepared-state freshness is therefore a benchmark invariant: reports must record whether `.grimoire/` was rebuilt, and reused-state results must not be treated as paired comparisons against fresh-state results. See [the declaration-alias ranking report](../../evaluation/results/declaration-alias-ranking-calibration-2026-07-24.md).

## Standalone location-specificity calibration

The next cycle isolated a duplicated intent-ranking bias in explicit source-location questions. Direct-location ranking was re-adding large bonuses for every filename, path, and leading-line match after lexical retrieval had already scored those fields. Generic path nouns could therefore outrank broader implementation evidence.

The accepted correction applies only to standalone, full-weight direct-location questions beginning with an explicit location form such as `where`, `find`, `locate`, or `which function`. Each query term contributes only its strongest evidence signal, with body and declaration evidence weighted above filename and path matches. Generated direct-location sub-facets inside mechanism and mixed queries retain their previous ranking behavior.

Calibration pass rate increased from 11.33% to 13.00%, required recall from 32.64% to 33.09%, R@10 from 35.78% to 37.45%, R@20 from 47.98% to 50.48%, and MRR from 0.3718 to 0.3759. Validation required recall increased from 17.01% to 19.10%, R@10 from 24.65% to 27.98%, R@20 from 39.09% to 39.61%, and MRR from 0.2450 to 0.2522. No calibration or validation repository lost pass rate, required recall, R@10, R@20, or MRR. Fresh-state GDQuest and Trilium results were neutral on every quality metric.

Broader use of the same formula for call-chain, mechanism, architecture, composite location facets, or every generated direct-location facet was rejected because it regressed recall or ranking. Repository-derived declaration co-occurrence expansion was also rejected because it added latency and cross-repository regressions. See [the standalone location-specificity report](../../evaluation/results/standalone-location-specificity-calibration-2026-07-24.md).

## Facet-preserving budget fitting

The next cycle targeted final exact-token fitting after attribution showed that budget fitting remained the largest source of missing required evidence. Assembled-stage diagnostics demonstrated that several failures already had an early chunk from the required file in the package while later same-file chunks containing additional required symbols were omitted.

Coverage-aware adaptive packages now protect one source candidate per available facet. Pure mechanism, call-chain, and direct-location owners may also protect one same-file companion when it contributes a BM25, declaration-alias, or leading-line term not already represented by the owner. Architecture and mixed-intent owners are excluded because their companions displaced stronger evidence during calibration.

Calibration required recall increased from 31.03% to 31.47% and budget-fitting losses fell from 54 to 53. Validation required recall increased from 19.10% to 19.37%, Space Rocks increased from 24.79% to 25.62%, and validation budget-fitting losses fell from 55 to 54. No calibration or validation repository lost a primary quality metric. Fresh-state GDQuest and Trilium quality metrics were unchanged. See [the facet-preserving fitting report](../../evaluation/results/facet-preserving-budget-fitting-calibration-2026-07-24.md).

## Calibration workflow

1. Pin and record the repository revision and evaluated scope.
2. Write expectations before changing retrieval behavior.
3. Validate paths and symbols against that revision.
4. Prepare one immutable source and vector state.
5. Run the baseline and candidate variant against the same state.
6. Compare ranking, final recall, package composition, and failure stages.
7. Inspect every regression at the case level.
8. Reduce confirmed defects into deterministic fixtures.
9. Change one responsible seam or weight family at a time.
10. Rerun the calibration split and reject material cross-repository regressions.
11. Run the validation split only after a candidate survives calibration.
12. Keep the test split sealed until the implementation and constants are frozen.

## Measured defects and sequencing

The first lexical attribution pass found that common words and substring matches could dominate scores; for example, a query token such as `and` could receive a filename boost against `command.go`. This established the need for stopword suppression and boundary-aware matching as measured corrections rather than subjective tuning.

Provider-attribution runs then showed that vector results were not equivalent to lexical results: vector ranking often improved required-evidence retrieval, while final package recall sometimes remained unchanged. Package fitting and candidate survival therefore require separate calibration from provider ranking.

The July 2026 query-shape reports added a further gate: adaptive assembly must introduce zero required source or structural losses before automatic target tuning is considered successful.

## Report discipline

Checked-in reports establish reproducible baselines and regression evidence. They do not prove that retrieval quality or automatic budgets are solved across arbitrary repositories. Future reports must preserve revision, corpus, provider, mode, state, and date rather than copying aggregate percentages into context-free documentation.
