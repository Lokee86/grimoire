# Space Rocks Room-Scale Architecture Benchmark

Date: 2026-07-28

Task: investigate the existing Space Rocks multiplayer architecture and recommend a scalable direction for larger rooms without assuming a proposed solution.

Revision: `460da4af05c44d1835401fa853f5fc6b718262c8`

Model/provider: `gpt-5.6-sol` through `openai-codex`

Conditions ran in parallel after condition-specific cold preparation. No call or token budget was imposed.

## Result

All three answers passed automatic grounding after the validator was corrected to support multiple noncontiguous ranges in one structured evidence item.

The three answers converged on the same sound direction:

- retain authoritative room simulation;
- publish receiver-independent immutable presentation facts once per generation;
- keep receiver-interest policy in the realtime protocol boundary;
- filter before receiver delta construction;
- preserve per-session baselines, lifecycle create/delete semantics, backpressure, and resync behavior;
- use toroidal spatial queries and hysteresis;
- preserve or explicitly communicate spectator view ownership before filtering remote ships;
- measure candidate construction, encoding, bytes, and client presentation separately.

No answer was materially wrong. CBM was marginally strongest on ownership precision, Grimoire was strongest on explicit end-to-end spectator/shared-contract planning, and plain search was broad and technically sound. The quality difference was small; the efficiency difference was not.

**Overall benchmark outcome: Grimoire produced a comparable implementation-grade answer substantially faster and with substantially lower interaction cost.**

## Efficiency

| Condition | Preparation | Agent elapsed | Logical total | Model calls | Total tokens | Answer bytes |
|---|---:|---:|---:|---:|---:|---:|
| Grimoire | 95.835 s | 396.143 s | 491.978 s | 16 | 1,053,805 | 26,732 |
| CBM | 6.052 s | 693.310 s | 699.362 s | 25 | 3,300,961 | 21,085 |
| Plain | none | 747.395 s | 747.395 s | 18 | 1,840,597 | 25,984 |

Total-token counts include cache reads:

| Condition | Input | Output | Reasoning | Cache reads |
|---|---:|---:|---:|---:|
| Grimoire | 132,672 | 11,309 | 2,601 | 909,824 |
| CBM | 201,364 | 12,749 | 4,876 | 3,086,848 |
| Plain | 154,920 | 12,973 | 3,970 | 1,672,704 |

Relative to CBM, Grimoire used:

- 36% fewer model calls;
- 68% fewer total tokens;
- 43% less agent elapsed time;
- about 30% less logical total time after including cold preparation.

Relative to plain search, Grimoire used:

- 11% fewer model calls;
- 43% fewer total tokens;
- 47% less agent elapsed time;
- about 34% less logical total time after including cold preparation.

## Grounding

| Condition | Valid | Inline citations | Structured evidence items | Canonical handles | Hidden path families found |
|---|---|---:|---:|---:|---|
| Grimoire | yes | 47 | 20 | 0 | `services/`, `client/` |
| CBM | yes | 45 | 12 | 0 | `services/`, `client/` |
| Plain | yes | 53 | 19 | 0 | `services/`, `client/` |

Every reported path and range passed validation after noncontiguous range support was added. None of the answers supplied canonical Grimoire handles. All three omitted `shared/` from their final structured evidence, although Grimoire's prose explicitly proposed `shared/packets/gameplay.toml` for the later spectator-view request.

This common omission means the hidden cross-stack evidence requirement was only partially satisfied. It did not invalidate factual grounding because hidden rubric coverage is deliberately non-gating.

## Qualitative rubric review

### Ownership

CBM was slightly strongest. It cleanly separated:

- immutable presentation spatial facts under `game`;
- receiver-interest policy under `protocol/realtime`;
- transport and commit mechanics under `networking`;
- unchanged client lifecycle application.

Plain reached nearly the same boundary and explicitly rejected concurrent reuse of the mutable collision index.

