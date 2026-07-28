# Agent benchmark findings

This page summarizes the current end-to-end agent benchmarks and the operational conclusions they support. Raw reports remain the source of truth for exact prompts, revisions, tool restrictions, state, and telemetry.

## Current conclusion

These runs are recorded as experimental evidence, but the multi-repository suite is **not a good neutral benchmark of discovery systems**. Its prompts name the subsystem, enumerate the expected ownership areas, and supply enough domain vocabulary to turn much of the work into a checklist of lexical searches. A strong model with `rg` and direct file reads is therefore operating in unusually favorable conditions: most of the ambiguity that discovery tooling is designed to remove has already been removed by the benchmark prompt.

The results also suggest a task-size inversion rather than a simple overhead rule:

- On smaller, direct, strongly named tasks, discovery output can introduce enough adjacent, stale, or weakly ranked context to create a lost-in-the-middle problem that plain `rg` avoids. The model's useful working set was already small, so adding structured context can displace or blur the obvious evidence.
- On longer, cross-cutting, architecturally ambiguous tasks, discovery can prevent the same failure by collapsing a much larger search space into a bounded set of ownership boundaries, relationships, and evidence handles.

This lost-in-the-middle explanation is a working hypothesis consistent with the observed answers, not a controlled causal result. The suite has one trial per condition and was not designed to isolate context-position effects.

The strongest valid assisted result remains the Space Rocks network-interest benchmark, where Grimoire reduced the amount of context and repeated searching required for a broad cross-language architecture task. HikariCP favored plain inspection because the requested lifecycle was compact and concretely named. The earlier prompt-shaped Detekt Grimoire run failed semantically, while the replacement unclear-ownership Detekt task completed successfully in all conditions but favored plain inspection and CBM on efficiency. The repaired Now in Android run completed, but extensive invalid citations show that additional discovery context did not improve grounding on that prompt-shaped, `rg`-friendly task.

The practical conclusion is to use Grimoire when it reduces an otherwise large or ambiguous working set, and stop using it when direct inspection has already narrowed the task sufficiently. The evidence supports **selective Grimoire-assisted normal inspection**, not mandatory discovery-tool use.

## Version 2 benchmark infrastructure

`evaluation/agent_benchmark_tasks.v2.json` replaces checklist-shaped prompts with natural problem reports and hidden rubrics covering architectural exploration, unclear ownership, cross-language change, impact analysis, and source-plus-rationale investigation.

The canonical runner automatically invalidates completed answers when citations or structured evidence reference missing files, invalid paths, or out-of-range lines. Any Grimoire evidence that includes an inspected source-range handle must match its audited canonical path and lines; handle coverage is reported without forcing agents to replace cheaper direct source reads. The runner also reports non-gating coverage of the hidden rubric's expected path families. Preparation timing and discovery-output volume are recorded separately so conclusions can distinguish provider startup cost, response size, process completion, and grounded answer quality.

Saved answers can be revalidated through `evaluation/revalidate_agent_benchmark.py` without rerunning agents. The validator accepts and individually checks noncontiguous structured ranges while rejecting a single canonical handle attached to multiple ranges.

## Version 2 room-scale architecture result

Report: [`evaluation/results/agent-benchmark-v2/space-rocks-room-scale-architecture/report.md`](../../evaluation/results/agent-benchmark-v2/space-rocks-room-scale-architecture/report.md)

The first version 2 task asked the agents to investigate larger-room networking architecture without supplying a subsystem checklist. All three answers were grounded and implementation-grade after a validator format gap for noncontiguous ranges was corrected.

| Condition | Preparation | Agent elapsed | Model calls | Total tokens | Grounding |
| --- | ---: | ---: | ---: | ---: | --- |
| Grimoire | 95.8s | **6m 36s** | **16** | **1,053,805** | Pass |
| CBM | **6.1s** | 11m 33s | 25 | 3,300,961 | Pass |
| Plain | none | 12m 27s | 18 | 1,840,597 | Pass |

All three converged on receiver-scoped interest management over immutable presentation facts while preserving per-session baseline, lifecycle, transport, and client semantics. CBM was marginally strongest on ownership precision, Grimoire was strongest on explicit spectator/shared-contract planning, and plain inspection remained broad and technically sound. The quality difference was small.

Grimoire made one 50,915-byte search call, then shifted to direct inspection. Including its much slower cold preparation, it still completed about 207 seconds before CBM and 255 seconds before plain. This is a strong positive result for broad architectural exploration, not a universal product verdict.

