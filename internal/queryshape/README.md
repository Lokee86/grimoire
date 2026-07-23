# Query shape

`internal/queryshape` deterministically observes prompt specificity, breadth,
ambiguity, candidate dispersion, and graph-region spread before context assembly.

`Analyze` returns two contracts:

- `Profile` records the measured query and retrieval shape.
- `RetrievalPolicy` recommends focused, bounded, or exploratory assembly.

The policy is currently shadow-only. Evaluation reports retain it, but retrieval,
curation, token budgeting, and package compilation do not consume it. A missing
budget is represented as `automatic-shadow` without selecting a token target.
