# Arcana

Arcana is the repository-graph foundation of the [**Warlock Toolchain**](https://github.com/Lokee86/warlock-toolchain).
It models repositories as queryable graphs and provides the storage, snapshot,
and traversal foundations used by higher-level Warlock tools such as Demon Docs,
Grimoire Context, and Pitlord.

## Ownership boundaries

- **Arcana** owns the factual repository graph, graph storage, snapshots,
  deterministic queries, and measurements of storage representations.
- **Demon Docs** owns documentation semantics, policy, review history, and
  Codemap decisions. It consumes Arcana facts without owning the graph
  engine.
- **Grimoire Context** owns task interpretation, relevance ranking, token
  budgets, and final context construction. It queries Arcana and Demon
  Docs without becoming either system's storage layer.

Arcana remains a standalone Rust process or CLI boundary. Go consumers do
not link it through cgo or FFI.

## Graph workload foundation

Arcana includes deterministic synthetic graph generation for exercising
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

Arcana snapshots are immutable compositions rather than mutable graph
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

## Repository ingestion

Arcana now accepts language-neutral repository facts and compiles them into
the same dense packed graph format used by its synthetic workloads. The compiled
output contains:

- `graph.arcana` — packed forward and reverse adjacency;
- `catalogue.tsv` — dense node IDs mapped back to stable keys, paths, names,
  kinds, content IDs, and source spans;
- `unresolved.tsv` — canonical version-2 unresolved-reference facts keyed back to
  catalogue nodes.

The first adapter uses Go's `go/parser` and `go/ast` packages for extraction and
`golang.org/x/tools/go/packages` with Go type information for semantic call
resolution. It emits repository, directory, file, package, import, type,
function, method, and test nodes, plus containment, definition, import, and
resolved call edges.

```text
go run ./adapters/go \
  -repo /path/to/go/module \
  -output target/repository.facts.tsv

cargo run --release -- import-facts \
  --facts target/repository.facts.tsv \
  --output target/repository-index

cargo run --release -- query \
  --graph target/repository-index/graph.arcana \
  --catalogue target/repository-index/catalogue.tsv \
  --name ExampleFunction \
  --reverse \
  --relation calls
```

For integrations, `arcana protocol` opens the verified repository snapshot once
and serves one JSON response for every JSON request line on standard input:

```text
cargo run --release -- protocol --snapshot target/repository-index
{"id":"stats","op":"stats"}
{"id":"symbol","op":"resolve_symbol","name":"ExampleFunction"}
{"id":"calls","op":"neighbors","node_id":42,"direction":"outgoing","relation":"calls"}
{"id":"unresolved","op":"unresolved","path":"internal/example.go","reason":"unsupported-form"}
{"id":"diff","op":"diff","other_snapshot":"target/previous-index"}
```

The stable protocol identifier is `arcana.query.v1`. Supported operations are
`resolve_symbol`, `resolve_file`, `list_nodes`, `neighbors`, `unresolved`,
`stats`, and `diff`. Request IDs are echoed unchanged, and malformed or failed
requests return JSON errors without terminating the stream.

The Go resolver handles concrete same-package and internal cross-package
functions and methods, including recursive self-calls. Built-ins, type
conversions, external APIs, dynamic dispatch, ambiguity, and missing targets are
retained as first-class unresolved-reference facts rather than becoming
speculative graph edges. Anonymous function bodies remain outside the graph
until closures have their own node model. See
[`docs/GO_ADAPTER_VALIDATION.md`](docs/GO_ADAPTER_VALIDATION.md) for measured
results on Demon Docs and the Space Rocks game server.

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

1. Resolve Go selector calls, methods, and internal cross-package calls without
   emitting speculative edges.
2. Convert file-level fact changes into overlays instead of rebuilding the full
   repository fact set.
3. Bind catalogues cryptographically to their packed graph artifacts and publish
   them through the snapshot manifest.
4. Define compaction triggers from real mutation rates and query mixes.
5. Add the Rust adapter through the same language-neutral fact boundary.

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

Arcana is available under the [MIT License](LICENSE.md).
