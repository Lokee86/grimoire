# Multi-repository agent benchmark — 2026-07-27

This report records three unfamiliar-repository tasks comparing plain inspection, CBM 0.9.0, and the current unified Grimoire build.

## Conditions

Every completed run used GPT-5.6 Sol through the `openai-codex` provider with high reasoning effort:

- **Plain:** normal shell, Git, search, and file inspection;
- **CBM:** normal tools plus CBM 0.9.0 and its installed skill;
- **Grimoire:** normal tools plus the current unified Grimoire build and `skills/grimoire/SKILL.md`.

The assisted conditions could not access the competing discovery system. Each condition used an isolated clean checkout at the recorded commit. HikariCP and Detekt ran their three conditions concurrently within each task. Now in Android was resumed later and ran sequentially: plain, CBM, then Grimoire. Reported agent time excludes separately recorded discovery-state preparation.

The suite has one trial per condition and task. The figures are end-to-end agent telemetry, not statistically stable product guarantees.

## Benchmark suitability warning

This suite is retained because negative and awkward results are still evidence, but it is not a good neutral benchmark of repository discovery. The prompts name the subsystem, enumerate the expected ownership areas, and provide enough concrete vocabulary for GPT-5.6 Sol to turn the task into a sequence of `rg` searches and direct reads. That specifically favors a strong model with ordinary shell discovery while forcing assisted conditions to pay for additional context that may not reduce ambiguity.

The results are consistent with a task-size inversion:

- on compact, direct, strongly named tasks, discovery output can expand an already sufficient working set and create a lost-in-the-middle problem;
- on broad, cross-cutting, architecturally ambiguous tasks, discovery can compress the working set and help prevent the same problem.

This is an interpretation of the observed answer patterns, not a controlled causal measurement of context position.

## HikariCP connection lifecycle

Repository commit: `a4d93f4f85517f90e632b795486d7102e933d7ff`

The agent traced connection acquisition, borrowing, proxy creation, close and recycle behavior, broken-connection handling, timeout and suspension paths, metrics, leak detection, maintenance, shutdown, and relevant tests.

| Condition | Time | API calls | Input tokens | Output tokens |
| --- | ---: | ---: | ---: | ---: |
| Plain | **6m 32s** | **12** | **110,392** | **14,763** |
| Grimoire | 13m 19s | 26 | 736,194 | 26,829 |
| CBM | 16m 18s | 42 | 579,920 | 24,718 |

Preparation excluded from agent time:

- CBM index: 0.9 seconds;
- Grimoire prewarm: 3.1 seconds.

Plain inspection won decisively. The repository is compact, the lifecycle has strong concrete identifiers, and the requested path is comparatively linear. Grimoire finished before CBM and used 38% fewer calls, but it still performed substantially more work than plain inspection and used 27% more input tokens than CBM.

A later answer-quality audit found that all three agents produced substantive lifecycle traces, but they were not equal. Plain was strongest: 178 valid prose citations across 26 files and 35 populated evidence items with no invalid paths or ranges. Grimoire was a valid close second: 71 valid prose citations and 33 populated evidence items with no invalid paths or ranges, though it omitted some of plain's build-time proxy-generation detail and had a narrower test inventory. CBM was broadly correct but had weaker evidence hygiene: four prose citations and three structured evidence items referenced ranges beyond the pinned files, including leak-task and metrics-test citations.

This result reinforces the earlier simple packet-trace observation: structured repository intelligence can be pure overhead when ordinary search already exposes the complete path cheaply.

## Detekt rule-set plugin flow

Repository commit: `f9e1d5cc239ab740ce499b1edb36b872012648e2`

The agent traced custom rule-set discovery through service registration, provider and rule factories, configuration and activation, analysis execution, findings and reporting, CLI and Gradle integration, documentation, tests, and failure modes.

| Condition | Time | API calls | Input tokens | Output tokens |
| --- | ---: | ---: | ---: | ---: |
| Plain | **10m 03s** | **20** | **170,965** | 18,523 |
| CBM | 18m 22s | 57 | 432,184 | 26,328 |
| Grimoire | **Invalid refusal after 5m 33s** | 19 | 176,003 | **4,599** |

Preparation excluded from agent time:

- CBM index: 6.6 seconds;
- Grimoire prewarm: 18.0 seconds.

The plain and CBM agents both produced strong, implementation-grade traces. Plain was strongest overall: 150 valid prose citations across 62 files and 37 populated evidence items with no missing paths or out-of-range citations. It covered the CLI and Gradle convergence, loader ownership, provider and rule construction, configuration validation, analyzer concurrency, suppression and baseline ordering, reporting, failure propagation, subtle loader-lifetime issues, duplicate-provider behavior, and a focused cross-entry test plan. CBM was close in substantive coverage, with 129 prose citations and 41 evidence items; one CLI citation extended one line beyond the pinned file.

The Grimoire agent did not produce a trace. It reported that prepared discovery returned unrelated files, Lexicon preparation lacked its adapter root, Arcana was absent, and no relevant handles could be obtained. Despite normal shell and file discovery being allowed by the benchmark prompt, it refused rather than falling back to direct inspection. It supplied zero source citations and left all nine evidence lanes empty. The recorded 5m 33s is time-to-semantic-failure, not a successful completion.

## Now in Android topic-notification muting

Repository commit: `7d45eae4f8720a0c77f507712ba2437ff974b6ed`

