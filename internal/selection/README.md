# Selection

`internal/selection` curates ranked `retrieve.Candidate` values before the exact-budget compiler.

## Contract

`Curate(snapshot, candidates)` returns a deterministic sequence with these phases:

1. Candidates are consumed in incoming merged order. Provider ranks are retained
   as provenance but are never compared across exact, vector, or lexical sources.
2. Duplicate chunk IDs and later overlapping ranges in the same file are removed.
3. Remaining primaries are reordered with calibrated soft diversity pressure.
   Each previously selected chunk from the same file adds 10 positions of penalty;
   each from the same subsystem adds 18. All unique, non-overlapping primaries remain.
4. The first three diversified primaries are emitted first, followed by their
   deduplicated immediate prepared neighbors, then all remaining primaries.
   Adjacent candidates use source `adjacent` and carry a directional reason.

Primary candidate scores, ranks, sources, and reasons are copied unchanged.
For paths below `internal/`, the child directory is the subsystem; other paths
use their first path component. Prepared chunk order is the adjacency order.

`CurateWithConfig` exposes the same implementation with explicit penalties and
neighbor-anchor count. The judged evaluator uses this seam for calibration; the
normal context command always uses `DefaultConfig`.

The current defaults were selected against the 716-file Grimoire lexical/adaptive
candidate stream after BM25 intent retrieval, query decomposition, and Arcana
semantic vectors landed. The older `4/10/2` and intermediate `10/10/3`
calibrations regressed or underperformed on the final candidate stream.
See the
[calibration comparison](../../evaluation/results/selection-calibration-comparison-2026-07-23.md).

This package deliberately does not know about budgets or serialized packages.
The compiler remains the final exact-budget boundary.
