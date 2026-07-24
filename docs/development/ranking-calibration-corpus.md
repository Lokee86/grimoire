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

`evaluation/retrieval/suite.json` freezes the current retrieval implementation at commit `c7cb6ee321ef9ccc630a054bb4315872137bf3d8`, the production selection configuration `10/18/3`, and three repository-level splits:

| Split | Repositories | Permitted use |
| --- | --- | --- |
| Calibration | Grimoire, Lexicon, Gum, HTTPie, fd | Diagnose failures and tune deterministic behavior or bounded constants. |
| Validation | Space Rocks, RuboCop, Actual `loot-core` | Choose between candidates and reject cross-repository regressions. Do not tune directly to individual cases. |
| Test | GDQuest 2D Space Game, Trilium | One final run after implementation and constants are frozen. |

The split is repository-level so closely related cases cannot leak between calibration and evaluation. Aggregate reports must include both per-repository results and an unweighted macro-average across repositories; a large corpus must not dominate merely because it contains more cases.

`evaluation/run_retrieval_suite.py` verifies pinned checkout revisions and rejects a run when retrieval implementation paths differ from the recorded baseline. The manifest permits only the separately measured production-default change in `internal/selection`; all evaluation commands still pass explicit selection values. The test split is sealed unless the caller explicitly passes `--allow-test`.

## July 24, 2026 calibration result

The suite contains 91 cases across ten repositories: 39 calibration cases, 42 validation cases, and 10 held-out test cases. The accepted production configuration is `10/18/4`, increasing adjacent-primary coverage from three to four while retaining the existing file and subsystem penalties.

Across the calibration split, required recall increased from 21.85% to 22.77% and irrelevant selection fell from 81.54% to 81.26%, with pass rate unchanged. Validation required recall increased from 11.19% to 11.47%, irrelevant selection fell slightly, and median latency decreased by 55.8 ms. On the once-consumed held-out test split, required recall increased from 32.13% to 33.98%, irrelevant selection fell from 80.06% to 78.38%, and median latency decreased by 42.8 ms.

The change improves final package assembly rather than initial ranking: R@10, R@20, MRR, and held-out pass rate were unchanged. The test split produced no fully passing cases, and Trilium required recall remained 5.0%; the result is a bounded curation improvement, not evidence that retrieval quality is solved. No further tuning was performed against the held-out cases.

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
