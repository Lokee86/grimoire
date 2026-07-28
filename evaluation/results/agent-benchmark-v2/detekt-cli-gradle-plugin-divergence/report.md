# Detekt CLI / Gradle plugin divergence benchmark

Date: 2026-07-28

Task category: unclear ownership

Repository: `corpus/kotlin-detekt`

Pinned revision: `f9e1d5cc239ab740ce499b1edb36b872012648e2`

Model/provider: GPT-5.6 Sol through OpenAI Codex

Conditions ran concurrently in separate clean detached worktrees. Every condition had normal shell, Git, and direct source inspection. CBM and Grimoire were optional, mutually exclusive additions.

## Task

> A third-party rule-set JAR works when detekt is run from the command line, but under the Gradle plugin it is missing or behaves as though it was configured differently. Determine where responsibility for the divergence most likely lives, identify the smallest correct fix boundary, and provide a verification plan.

The prompt did not disclose the hidden ownership checklist or expected files.

## Result summary

| Condition | Preparation | Agent elapsed | Total elapsed | Model calls | Total tokens | Answer bytes | Grounding |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Plain | none | **5m 04s** | **5m 04s** | **14** | **767,334** | 16,747 | Pass |
| CBM | **5.3s** | 5m 00s | 5m 05s | 16 | 831,264 | 15,022 | Pass |
| Grimoire | 31.8s | 5m 54s | 6m 26s | 16 | 1,457,421 | 16,043 | Pass |

All three conditions completed normally, supplied valid evidence JSON, and covered the hidden `detekt-cli/`, `detekt-gradle-plugin/`, and `detekt-core/` path families.

Grounding details:

| Condition | Checked inline citations | Structured evidence items | Findings |
| --- | ---: | ---: | --- |
| Plain | 39 | 21 | none |
| CBM | 32 | 15 | none |
| Grimoire | 47 | 12 | none |

None of the answers used canonical Grimoire handles.

## Shared technical conclusion

All three independently identified the same implemented divergence:

1. The standalone CLI accepts `--plugins` and transfers those JAR paths into the core extension specification.
2. Core creates a dedicated plugin classloader and uses it for `ServiceLoader` provider discovery and plugin default configuration.
3. The Gradle plugin resolves `detektPlugins` into `pluginClasspath`, but its actual CLI argument list does not include `--plugins`.
4. Instead, Gradle combines `pluginClasspath` with detekt's own execution classpath for both the direct invoker and Worker API paths.
5. The debug-only reproduction command prints `--plugins`, so the printed command does not match the real invocation.
6. The direct invoker's global classloader cache is keyed by classpath file paths rather than JAR contents. Replacing a plugin JAR at the same path can therefore retain stale classes and `config/config.yml` resources in a Gradle daemon while a fresh CLI process observes the new artifact.

All three proposed the same smallest correct boundary:

- add one canonical Gradle-side plugin CLI argument using `File.pathSeparator`;
- include it in analysis, baseline creation, and config generation;
- invoke detekt with `detektClasspath` only rather than co-loading plugin JARs into the host CLI classloader;
- make the debug command print the actual argument list;
- verify direct and Worker API paths with the same real ServiceLoader/config fixture, including same-path JAR replacement without stopping the daemon.

No answer proposed changing detekt-core's provider discovery or the third-party extension API.

## Qualitative comparison

### Plain

Plain inspection produced the strongest overall answer by a small margin. It gave the most complete evidence set and the clearest causal chain from Gradle task input, through the missing CLI argument, to the path-keyed classloader cache and stale plugin configuration. Its verification plan covered all three Gradle task types, same-path replacement, embedded configuration, both invocation modes, and classpath isolation.

### CBM

CBM produced a close second-place answer. It was slightly shorter, added a useful explicit version-skew check, and included a dependency-conflict regression case. It reached essentially the same fix and verification plan as plain inspection, but with two additional model calls and slightly more total context.

### Grimoire

Grimoire was technically correct and well grounded, but it did not add a unique ownership finding or safer fix. It emphasized the same-path classloader-cache defect and correctly rejected a cache-key-only repair in favor of restoring the explicit plugin contract. Its answer was comparable to CBM, but the investigation was slower and substantially more expensive.

All three covered the central hidden rubric dimensions: separate entry paths, classloader/process ownership, the convergence boundary, and parity verification. None exhaustively enumerated every possible malformed-service, duplicate-provider, incompatible-plugin, and propagated-error combination, so the broad failure-mode dimension remains only partially exercised by this task result.

## Efficiency interpretation

Compared with plain inspection, Grimoire used:

- 2 more model calls;
- 690,087 more total tokens, about 90% more;
- 50 seconds more agent time;
- about 82 seconds more total time after cold preparation.

Compared with CBM, Grimoire used:

- the same number of model calls;
- 626,157 more total tokens, about 75% more;
- about 81 seconds more total time including preparation.

CBM and plain were effectively tied on wall time. CBM's 5.3-second preparation offset its four-second faster agent completion.

This is a negative result for mandatory Grimoire use. Although the prompt did not provide a subsystem checklist, it supplied a concrete divergence between two named entry points. Direct search could rapidly locate `--plugins`, `detektPlugins`, `pluginClasspath`, the invocation classpath, and the classloader cache. The working set became compact early enough that structured discovery did not improve the final plan.

## Grimoire behavior

The Grimoire agent made one search request:

- query: `third party ruleset plugin classpath config CLI Gradle plugin Detekt task`;
- response size: 44,867 bytes;
- 26 nodes;
- 21 source ranges;
- 8 documents;
- 8 graph paths;
- no tool errors.

The response began with low-value import nodes before the useful ownership evidence. The agent then switched to direct source inspection, which was appropriate, but the initial response had already introduced substantial context. This result strengthens the case for tighter broad-search shaping and suppression of weak structural-adjacent results without introducing a global relevance score or preassembled context package.

## Preparation breakdown

CBM indexing: 5.254 seconds.

Grimoire cold preparation: 31.788 seconds externally, 31.211 seconds in internal timing buckets:

| Stage | Time |
| --- | ---: |
| Lexicon | 18.886s |
| Documentation | 7.346s |
| Reinspection | 2.188s |
| Source index | 1.771s |
| Arcana | 0.542s |
| Initial inspection | 0.284s |
| Final verification | 0.183s |
| Marker writes | 0.003s |

Lexicon remained the largest preparation component, but documentation preparation was proportionally significant on this repository.

## Harness findings

The run exposed two post-completion harness defects, neither of which affected the agent outputs:

1. A relative `--output` path was forwarded to subprocesses running from detached checkouts, causing usage files and the MCP audit log to be written inside those worktrees.
2. Successful cleanup referenced `shutil` without importing it, so the runner exited after writing the completed summary.

The output path is now resolved to an absolute path before launching subprocesses, `shutil` is imported, and the displaced artifacts were imported into the canonical result set without rerunning agents. The combined v2 summary retains both completed tasks.

## Retained artifacts

- `plain.stdout.txt`, `cbm.stdout.txt`, `grimoire.stdout.txt`: full unedited answers;
- matching `.usage.json`: model telemetry;
- matching `.grounding.json`: automatic grounding reports;
- `grimoire.mcp-audit.jsonl`: exact Grimoire request and structured response;
- `grimoire-prewarm.stdout.txt`: preparation state and timing breakdown;
- CBM indexing stdout/stderr;
- `evaluation/results/agent-benchmark-v2/summary.json`: combined machine-readable suite summary.

Transient CBM databases and detached benchmark worktrees are not part of the retained evidence.
