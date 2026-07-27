# Space Rocks Grimoire vs CBM Agent Benchmark — 2026-07-27

## Decision

Grimoire and CBM both completed every scored repository-investigation task correctly: **12/12 each**.

Grimoire did not produce a task-quality advantage large enough to justify its current operating cost. On this corpus, CBM is the practical efficiency winner:

- CBM prepared the repository in about **6.3 seconds**; Grimoire required about **192 seconds** for source, Lexicon, Arcana, and context state.
- CBM used a median **105,272 noncached tokens** per trial versus **143,515** for Grimoire.
- CBM's median end-to-end time was **338.1 seconds** versus **360.8 seconds** for Grimoire.
- Median model tool calls were effectively tied: **29.5 CBM** versus **30.5 Grimoire**.
- Grimoire recovered slightly more of the exact judged evidence: **52/54** required items versus **50/54** for CBM. Both recovered **2/3** judged structural items.

The lexical-first architecture should remain. Further global ranking and owner-retrieval tuning is not justified by these results. Arcana and Lexicon remain useful as explicit structural inspection systems, but the benchmark does not show that their default use gives Grimoire better final answers than CBM.

## What was benchmarked

The benchmark measured whether the same agent could correctly investigate realistic repository questions using Grimoire or CBM as its only repository-intelligence system.

Four Space Rocks tasks were used:

1. Trace the Godot Discord login-session flow into Rails.
2. Trace a Godot gameplay input packet into the authoritative Go handler.
3. Map the complete respawn and match-classification lifecycle.
4. Find every production reader of `SPACE_ROCKS_LOCAL_SERVER_PORT` and explain its endpoint responsibility.

Each task ran three independent times per system: **4 tasks × 2 systems × 3 trials = 24 effective trials**.

Controls:

- Repository pinned to Space Rocks revision `ff882f636706e4917f86e156ce1ed7f40b467e83`.
- Same Hermes model for every trial: `gpt-5.6-sol` through OpenAI Codex.
- Read-only investigation.
- Maximum 55 discovery operations.
- No `rg`, `grep`, recursive listing, Git search, IDE search, or competing intelligence system.
- Direct source reads were allowed only after the assigned system surfaced the path.
- Every answer required line citations and machine-readable evidence.
- Noncached tokens are Hermes input plus output tokens; cache-read tokens are reported separately in the JSON artifact.

Task correctness and exact judged-evidence recovery were scored separately. An answer could complete the task correctly through equivalent current-code evidence without naming every preferred judgment symbol.

Elapsed agent time is directionally useful rather than a strict microbenchmark because trials ran concurrently and include model/API latency.

## Preparation cost

| System | Preparation | Result |
|---|---:|---|
| CBM 0.9.0 | 6.3 s | Repository indexed; 26,537 nodes and 97,202 edges reported. |
| Grimoire source index | 35.6 s | Lexical source state prepared. |
| Grimoire Lexicon/Arcana/context refresh | 156.7 s | Deterministic lexical-first structural path ready. |
| Grimoire total | ~192.3 s | Approximately 30× CBM preparation time. |

Vector retrieval was not used. This benchmark exercised Grimoire's deterministic lexical-first context path plus scoped Lexicon/Arcana inspection and direct global query operations.

## End-to-end agent results

| System | Task | Correct | Exact required evidence | Median time | Median tool calls | Median noncached tokens |
|---|---|---:|---:|---:|---:|---:|
| Grimoire | Auth | 3/3 | 7/9 | 335.8 s | 32 | 122,705 |
| CBM | Auth | 3/3 | 7/9 | 390.6 s | 47 | 140,842 |
| Grimoire | Gameplay packet | 3/3 | 15/15 | 339.5 s | 29 | 164,325 |
| CBM | Gameplay packet | 3/3 | 15/15 | 298.2 s | 14 | 90,822 |
| Grimoire | Respawn lifecycle | 3/3 | 18/18 | 595.8 s | 49 | 273,408 |
| CBM | Respawn lifecycle | 3/3 | 16/18 | 340.0 s | 49 | 146,090 |
| Grimoire | Config readers | 3/3 | 12/12 | 305.6 s | 17 | 96,601 |
| CBM | Config readers | 3/3 | 12/12 | 298.9 s | 23 | 59,556 |

