# Sol High Fast agent benchmark rerun

Date: July 29, 2026  
Harness commit: `0375e5f6e541970e6ef509ac655961405c82e0f3`  
Model: `gpt-5.6-sol`  
Reasoning: `high`  
Service tier: Fast (`priority` on audited Hermes requests)  
Execution: sequential tasks and sequential conditions

## Executive conclusion

The benchmark supports Grimoire as a useful context-reduction and grounding system for broad repository investigations. Across the four directly comparable Codex tasks, Grimoire used **32.4% less processed input than Plain** and **21.2% less than CBM**, while grounding all four answers and losing only one rubric point.

The LevelDB task exposed an important boundary. It was a comparatively narrow investigation with ownership concentrated in a few `DBImpl` methods. Grimoire found the correct evidence with the least fresh input, but Hermes repeatedly continued investigating after the answer was effectively established. A second isolated Grimoire/Hermes run repeated and worsened that behavior. Grimoire may therefore provide little benefit, or become counterproductive, when direct repository inspection is already cheap and the task does not require broad context.

This is not a contradiction. Retrieval quality, context reduction, answer judgment, and agent stopping behavior are separate concerns. Grimoire performed strongly at finding grounded evidence; the LevelDB inefficiency came from the interaction between structured retrieval and the model's investigation trajectory.

## Completion status

All five primary tasks are complete: **15 of 15 planned runs**.

The first four tasks used Codex CLI. Codex then exhausted its account allowance, so the three LevelDB conditions were completed with Hermes one-shot using the same Sol/High/Fast configuration. Hermes initially exposed a one-shot routing defect that dropped High and Fast settings; those trial runs were archived and excluded. The accepted LevelDB runs all report `service_tier: priority` and nonzero reasoning-token usage.

A supplemental second Grimoire/Hermes LevelDB run was performed after the primary benchmark. It is documented separately and is **not included** in the original 15-run totals or manual quality scores.

## Primary run results

| Task | Plain | CBM | Grimoire |
|---|---:|---:|---:|
| grimoire-state-maintenance-ownership | 286.0s (valid) | 400.2s (valid) | 266.3s (valid) |
| space-rocks-room-scale-architecture | 501.0s (valid) | 462.7s (**invalid**) | 479.9s (valid) |
| detekt-cli-gradle-plugin-divergence | 295.4s (valid) | 307.0s (valid) | 345.0s (valid) |
| space-rocks-distant-player-locator | 425.5s (valid) | 502.6s (valid) | 494.0s (valid) |
| leveldb-background-compaction-pause (Hermes) | 399.1s (valid) | 387.6s (valid) | 422.0s (valid) |

## Raw elapsed totals

| Condition | Total time | Mean time | Grounded runs |
|---|---:|---:|---:|
| Plain | 1907.2s | 381.4s | 5/5 |
| CBM | 2060.0s | 412.0s | 4/5 |
| Grimoire | 2007.3s | 401.5s | 5/5 |

CBM's total includes the invalid room-scale run because this table reports raw elapsed work. Its quality score excludes that run.

## Manual rubric assessment

| Condition | Score | Eligibility |
|---|---:|---|
| Plain | **40/40** | Five valid runs |
| CBM | **32/32** among eligible runs | Four valid; one disqualified |
| Grimoire | **39/40** | Five valid runs |

The one-point Grimoire deduction remains the distant-player locator task: it proposed low-rate updates through the existing hot `ship_delta` path rather than the requested separate low-cadence locator seam.

CBM's room-scale answer remains disqualified because it cited five nonexistent paths and omitted required `shared/` evidence.

## Token-accounting terminology

For the first four Codex runs:

- **Processed input** is the reported aggregate input across model calls, including cached context reread on later calls.
- **Cached input** is the cached portion of that processed input.
- **Fresh input** is calculated as processed input minus cached input.
- Output and reasoning are shown separately. Reasoning tokens are a subset of model output accounting and are not added again to total tokens.

