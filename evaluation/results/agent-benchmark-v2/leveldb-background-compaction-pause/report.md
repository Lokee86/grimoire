# LevelDB background-compaction pause benchmark

Date: 2026-07-28

Task category: impact analysis

Repository: `corpus/leveldb`

Pinned revision: `99b3c03b3284f5886f9ef9a4ef703d57373e61be`

Model/provider: GPT-5.6 Sol through OpenAI Codex

Conditions ran concurrently in separate clean detached worktrees. Every condition had normal shell, Git, and direct source inspection. CBM and Grimoire were optional, mutually exclusive additions.

## Task

> We want callers to pause and resume automatic background compaction while explicit manual compaction remains available. Investigate the safest ownership seam, explain what behavior would be dangerous, and produce a minimal implementation and test plan.

The prompt did not disclose the hidden scheduling, manual-compaction, write-pressure, concurrency, or lifecycle checklist.

## Result summary

| Condition | Preparation | Agent elapsed | Total elapsed | Model calls | Total tokens | Answer bytes | Grounding |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| CBM | 1.2s | **4m 51s** | **4m 52s** | **10** | **494,724** | 18,724 | Pass |
| Plain | none | 6m 43s | 6m 43s | 12 | 597,777 | 19,399 | Pass |
| Grimoire | 11.3s | 6m 30s | 6m 41s | 16 | 1,566,181 | 15,989 | **Invalid** |

Grounding details:

| Condition | Checked inline citations | Structured evidence items | Handle items | Canonical handles | Findings |
| --- | ---: | ---: | ---: | ---: | --- |
| CBM | 60 | 18 | 0 | 0 | none |
| Plain | 62 | 23 | 0 | 0 | none |
| Grimoire | 48 | 15 | 4 | 1 | three handle/range mismatches |

Every condition covered the hidden `db/db_impl*`, `include/leveldb/`, and `db/` path families. Grimoire's cited files and line ranges existed, but three evidence items paired narrower model-authored ranges with broader canonical inspected handles. The harness correctly marked the completed run invalid rather than accepting process success.

## Shared technical conclusion

All three found the same central ownership boundary:

- `DBImpl` owns the per-database mutex, condition variable, scheduled/running flag, immutable memtable, manual request, version set, shutdown state, and background error.
- `MaybeScheduleCompaction()` is the common scheduling seam for immutable-memtable flushes, explicit manual requests, and automatic `VersionSet::NeedsCompaction()` work.
- The pause must apply only to automatic table-to-table compaction.
- Immutable-memtable flushing must remain available or ordinary writes and `CompactRange()` can deadlock.
- Explicit manual compaction must remain available because it installs `manual_compaction_` and waits on the same scheduler/condition-variable path.
- Resume must call `MaybeScheduleCompaction()` so accumulated automatic demand is not left dormant.
- Pausing for long enough can still reach Level-0 slowdown and stop thresholds; preserving that backpressure is safer than bypassing it.
- The control belongs per `DBImpl`, not in `Env` or `VersionSet`.

All three also rejected mid-compaction cancellation and global `Env` pausing.

## Qualitative comparison

### CBM

CBM produced the strongest answer.

It identified an important queued-callback race that the other answers did not fully close: checking only `MaybeScheduleCompaction()` does not stop an automatic callback that was queued before the pause. Its design applies one automatic-work predicate in both `MaybeScheduleCompaction()` and `BackgroundCall()`. A pre-pause callback can then acquire the mutex, observe that only paused automatic work remains, clear the scheduled flag, and wake the pause waiter without starting another compaction.

CBM also:

- used a balanced per-database pause count rather than an unowned boolean;
- preserved immutable and manual work in both scheduling and callback execution;
- accounted for the test-only `ModelDB` subclass and optional C API parity;
- proposed a shared-`Env`, two-database regression test to catch implementations that park the single environment worker;
- covered read-triggered and write-triggered automatic compaction, queued/running work, write stalls, manual progress, resume, and API compatibility.

Its plan is the smallest one that cleanly distinguishes automatic policy from shared execution machinery.

### Plain

Plain produced a strong second-place answer.

It used a balanced pause depth, gave the clearest public contract, and had the strongest nested-pause and deterministic barrier test plan. It correctly set the pause state before waiting, preserved immutable/manual scheduling, and resumed accumulated work explicitly.