Cold preparation identifies Lexicon as the clear optimization target: 69.3 of 95.2 internal preparation seconds. The single 50.9 KB discovery response also confirms that broad response shaping remains unfinished.

## Version 2 Detekt ownership result

Report: [`evaluation/results/agent-benchmark-v2/detekt-cli-gradle-plugin-divergence/report.md`](../../evaluation/results/agent-benchmark-v2/detekt-cli-gradle-plugin-divergence/report.md)

The second version 2 task asked where responsibility belongs when a third-party rule-set JAR works through the CLI but is missing or configured differently through the Gradle plugin. The prompt named the two entry points but did not supply the classloader, task, core, or fix checklist.

| Condition | Preparation | Agent elapsed | Total elapsed | Model calls | Total tokens | Grounding |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Plain | none | 5m 04s | **5m 04s** | **14** | **767,334** | Pass |
| CBM | **5.3s** | **5m 00s** | 5m 05s | 16 | 831,264 | Pass |
| Grimoire | 31.8s | 5m 54s | 6m 26s | 16 | 1,457,421 | Pass |

All three found the same real boundary defect: Gradle resolves `detektPlugins` into `pluginClasspath` but does not pass the CLI's explicit `--plugins` argument. Instead it co-loads plugin JARs with detekt's host classpath, including a path-keyed daemon classloader cache. All three proposed restoring the explicit plugin argument across analysis, baseline, and config generation while keeping the host classpath detekt-only.

Plain inspection produced the strongest overall answer by a small margin; CBM was a close second; Grimoire remained technically correct but added no unique ownership finding. Grimoire used about 90% more total tokens than plain and completed about 82 seconds later including preparation.

The Grimoire agent made one 44,867-byte search call that emitted 26 nodes, 21 source ranges, eight documents, and eight graph paths. The response began with low-value import nodes. This is a concrete negative result for broad default response shaping on a task whose working set becomes compact after locating the two named entry paths.

## Network-interest architecture benchmark

Report: [`evaluation/results/network-interest-agent-benchmark-2026-07-27-v4/report.md`](../../evaluation/results/network-interest-agent-benchmark-2026-07-27-v4/report.md)

Three GPT-5.6 Sol agents investigated separate clean Space Rocks worktrees at commit `460da4af05c44d1835401fa853f5fc6b718262c8` in parallel:

- plain normal repository tools;
- normal tools plus CBM 0.9.0 and its installed skill;
- normal tools plus Grimoire and `skills/grimoire/SKILL.md`.

Each assisted agent lacked access to the competing discovery system. The task required a read-only implementation plan for per-client network interest across visibility math, receiver-specific projection, baseline and delta membership, hysteresis, spectator behavior, distant-player locators, packet budgets, generated Go/GDScript contracts, client application, tests, and failure modes.

| Condition | Technical rubric | Required evidence contract | Discovery calls | Model API calls | Noncached tokens | Completion |
| --- | ---: | --- | ---: | ---: | ---: | ---: |
| Plain | 10/10 | Pass | 0 | 53 | 471,723 | 16m 54s |
| CBM | 10/10 | Fail | 20 | 54 | 543,471 | 17m 31s |
| Grimoire | 10/10 | Pass | 2 | 22 | 261,631 | 9m 16s |

Compared with plain inspection, Grimoire used:

- 44.5% fewer noncached tokens;
- 58.5% fewer model API calls;
- 58.8% fewer total tokens including cache reads;
- about 45% less completion time.

The Grimoire agent loaded the new skill, made two unified discovery calls, then shifted to targeted source reads. This is the intended behavior: use Grimoire to collapse broad discovery, then use direct inspection for proof.

All three plans covered the ten technical areas. Grimoire's answer remained implementation-grade and complied with the required evidence schema. The result is a strong task-specific win, not a universal performance claim.

## Earlier Space Rocks agent benchmark

Report: [`evaluation/results/space-rocks-grimoire-cbm-agent-benchmark-2026-07-27.md`](../../evaluation/results/space-rocks-grimoire-cbm-agent-benchmark-2026-07-27.md)

The earlier benchmark used four narrower repository tasks, restricted normal discovery tools, and exercised the now-retired context-package path before interactive follow-up. Both Grimoire and CBM completed all 12 effective task trials correctly, while CBM used fewer median noncached tokens and slightly less median elapsed time.

That result remains useful historical evidence, but it should not be generalized to the current progressive-discovery product because:

