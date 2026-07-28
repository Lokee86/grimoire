# LevelDB background-compaction pause benchmark

Date: 2026-07-28

Task category: impact analysis

Repository: `corpus/leveldb`

Pinned revision: `99b3c03b3284f5886f9ef9a4ef703d57373e61be`

Model/provider: GPT-5.6 Sol through OpenAI Codex

Plain, CBM, and the original Grimoire condition ran concurrently in separate clean detached worktrees. After fixing the Lexicon-to-Arcana compatibility defect, only the Grimoire condition was rerun against the same pinned revision. The agent investigation used fully prepared current Lexicon, Arcana, Grimoire, and documentation state.

## Task

> We want callers to pause and resume automatic background compaction while explicit manual compaction remains available. Investigate the safest ownership seam, explain what behavior would be dangerous, and produce a minimal implementation and test plan.

The prompt did not disclose the hidden scheduling, manual-compaction, write-pressure, concurrency, or lifecycle checklist.

## Current comparison

| Condition | Preparation | Agent elapsed | Total elapsed | Model calls | Total tokens | Answer bytes | Grounding |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| CBM | 1.2s | **4m 51s** | **4m 52s** | **10** | **494,724** | 18,724 | Pass |
| Grimoire, healthy rerun | 12.1s | 5m 33s | 5m 45s | 17 | 1,287,857 | 20,491 | **Invalid** |
| Plain | none | 6m 43s | 6m 43s | 12 | 597,777 | 19,399 | Pass |

Grounding details:

| Condition | Checked inline citations | Structured evidence items | Handle items | Canonical handles | Findings |
| --- | ---: | ---: | ---: | ---: | --- |
| CBM | 60 | 18 | 0 | 0 | none |
| Plain | 62 | 23 | 0 | 0 | none |
| Grimoire, healthy rerun | 50 | 17 | 4 | 3 | one handle/range mismatch |

Every condition covered the hidden `db/db_impl*`, `include/leveldb/`, and `db/` path families. The healthy Grimoire answer cited existing files and valid line ranges, but one evidence item claimed `db/db_impl.cc:702-727` while its canonical handle resolved to `db/db_impl.cc:702-781`. The hardened harness therefore rejected the answer.

## Shared technical conclusion

All conditions found the central ownership boundary:

- `DBImpl` owns the per-database mutex, condition variable, scheduled/running flag, immutable memtable, manual request, version set, shutdown state, and background error.
- `MaybeScheduleCompaction()` is the common scheduling seam for immutable-memtable flushes, explicit manual requests, and automatic `VersionSet::NeedsCompaction()` work.
- The pause must apply only to automatic table-to-table compaction.
- Immutable-memtable flushing must remain available or ordinary writes and `CompactRange()` can deadlock.
- Explicit manual compaction must remain available because it installs `manual_compaction_` and waits on the same scheduler and condition-variable path.
- Resume must call `MaybeScheduleCompaction()` so accumulated automatic demand is not left dormant.
- Long pauses may still reach Level-0 slowdown and stop thresholds; preserving that backpressure is safer than bypassing it.
- The control belongs per `DBImpl`, not in `Env` or `VersionSet`.

All answers rejected mid-compaction cancellation and global `Env` pausing.

## Qualitative comparison

### CBM

CBM remains the strongest answer.

It closes the queued-callback race by applying one automatic-work predicate in both `MaybeScheduleCompaction()` and `BackgroundCall()`. A callback queued before pause can acquire the mutex, observe that only paused automatic work remains, clear the scheduled flag, and wake waiters without beginning another table compaction.

CBM also:

- uses balanced per-database pause ownership rather than an unowned boolean;
- preserves immutable and manual work in scheduling and callback execution;
- accounts for the test-only `ModelDB` subclass and optional C API parity;
- proposes a shared-`Env`, two-database regression test;
- covers read-triggered and write-triggered automatic compaction, queued and running work, write stalls, manual progress, resume, shutdown, and API compatibility.

### Plain

Plain remains a strong valid answer.

It uses balanced pause depth, gives the clearest synchronous public contract, and has strong nested-pause and deterministic barrier tests. Its main omission is that it changes only `MaybeScheduleCompaction()`. A callback already queued before pause may still enter automatic `BackgroundCompaction()` while pause waits for the scheduled flag to clear.

### Grimoire healthy rerun

The healthy rerun is substantially better than the original degraded Grimoire answer.

