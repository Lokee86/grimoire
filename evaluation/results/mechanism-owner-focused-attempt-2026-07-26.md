# Mechanism-owner focused attempt

Date: 2026-07-26

The Arcana corpus labels are human mechanism-owner judgments used as a common benchmark. They are not treated as absolute ground truth.

## Question

Can one-hop graph evidence resolve implementation owners that semantic and lexical retrieval rank behind nearby wrappers or helpers?

## Attempt

The experiment added bounded incoming/outgoing neighbor analysis to the existing 256-candidate graph pass. Two variants were measured:

1. Inject qualifying one-hop functions as new candidates.
2. Refine the design so graph evidence may only boost candidates already present in the Lexicon/vector pool.

No additional Arcana protocol round trip was added.

## Results

| Variant | Corpus | Required seed recall | MRR | Median vector-path latency |
| --- | --- | ---: | ---: | ---: |
| Owner scoring disabled | Grimoire | 10.0% | 0.200 | 6,597.7 ms |
| Inject one-hop owners | Grimoire | 0.0% | 0.000 | 8,152.8 ms |
| Boost existing candidates only | Grimoire | 0.0% | 0.000 | 11,778.2 ms |
| Owner scoring disabled | Space Rocks | 22.2% | 0.074 | 13,174.4 ms |
| Inject one-hop owners | Space Rocks | 22.2% | 0.130 | 14,452.7 ms |

The Space Rocks MRR gain came from promoting an already retrieved WebRTC recovery owner from rank 2 to rank 1. However, both owner variants regressed Grimoire exact-owner recall from 10% to 0%. Newly injected candidates also displaced useful seeds with plausible but non-owning helpers.

## CBM comparison

Codebase Memory MCP 0.9.0 was indexed against the same working trees and scored at top six using its natural-language BM25 mode and its keyword-array semantic mode. Grimoire was re-indexed after the experimental implementation was removed, with the benchmark harness temporarily absent from the indexed tree.

| Corpus | CBM mode | Required seed recall | MRR | Median latency |
| --- | --- | ---: | ---: | ---: |
| Grimoire | BM25 | 0.0% | 0.000 | 51.0 ms |
| Grimoire | Semantic | 0.0% | 0.000 | 45.2 ms |
| Space Rocks | BM25 | 0.0% | 0.000 | 139.5 ms |
| Space Rocks | Semantic | 0.0% | 0.000 | 62.5 ms |

CBM is substantially faster on this benchmark, but it does not currently provide a stronger exact-owner quality target.

## Decision

Reject and revert the mechanism-owner stage. The existing calibration cases already triggered the no-regression stop condition, so no sealed holdout run was used to rationalize a failed design. One-hop graph proximity is useful supporting evidence, but it is not sufficiently specific to infer ownership at seed-selection time. Continue using deterministic retrieval as a reproducible candidate generator, with downstream inspection and verification responsible for final ownership conclusions.