- agents were prevented from using normal shell discovery;
- Grimoire paid for a preassembled context package that is no longer part of the active interface;
- several tasks had strong lexical anchors and low architectural ambiguity;
- preparation dominated a portion of the measured cost.

It nevertheless identified a real failure mode: Grimoire can become an additional layer of work when a direct lookup already supplies nearly all useful evidence.

## Simple packet-trace follow-up

A later one-run packet trace allowed normal repository inspection alongside each assigned discovery system. All three agents produced the required call path and server-authority explanation, but plain shell inspection completed fastest and used the fewest noncached tokens. CBM and especially Grimoire added response and interaction cost without improving the final evidence.

This run had one trial per system and is not statistically strong. Its main value is qualitative: the task was a nearly linear trace with concrete symbol names, so ordinary search had little ambiguity to remove.

## Unfamiliar-repository follow-up

Report: [`evaluation/results/multi-repo-agent-benchmark-2026-07-27-v1/report.md`](../../evaluation/results/multi-repo-agent-benchmark-2026-07-27-v1/report.md)

A three-repository suite compared plain inspection, CBM, and Grimoire on HikariCP, Detekt, and Now in Android. All conditions used GPT-5.6 Sol with normal repository tools available; CBM and Grimoire were optional, mutually exclusive additions. HikariCP and Detekt ran concurrently within each task. Now in Android was resumed later and ran sequentially.

| Task | Plain | CBM | Grimoire | Outcome |
| --- | ---: | ---: | ---: | --- |
| HikariCP connection lifecycle | **6m 32s**, 12 calls | 16m 18s, 42 calls | 13m 19s, 26 calls | Plain won; the trace was too direct to benefit from codebase intelligence |
| Detekt rule-set plugin flow | **10m 03s**, 20 calls | 18m 22s, 57 calls | Invalid refusal after 5m 33s | Plain produced the strongest valid answer; Grimoire returned empty evidence |
| Now in Android topic-notification muting | **9m 34s**, 14 calls | 15m 06s, 27 calls | 9m 47s, 22 calls on repaired rerun; initial 2m 34s refusal retained separately | Plain remained strongest; repaired Grimoire completed but had extensive grounding errors |

The HikariCP result is an important negative result for mandatory Grimoire use. It was a compact repository with concrete lifecycle identifiers, so direct search found the relevant path with little ambiguity. A quality audit ranked plain first, Grimoire a valid close second, and CBM third. Plain had 178 valid prose citations and 35 evidence items with no invalid ranges; Grimoire had 71 valid citations and 33 evidence items with no invalid ranges; CBM had four out-of-range prose citations and three out-of-range evidence items.

The Detekt telemetry initially appeared to show a Grimoire speed win, but the answer audit invalidated it. Plain and CBM both produced implementation-grade traces. Plain had 150 valid prose citations across 62 files and 37 populated evidence items with no invalid ranges. CBM had 129 prose citations and 41 evidence items, with one citation extending one line beyond its file. Grimoire refused after discovery preparation failed, supplied no source citations, and left every evidence lane empty. Its 5m 33s is time-to-failure.

The first Now in Android Grimoire attempt was an invalid 2m 34s refusal caused by missing packaged Lexicon adapters, ambiguous MCP arguments, and Hermes treating ordinary tool errors as server failures. After fixing Grimoire's bundle-relative adapter discovery and mode-specific MCP schema, the rerun completed in 9m 47s with 22 model calls, 401,419 input tokens, and 16,645 output tokens. Lexicon, Arcana, and Grimoire were all prepared successfully.

The repaired answer was comprehensive but not implementation-grade. It populated all ten evidence lanes, yet 22 of 58 prose citations and 11 of 28 structured evidence items pointed to nonexistent paths or ranges beyond the pinned files. Plain remained strongest with 104 valid prose citations and 32 evidence items; CBM remained a close second with 80 prose citations, 42 evidence items, and two one-line range overruns. The repaired Grimoire run is useful evidence that infrastructure success and detailed output do not guarantee reliable grounding.

This task is also unusually favorable to a strong model using `rg`: the prompt names persistence, repositories, synchronization, notifications, two UI areas, DI, analytics, accessibility, and tests. It effectively supplies the search plan. The Grimoire condition then adds a larger body of related context, which may have created the small-task lost-in-the-middle failure seen in its stale and invented citations.

Together with the Space Rocks architecture result, the current working conclusion is:

