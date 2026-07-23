# Query-shape package

`internal/queryshape` deterministically classifies a prompt and emits the retrieval policy used by automatic context construction.

## Inputs

Analysis receives:

- the raw query;
- any caller-supplied budget;
- exact, ranked, merged, and structural candidates; and
- ranking confidence and path/graph dispersion derived from those candidates.

Prompt semantics remain separate from candidate ranking. Retrieval evidence may refine breadth or ambiguity, but it does not silently rewrite candidate scores.

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

This package owns classification and policy selection. It does not retrieve, rank, curate, assemble, or compile evidence. `internal/assembly` decides when scope-specific evidence coverage is sufficient, and `internal/compiler` enforces the final token boundary.
