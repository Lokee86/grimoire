# ArcanaGraph

ArcanaGraph is the repository-graph foundation of the [**Warlock Toolchain**](https://github.com/Lokee86/warlock-toolchain).
It models repositories as queryable graphs and provides the storage, snapshot,
and traversal foundations used by higher-level Warlock tools such as Demon Docs,
Grimoire Context, and Pitlord.

## Ownership boundaries

- **ArcanaGraph** owns the factual repository graph, graph storage, snapshots,
  deterministic queries, and measurements of storage representations.
- **Demon Docs** owns documentation semantics, policy, review history, and
  Codemap decisions. It consumes ArcanaGraph facts without owning the graph
  engine.
- **Grimoire Context** owns task interpretation, relevance ranking, token
  budgets, and final context construction. It queries ArcanaGraph and Demon
  Docs without becoming either system's storage layer.

ArcanaGraph remains a standalone Rust process or CLI boundary. Go consumers do
not link it through cgo or FFI.

## Graph workload foundation

ArcanaGraph includes deterministic synthetic graph generation for exercising
its packed storage, snapshot, overlay, and compaction systems across more than a
single toy topology.

The workload foundation currently includes five topology families:

- **Modular** — cohesive clusters with a configurable cross-cluster edge share.
- **Entangled** — hubs, cross-cluster relationships, cycles, and local edges.
- **Hub-heavy** — a small set of nodes owns most incoming and outgoing edges.
- **Layered** — deep, mostly forward relationships with a smaller irregular
  edge set.
- **Dense subsystem** — a tightly connected subsystem inside a larger sparse
  graph.

Standard scale tiers range from 10,000 nodes and 100,000 edges to 5,000,000
nodes and 50,000,000 edges. Generation scales with requested edges rather than
enumerating every possible node pair.

Mutation plans cover single-node, local-range, scattered, hub-focused, and
percentage updates. A plan contains exact removed and replacement edges so the
overlay and rebuilt packed snapshot receive the same logical update.

## Determinism and invariants

Synthetic datasets and mutations are controlled by explicit seeds. Generated
and mutated graphs guarantee:

- exact requested edge counts;
- unique directed non-self edges;
- canonical source/target/kind ordering;
- topology-specific edge distributions;
- stable output for the same specification and seed; and
- preserved edge-kind counts across mutations.

The generator uses a small internal permutation sampler and has no third-party
dependencies. Dataset construction is outside the measured update and query
windows and is reused by both benchmark paths.

## Packed adjacency format

The immutable packed format uses a fixed, versioned, little-endian header
followed by aligned forward offsets, forward targets, forward edge kinds,
reverse offsets, reverse sources, and reverse edge kinds. The writer
canonicalizes logical edges, streams deterministic bytes through a temporary
file, syncs them, and atomically commits a new snapshot path.

Opening a packed graph validates the header, exact section layout, file length,
payload checksum, logical dataset checksum, offset tables, node bounds, and
adjacency ordering. Queries read directly from the packed byte buffer without
rebuilding an in-memory graph. A separate in-memory implementation provides the
correctness oracle used by round-trip tests.

## Snapshots and overlays

ArcanaGraph snapshots are immutable compositions rather than mutable graph
files. A snapshot manifest identifies one validated packed base plus an optional
immutable overlay. The manifest is published last, so readers never observe a
partially assembled snapshot.

Overlay v1 stores canonical added edges and removed-edge tombstones bound to the
exact node count, edge count, and dataset checksum of its packed base. Opening a
snapshot validates the base, overlay, manifest identities, visible edge count,
and visible dataset checksum before queries are allowed. Forward and reverse
queries merge the base adjacency with small in-memory overlay indexes without
rewriting the packed base.

Snapshot and overlay files are content-identifiable and refuse in-place
replacement. `compact_snapshot` materializes the visible graph into a new packed
base, verifies its edge count and checksum, and publishes a new base-only
manifest last. The source snapshot remains untouched.

## Benchmarks

The benchmark harness compares immutable overlays against rebuilding a complete
packed replacement while treating the existing packed base as shared storage:

```text
cargo run --release -- benchmark \
  --tier small \
  --topology modular \
  --queries 10000 \
  --samples 3 \
  --csv target/benchmarks/small-modular-mutations.csv
```

Supported tiers are `small`, `medium`, `large`, and `stress`. Supported
synthetic topologies are `modular`, `entangled`, `hub-heavy`, `layered`, and
`dense-subsystem`.

The five mutation workloads cover one hot node, a local range, scattered
changes, hub-focused changes, and one percent of all edges. Each run measures
creation, fully validated reopen, warm forward/reverse queries, and incremental
file size. Overlay and rebuilt-packed results must produce identical visible
graph checksums and query fingerprints.

## Next implementation steps

1. Validate mutation and compaction policy against captured Demon Docs and
   Space Rocks repository graphs.
2. Define compaction triggers from overlay size, mutation rate, reopen frequency,
   and expected query mix rather than one fixed percentage.
3. Add separate-process cold and mixed-cache benchmark modes.
4. Stream compaction directly into the packed writer if materialization becomes
   a measured memory bottleneck.
5. Measure ordinary buffered reads before introducing memory mapping.

## Development

The package uses Rust edition 2024.

```text
cargo fmt -- --check
cargo check --all-targets
cargo test --all-targets
cargo run -- --help
cargo run -- --version
```

## License

ArcanaGraph is available under the [MIT License](LICENSE.md).
