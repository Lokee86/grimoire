# Query-shape package

`internal/queryshape` deterministically classifies a prompt and emits the retrieval policy used by automatic context construction.

## Inputs

Analysis receives:

- the raw query;
- any caller-supplied budget;
- exact, ranked, merged, and structural candidates; and
- ranking confidence and path/graph dispersion derived from those candidates.

Query-only retrieval intents are emitted before candidate generation. The app layer consumes them to run bounded BM25 and exact passes, reserve mixed-query coverage, attach candidate roles, and choose the most graph-relevant structural query. Retrieval evidence may then refine breadth or ambiguity without changing the original intent plan.

## Profile

The emitted profile records intent, specificity, breadth, ambiguity, cross-system scope, evidence needs, and reasons.

## Current policy tiers

| Scope | Minimum | Target | Maximum |
| --- | ---: | ---: | ---: |
| Focused | 2,000 | 3,000 | 6,000 |
| Bounded | 3,000 | 6,000 | 10,000 |
| Exploratory | 6,000 | 12,000 | 18,000 |

When the requested budget is zero, `app` activates the policy and uses the target for adaptive compilation. When a caller supplies a positive budget, the policy is emitted in shadow form and fixed-budget behavior remains authoritative.

## Boundary

This package owns classification, query-only intent planning, and policy selection. It does not execute providers, rank candidates, curate, assemble, or compile evidence. `internal/app` consumes intent plans, `internal/assembly` decides when scope-specific evidence coverage is sufficient, and `internal/compiler` enforces the final token boundary.