Its main omission relative to CBM is that it changed only `MaybeScheduleCompaction()`. A callback already queued before pause may still enter `BackgroundCompaction()` and start automatic work after the pause call begins. Because pause waits for `background_compaction_scheduled_` to clear, the proposed synchronous contract remains safe when the method returns, but the implementation performs avoidable automatic work during the pause transition and does not close the scheduling/execution race as precisely.

### Grimoire

The Grimoire answer found the correct broad ownership seam and most major failure modes, but it was weaker and cannot be accepted as a valid benchmark answer.

Its proposed state was an idempotent boolean. `DB` is explicitly safe for concurrent use, so two independent pause owners could race: one caller's resume can restart automatic compaction while another still expects it paused. The answer acknowledged that manual work might conservatively extend pause time but did not define balanced ownership or an error path for unmatched resume.

Like plain, it gated `MaybeScheduleCompaction()` but did not add the CBM answer's explicit callback-execution predicate. It also submitted three structured claims with handles that resolved to different canonical ranges, causing automatic invalidation.

## Efficiency interpretation

Compared with CBM, Grimoire used:

- six more model calls;
- 1,071,457 more total tokens, about 217% more;
- about 99 seconds more agent time;
- about 109 seconds more total time including preparation.

Compared with plain, Grimoire used:

- four more model calls;
- 968,404 more total tokens, about 162% more;
- roughly the same total wall time;
- an invalid final evidence submission.

CBM won on wall time, token use, call count, grounding, and technical completeness.

## Grimoire preparation failure

Grimoire cold preparation took 11.285 seconds externally and 11.258 seconds in internal timing buckets:

| Stage | Time |
| --- | ---: |
| Lexicon | 6.887s |
| Source index | 2.031s |
| Reinspection | 1.241s |
| Documentation | 0.398s |
| Initial inspection | 0.301s |
| Arcana | 0.232s, failed |
| Final verification | 0.152s |
| Marker writes | 0.006s |

Arcana synchronization failed with:

> `Lexicon snapshot is malformed: unresolved reason`

The fallback kept deterministic source discovery available, so the agent run completed, but this was not a healthy full Grimoire stack. The failure is still a product result: a supported C/C++ repository produced Lexicon state that Arcana rejected, and the degradation was visible only as a preparation warning.

The benchmark therefore does not measure whether healthy Arcana graph traversal would improve this task. It does measure current fallback behavior and preparation reliability.

## Discovery behavior

The degraded Grimoire condition made four calls:

- three searches;
- one inspect;
- 137,982 bytes of structured responses;
- 89 new nodes;
- 92 new source ranges;
- 10 documents;
- 26 graph-path records;
- no MCP execution errors after preparation.

This is the largest discovery expansion in the first four version 2 tasks despite LevelDB being smaller than Space Rocks. The agent consumed over three times CBM's total token volume without producing a superior or valid answer.

## Conclusions

This is a strong negative result for Grimoire in its current form.

The task required impact analysis across scheduling, manual progress, write pressure, concurrency, shutdown, public API, and tests. It was not a trivial symbol lookup, yet CBM and plain inspection both outperformed Grimoire. The result exposes three separate issues:

1. C/C++ prepared-state compatibility is not reliable enough: Arcana rejected the Lexicon snapshot.
2. Broad lexical/source fallback can expand aggressively when structural graph support is unavailable.
3. Canonical handle use is not yet reliably transferred into exact final evidence ranges.

It also validates the benchmark hardening: without automatic handle/range checking, the Grimoire answer would have appeared grounded and might have been manually accepted.

## Retained artifacts

- `plain.stdout.txt`, `cbm.stdout.txt`, `grimoire.stdout.txt`: full unedited answers;
- matching `.usage.json`: model telemetry;
- matching `.grounding.json`: automatic grounding reports;
- `grimoire.mcp-audit.jsonl`: exact Grimoire requests and structured responses;
- `grimoire-prewarm.stdout.txt`: preparation state, warning, and timing breakdown;
- CBM indexing stdout/stderr;
- `evaluation/results/agent-benchmark-v2/summary.json`: combined machine-readable suite summary.

Transient CBM databases and detached benchmark worktrees are not part of the retained evidence.
