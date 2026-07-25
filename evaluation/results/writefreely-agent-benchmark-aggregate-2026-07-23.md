# WriteFreely agent benchmark aggregate

Generated July 23, 2026.

## Scope

This report combines the two compatible WriteFreely agent-discovery benchmark campaigns for the task:

> Trace how a WriteFreely post is created and persisted.

Both campaigns used three independent trials per configuration, a 12,000-token context-package target where applicable, and a 55 repository-discovery-call ceiling. Package-level retrieval benchmarks are summarized separately because recall, irrelevance, and millisecond package-generation metrics are not directly comparable with agent time, calls, and model usage.

## Agent discovery results

| Campaign | Configuration | Trials | Result | Discovery calls | Time | API calls | Token accounting | Tokens |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| Earlier comparison | Bare agent | 3 | 3/3 success | 54 | 250.8 s | 16 | session tokens | 1,210,653 |
| Earlier comparison | Semantic Grimoire | 3 | 3/3 success | 54 | 307.6 s | 19 | session tokens | 1,309,755 |
| Earlier comparison | Grimoire + Lexicon | 3 | 2/3 success | 44 | 277.8 s | 27 | session tokens | 2,661,177 |
| Earlier comparison | Grimoire + Lexicon + Arcana | 3 | 3/3 success | 54 | 237.3 s | 14 | session tokens | 866,947 |
| Hermes comparison | Grimoire + Arcana | 3 | 10/10 average; 3/3 | 48.3 | 268.8 s | 16.0 | noncached input + output | 126,579 |
| Hermes comparison | Grimoire + Lexicon + Arcana | 3 | 10/10 average; 3/3 | 50.3 | 273.3 s | 21.3 | noncached input + output | 102,427 |

The earlier campaign's values are its reported configuration-level aggregates. The Hermes rows are arithmetic means from the three recorded trials. Token values must not be compared directly across campaigns because the earlier report used session-token accounting, while the Hermes report separates noncached input/output from cached tokens.

## Repeated full-stack result

`Grimoire + Lexicon + Arcana` completed successfully in all six trials across the two campaigns.

| Campaign | Success | Discovery calls | Time | API calls |
| --- | ---: | ---: | ---: | ---: |
| Earlier comparison | 3/3 | 54 | 237.3 s | 14 |
| Hermes comparison | 3/3 | 50.3 | 273.3 s | 21.3 |

The repeated result supports full-stack reliability on this task, but not a stable speed or discovery-call advantage. One campaign made it the fastest configuration; the Hermes campaign made it slightly slower and more discovery-heavy than Arcana alone.

## Arcana-enabled aggregate

Across all Arcana-enabled configurations in both campaigns:

- 9/9 trials completed successfully.
- The six newly rubric-scored Hermes trials all achieved 10/10 correctness.
- Arcana alone matched the full stack's correctness while averaging 2.0 fewer discovery calls and finishing 4.5 seconds faster.
- Adding Lexicon to Arcana reduced Hermes noncached model tokens by 19.1%, from 126,579 to 102,427 per trial.

This isolates Lexicon's demonstrated benefit in the Hermes campaign: token efficiency, not correctness, speed, or reduced discovery.

## Related package-level benchmarks

These results describe the retrieval and package-building machinery beneath the agent tests. They are included for interpretation, not pooled with agent metrics.

| Benchmark | Result |
| --- | --- |
| Standalone vs Lexicon-assisted retrieval, July 22 | Required and supporting recall were effectively unchanged across modes. Lexicon assistance reduced some median latencies but did not improve evidence recall. |
| Lexical/vector/hybrid attribution, July 23 | On the Grimoire corpus, vector improved early ranking over lexical at R@10 and R@20, but all three modes finished at 2.2% final required-evidence recall. Useful vector evidence was discarded during final package fitting. |
| Fixed vs adaptive query shape, July 23 | Adaptive sizing raised required recall from 0.0% to 2.2% and reduced irrelevant selections from 98.9% to 96.9%. That run validated adaptive assembly mechanics on a pre-calibration code state; it was not a final retrieval-quality result. |

## Aggregate conclusion

1. **Arcana is the strongest demonstrated contributor to agent correctness on the WriteFreely task.** Every Arcana-enabled trial succeeded, and every newly scored answer achieved 10/10.
2. **Lexicon alone is not yet validated as an agent-performance improvement.** Its earlier configuration completed only 2/3 trials, used the most reported session tokens, and made the most API calls despite reducing discovery calls.
3. **Lexicon can complement Arcana by reducing model-token consumption.** In the controlled Hermes comparison it cut noncached token use by approximately one-fifth, but did not improve correctness, elapsed time, or discovery calls.
4. **Plain semantic Grimoire did not beat the bare-agent baseline in the earlier campaign.** It used the same discovery-call count and took longer with more session tokens and API calls.
5. **The package-level bottleneck remains final evidence selection and budget fitting.** Vector retrieval is finding distinct useful candidates, but the final package frequently discards them.
6. **The agent result and retrieval-corpus result are not contradictory.** Arcana supplies high-value structural facts for this particular call-chain task, allowing perfect answers even while Grimoire's general-purpose fixed-corpus evidence recall remains poor.

The current evidence supports retaining Arcana integration, keeping Lexicon as a structural seed and possible token-efficiency aid, and focusing Grimoire optimization on package fitting rather than further provider-weight tuning.