- the multi-repository suite is biased toward a strong model with `rg` because the prompts expose most of the needed search vocabulary and subsystem checklist;
- discovery can create a lost-in-the-middle problem on compact, direct tasks by expanding an already sufficient working set;
- discovery can prevent lost-in-the-middle on broad, ambiguous tasks by compressing a working set that would otherwise sprawl across many searches and files;
- the Space Rocks network-interest result strongly favors Grimoire for the latter task class, while HikariCP favors plain inspection for the former;
- Detekt records a semantic Grimoire failure; the repaired Now in Android run records a completed but poorly grounded Grimoire answer;
- process completion, tool availability, citation validity, and answer usefulness must be scored separately;
- the correct agent behavior is to use Grimoire selectively and return to direct inspection once discovery is no longer the bottleneck.

## Interpretation by task class

| Task class | Expected starting strategy |
| --- | --- |
| Exact identifier, path, configuration key, or short named call chain | Shell search or Grimoire exact search; use whichever is immediately available |
| Unfamiliar repository with no concrete anchor | Grimoire `orient`, then narrow search |
| Cross-file or cross-language ownership question | Grimoire `search`, then handle-based inspection and trace |
| Transitive dependents or bounded impact | Grimoire `impact` from a verified handle |
| Source plus rationale or architecture documentation | Grimoire search with documents enabled, then verify source separately |
| Broad implementation plan spanning protocol, state, generation, and client application | Grimoire-assisted normal inspection |

Do not force agents to make a fixed number of Grimoire calls. Choosing not to use Grimoire for a cheap literal lookup is correct behavior. On small tasks, every additional result must justify the context it occupies; otherwise discovery may reduce answer quality even when retrieval itself is relevant.

## Benchmark requirements

A fair assisted-agent comparison should preserve:

- identical repository revisions and clean checkout state;
- identical task wording, model, completion criteria, and evidence contract;
- normal shell, Git, search, and file-reading tools in every condition;
- exactly one optional discovery system per assisted condition;
- the product's real installed skill rather than ad hoc prompt instructions;
- equivalent warm or cold state, reported explicitly;
- all setup, refresh, discovery, direct-read, model-call, token, and elapsed costs;
- validation of final citations and structured deliverables;
- separate reporting for preparation and agent investigation when both matter.

Parallel runs reduce wall-clock test time but introduce resource contention. Use per-process completion timestamps and record that the runs were concurrent.

## Product implications

The current evidence supports these operating rules:

1. Keep Grimoire progressive and handle-based; do not restore mandatory preassembled context packages.
2. Keep source, documentation, symbols, and relationships in independent lanes.
3. Keep initial requests narrow and responses bounded.
4. Reuse investigation sessions to avoid replaying evidence.
5. Avoid unnecessary state refreshes.
6. Let agents stop using Grimoire when direct inspection becomes cheaper.
7. Judge changes by end-to-end agent outcomes, not internal retrieval metrics alone.
8. Maintain benchmark coverage across lookup, focused trace, architecture, impact, and source-plus-document tasks.

## Reproduction artifacts

Current architecture benchmark:

- runner: `evaluation/run_network_interest_agent_benchmark.py`;
- report: `evaluation/results/network-interest-agent-benchmark-2026-07-27-v4/report.md`;
- machine summary: `evaluation/results/network-interest-agent-benchmark-2026-07-27-v4/summary.json`;
- raw outputs and usage files: the same result directory;
- Grimoire skill: `skills/grimoire/SKILL.md`.

Unfamiliar-repository benchmark:

- runners: `evaluation/run_multirepo_agent_benchmark.py` and `evaluation/run_nowinandroid_sequential.py`;
- report: `evaluation/results/multi-repo-agent-benchmark-2026-07-27-v1/report.md`;
- machine summary: `evaluation/results/multi-repo-agent-benchmark-2026-07-27-v1/summary.partial.json`;
- completed tasks: HikariCP connection lifecycle, Detekt rule-set plugin flow, and Now in Android topic-notification muting;
- Now in Android Grimoire conditions: initial unreachable-MCP refusal plus the repaired 9m 47s rerun and its citation audit.

Earlier benchmark:

- corpus: `evaluation/agent_discovery/space-rocks.v1.json`;
- report: `evaluation/results/space-rocks-grimoire-cbm-agent-benchmark-2026-07-27.md`;
- metrics: `evaluation/results/space-rocks-grimoire-cbm-agent-benchmark-2026-07-27.json`.

Every report is evidence only for its recorded conditions and date.

