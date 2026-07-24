# Query decomposition proof

Date: 2026-07-23  
Branch: `codex/query-decomposition`  
Corpus: `evaluation/retrieval/grimoire.json`  
Cases: 12  
Mode: lexical  
Structural providers: disabled

## Implemented

- Decomposes long prose requests at action boundaries instead of relying on a small leading-cue vocabulary.
- Recognizes coordinated repository actions such as planning/search, exact recovery, merge/curation, package fitting, and evaluation reporting.
- Parses Markdown requests by implementation section and removes fenced examples before retrieval planning.
- Excludes generic goal/deliverable sections and unrelated external corpus or snapshot phases from the bounded clause set.
- Keeps the complete request in the exposed retrieval policy as a low-weight context intent.
- Skips that low-weight complete-request pass for BM25 and exact providers when specific clauses are available, preventing long prompt text from dominating lexical ranking.
- Preserves focused single-purpose queries unchanged.

## Comparison

Both runs used the same prepared snapshot from the consolidated Grimoire tree.

| Metric | BM25 + initial intents | Query decomposition | Change |
| --- | ---: | ---: | ---: |
| Corpus pass rate | 8.3% | 8.3% | unchanged |
| Required source recall | 8.9% | 11.1% | +2.2 pp |
| Required R@10 | 11.1% | 13.5% | +2.4 pp |
| Required R@20 | 13.9% | 22.5% | +8.6 pp |
| MRR | 0.049 | 0.073 | +0.024 |
| Irrelevant source selections | 89.5% | 83.2% | -6.3 pp |
| Median latency | 1216.1 ms | 1316.3 ms | +100.2 ms / +8.2% |
| p95 latency | 1616.5 ms | 2233.1 ms | +616.6 ms / +38.1% |

## Long mixed queries

| Metric | BM25 + initial intents | Query decomposition |
| --- | ---: | ---: |
| Required source recall | 0.0% | 23.1% |
| Irrelevant source selections | 96.6% | 73.3% |

The structured long request recovered 16.7% of required evidence, while the coordinated prose request recovered 28.6%. Neither case fully passed; remaining losses are mostly ranking and package-fitting failures after relevant clauses are now represented.

## Tradeoff

The decomposition materially improves long-query evidence coverage and precision, but executes several bounded BM25 passes. Median latency rises modestly; p95 latency rises more substantially on the two longest requests. Further latency reduction should optimize multi-query scoring rather than collapse the clauses back into one prompt.

## Verification

The implementation is covered by deterministic regression tests for:

- coordinated action decomposition;
- Markdown section extraction;
- fenced-example exclusion;
- generic/external section exclusion;
- focused-query preservation;
- low-weight full-prompt provider suppression.
