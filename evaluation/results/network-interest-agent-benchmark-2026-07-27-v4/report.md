# Space Rocks Network-Interest Agent Benchmark

Date: 2026-07-27

## Setup

All three agents investigated separate clean detached worktrees at commit `460da4af05c44d1835401fa853f5fc6b718262c8` using GPT-5.6 Sol through OpenAI Codex. The runs were started in parallel.

- **Plain:** normal Hermes shell, Git, search, and file-reading tools; no CBM or Grimoire access and no discovery skill.
- **CBM:** the same normal tools plus CBM 0.9.0 and its exact embedded `codebase-memory/SKILL.md`; no Grimoire access.
- **Grimoire:** the same normal tools plus Grimoire and `skills/grimoire/SKILL.md`; no CBM access.

The task was a read-only implementation plan for per-client network interest. It required tracing camera visibility, receiver-specific world projection, baseline/delta membership, hysteresis, self and spectate behavior, a low-cadence offscreen-player locator path, budgets and prioritization, generated Go/GDScript contracts, client application, tests, failure modes, and an ordered implementation sequence.

The active `codex/network-interest` worktree was excluded. All three benchmark worktrees remained clean.

CBM was indexed from the clean detached checkout into an isolated cache as project `space-rocks-network-interest-clean-460da4af`. Searches for implementation-only names such as `player_locator`, `network_interest`, `PlayerLocator`, and `interestExitMargin` returned no results before the run.

Hermes could not consume CBM 0.9.0's paginated `tools/list` handshake directly. A transparent stdio compatibility proxy flattened only those tool-list pages; all actual CBM requests and responses were forwarded unchanged to the CBM 0.9.0 process.

## Results

| Condition | Technical rubric | Required evidence JSON | Discovery calls | Model API calls | Noncached tokens | Total tokens including cache | Actual completion* |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: |
| Plain | 10/10 | Pass | 0 | 53 | 471,723 | 7,404,203 | 16m 54s |
| CBM | 10/10 | **Fail** | 20 | 54 | 543,471 | 7,009,519 | 17m 31s |
| Grimoire | 10/10 | Pass | 2 | 22 | 261,631 | 3,052,543 | 9m 16s |

\* Completion time is based on each final output file's modification time relative to the common launch. The runner's stored elapsed value for later-collected parallel processes was not an independent completion timestamp.

Noncached tokens are input plus output tokens. Reasoning tokens are already represented within the provider's token accounting and are shown separately in `summary.json`.

## Technical grading

Each answer received one point for every required technical area:

1. Correctly finding and interpreting the wrapped camera-visibility seam.
2. Keeping authoritative simulation complete and filtering receiver-specific presentation/network projection instead.
3. Filtering before quantization, candidate construction, and encoding.
4. Using prior committed baseline membership so interest exits become reliable deletes and re-entry becomes a full create.
5. Applying distinct enter and exit margins for hysteresis.
6. Forcing the receiver and resolved spectate target to remain relevant.
7. Separating low-cadence locator data from full hot ship movement.
8. Addressing cadence, hard caps, budgets, chunking, and prioritization.
9. Tracing source-of-truth packet generation through Go and Godot outputs and client application.
10. Providing meaningful tests and failure modes.

All three plans covered all ten areas. They independently converged on the important architecture: immutable authoritative world state, recipient filtering at the realtime world-projection boundary, committed-baseline-driven hysteresis, lifecycle deletes on interest exit, forced self/view target, and a low-cadence locator source for distant-player indicators.

The active implementation eventually used a dedicated locator lane/packet. All three benchmark answers instead proposed carrying locator information through the existing reliable session lane. That is a defensible design, but it would need explicit encoded-size, cadence-isolation, and session-growth validation before adoption.

## Deliverable compliance

Plain and Grimoire ended with the exact required `BENCHMARK_EVIDENCE_JSON:` line, correct key names, and path/symbol/line/claim objects.

CBM ended with a different bare JSON object. Its technical prose was strong, but it did not follow the required evidence schema or prefix. This is the only material answer-quality failure in the final run.

## Efficiency

Compared with the plain baseline, Grimoire used:

- **44.5% fewer noncached tokens**.
- **58.5% fewer model API calls**.
- **58.8% fewer total tokens including cache reads**.
- About **45% less completion time** in the parallel run.

CBM compared with the plain baseline used:

- **15.2% more noncached tokens**.
- One additional model API call.
- **5.3% fewer total tokens including cache reads**.
- About 37 seconds more completion time.

Compared with Grimoire, CBM used 107.7% more noncached tokens and 145.5% more model API calls.

## Discovery behavior

The Grimoire agent loaded its skill, made two unified discovery calls, and then switched to targeted source reads and searches. This closely followed the new skill's guidance: narrow discovery, reuse returned evidence, and stop using Grimoire when direct inspection becomes cheaper.

The CBM agent loaded its skill and made 20 structural calls: one `list_projects`, two `index_status`, one `get_architecture`, and sixteen `search_graph` calls. It then performed extensive direct verification. Access to CBM did not reduce the model's overall search/verification workload in this task.

The plain agent reached an equally strong technical plan through exhaustive normal repository inspection, but at substantially greater token and API-call cost than Grimoire.

## Conclusion

**Grimoire won this benchmark.** It preserved full technical correctness and evidence compliance while using far fewer model calls and tokens and reaching its final answer much earlier.

The plain baseline remained highly capable but expensive. CBM produced a technically strong plan, but it neither improved efficiency over plain inspection nor followed the required evidence-output contract.
