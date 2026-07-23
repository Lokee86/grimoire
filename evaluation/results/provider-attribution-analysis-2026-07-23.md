# Lexical, vector, and hybrid provider attribution

Generated from the frozen Gum and Grimoire retrieval corpora on July 23, 2026.
Hybrid uses neutral provider-rank interleaving; lexical scores and cosine scores are not compared directly.

## Gum

| Mode | Pass | Final required | R@10 | R@20 | MRR | Final irrelevant |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 0.0% | 28.6% | 26.7% | 31.7% | 0.476 | 74.1% |
| vector | 20.0% | 50.0% | 45.0% | 60.0% | 0.477 | 70.7% |
| hybrid | 0.0% | 42.9% | 50.0% | 55.0% | 0.557 | 69.0% |

### Required-evidence survival

| Mode | Required | Retrieved | Curated | Included | Budget-fitting losses |
| --- | ---: | ---: | ---: | ---: | ---: |
| lexical | 14 | 14 | 14 | 4 | 10 |
| vector | 14 | 14 | 14 | 7 | 7 |
| hybrid | 14 | 14 | 14 | 6 | 8 |

### Per-query provider contribution

| Case | Lexical R@20 | Vector R@20 | Hybrid R@20 | Vector-only required in top 20 | Lexical/vector top-20 overlap | Hybrid vector required retrieved/curated/included |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `gum-ao-01` | 0.0% | 66.7% | 66.7% | 2 | 10/20 | 3/3/2 |
| `gum-cc-01` | 33.3% | 33.3% | 33.3% | 0 | 8/20 | 3/3/1 |
| `gum-dl-01` | 50.0% | 100.0% | 100.0% | 1 | 8/20 | 2/2/1 |
| `gum-lm-01` | 75.0% | 50.0% | 75.0% | 0 | 7/20 | 2/2/1 |
| `gum-me-01` | 0.0% | 50.0% | 0.0% | 2 | 9/20 | 3/3/0 |

Vector contributed 5 required top-20 candidates that lexical did not retrieve in the same top-20 window.
Hybrid carried 13 required vector candidates into retrieval and 13 through curation, but only 5 reached the final package.
Lexical and vector top-20 lists overlapped in 42 of 100 compared positions, so the providers were not merely returning identical candidates.

## Grimoire

| Mode | Pass | Final required | R@10 | R@20 | MRR | Final irrelevant |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 0.0% | 2.2% | 10.4% | 25.0% | 0.058 | 87.3% |
| vector | 0.0% | 2.2% | 16.5% | 29.0% | 0.080 | 87.6% |
| hybrid | 0.0% | 2.2% | 12.5% | 20.6% | 0.059 | 87.4% |

### Required-evidence survival

| Mode | Required | Retrieved | Curated | Included | Budget-fitting losses |
| --- | ---: | ---: | ---: | ---: | ---: |
| lexical | 45 | 34 | 33 | 1 | 32 |
| vector | 45 | 35 | 34 | 1 | 33 |
| hybrid | 45 | 33 | 33 | 1 | 32 |

### Per-query provider contribution

| Case | Lexical R@20 | Vector R@20 | Hybrid R@20 | Vector-only required in top 20 | Lexical/vector top-20 overlap | Hybrid vector required retrieved/curated/included |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `grimoire-ao-01` | 25.0% | 0.0% | 0.0% | 1 | 9/20 | 3/3/0 |
| `grimoire-ao-02` | 25.0% | 25.0% | 25.0% | 1 | 14/20 | 5/5/0 |
| `grimoire-cc-01` | 0.0% | 0.0% | 0.0% | 0 | 9/20 | 1/1/0 |
| `grimoire-cc-02` | 33.3% | 0.0% | 16.7% | 0 | 9/20 | 2/2/0 |
| `grimoire-dl-01` | 0.0% | 100.0% | 0.0% | 1 | 11/20 | 3/3/0 |
| `grimoire-dl-02` | 100.0% | 100.0% | 100.0% | 0 | 9/20 | 2/2/0 |
| `grimoire-dl-03` | 0.0% | 0.0% | 0.0% | 0 | 12/20 | 4/4/1 |
| `grimoire-lm-01` | 0.0% | 0.0% | 0.0% | 0 | 5/20 | 10/7/0 |
| `grimoire-lm-02` | 0.0% | 14.3% | 14.3% | 2 | 8/20 | 9/9/1 |
| `grimoire-me-01` | 50.0% | 50.0% | 0.0% | 1 | 11/20 | 2/2/0 |
| `grimoire-me-02` | 0.0% | 25.0% | 25.0% | 1 | 8/20 | 5/5/1 |
| `grimoire-me-03` | 66.7% | 33.3% | 66.7% | 1 | 8/20 | 3/3/1 |

Vector contributed 8 required top-20 candidates that lexical did not retrieve in the same top-20 window.
Hybrid carried 49 required vector candidates into retrieval and 46 through curation, but only 4 reached the final package.
Lexical and vector top-20 lists overlapped in 113 of 240 compared positions, so the providers were not merely returning identical candidates.

## Conclusion

Vector retrieval is measurably different and often better before package fitting. The earlier apparent lack of improvement came from evaluating the final package, where useful vector evidence was discarded and the modes converged.

Gum demonstrates provider value directly: vector-only materially improves ranking and final recall, while hybrid improves early ranking but loses some of that gain downstream.

Grimoire isolates the masking effect: vector improves R@10, R@20, and MRR, yet lexical, vector, and hybrid finish with the same final required recall. Required evidence generally survives retrieval and curation, then disappears during package fitting.

The next tuning target is therefore package fitting and final selection, not another vector-weight guess. Lexical stopword and boundary defects remain real, but they are not the explanation for the vector benchmark result.
