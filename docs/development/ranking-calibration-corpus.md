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
10. Rerun all frozen corpora and reject material cross-corpus regressions.

## Measured defects and sequencing

The first lexical attribution pass found that common words and substring matches could dominate scores; for example, a query token such as `and` could receive a filename boost against `command.go`. This established the need for stopword suppression and boundary-aware matching as measured corrections rather than subjective tuning.

Provider-attribution runs then showed that vector results were not equivalent to lexical results: vector ranking often improved required-evidence retrieval, while final package recall sometimes remained unchanged. Package fitting and candidate survival therefore require separate calibration from provider ranking.

The July 2026 query-shape reports added a further gate: adaptive assembly must introduce zero required source or structural losses before automatic target tuning is considered successful.

## Report discipline

Checked-in reports establish reproducible baselines and regression evidence. They do not prove that retrieval quality or automatic budgets are solved across arbitrary repositories. Future reports must preserve revision, corpus, provider, mode, state, and date rather than copying aggregate percentages into context-free documentation.
