# Query-shape fixed vs adaptive comparison

Generated from the same prepared Grimoire index and the 12-case repository-owned lexical corpus on 2026-07-23.

> This worktree was created from the committed `main` tip and deliberately excludes the active uncommitted retrieval-ranking calibration in the primary checkout. The source-recall figures therefore validate the adaptive budgeting and assembly mechanics, not final retrieval quality.

## Aggregate comparison

| Variant | Profile agreement | Required recall | Irrelevant selections | Median tokens | p95 tokens | Median chunks | Median budget use |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| fixed | 100.0% | 0.0% | 98.9% | 5937 | 11900 | 8.0 | 99.0% |
| adaptive | 100.0% | 2.2% | 96.9% | 8715 | 11990 | 9.5 | 99.2% |

## Category comparison

| Category | Fixed median tokens | Adaptive median tokens | Fixed required recall | Adaptive required recall |
| --- | ---: | ---: | ---: | ---: |
| architecture-ownership | 5930 | 11830 | 0.0% | 0.0% |
| call-chain-investigation | 7980 | 11932 | 0.0% | 0.0% |
| direct-location | 2964 | 5947 | 0.0% | 33.3% |
| long-mixed-query | 11903 | 11712 | 0.0% | 0.0% |
| mechanism-explanation | 5932 | 5953 | 0.0% | 0.0% |

## Per-case policy results

| Case | Scope | Fixed budget | Adaptive budget | Fixed tokens | Adaptive tokens | Assembled candidates | Stop reason | Assembly losses |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| grimoire-dl-01 | bounded | 3000 | 6000 | 2964 | 5947 | 151 | bounded evidence coverage satisfied | 0 |
| grimoire-dl-02 | bounded | 3000 | 6000 | 2786 | 5941 | 156 | bounded evidence coverage satisfied | 0 |
| grimoire-dl-03 | bounded | 3000 | 6000 | 2971 | 5998 | 134 | bounded evidence coverage satisfied | 0 |
| grimoire-me-01 | bounded | 6000 | 6000 | 5905 | 5930 | 122 | bounded evidence coverage satisfied | 0 |
| grimoire-me-02 | bounded | 6000 | 6000 | 5980 | 6000 | 130 | bounded evidence coverage satisfied | 0 |
| grimoire-me-03 | bounded | 6000 | 6000 | 5932 | 5953 | 127 | bounded evidence coverage satisfied | 0 |
| grimoire-ao-01 | exploratory | 6000 | 12000 | 5942 | 11718 | 98 | exploratory evidence coverage satisfied | 0 |
| grimoire-ao-02 | exploratory | 6000 | 12000 | 5919 | 11941 | 102 | exploratory evidence coverage satisfied | 0 |
| grimoire-cc-01 | exploratory | 8000 | 12000 | 7966 | 11988 | 91 | exploratory evidence coverage satisfied | 0 |
| grimoire-cc-02 | exploratory | 8000 | 12000 | 7995 | 11876 | 64 | exploratory evidence coverage satisfied | 0 |
| grimoire-lm-01 | exploratory | 12000 | 12000 | 11878 | 11993 | 64 | exploratory evidence coverage satisfied | 0 |
| grimoire-lm-02 | exploratory | 12000 | 12000 | 11928 | 11430 | 68 | exploratory evidence coverage satisfied | 0 |

## Result

- Query-profile agreement remained 12/12 cases.
- Adaptive source assembly losses: 0.
- Adaptive structural assembly losses: 0.
- Every adaptive run stopped through deterministic evidence coverage rather than exhausting the full curated candidate set.
- Median package size increased because the corpus intentionally maps architecture and call-chain queries to 12,000-token exploratory targets and all current direct-location cases classify as 6,000-token bounded queries. This confirms variable sizing; it is not yet evidence that the selected targets are optimal.
- The next calibration should be repeated after the active retrieval-ranking work is merged, using the same fixed/adaptive comparison to tune target budgets and reserve multipliers against representative recall.