The LevelDB runs used Hermes usage reports with explicit `input_tokens`, `cache_read_tokens`, `output_tokens`, and `api_calls`. Because the runner and accounting schema differ, LevelDB is reported separately rather than folded into the four-task Codex percentages.

## Per-task Codex input usage

Grimoire used the least processed input on every directly comparable Codex task.

| Task | Plain | CBM | Grimoire | Grimoire reduction vs Plain | Grimoire reduction vs CBM |
|---|---:|---:|---:|---:|---:|
| State maintenance ownership | 3,514,755 | 4,417,439 | **2,340,752** | **33.4%** | **47.0%** |
| Room-scale architecture | 10,942,309 | 7,821,228 | **6,492,845** | **40.7%** | **17.0%** |
| Detekt CLI/Gradle divergence | 2,844,203 | 2,396,967 | **2,205,915** | **22.4%** | **8.0%** |
| Distant-player locator | 8,801,965 | 7,763,069 | **6,612,396** | **24.9%** | **14.8%** |
| **Four-task total** | **26,103,232** | **22,398,703** | **17,651,908** | **32.4%** | **21.2%** |

CBM's room-scale token use is retained for accounting even though its answer was disqualified for fabricated evidence.

### Fresh and cached input

| Condition | Fresh input | Cached input | Processed input |
|---|---:|---:|---:|
| Plain | 890,048 | 25,213,184 | 26,103,232 |
| CBM | 889,583 | 21,509,120 | 22,398,703 |
| Grimoire | **842,692** | **16,809,216** | **17,651,908** |

Grimoire reduced fresh input by only about **5.3%** versus Plain. Most of the total gain came from carrying and rereading substantially less accumulated context across later calls. This is the central positive result: Grimoire did not merely return shorter snippets; it reduced repeated context processing while preserving investigation depth and grounding.

### Output and reasoning

| Condition | Output tokens | Reasoning tokens |
|---|---:|---:|
| Plain | 64,221 | 32,613 |
| CBM | 71,241 | 35,821 |
| Grimoire | 65,987 | 33,899 |

Grimoire's savings were not produced by shallow or abbreviated answers. Its aggregate output and reasoning remained close to Plain while processed input fell by nearly one-third.

## Primary LevelDB Hermes comparison

All three primary answers scored **8/8**. Each correctly:

- placed pause state in mutex-protected `DBImpl`;
- gated only automatic `VersionSet::PickCompaction` work;
- preserved immutable-memtable flushing and explicit manual compaction;
- rescheduled accumulated debt on final resume;
- addressed queued callbacks, in-flight compaction, L0 write stalls, nested callers, shutdown, and deterministic tests.

| Condition | Time | Fresh input | Cached input | Processed input | Output | Total tokens | Reasoning | API calls |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Plain | 399.1s | 113,789 | 657,408 | 771,197 | 17,584 | 788,781 | 10,692 | 13 |
| CBM | 387.6s | 112,327 | 570,880 | 683,207 | 16,377 | 699,584 | 9,819 | 12 |
| Grimoire | 422.0s | **100,796** | 1,084,928 | 1,185,724 | 17,509 | **1,203,233** | 11,204 | **18** |

CBM was fastest on LevelDB. Grimoire used the least fresh input but the most cached context and was slowest. Answer quality was tied.

### Why LevelDB is a useful boundary case

LevelDB was not trivial: the task involved concurrency, explicit versus automatic compaction, write stalls, immutable memtables, shutdown, and API lifetime semantics. Relative to the other benchmark tasks, however, it was narrow and structurally straightforward:

- the repository was small;
- ownership was concentrated in `DBImpl`;
- the central control flow lived in `MaybeScheduleCompaction`, `BackgroundCompaction`, and `CompactRange`;
- only a few public API and test files were required;
- all three conditions independently reached the same correct design.