The agent designed a cross-module implementation plan covering user-data persistence, independent follow and mute state, synchronization and notification filtering, topic and settings UI, demo and production wiring, analytics, accessibility, migrations, and tests.

| Condition | Time | API calls | Input tokens | Output tokens |
| --- | ---: | ---: | ---: | ---: |
| Plain | **9m 34s** | **14** | **137,640** | 20,739 |
| Grimoire repaired rerun | 9m 47s | 22 | 401,419 | **16,645** |
| CBM | 15m 06s | 27 | 753,798 | 25,708 |
| Grimoire initial attempt | **Invalid refusal after 2m 34s** | 14 | 166,977 | 4,191 |

Preparation excluded from agent time:

- CBM index: 1.5 seconds;
- initial Grimoire prewarm: 9.4 seconds;
- repaired Grimoire prewarm: 17.9 seconds, including successful Lexicon and Arcana preparation.

The plain and CBM agents both produced substantial implementation plans. A citation audit found that plain supplied 104 prose citations and 32 structured evidence items with no missing paths or invalid line ranges. CBM supplied 80 prose citations and 42 evidence items; two build-file ranges exceeded the files by one line. Both answers correctly identified the local Proto DataStore ownership, the unchanged For You follow-state invariant, `OfflineFirstNewsRepository.syncWith` as the notification-policy boundary, genuine-newness and first-sync concerns, independent topic and settings UI state, demo/prod behavior, and the required test matrix.

Plain was the strongest valid answer overall. It had the cleanest citation contract and a complete implementation sequence, although it broadened the proposal into adjacent synchronization and cross-reference cleanup. CBM was also implementation-grade and caught an important persistence detail that plain did not emphasize: the existing follow setters swallow `IOException`, so analytics must not claim success unless the new mute mutation reports or propagates persistence failure.

The initial Grimoire process did not produce an implementation plan. After 2m 34s it reported that the Grimoire MCP server was unreachable, returned zero evidence items, and asked for a retry. Its 14 calls and token usage are time-to-semantic-failure. The process exit code and usage file described the run as completed only because Hermes exited normally after emitting the refusal.

The failure chain was subsequently traced to two Grimoire-side issues and one Hermes issue. The benchmark bundle placed Lexicon adapters at `adapters/` while the executable initially searched only `bin/adapters/`; the MCP schema also allowed the model to send `inspect` with a file-valued `target` even though inspect required handles or an anchor. Three ordinary validation errors then triggered Hermes's server-level circuit breaker. Hermes was not changed, but Grimoire was updated to discover bundle-sibling adapters and expose mode-specific MCP argument constraints.

The repaired rerun prepared Lexicon, Arcana, and Grimoire successfully and produced a complete plan in 9m 47s. It populated all ten evidence lanes and did not refuse. However, a mechanical audit found 22 invalid citations among 58 prose citations and 11 invalid items among 28 structured evidence items. Problems included nonexistent or stale paths such as `GetUserNewsResourcesUseCase.kt`, `SettingsScreen.kt`, backup XML files, `TopicAnalytics.kt`, and `SystemTrayNotifierTest.kt`, plus several ranges beyond file ends. The answer was detailed and broadly plausible, but not reliably grounded enough to be implementation-grade.

Plain therefore remained the strongest answer, CBM remained a close second, and the repaired Grimoire answer ranked third. The task prompt itself is highly `rg`-friendly: it names every major layer and feature area to inspect. The extra related context supplied through discovery appears to have enlarged the model's working set without resolving much uncertainty, which is consistent with a small-task lost-in-the-middle failure.

A still earlier setup attempt had failed before model execution because the long detached-checkout path exceeded Windows filename limits. The short worktree fixed that unrelated harness problem.

## Current interpretation

The suite primarily measures how effectively a strong model can exploit prompt-provided vocabulary with `rg`; it does not isolate the value of discovery tooling under realistic uncertainty. The tasks disclose most of the expected components and ask the model to walk a checklist, so ordinary search begins with unusually good anchors.

The observed pattern is better described as a task-size inversion:

1. **Compact and direct:** plain inspection keeps a small working set. Additional discovery context can create competition among related results and produce lost-in-the-middle grounding failures.
2. **Broad and ambiguous:** repeated direct searching creates a large, fragmented working set. Structured discovery can reduce and organize that context, preventing the model from losing important relationships.

The Space Rocks network-interest benchmark is evidence for the second class. HikariCP is evidence for the first. Detekt records a semantic Grimoire failure. Now in Android records both an initial infrastructure-driven refusal and a repaired, completed answer whose extensive invalid citations made it weaker than plain and CBM.

This remains preliminary. There is one trial per condition, and context-position effects were not independently controlled. Nevertheless, final answer quality must be audited separately from process exit, tool availability, latency, and token counts.

The practical product rule is not to force Grimoire into every repository question. Use ordinary inspection when prompt vocabulary and direct search already reduce the problem to a narrow path. Use Grimoire when it materially reduces a broad or ambiguous search space, then return to exact source reads for proof.

## Artifacts

The result directory contains:

- `summary.partial.json` — commits, preparation time, completion time, and usage telemetry;
- per-condition `stdout.txt`, `stderr.txt`, and `usage.json` files;
- CBM indexing and Grimoire prewarm logs.

The benchmark runners are:

- `evaluation/run_multirepo_agent_benchmark.py` for the original task suite;
- `evaluation/run_nowinandroid_sequential.py` for the resumed sequential Now in Android runs.

Raw outputs and telemetry remain the source of truth for exact model usage and completion data.