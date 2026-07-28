# Space Rocks distant-player locator benchmark

Date: 2026-07-28

Task category: cross-language change

Repository: `space-rocks`

Pinned revision: `460da4af05c44d1835401fa853f5fc6b718262c8`

Model/provider: GPT-5.6 Sol through OpenAI Codex

Conditions ran concurrently in separate clean detached worktrees. Every condition had normal shell, Git, and direct source inspection. CBM and Grimoire were optional, mutually exclusive additions.

## Task

> Players still need reliable offscreen indicators for distant players after full-rate movement updates become receiver-filtered. Plan the smallest end-to-end change that preserves those indicators without restoring full hot updates for distant players.

The prompt did not disclose the expected server seam, generated contract, client ownership boundary, cadence policy, or failure checklist.

## Result summary

| Condition | Preparation | Agent elapsed | Total elapsed | Model calls | Total tokens | Answer bytes | Grounding |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Plain | none | **5m 27s** | **5m 27s** | **21** | **2,176,574** | 17,083 | Pass |
| CBM | **5.3s** | 6m 31s | 6m 37s | 25 | 2,401,151 | 20,834 | Pass |
| Grimoire | 1m 35s | 8m 06s | 9m 41s | **21** | 2,379,253 | 20,275 | Pass |

All three conditions completed normally and supplied valid evidence JSON.

Grounding details:

| Condition | Checked inline citations | Structured evidence items | Hidden path-family coverage | Canonical handles |
| --- | ---: | ---: | --- | ---: |
| Plain | 37 | 12 | `services/`, `client/`; missing `shared/` | 0 |
| CBM | 52 | 17 | `services/`, `shared/`, `client/` | 0 |
| Grimoire | 56 | 12 | `services/`, `client/`; missing `shared/` | 1 |

The missing `shared/` coverage in plain and Grimoire is limited to their final structured evidence. Both prose answers discussed generated packet source files. Grimoire supplied the suite's first canonical inspected source-range handle, resolving correctly to the overlay projection.

## Current architecture found by all three

All answers correctly established that:

- immutable server presentation currently publishes every ship;
- receiver snapshots do not yet filter entity membership;
- full ship transform changes are projected into the hot `ship_delta` path;
- the ships transport is unordered, unreliable, and supersedable;
- client offscreen indicators currently consume rendered remote-player positions;
- removing a filtered player from world/render state also removes its target and hue cache;
- `OSIndicatorController` can already render any ID-to-position dictionary and removes indicators when IDs disappear;
- authoritative server coordinates must be converted through the existing wrapped/ViewAnchor visual-position API.

The task therefore needs an independent, server-owned coarse position read model rather than retained distant render nodes or restored full-rate ship movement.

## Proposed designs

### Plain and CBM

Plain and CBM independently proposed the same sound transport direction:

- add a dedicated low-cadence `player_locator` packet family;
- carry it on the existing unreliable ships transport rather than reliable control traffic;
- keep packet-family sequence/projection state separate from `ship_delta` despite sharing the physical lane;
- publish a complete replacement set at roughly 5 Hz, with immediate membership/lifecycle changes;
- include authoritative position, velocity, and active state;
- add generated Go/GDScript and compact-wire descriptors from the shared TOML sources;
- maintain independent client locator state;
- merge rendered positions with locator-only positions through a new indicator-specific WorldSync read model;
- use bounded extrapolation and stale expiry;
- leave targeting and normal rendered-player state on the precise world/hot path.

This satisfies the intended low-cadence server seam and avoids using reliable control traffic for continuously changing coordinates.

### Grimoire

Grimoire instead proposed adding `player_locators` to the reliable overlay projection and delta stream. It argued that this was the smallest change because overlay is already receiver-local and has full/delta client state.

That design is internally coherent, but it is the wrong transport boundary for this task. Coordinates change continuously even when sampled at 5 Hz. Placing complete locator position arrays on the ordered/reliable overlay lane can create head-of-line pressure behind superseded movement samples and couples coarse movement to required HUD/control delivery. The hidden rubric explicitly disallowed reusing reliable control traffic for hot state.

Grimoire also proposed replacing the existing broad `get_remote_player_visual_positions()` read model once a locator baseline exists rather than adding an indicator-specific method. That risks changing other consumers of precise rendered-player positions. Plain and CBM kept the new coarse read model scoped to offscreen indicators.

## Shared answer gaps

Automatic grounding passed for all three, but none fully satisfied the hidden implementation rubric.

### Indicator color was not preserved

All three correctly observed that filtered player removal erases the existing remote hue state. Nevertheless, their locator records contained identity and movement fields only. They then relied on `OSIndicatorController`'s fallback hue when no rendered-player hue exists.

That recreates the known failure mode where distant indicators flatten to one fallback color. A complete plan must either:

