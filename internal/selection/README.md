# Selection

`internal/selection` curates ranked `retrieve.Candidate` values before the exact-budget compiler.

## Contract

`Curate(snapshot, candidates)` returns a deterministic sequence with these phases:

1. Candidates are consumed in incoming merged order. Provider ranks are retained
   as provenance but are never compared across exact, vector, or lexical sources.
2. Duplicate chunk IDs and later overlapping ranges in the same file are removed.
3. Remaining primaries are reordered with soft diversity pressure. Repeated files
   cost more than repeated subsystems; all unique, non-overlapping primaries remain.
4. The first four diversified primaries are emitted first, followed by their
   deduplicated immediate prepared neighbors, then all remaining primaries.
   Adjacent candidates use source `adjacent` and carry a directional reason.

Primary candidate scores, ranks, sources, and reasons are copied unchanged.
For paths below `internal/`, the child directory is the subsystem; other paths
use their first path component. Prepared chunk order is the adjacency order.

This package deliberately does not know about budgets or serialized packages.
The compiler remains the final exact-budget boundary.