It now identifies the queued-callback race explicitly and proposes gating both `MaybeScheduleCompaction()` and `BackgroundCall()`. It preserves immutable-memtable and manual work, rejects cancellation, accounts for `ModelDB` and optional C wrappers, and provides deterministic tests for queued work, in-flight work, manual progress, resume, and write pressure.

Its remaining architectural weakness is pause ownership. It recommends idempotent non-blocking boolean semantics and explicitly tests that two pauses followed by one resume restart work. Because `DB` is safe for concurrent callers, two independent owners can both pause and one owner can resume while the other still expects automatic compaction suppressed. A balanced count or explicit ownership token is safer unless the API deliberately defines global last-writer-wins state.

The healthy Grimoire plan is more complete than plain on callback execution, but weaker than plain on overlapping-caller semantics. CBM combines both strengths and remains the best plan.

## Efficiency interpretation

Compared with CBM, the healthy Grimoire rerun used:

- seven more model calls;
- 793,133 more total tokens, about 160% more;
- about 42 seconds more agent time;
- about 53 seconds more total time including preparation;
- an invalid final evidence submission.

Compared with plain, the healthy Grimoire rerun:

- finished about 70 seconds earlier at the agent stage and 58 seconds earlier including preparation;
- used five more model calls;
- used 690,080 more total tokens, about 115% more;
- produced a stronger queued-callback analysis but weaker pause-ownership semantics;
- remained grounding-invalid.

This is no longer the severe efficiency failure shown by the degraded run, but it is not a Grimoire win. CBM still wins on wall time, calls, tokens, grounding, and overall technical completeness.

## Healthy Grimoire preparation

The rerun used the rebuilt packaged binaries after the Arcana compatibility fix. Preparation completed successfully:

| Stage | Time |
| --- | ---: |
| Lexicon | 7.532s |
| Source index | 2.122s |
| Reinspection | 1.287s |
| Documentation | 0.402s |
| Arcana | 0.328s |
| Initial inspection | 0.300s |
| Final verification | 0.138s |
| Marker writes | 0.004s |

Lexicon and Arcana were current and aligned on snapshot `sha256:47d675e03a73f00914e07d9e2c69724b34e2e0d5996ddd1e4eb33be9c9fdb4e4`. No provider warning or fallback occurred.

The agent used one Grimoire search:

- 31,976 response bytes;
- 26 new nodes;
- 20 new source ranges;
- eight graph paths;
- no documents;
- no tool errors or malformed records.

It then switched to direct source inspection. This is materially healthier than the degraded run's four calls and 137,982 response bytes.

## Original degraded run

The original Grimoire condition ran before the compatibility fix. Arcana rejected the valid C-family `unsupported-macro-expansion` label as an unknown enum value and Grimoire continued in source and lexical fallback mode.

That run took 6m 30s of agent time, used 16 calls and 1,566,181 tokens, emitted 137,982 bytes across four discovery calls, proposed only a scheduler-time gate, and was invalidated by three canonical handle/range mismatches.

Its full artifacts remain retained with `grimoire.degraded-2026-07-28` filename prefixes. They are historical evidence of fallback behavior and provider compatibility failure, not the current graph-assisted result.

## Conclusions

The healthy rerun changes the LevelDB interpretation:

1. The fixed Arcana graph can support a materially better impact-analysis answer. The agent found the prequeued-callback race that the degraded run missed and reduced discovery output from 138 KB to 32 KB.
2. Grimoire still did not beat CBM. CBM produced the safest ownership semantics, the most complete test surface, lower latency, fewer calls, fewer tokens, and valid evidence.
3. Grimoire and plain have mixed qualitative advantages. Grimoire is stronger on callback execution; plain is stronger on balanced caller ownership.
4. Canonical handle transfer remains defective. Even with a healthy graph and only one discovery call, the final answer attached a narrower range to a broader handle and was automatically invalidated.
5. This result no longer supports a claim that graph-assisted Grimoire is ineffective on this task. It supports a narrower claim: healthy structural discovery improved the answer and response shape, but did not establish superiority over CBM or plain inspection.

## Retained artifacts

Current healthy rerun:

- `grimoire.stdout.txt`: full unedited answer;
- `grimoire.usage.json`: model telemetry;
- `grimoire.grounding.json`: automatic grounding report;
- `grimoire.mcp-audit.jsonl`: exact search request and structured response;
- `grimoire-prewarm.stdout.txt`: healthy preparation state and timings.

Original degraded condition:

- files prefixed `grimoire.degraded-2026-07-28` in the same task directory.

The canonical `evaluation/results/agent-benchmark-v2/summary.json` now records the healthy rerun while preserving plain and CBM. Transient benchmark worktrees are not retained.
