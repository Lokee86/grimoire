# Recent changes — July 2026

This record summarizes the current product changes merged through `d75a81a` plus the agent-skill, benchmark, installation, and documentation work prepared afterward. It is a change summary, not a substitute for reference documentation.

## Unified progressive discovery

Grimoire now owns the normal repository-investigation interface across:

- exact source matches;
- BM25 source matches;
- separately ranked documentation;
- Lexicon declarations and definitions;
- Arcana and Lexicon relationships;
- handle-based inspect, trace, and impact operations.

Consumers no longer choose a provider. Grimoire routes each operation and preserves provider provenance in independent lanes.

The former answer-shaped context-package compiler, query-shape policy, evidence assembly, curation, package fitting, diff-context, graph-ranking, and package-oriented source evaluator were removed from the active CLI and MCP path. Historical reports remain only as calibration evidence.

## Lexical-first source discovery

Prepared source state now includes a persistent identifier-aware lexical index. Source discovery uses postings to localize candidate chunks before BM25 scoring rather than rescanning the entire prepared corpus for each query.

Recent retrieval work added:

- file-level discovery scopes;
- source search restricted to bounded path sets;
- declaration vocabulary and bounded lexical aliases;
- compact source excerpts in search results;
- independent per-lane limits;
- direct relationship results alongside exact, source, document, and symbol lanes.

## Lexicon and Arcana integration

Lexicon and Arcana are consolidated into the Grimoire repository while retaining independent binaries, state, protocols, tests, and specialist use.

Grimoire now exposes their commands through namespaces:

```bash
grimoire lexicon <command>
grimoire arcana <command>
```

`grimoire lexicon check` and `grimoire arcana check` report provider resolution. Other arguments are forwarded to the owning executable with process I/O and exit status preserved.

Arcana gained repository-graph protocol export and Grimoire consumes direct graph relationships through the unified discovery interface. Lexicon facts remain the fallback relationship source when Arcana is unavailable.

## Repository state and runtime packaging

The root workflow now builds and packages:

- Grimoire, Lexicon, and Arcana executables;
- the Lodestone native library;
- Lexicon runtime adapters;
- the canonical Grimoire agent skill.

Lexicon can discover packaged adapters beside the installed executable or in the combined bundle's sibling `adapters/` directory. Repository state supports explicit paths and environment/configuration overrides rather than requiring callers to invoke providers from their source checkouts. The Grimoire MCP schema now also exposes mode-specific argument requirements so invalid calls such as file-valued `inspect target` requests can be rejected before execution.

The root workflow remains CPU-bounded by default across Go and Cargo. Release archives are deterministic and the GitHub release workflow publishes Windows x86_64 and Linux x86_64 combined bundles. The embedded bundle installer now copies Lexicon runtime adapters, and a release-consumer smoke packages a real combined ZIP, installs it into a clean temporary location, launches the installed MCP server, prepares managed provider state, and verifies opaque handles through inspect and trace.

## Lodestone extraction

The native vector engine moved into the separate Lodestone repository. Grimoire retains the Go compatibility boundary and packages the Lodestone native library with Grimoire releases.

Optional vectors remain limited to the documentation lane and Arcana semantic graph entry points. Exact source, BM25 source, symbols, relationships, trace, and impact do not require repository-wide source embeddings.

## Agent skill and installation

The canonical agent skill now lives at:

```text
skills/grimoire/SKILL.md
```

Builds and release archives include it. Installing Grimoire writes it by default to:

```text
~/.agents/skills/grimoire/SKILL.md
~/.hermes/skills/grimoire/SKILL.md
```

Installers support repeated `--skills-dir` destinations and `--skip-skills`.

The skill encodes the measured efficient workflow:

- use Grimoire alongside normal shell and file inspection;
- start concrete tasks with narrow search;
- reuse one investigation session;
- keep result limits small;
- omit documentation for code-only work;
- follow handles instead of repeating broad searches;
- avoid unnecessary refreshes;
- verify source and stop querying when direct inspection is cheaper.

## Agent benchmark results

The current benchmark evidence separates two questions: whether codebase intelligence is useful for a task, and which system performs better when it is useful.

- The current multi-repository suite is specifically favorable to GPT-5.6 Sol with `rg`: its prompts name the subsystem and enumerate the expected ownership areas, supplying much of the search plan in advance.
- On compact, direct tasks, additional discovery context can create a lost-in-the-middle problem by expanding an already sufficient working set.
- On broad, cross-system architecture tasks, structured discovery can prevent lost-in-the-middle by compressing and organizing a working set that would otherwise sprawl across many searches.
- The strongest valid assisted result remains the Space Rocks network-interest benchmark. HikariCP favored plain inspection. Detekt's Grimoire run failed semantically. The repaired Now in Android rerun completed but had extensive grounding and citation errors.

In the final Space Rocks network-interest benchmark, Grimoire preserved a 10/10 technical plan and evidence compliance while completing in 9m 16s, compared with 16m 54s for plain inspection and 17m 31s for CBM assistance. It used 22 model API calls and 261,631 noncached tokens, compared with 53 calls and 471,723 tokens for plain inspection.

A later unfamiliar-repository follow-up produced three completed tasks:

- HikariCP connection lifecycle: plain 6m 32s, Grimoire 13m 19s, CBM 16m 18s;
- Detekt rule-set plugin flow: plain produced the strongest valid answer in 10m 03s; CBM produced a strong answer in 18m 22s; Grimoire refused after 5m 33s with empty evidence;
- Now in Android topic-notification muting: plain produced the strongest answer in 9m 34s; CBM produced a strong answer in 15m 06s; the initial Grimoire attempt refused after 2m 34s, while the repaired rerun completed in 9m 47s but ranked third because 22 of 58 prose citations and 11 of 28 evidence items were invalid.