Grimoire correctly separated simulation, projection, transport, and client application, but its later proposed `realtime.RoomProjectionCache` is less precise than CBM's presentation-owned immutable spatial snapshot. The cache may still be viable, but the owner needs clarification so receiver-independent presentation facts do not drift into protocol policy.

### Membership and lifecycle

All three were strong. They covered:

- filtered full baselines;
- reliable create on interest entry;
- reliable delete on interest exit;
- hot updates only after establishment;
- hysteresis;
- self and spectator anchoring;
- resync;
- backpressure-safe commit semantics;
- re-entry and stale membership cleanup.

CBM supplied an especially useful detail: derive retained membership from the successfully committed receiver projection rather than mutating an independent interest set before transport success.

Grimoire supplied the clearest explicit server-visible spectator-view request and separated it from gameplay targeting.

Plain proposed the safest staged rollout: filter bullets, asteroids, and pickups first; preserve all ships until the spectator-view seam exists.

### Prioritization and packet budgeting

All three correctly treated interest selection as separate from packet budgeting and cadence. Each recognized that the existing scheduler and chunking mechanisms bound packet shape without reducing the number of entities considered per receiver.

### Cross-stack change

Grimoire was strongest in prose because it named the shared packet source, generated outputs, client spectate flow, server inbound handling, and session state required for a spectator-view request.

Plain and CBM both traced server and Godot behavior well, but their first slices intentionally avoided a wire change. None included a `shared/` path in structured evidence.

### Verification and rollout

All three proposed strong failure-oriented testing and measurement. Common strengths included toroidal seams, hysteresis churn, blocked reliable groups, hot-before-create behavior, resync, spectator transitions, client node counts, bytes, CPU, allocations, and increasing receiver counts.

## First-use preparation

Grimoire cold preparation took 95.246 seconds internally:

| Stage | Time |
|---|---:|
| Lexicon | 69.342 s |
| Grimoire source index | 14.046 s |
| Reinspection | 4.645 s |
| Documentation index | 3.387 s |
| Arcana | 2.112 s |
| Initial inspection | 1.506 s |
| Final verification | 0.194 s |
| Marker writes | 0.004 s |
| Lock wait | 0.001 s |

Lexicon accounted for roughly 73% of Grimoire cold preparation. This is the clear first optimization target. Arcana and documentation preparation were comparatively small.

CBM cold indexing took 6.052 seconds.

Despite the preparation disadvantage, Grimoire still completed its condition about 207 seconds before CBM and 255 seconds before plain search when preparation is included.

## Response-size finding

Grimoire made only one audited discovery call, but that response was 50,915 bytes and contained:

- 25 new nodes;
- 22 new source ranges;
- 8 documents;
- 8 graph paths.

The agent then used direct source inspection and produced a valid answer in 16 model calls. This supports the discovery-interface model, but it also confirms the remaining response-shaping concern: one broad architecture search still returned a large context payload. Future tuning should reduce broad-response duplication and adjacent evidence without reintroducing a global ranking or preassembled context package.

## Harness correction discovered by the run

The initial validation marked CBM and plain invalid because their structured evidence combined noncontiguous ranges such as `9-20,81-113`. The paths and ranges themselves were legitimate.

The validator now:

- accepts comma-separated and list-form noncontiguous ranges;
- validates every individual range;
- rejects one canonical handle attached to multiple ranges;
- supports revalidating saved answers without rerunning agents.

The existing outputs were revalidated in place. No agent condition was rerun.

## Interpretation

This task supports the working hypothesis that Grimoire is useful on open-ended architectural exploration. It did not produce a uniquely better plan, but it reached the same quality class with substantially less interaction than both alternatives.

The result does not establish universal superiority. It is one task, all answers converged strongly, no answer used canonical Grimoire handles, and all three missed the hidden `shared/` evidence family. The remaining four v2 tasks are needed before drawing a broader conclusion.
