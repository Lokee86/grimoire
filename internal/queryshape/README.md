# Query shape

`internal/queryshape` deterministically observes prompt specificity, breadth,
ambiguity, candidate dispersion, and graph-region spread before context assembly.

`Analyze` returns two contracts:

- `Profile` records the measured query and retrieval shape.
- `RetrievalPolicy` recommends focused, bounded, or exploratory assembly.

Retrieval evaluation retains the policy in shadow mode. The normal `context`
command activates automatic budget and assembly recommendations only when
`--budget` is omitted or zero. Explicit positive budgets retain fixed assembly.
Automatic requests use the policy scope to select bounded evidence coverage and
emit the resulting assembly decision in the context package.

Automatic target and maximum recommendations are:

| Scope | Target | Maximum |
| --- | ---: | ---: |
| focused | 3,000 | 4,000 |
| bounded | 6,000 | 8,000 |
| exploratory | 12,000 | 16,000 |
