# Arcana Semantic Graph Vector Calibration — 2026-07-26

## Decision

Keep the Arcana semantic graph index as a supported **conditional recall-expansion** lane.

It is not an equal mandatory provider and it does not replace Lexicon or Arcana's authoritative graph. Production integration supports:

- `--arcana-semantic=auto` — default; skip the embedding request when the query explicitly names a compound Lexicon seed or its path, otherwise attempt semantic recall expansion;
- `--arcana-semantic=on` — force semantic seed lookup for quality-sensitive or diagnostic requests; and
- `--arcana-semantic=off` — use deterministic Lexicon-seeded Arcana traversal only.

Index construction remains explicit. Context queries never build vector state. Missing, stale, corrupt, policy-incompatible, or unavailable vector state falls back to deterministic traversal.

Implementation revision: `aa010e975abad3de63a60f2c7431807e5849e39d`.

## Final policy-v5 contract

Semantic policy version 5 retains declaration-level graph entry points and excludes variables, parameters, fields, imports, exports, constants, tests, synthetic `@...` paths, and anonymous closures or lambdas. Each embedding document contains exact graph identity plus deterministic natural-language terms derived from camel case, snake case, acronyms, and path separators.

Neighbourhood text is restricted to semantic entry points, ordered with operational relationships first, bounded to 12 outgoing and 12 incoming relationships, and capped at 6,000 UTF-8 bytes. Oversized provider batches are recursively split, successful cache objects are synchronized immediately, and snapshot publication remains deterministic and rollback-safe.

## Evaluated snapshots

| Corpus | Arcana snapshot | Cases | Indexed documents |
| --- | --- | ---: | ---: |
| Grimoire mixed Go/Rust | `sha256:4a9cefb03363fb367355195aa56212caa7ee0babb5cad795010fe5f799945b75` | 5 | 5,472 |
| Space Rocks | `sha256:1e7c96b463135cbcd7cf92860fe18e3a3445129407c94956a2b87cb6f7cf1177` | 9 | 14,077 |

Both indexes use `qwen3-embedding-0.6b-q8_0-512d`. The Space Rocks policy-v5 publication embedded 14,077 documents through 1,760 successful provider requests, produced a 30,829,494-byte snapshot, and reported no failed batches.

## Paired production-bound results

The paired modes use the same prepared source, Lexicon export, Arcana snapshot, six-seed bound, and deterministic graph expansion. `lexicon-plus-vector` interleaves semantic and Lexicon seeds under the same production limit.

### Grimoire

| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | Median payload | Provider calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Lexicon seeds | 0.0% | 0.0% | 0.000 | 0.0% | 4,892.8 ms | 23,835 B | 2.00 |
| Lexicon + vector | 0.0% | 0.0% | 0.000 | 0.0% | 5,894.9 ms | 28,611 B | 3.00 |

Vector delta: no quality gain, +1,002.1 ms median latency, +4,776 B median payload, and one additional provider call.

The exact required owners are present in the mixed-language index, but raw semantic ranks remain 188–1,056 with a mean required-owner rank of 498.7. This self-referential corpus asks for mechanisms whose intent is not represented strongly enough by graph identity and immediate relationships alone. Identifier term splitting improved document readability but did not move those owners into the six-seed production bound.

### Space Rocks

| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | Median payload | Provider calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Lexicon seeds | 0.0% | 0.0% | 0.000 | 0.0% | 9,125.1 ms | 13,823 B | 2.00 |
| Lexicon + vector | 22.2% | 22.2% | 0.059 | 22.2% | 10,954.3 ms | 27,176 B | 3.00 |

Vector delta: +22.2 percentage points pass rate, required seed recall, and required structural recall; +0.059 MRR; +1,829.2 ms median latency; +13,353 B median payload; and one additional provider call.

The successful conceptual cases were:

- client gameplay reset — required owner retained at production rank 5; and
- client process shutdown — required owner retained at production rank 3.

The vector lane also surfaced useful supporting owners for WebRTC recovery and several cross-language mechanisms without satisfying the full required-owner judgments.

## Interpretation

The final measurements support the conditional design:

- vectors recover owners that lexical matching misses on a large cross-language repository;
- they do not reliably solve concept-heavy self-description when intent is absent from graph text;
- the additional request and graph expansion have measurable latency and payload cost; and
- explicit symbol or path queries should continue to bypass semantic expansion in `auto` mode.

The six-seed production bound remains unchanged. Raising it would increase graph expansion and payload on every semantic request without evidence of broad enough recall gains.

## Reliability findings retained

The calibration exercised and retained:

- exact snapshot pinning and concurrent build/query locking;
- checksum and corruption validation;
- rollback-safe publication;
- resumable content-addressed cache objects;
- adaptive splitting of provider requests that exceed context limits;
- deterministic 6,000-byte document bounding;
- mixed-language Go/Rust snapshot alignment;
- declaration-preferred seed resolution across callable kind differences; and
- deterministic fallback to Lexicon-seeded traversal.

Authoritative detailed reports:

- `arcana-grimoire-paired-v5-final-2026-07-26.{json,md}`
- `arcana-space-rocks-paired-v5-final-2026-07-26.{json,md}`
