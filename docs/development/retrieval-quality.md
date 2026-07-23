# Retrieval quality

Grimoire evaluates the entire context-construction pipeline rather than treating a search score as proof of useful context.

## Deterministic fixtures

Small checked-in fixtures under `internal/app/testdata/retrieval-quality/` exercise real prepared indexing, targeted exact recovery, provider merging, curation, package compilation, and deterministic serialization. They cover identifiers, quoted error codes, configuration keys, version strings, adjacent context, provider deduplication, and exact package budgets.

These fixtures are regression gates, not a claim of general retrieval quality. New production failures should be reduced to the smallest deterministic fixture before changing ranking or selection rules.

## Judged repository corpora

Repository-owned corpora under `evaluation/retrieval/` define real development questions with required, supporting, forbidden, structural, and query-profile expectations. The current report set includes Grimoire self-evaluation and external repository comparisons. Every report is tied to the exact repository revision, prepared/vector state, mode, provider set, and date used to produce it.

## Evaluation stages

For source evidence, the evaluator records whether each judged item was:

1. present in the prepared index;
2. found by the broad diagnostic probe;
3. retrieved by the production candidate limit;
4. recovered by exact matching;
5. retained during candidate merge;
6. retained during curation;
7. retained during adaptive assembly; and
8. included after final package fitting.

Structural evidence has parallel stages for provider production, composition, adaptive assembly, and final inclusion. This separates search failure from loss introduced later in the package pipeline.

## Failure classes

Source failures include stale/incomplete index, embedding or vector-ranking miss, exact-recovery miss, candidate-merge loss, curation loss, adaptive-assembly loss, and budget-fitting loss.

Structural failures include provider miss, composition loss, structural adaptive-assembly loss, and structural budget-fitting loss. The broad diagnostic probe is excluded from reported production latency.

## Corpus schema

Each case may define:

- `required`, `supporting`, and `forbidden` source evidence;
- `required_structural`, `supporting_structural`, and `forbidden_structural` evidence;
- a fixed case budget;
- category and rationale; and
- `expected_query_profile` assertions.

Source evidence identifies paths and optionally symbols. Structural evidence identifies provider and kind, then may constrain subject symbol/path, relation, direction, certainty, target, ordered call-chain subsequence, or unresolved expression.

Before execution, the runner validates referenced source paths and symbols paired with paths. Incorrect expectations must be corrected rather than counted as retrieval failures.

## Metrics

Final-package metrics include required/supporting source and structural recall, forbidden violations, irrelevant-selection rates, package tokens, selected chunks, and budget utilization.

Pre-curation ranking metrics include required recall at 10 and 20, mean reciprocal rank, and judged relevance at 10 and 20.

Policy metrics include profile agreement, selected scope and target, curated versus assembled counts, represented regions and roles, stop reason, and evidence lost specifically during assembly.

## Fixed versus adaptive runs

A normal corpus run uses each case's fixed budget unless overridden. `--adaptive` discards case budgets for execution and activates the same automatic policy used by `grimoire context` with no positive budget.

A paired comparison must answer separately:

1. Did policy and assembly preserve judged evidence?
2. Did selected package size and composition improve task suitability?

Zero assembly loss is a necessary regression gate, not proof that retrieval or the automatic target is optimal.

## Current measured interpretation

The July 23, 2026 provider-attribution runs established that vector retrieval produced materially different and often stronger pre-curation results than lexical retrieval, while final package recall could still remain unchanged because useful candidates were lost later. That result is why Grimoire now reports ranking, curation, adaptive assembly, and final fitting as separate stages.

The query-shape comparison demonstrated deterministic profile agreement and zero adaptive-assembly loss on its judged 12-case run after reserve calibration. Its package targets remain calibration choices, not universal optima.

The detailed dated JSON and Markdown files under `evaluation/results/` are authoritative for those measurements. Do not copy percentages into product claims without the corresponding corpus and variant.

## Interpreting failures

A case that never entered production candidates requires retrieval or ranking work. A candidate lost at merge or curation requires pipeline work. A candidate preserved through assembly but omitted by the compiler requires budget or fitting-order work. A structural provider miss may originate in Lexicon seed discovery before Arcana is queried.

Low irrelevant-selection rate is useful only when required recall remains acceptable. Smaller packages are not automatically better when they remove answer-critical evidence.

See [Ranking calibration corpus](ranking-calibration-corpus.md) and [Testing and benchmarks](testing-and-benchmarks.md).