This made LevelDB a task where normal search and direct inspection were already inexpensive. Grimoire had less opportunity to reduce a genuinely large context surface.

### Primary Grimoire retrieval trace

The first Grimoire run did not fail to retrieve the answer:

- two Grimoire calls total: one search and one inspect;
- 28,917 response bytes total;
- zero tool errors;
- the first search immediately found the correct scheduling and ownership methods.

The initial response was nevertheless redundant for such a narrow task. A search requested with `limit=8` returned 24 nodes, 24 source ranges, and 24 retrieval hits, totaling about 22.2 KB before the follow-up inspect. The same evidence was represented in multiple structural forms.

The larger inefficiency occurred after retrieval. Hermes made 18 model calls, compared with 12 for CBM and 13 for Plain. Evidence supplied early in the conversation was therefore repeatedly included in later cached prompts. A modest retrieval payload became more than one million cache-read tokens.

## Counterfactual normal-length LevelDB estimate

Plain and CBM averaged **12.5 API calls** on LevelDB. Two approximations were considered for what the first Grimoire run might have consumed if Hermes had stopped after a similar number of calls:

1. **Simple proportional scaling:**  
   `1,203,233 × 12.5 / 18 = approximately 835,578 total tokens`.

2. **Accumulating-history approximation:** cached context generally grows with each successive call, so removing the final calls should save more than linear scaling predicts. Treating cache growth as roughly triangular produces approximately **547,000 to 639,000 total tokens** for 12 to 13 calls.

The exact counterfactual cannot be recovered without per-call token records. A defensible range is therefore approximately **0.55M to 0.84M total tokens**. The important conclusion is not the precise midpoint: a normal-length trajectory likely would have placed Grimoire near the Plain/CBM range rather than at the observed 1.20M total.

## Supplemental second Grimoire/Hermes LevelDB run

A second isolated run used the same task, revision, model, reasoning level, Fast/priority service tier, Grimoire build, and read-only benchmark prompt. It was deliberately excluded from the original benchmark totals.

| Metric | Primary Grimoire run | Supplemental run | Change |
|---|---:|---:|---:|
| API calls | 18 | **19** | +1 |
| Fresh input | 100,796 | **151,660** | +50,864 |
| Cached input | 1,084,928 | **1,536,000** | +451,072 |
| Output | 17,509 | 15,611 | -1,898 |
| Reasoning | 11,204 | 7,775 | -3,429 |
| **Total tokens** | **1,203,233** | **1,703,271** | **+500,038 / +41.6%** |
| Grimoire calls | 2 | **3** | +1 |
| Grimoire response bytes | 28,917 | **53,024** | +24,107 |
| Tool errors | 0 | 0 | none |

The supplemental run again followed an 18-to-19-call trajectory and consumed even more context. This weakens the hypothesis that the first result was merely a one-off stochastic outlier. Model variance still mattered—the second run used an additional retrieval call and 41.6% more total tokens—but the repeated long trajectory indicates a reproducible interaction between this narrow task, Grimoire's structured evidence, and Hermes's stopping behavior.

The supplemental model run completed successfully and its audited usage is valid (`service_tier: priority`, `completed: true`, `failed: false`). The wrapper then encountered a Windows CP1252 decoding error while capturing one character from the final answer. The answer text was lost, so the supplemental run cannot be honestly rescored for quality or grounding and is retained only as usage and trajectory evidence. Its elapsed time is not treated as an authoritative benchmark measurement for the same reason.

The temporary Hermes one-shot bridge was restored byte-for-byte after the run. Restored `oneshot.py` SHA-256: `3c8b2233b803aa873538049e911d21bd12ca9a75c453744e54fff57a57e1244e`. The isolated checkout and profile were removed.

## Conclusions reached from the benchmark and follow-up analysis

1. **Grimoire's strongest demonstrated value is broad-context efficiency.** On large, distributed, cross-layer investigations, it consistently reduced processed input while maintaining strong grounding and answer depth.