The HikariCP result confirms that discovery can be counterproductive when direct search already yields a small, obvious working set. The Now in Android rerun strengthens that warning: infrastructure and schema fixes restored a complete Grimoire answer, but the extra related context did not improve grounding and may have contributed to a small-task lost-in-the-middle failure. Conversely, the Space Rocks network-interest result shows the value of discovery when it compresses a genuinely large, ambiguous, cross-language investigation. Benchmark reporting now distinguishes process completion, tool health, citation validity, and practical answer quality.

See [Agent benchmark findings](agent-benchmark-findings.md) for methodology, limitations, and raw report links.

## Benchmark hardening and preparation telemetry

The next agent benchmark suite is defined in `evaluation/agent_benchmark_tasks.v2.json`. Its room-scale architecture, Detekt unclear-ownership, Space Rocks cross-language locator, and LevelDB impact-analysis tasks have now been executed; only the source-plus-rationale task remains. It replaces prompt-supplied subsystem checklists with natural problem reports and hidden rubrics across five intended task classes.

Benchmark grounding is now automatic. Every inline and structured path/range is checked against the pinned checkout, refusals and empty evidence are invalid, and Grimoire evidence handles are checked against exact inspected source ranges through an optional `grimoire.mcp.audit.v1` log, while handle coverage remains a metric rather than a mandatory tool-usage quota. Summaries distinguish process completion from grounded validity.

Repository preparation status now separates initial inspection, lock wait, reinspection, Lexicon, Arcana, source indexing, documentation indexing, marker writes, final source verification, and total time. The refresh path reuses one post-lock source fingerprint through provider preparation instead of repeatedly walking the repository after each action, then performs one final fingerprint check before marking state current.

Search and orient default to six results per lane, discovery excerpts are capped at 700 bytes, and duplicate exact/lexical source ranges preserve both independent lane entries while replacing the repeated lexical excerpt with a `duplicate_of` handle reference.

The first version 2 architecture run produced grounded implementation-grade answers in all three conditions. Grimoire completed in 6m 36s with 16 calls and 1.05M total tokens, versus CBM at 11m 33s/25 calls/3.30M tokens and plain at 12m 27s/18 calls/1.84M tokens. Grimoire cold preparation took 95.2 seconds, dominated by 69.3 seconds of Lexicon work. The run also exposed and fixed validator support for noncontiguous structured ranges; saved outputs can now be revalidated without rerunning agents.

The second version 2 Detekt run also produced grounded implementation-grade answers in all three conditions, but plain inspection won: 5m 04s, 14 calls, and 767k total tokens. CBM was effectively tied at 5m 05s including preparation, 16 calls, and 831k tokens. Grimoire took 6m 26s including preparation, 16 calls, and 1.46M tokens. All three found the same Gradle-to-CLI plugin-classloader defect and proposed the same adapter-layer repair. Grimoire's single 44.9 KB search response began with low-value import nodes, confirming that broad response shaping can still add cost after the task's working set becomes compact.

The Detekt run also exposed two post-completion harness defects: relative output paths placed subprocess telemetry inside detached worktrees, and cleanup referenced `shutil` without importing it. Output paths are now absolute before launch, cleanup imports are complete, and isolated result roots can be imported into the canonical summary without rerunning agents.

The third version 2 locator run favored plain inspection and CBM over Grimoire. Plain completed in 5m 27s with 21 calls and 2.18M total tokens; CBM completed in 6m 37s including preparation with 25 calls and 2.40M tokens; Grimoire completed in 9m 41s including preparation with 21 calls and 2.38M tokens. Plain and CBM chose a dedicated unreliable locator packet, while Grimoire reused the reliable overlay lane for changing coordinates and omitted stale expiry. All three missed durable per-player color metadata and complete spectating verification despite passing citation grounding.

That run also exposed summary replacement when running one selected task into an existing canonical output. The runner now preserves compatible existing task entries, records the latest run start separately, and rejects incompatible suite/model/provider metadata. Summary-only recovery can merge a selected task back onto a committed base without rerunning agents.

The fourth version 2 LevelDB run was won by CBM: 4m 52s including preparation, 10 calls, and 495k total tokens. Plain produced a strong valid answer in 6m 43s with 12 calls and 598k tokens. Grimoire took 6m 41s including preparation, 16 calls, and 1.57M tokens, then failed grounding because three model-authored ranges did not match their canonical handles.

LevelDB also exposed a provider-compatibility failure. Lexicon completed in 6.9 seconds, but Arcana rejected the C/C++ snapshot as malformed with `unresolved reason`; Grimoire continued in degraded source/lexical mode. Four discovery calls returned 138 KB of structured context. This is now documented as both a C/C++ Lexicon-to-Arcana integration defect and a degraded-mode response-shaping problem.

## Documentation and verification

The README now starts with release installation, first use, and agent configuration rather than developer build internals. New documentation covers:

- release and source installation;
- PATH and provider verification;
- skill roots;
- generic stdio MCP configuration;
- first-run state preparation;
- efficient request patterns;
- task-shape guidance;
- benchmark controls and conclusions;
- current roadmap after retirement of the package pipeline.

The packaging smoke suite verifies skill inclusion and installation.

## Reference points

- [README](../../README.md)
- [Installation and agent setup](../reference/installation.md)
- [Agent and MCP guide](../reference/agent-mcp.md)
- [Unified discovery contract](../reference/agent-query.md)
- [Release workflow](release-workflow.md)
- [Testing and benchmarks](testing-and-benchmarks.md)
- [Agent benchmark findings](agent-benchmark-findings.md)
- [Roadmap](../planning/roadmap.md)