- include stable presentation/color metadata in the locator record; or
- preserve an independent player-ID-to-hue/team-color read model outside filtered render membership.

Rendered positions should still override coarse locator positions inside interest, while durable color identity remains available outside interest.

### Spectating was not fully verified

The hidden verification rubric required spectating behavior. Plain and CBM mentioned receiver/view-target interest boundaries, but none supplied a complete acceptance case proving that locator coordinates, indicator identity, and camera/view-anchor conversion remain correct while spectating another player.

### Grimoire omitted stale-state protection

Plain and CBM defined bounded extrapolation and stale expiry. Grimoire's reliable overlay design defined replacement and empty-clear semantics but no client stale timeout. A stalled stream could therefore leave locator positions indefinitely visible until a lifecycle replacement arrives or the session resets.

## Qualitative comparison

### CBM

CBM produced the strongest plan by a small margin. It chose the correct unreliable ships transport, traced the full generated contract, covered all three hidden source families, kept coarse positions in an indicator-specific read model, and defined independent sequence state, bounded extrapolation, stale handling, and transition tests.

It still missed durable indicator color and complete spectating verification.

### Plain

Plain was close to CBM in quality and completed much faster. It chose the same correct packet-family seam, gave a particularly clear server publication and independent commit-state design, and supplied strong loss, cadence, transition, wraparound, and stale-state tests.

Its structured evidence omitted `shared/`, although its prose named the required shared contract sources. It also missed durable indicator color and complete spectating verification.

### Grimoire

Grimoire was well grounded and correctly mapped the server, generated-wire, client state, and presentation boundaries. It also used a canonical Grimoire handle successfully.

However, it selected the reliable overlay lane for movement-like locator coordinates, broadened an existing WorldSync read model rather than adding an indicator-specific one, omitted stale expiry, and still did not preserve distant-player color. Those are substantive design defects rather than merely presentation differences.

## Efficiency interpretation

Compared with plain inspection, Grimoire used:

- the same number of model calls;
- 202,679 more total tokens, about 9% more;
- 159 seconds more agent time, about 49% slower;
- about 254 seconds more total time after cold preparation, about 78% slower.

Compared with CBM, Grimoire used:

- four fewer model calls;
- about 22,000 fewer total tokens;
- about 185 seconds more total time including preparation.

This is a negative result for Grimoire on this cross-language task. The task was broader than Detekt, but its vocabulary still gave direct anchors into offscreen indicators, player movement, and receiver filtering. Plain search located the key server, shared-contract, and client paths quickly enough that Grimoire's extra context did not improve the architecture decision.

## Grimoire behavior

The Grimoire agent made two calls:

- one broad search;
- one handle-based inspect.

Combined structured output:

- 50,178 bytes;
- 25 nodes;
- 19 source ranges;
- eight documents;
- eight graph paths;
- no tool errors.

The initial response began with low-value fixture and test-helper symbols before the core protocol and presentation ownership evidence. The agent then inspected the overlay projection and anchored one final evidence item to the resulting canonical handle. Handle continuity worked, but broad-response shaping again returned substantial adjacent context.

The overlay result may also have biased the model toward reusing that lane because it was a prominent receiver-local projection seam. This is an inference, not a controlled causal conclusion; the final design remains the model's responsibility.

## Preparation breakdown

CBM indexing: 5.349 seconds.

Grimoire cold preparation: 94.770 seconds externally, 94.727 seconds in internal timing buckets:

| Stage | Time |
| --- | ---: |
| Lexicon | 66.293s |
| Source index | 14.675s |
| Documentation | 5.651s |
| Reinspection | 4.637s |
| Arcana | 2.262s |
| Initial inspection | 1.008s |
| Final verification | 0.187s |
| Marker writes | 0.004s |

Lexicon remains the dominant first-use cost on Space Rocks.

## Harness finding

The run initially replaced the combined suite summary with the selected task instead of preserving the first two task entries. Raw task artifacts were unaffected.

The canonical summary was recovered from the committed two-task summary and merged with this run without rerunning agents. The runner now loads a compatible existing summary and appends or replaces only selected task entries. It rejects incompatible model, provider, task-suite, or schema metadata rather than silently combining unlike runs. The import utility also supports summary-only recovery from a Git revision.

## Retained artifacts

- `plain.stdout.txt`, `cbm.stdout.txt`, `grimoire.stdout.txt`: full unedited answers;
- matching `.usage.json`: model telemetry;
- matching `.grounding.json`: automatic grounding reports;
- `grimoire.mcp-audit.jsonl`: exact Grimoire requests and structured responses;
- `grimoire-prewarm.stdout.txt`: preparation state and timing breakdown;
- CBM indexing stdout/stderr;
- `evaluation/results/agent-benchmark-v2/summary.json`: combined three-task suite summary.

Transient CBM databases and detached benchmark worktrees are not part of the retained evidence.