2. **Grimoire is not automatically beneficial on every repository task.** When ownership is obvious, the repository is small, and a few direct searches can establish the answer, Grimoire's fixed retrieval structure and duplicated evidence forms may cost more than they save.

3. **The LevelDB outlier contains both model and system effects.** Stochastic model behavior changed the exact number of searches and tokens, but two consecutive 18-to-19-call runs show that the excessive trajectory was not simply random bad luck.

4. **Retrieval succeeded; stopping failed.** Grimoire found the correct evidence early and used the least fresh input in the primary LevelDB comparison. Hermes did not recognize that the task was sufficiently grounded and repeatedly reprocessed the accumulated context.

5. **Retrieval quality and orchestration quality must be evaluated separately.** A retrieval system can provide excellent evidence while still producing poor end-to-end efficiency if the agent responds by exploring more instead of concluding sooner.

6. **Grounding remains a material advantage.** Grimoire and Plain grounded every primary run. CBM produced the benchmark's only invalid answer by fabricating paths on the broad room-scale task.

7. **Answer judgment remains separate from evidence quality.** Grimoire's only primary quality deduction was choosing the wrong architectural seam for the distant-player locator despite retrieving strong evidence.

8. **Cold preparation is still a real cost.** Across the first four tasks, Grimoire prewarming consumed 452.3s versus 20.9s for CBM. LevelDB preparation was much smaller—9.3s initially and 9.5s on the supplemental run—but still exceeded CBM's 0.9s.

9. **The benchmark should not be reduced to a single winner.** Plain was fastest overall and perfect in quality; Grimoire was the most context-efficient on broad tasks and matched Plain's grounding reliability; CBM was cheap to prepare and sometimes fast, but had a serious grounding failure.

## Practical product implications

The evidence suggests a task-sensitive operating model rather than forcing Grimoire into every investigation:

- Prefer Grimoire for broad ownership questions, cross-language or cross-service changes, unfamiliar repositories, impact analysis with many potential lanes, and investigations requiring source plus architectural rationale.
- Prefer direct inspection or a lightweight lexical path for narrow symbol-local questions where the likely owner is already obvious.
- Add a compact/narrow Grimoire response mode that deduplicates overlapping node, range, and retrieval-hit evidence.
- Treat result limits as semantic budgets rather than allowing one requested limit to expand into multiple parallel representations of the same evidence.
- Improve agent instructions or orchestration so discovery stops after the owner, relevant control flow, public boundary, and test seam are grounded.
- Benchmark narrow and broad task classes separately and run multiple seeds before attributing an outlier solely to model variance.

## Benchmark limitations

- The first four tasks and LevelDB used different runners and token-accounting schemas.
- The task set contains only five primary tasks; LevelDB is the only intentionally narrow small-repository boundary case in this suite.
- Only one model and reasoning level were tested in the completed rerun.
- The supplemental LevelDB answer could not be captured or rescored after a wrapper decoding failure.
- Token totals measure model processing, not monetary cost, and cached tokens may be billed or weighted differently by provider policy.
- Preparation time and investigation time are both relevant operational costs but should not be conflated.

## Artifacts

- `summary.json`: primary run metadata, timings, usage, grounding, and runner provenance.
- `manual-quality.json`: weighted primary-run rubric assessment.
- `supplemental-analysis.json`: machine-readable token analysis, LevelDB counterfactuals, rerun metrics, and conclusions.
- Per-task directories: final answers, grounding reports, preparation output, usage reports, and Grimoire audit logs.
- `leveldb-background-compaction-pause/rerun2/README.md`: supplemental-run methodology, capture limitation, metrics, and cleanup record.
- `leveldb-background-compaction-pause/rerun2/recovered-result.json`: recovered audited usage and first-versus-second comparison.
- Rejected Codex-limit and unconfigured-Hermes trials are retained with explicit archival filename suffixes and are not included in scoring.
