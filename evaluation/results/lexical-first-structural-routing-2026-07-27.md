# Lexical-first structural routing

Date: 2026-07-27

## Decision

Natural-language `grimoire context` queries now use deterministic lexical discovery before structural inspection:

1. Rank semantic chunks with BM25.
2. Rank whole production files with BM25.
3. Preserve candidates from both granularities.
4. Resolve Lexicon declarations only inside the selected lexical ranges.
5. Expand those declarations through Arcana for bounded graph evidence.
6. Protect the first 12 source candidates, admit at most eight interleaved structural facts, then continue fitting source candidates.

Lexicon and Arcana do not inject or reorder source candidates in the default route. Structural evidence is supporting evidence attached after lexical discovery.

The previous repository-wide structural discovery route remains available through `--structural-scope global`. Direct `grimoire query search|trace|impact|inspect` operations remain repository-wide and are unaffected by context routing.

## Paired retrieval result

Corpus: `evaluation/retrieval/grimoire.json`  
Cases: 12  
Prepared index: 962 files, 4,828 semantic chunks

| Metric | Source only | Scoped Lexicon + Arcana |
| --- | ---: | ---: |
| Final required-source recall | 11.1% | 11.1% |
| Required R@10 | 26.4% | 26.4% |
| Required R@20 | 50.0% | 50.0% |
| MRR | 0.184 | 0.184 |
| Irrelevant source selections | 72.4% | 69.9% |
| Median latency | 986.7 ms | 3,330.0 ms |
| p95 latency | 3,686.0 ms | 7,900.5 ms |

The structural route preserves every measured source-ranking and final-recall metric. Its current cost is approximately 2.34 seconds of median provider overhead on this evaluator path.

The corpus pass rate remains zero because 23 of its 45 required-source judgments refer to files removed by subsequent repository consolidation. Those stale expectations are recorded as `incorrect evaluation expectation`; pass rate is therefore not treated as an absolute quality measure here.

## Context-package smoke test

Query: `Where is SearchDetailed implemented and what calls it?`  
Budget: 12,000 tokens

| Route | Source selections | Structural facts | Structural providers | Tokens | Wall time |
| --- | ---: | ---: | --- | ---: | ---: |
| Default lexical scope | 13 | 2 | Lexicon, Arcana | 11,888 | 11.165 s |
| Global structural scope | 4 | 7 | Lexicon, Arcana | 11,956 | 10.381 s |

The smoke run confirms that scoped context includes graph evidence without allowing it to replace the source-first package. The global compatibility route retains its former graph-heavy composition.

## Global graph verification

The following direct query continued to search the complete graph and returned repository-wide incoming dependents for `SearchDetailed`:

```text
grimoire query impact --root . --anchor SearchDetailed --direction incoming --depth 2 --limit 6
```

Returned dependents included `collectStructuralContext`, the public `Search` wrapper, and relevant tests/modules. Direct graph access is independent of `context --structural-scope`.

## Commands

```text
grimoire eval retrieval --root . \
  --cases evaluation/retrieval/grimoire.json \
  --modes lexical --adaptive \
  --structural-providers none \
  --structural-scope lexical \
  --variant lexical-first-source-only-final \
  --output-prefix lexical-first-source-only-final-2026-07-27

grimoire eval retrieval --root . \
  --cases evaluation/retrieval/grimoire.json \
  --modes lexical --adaptive \
  --structural-providers lexicon,arcana \
  --structural-scope lexical \
  --arcana-semantic off \
  --variant lexical-first-scoped-structural-final \
  --output-prefix lexical-first-scoped-structural-final-2026-07-27
```

## Retained reports

- `evaluation/results/lexical-first-source-only-final-2026-07-27.json`
- `evaluation/results/lexical-first-source-only-final-2026-07-27.md`
- `evaluation/results/lexical-first-scoped-structural-final-2026-07-27.json`
- `evaluation/results/lexical-first-scoped-structural-final-2026-07-27.md`