### Overall

| System | Correct | Exact required evidence | Structural evidence | Median time | Median tool calls | Median noncached tokens |
|---|---:|---:|---:|---:|---:|---:|
| Grimoire | **12/12** | **52/54** | 2/3 | 360.8 s | 30.5 | 143,515 |
| CBM | **12/12** | 50/54 | 2/3 | **338.1 s** | **29.5** | **105,272** |

The aggregate hides task-specific differences:

- Grimoire was stronger on the auth trace: fewer calls, fewer tokens, and lower elapsed time.
- CBM was substantially more efficient on gameplay-packet and respawn tracing.
- Config-reader correctness was identical; CBM used fewer tokens, while Grimoire used fewer tool calls.
- Grimoire's slight exact-evidence advantage did not translate into better task success.

## Initial Grimoire context package

Before interactive follow-up queries, each Grimoire trial began with one 12,000-token lexical-scoped context package.

| Task | Package time | Exact required evidence | Structural evidence | Complete from package alone |
|---|---:|---:|---:|---:|
| Auth | 17.7 s | 2/3 | 0/1 | No |
| Gameplay packet | 25.3 s | 3/5 | — | No |
| Respawn lifecycle | 23.4 s | 1/6 | — | No |
| Config readers | 24.2 s | 4/4 | — | Yes |

The initial package completed only **1/4 tasks**, recovering **10/18 required evidence items** and **0/1 structural items**. Interactive Grimoire queries recovered the missing evidence and reached 12/12 final task success, but the one-shot package is not yet a reliable substitute for investigation.

CBM has no directly equivalent preassembled context-package operation, so this section is diagnostic rather than a head-to-head package comparison.

## Ground-truth corrections found by the benchmark

The benchmark exposed stale judgments in its own corpus. They were corrected before final scoring:

- `client/scripts/boot/local_alpha_profile_smoke.gd` is a third production Godot reader of `SPACE_ROCKS_LOCAL_SERVER_PORT`.
- `services/game-server/cmd/game-server/listen_address_localpackage.go` is a build-tagged Go reader of the same variable.
- The active gameplay-packet producer is `Player.get_input_packet` using the generated `Packets.input_packet` factory. The prior judgment named an unused wrapper as the production owner.

One Grimoire respawn trial was also replaced because its shortened benchmark prompt accidentally omitted the executable path. The failed harness run was excluded; the corrected replacement trial is included in the 24 effective runs.

## Interpretation

The architectural correction was still worthwhile. Grimoire now has a coherent separation:

- lexical search discovers likely source areas;
- Lexicon resolves symbols inside those areas;
- Arcana inspects relationships from resolved symbols;
- explicit global graph queries remain available.

What the benchmark does **not** show is an end-to-end advantage over a simpler competitor. Both systems reached perfect task success, while CBM required much less preparation and materially fewer tokens overall. Grimoire's graph-backed path produced a small exact-evidence gain, but not a correctness gain.

The current recommendation is:

1. Keep lexical-first discovery and the separate explicit graph-query surface.
2. Stop tuning global deterministic ranking and mechanism-owner heuristics.
3. Treat the current result as the regression baseline.
4. Improve Grimoire only where the benchmark shows concrete cost: preparation time, context-package completeness, scoped structural latency, and token consumption.
5. Require a future change to improve end-to-end task outcomes or efficiency, not merely an internal retrieval metric.

## Reproduction artifacts

- Corpus and judgments: `evaluation/agent_discovery/space-rocks.v1.json`
- Hermes telemetry summarizer: `evaluation/summarize_space_rocks_agent_benchmark.py`
- Compact per-run metrics: `evaluation/results/space-rocks-grimoire-cbm-agent-benchmark-2026-07-27.json`
