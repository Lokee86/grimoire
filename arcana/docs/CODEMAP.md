# Arcana codemap

This codemap describes the current implementation under `arcana/`. It is a starting-point guide, not a replacement for reading the owning module and its tests. “Implemented” below means there is a current source path for the behavior. Explicit future boundaries are collected at the end so they are not mistaken for working features.

## Top-level boundaries

- `arcana/Cargo.toml` defines the independently buildable Rust library and `arcana` binary. The crate uses Rust edition 2024 and forbids unsafe code.
- `arcana/build.rs` supplies `GRIMOIRE_RELEASE_VERSION` to the crate, falling back to the Cargo package version.
- `arcana/src/lib.rs` is the reusable library boundary. It exports `benchmark`, `lexicon`, `protocol`, `repository`, `snapshot`, `storage`, `synthetic`, and `vector`.
- `arcana/src/main.rs` is the executable entry point and the final command dispatcher.
- `arcana/README.md` describes product boundaries and end-to-end usage.
- `arcana/docs/LEXICON_CONTRACT.md` documents the Lexicon consumer contract and incremental ownership policy.
- `arcana/docs/repository-snapshots.md` documents repository snapshot artifacts and changed-file updates.
- `arcana/docs/vector-index.md` documents the optional semantic index. For implementation details and format/policy constants, follow `arcana/src/vector/`.

Arcana owns ingestion of language-neutral facts, dense repository compilation, packed graph storage, immutable graph/repository snapshots, overlays, deterministic graph queries, optional graph vectors, and graph-storage benchmarks. It does not own language parsing, source adapters, documentation policy, or Grimoire’s provider-neutral discovery API.

## CLI entry points

Command parsing is centralized in `arcana/src/cli.rs`; execution remains in command-specific owners. `arcana/src/main.rs` should stay thin: it parses, dispatches, formats top-level errors, and writes benchmark CSV output.

| Command | Owning execution path | What it currently does |
| --- | --- | --- |
| `arcana import-facts` | `arcana/src/cli_commands.rs` | Parses canonical TSV repository facts, compiles them, writes a packed base plus repository metadata, then publishes manifests. |
| `arcana update-facts` | `arcana/src/cli_update.rs` | Replaces facts owned by declared changed paths and writes a cumulative edge overlay when the dense node set is unchanged. |
| `arcana sync` | `arcana/src/cli_sync.rs`, `arcana/src/cli_sync_state.rs` | Reads Lexicon `CURRENT`, chooses an existing snapshot, overlay update, or packed rebuild, serializes writers, and atomically publishes Arcana `CURRENT`. |
| `arcana query` | `arcana/src/cli_query.rs` | Performs exact-name lookup and forward/reverse neighbor listing against an explicitly supplied `graph.arcana` and `catalogue.tsv`. It opens the packed graph directly; it is not the overlay-aware repository protocol. |
| `arcana protocol` | `arcana/src/cli_protocol.rs`, `arcana/src/protocol/` | Opens a validated repository snapshot and serves one JSON response per JSONL request on stdin/stdout. |
| `arcana vectorize` | `arcana/src/cli_vectors.rs`, `arcana/src/vector/build.rs` | Explicitly builds or reuses the semantic graph index for Arcana `CURRENT`. |
| `arcana semantic-query` | `arcana/src/cli_vectors.rs`, `arcana/src/vector/search.rs` | Embeds a query and scores the current or expected-snapshot vector index. |
| `arcana benchmark` | `arcana/src/main.rs`, `arcana/src/benchmark/` | Compares overlay mutation/reopen/query work with rebuilt packed graphs and optionally writes CSV. |

CLI tests live in `arcana/src/cli_tests.rs`, `arcana/src/cli_update_tests.rs`, and `arcana/src/cli_sync_tests.rs`. `arcana/src/cli_sync_state.rs` also has focused lock and atomic-replacement tests.

## Lexicon ingestion

The owner is `arcana/src/lexicon/`. This module consumes immutable Lexicon storage; it does not invoke language adapters.

- `arcana/src/lexicon/mod.rs` defines `LexiconSnapshot`, compatibility warnings, and comparisons of file/shared-object identities.
- `arcana/src/lexicon/snapshot.rs` resolves `.lexicon/CURRENT`, validates content-addressed snapshot and object IDs, verifies hashes and metadata, reads per-language shared and file objects, and assembles one fact set.
- `arcana/src/lexicon/format.rs` defines the decoded Lexicon manifest structures.
- `arcana/src/lexicon/binary.rs` decodes binary v1 string tables and node, edge, and unresolved sections with explicit size/count bounds.
- `arcana/src/lexicon/object.rs` decodes the legacy canonical JSON fact-object form.
- `arcana/src/lexicon/records.rs` converts external SHA-256 identities and labels into Arcana `RepositoryFacts`, detects compact-ID collisions, normalizes paths, and emits compatibility warnings for safely degraded unknown labels.

Implemented compatibility behavior is deliberate: unknown node kinds become `symbol`; unknown edge and unresolved-relation labels are skipped; unknown unresolved-reason labels are preserved. The same warnings are printed by sync and persisted in `compatibility.warnings` by `arcana/src/cli_sync.rs`.

Relevant tests are `arcana/src/lexicon/tests.rs`, `arcana/src/lexicon/binary_tests.rs`, `arcana/src/lexicon/format_tests.rs`, and inline conversion tests in `arcana/src/lexicon/records.rs`. End-to-end sync coverage is in `arcana/src/cli_sync_tests.rs`.

## Repository facts and compilation

The owner is `arcana/src/repository/`.

- `arcana/src/repository/model.rs` owns `NodeKey`, `ContentId`, node/edge facts, node kinds, relation kinds, and source spans.
- `arcana/src/repository/unresolved.rs` owns unresolved-reference reasons and records.
- `arcana/src/repository/path.rs` owns repository-relative path normalization used by ingestion, updates, and queries.
- `arcana/src/repository/fact_file.rs` and `arcana/src/repository/fact_file_error.rs` own the canonical TSV fact format used by manual import, persisted `facts.tsv`, and update input.
- `arcana/src/repository/compiler.rs` is the dense compilation seam. It validates nodes and endpoints, assigns deterministic snapshot-local `NodeId` values in `NodeKey` order, deduplicates repeated relationship occurrences into reachability edges, converts relation labels to stable edge codes, and builds the catalogue.
- `arcana/src/repository/catalogue.rs` and `arcana/src/repository/catalogue_file.rs` own dense-ID-to-fact metadata and its persisted/indexed form.
- `arcana/src/repository/ownership.rs` partitions nodes, edges, and unresolved records into shared and source-file-owned facts and replaces declared file partitions.
- `arcana/src/repository/incremental.rs` recompiles the replacement view, requires an unchanged node-key-to-dense-ID map, and computes additions/removals relative to the packed base.
- `arcana/src/repository/lexicon_fact_file.rs` is the complete Lexicon JSONL importer retained for migration/diagnostic paths; normal snapshot sync uses `arcana/src/lexicon/` instead.

Compilation produces a `GraphDataset`, `RepositoryCatalogue`, stable-to-dense node mapping, and unresolved facts. `NodeId` is local to a compiled snapshot; `NodeKey`/Lexicon identity is the cross-snapshot identity seam.

Repository tests are grouped by responsibility: `compiler_catalogue_tests.rs`, `fact_file_tests.rs`, `lexicon_fact_file_tests.rs`, `ownership_tests.rs`, and `incremental_tests.rs` under `arcana/src/repository/`.

## Packed storage

The owner is `arcana/src/storage/`.

- `arcana/src/storage/format.rs` owns the packed header, section layout, little-endian primitives, and stable checksums.
- `arcana/src/storage/dataset.rs` validates/canonicalizes logical datasets and computes dataset checksums.
- `arcana/src/storage/writer.rs` writes forward and reverse adjacency sections through a synchronized temporary file and refuses to replace an existing packed path.
- `arcana/src/storage/reader.rs` reads the immutable file into a shared byte buffer, validates layout/checksums/offsets/bounds/order, and exposes forward and reverse neighbor iterators. It does not memory-map because the crate forbids unsafe code.
- `arcana/src/storage/oracle.rs` provides the in-memory correctness oracle used to compare packed behavior.
- `arcana/src/storage/error.rs` owns dataset, packed-open, and query errors.

Start in `format.rs` plus both `writer.rs` and `reader.rs` for a format change; update corruption and round-trip coverage together. Tests live in `arcana/src/storage/tests.rs`, `arcana/src/storage/corruption_tests.rs`, and small inline reader/writer tests.

## Graph snapshots, overlays, and repository snapshots

There are two nested ownership layers.

### Graph composition: `arcana/src/snapshot/`

- `graph.rs` opens a `graph.manifest`, validates the packed base and optional overlay, exposes overlay-aware visible neighbors, derives graph snapshot IDs, and publishes graph manifests.
- `manifest.rs` and `manifest_io.rs` own the graph manifest schema and immutable read/write rules.
- `overlay_format.rs`, `overlay_validation.rs`, and `overlay_error.rs` own overlay layout, change validation, and failures.
- `overlay_writer.rs` writes immutable added-edge and removed-edge operations bound to an exact packed base.
- `overlay.rs` opens those operations, verifies their base/visible identities, builds operation indexes, and merges them with base adjacency for reads.
- `compaction.rs` implements library-level compaction: materialize the visible graph, write a new packed base, verify count/checksum equivalence, and publish a new base-only manifest without modifying the source snapshot.

Graph snapshot tests live in `graph_tests.rs`, `overlay_tests.rs`, `manifest_tests.rs`, and `compaction_tests.rs` under `arcana/src/snapshot/`.

### Repository generation: `arcana/src/repository/`

- `repository_snapshot_format.rs` owns `repository.manifest` fields/versioning.
- `repository_snapshot.rs` publishes and opens the complete generation binding `graph.manifest`, `catalogue.tsv`, `unresolved.tsv`, and `facts.tsv`. Opening verifies artifact checksums and validates that graph, catalogue, facts, and unresolved data describe one generation.
- `repository_snapshot_validation.rs` owns component checksums, immutable writes, identity derivation, and cross-artifact validation.
- `repository_snapshot_error.rs` owns publication/open errors.
- `repository_snapshot_tests.rs` covers publication, reopening, corruption rejection, and component consistency.

`arcana/src/cli_commands.rs` writes complete base generations. `arcana/src/cli_update.rs` copies the original packed base, writes an optional cumulative overlay, publishes `graph.manifest`, and writes a new repository generation. `arcana/src/cli_sync.rs` places immutable generations under `.arcana/snapshots/<lexicon-digest>/`; `arcana/src/cli_sync_state.rs` owns `.arcana/LOCK` and atomic `CURRENT` replacement.

## Protocol and query handling

The owner is `arcana/src/protocol/`. The stable response protocol identifier is `arcana.query.v1`.

- `request.rs` defines the JSON request envelope and all operation fields.
- `session.rs` opens `repository.manifest`, transfers the validated graph/catalogue/unresolved components into `ProtocolSnapshot`, builds unresolved indexes, parses each line, and routes operations.
- `server.rs` owns the stdin/stdout JSONL loop and flushes one response per request.
- `response.rs` owns success/failure envelopes and shared node, relationship, and unresolved JSON shapes.
- `error.rs` owns protocol startup/serving errors; operation failures are converted to structured error responses in `session.rs`.
- `queries.rs` owns node search, symbol/file resolution, paged node listing, direct neighbor queries, unresolved queries, and common query limits/parsers.
- `traversal.rs` owns relation masks, bounds, deterministic neighbor expansion, breadth-first distance/path primitives, and shared traversal response building.
- `path_queries.rs` owns bounded multi-path and shortest-call-chain operations.
- `analysis_queries.rs` owns reachability, impact, dead-symbol, and operational-role operations.
- `architecture_queries.rs` owns deterministic architecture-community summaries.
- `stats.rs` owns snapshot statistics.
- `diff.rs` opens and compares another repository snapshot.
- `graph_export.rs` owns bounded, deterministic graph export. It pages path-filtered catalogue nodes, optionally adds validated pinned nodes, and emits visible edges whose endpoints are both returned.

Protocol queries are overlay-aware because `ProtocolSnapshot` contains a validated `GraphSnapshot`. Add an operation by changing `request.rs`, routing it in `session.rs`, placing behavior in the narrow owning query file, reusing `response.rs` shapes where applicable, and adding focused cases to `arcana/src/protocol/tests.rs`.

## Optional vectors

The owner is `arcana/src/vector/`. This path is explicit and optional; ordinary sync, packed storage, snapshots, and deterministic protocol queries do not call it.

- `documents.rs` owns semantic eligibility and deterministic bounded graph-document rendering. The implementation currently uses semantic eligibility policy version 6.
- `client.rs` owns the `Embedder` contract, defaults, OpenAI-compatible embeddings request/response handling, truncation to retained dimensions, and vector normalization.
- `http.rs` is the small HTTP transport used by the embedding client.
- `cache.rs` owns content-addressed `.avec` objects under `.arcana/vector-cache/<embedding-identity>/`, including object keys, checksums, finite-value validation, and atomic repair.
- `build.rs` opens Arcana `CURRENT`, validates/reuses exact indexes, reuses cache objects across snapshots, embeds missing documents in bounded concurrent batches, materializes deterministic index files, and refuses publication if `CURRENT` changes.
- `index.rs` owns index format version 3, paths, manifest parsing, snapshot/model/policy identity matching, and full build/status validation of `manifest.json`, `nodes.jsonl`, and `vectors.f32`.
- `search.rs` pins the current or expected snapshot, performs cheap open checks, embeds the query, streams records/vectors in one scoring pass, finite-checks values, and sorts hits deterministically by score then node key.
- `error.rs` owns embedding-service errors, while `http.rs` owns HTTP-specific failures and `index.rs` owns vector-index state/corruption errors.

Vector tests are `arcana/src/vector/index_tests.rs`, `arcana/src/vector/documents_tests.rs`, and inline tests beside narrow implementation details. Start in `documents.rs` for indexed-object policy or rendering, `cache.rs` for reuse identity/object format, `index.rs` plus `build.rs` for index format/publication, and `search.rs` for ranking/open behavior.

## Synthetic graphs and benchmarks

`arcana/src/synthetic/` owns deterministic graph data used by storage/snapshot benchmarking and tests:

- `spec.rs` owns scale tiers, topology specifications, and validation.
- `generator.rs` dispatches generation to `modular.rs`, `entangled.rs`, `hub_heavy.rs`, `layered.rs`, or `dense.rs`.
- `sampling.rs` supplies deterministic sampling primitives.
- `mutation/` owns deterministic mutation selection, replacement, application, and invariants.

`arcana/src/benchmark/` owns the measured comparison:

- `cli.rs` parses benchmark tiers, topologies, seeds, sample/query counts, work paths, CSV paths, and retention flags.
- `common.rs` owns benchmark configuration and shared workload setup.
- `mutation_plan.rs` defines the standard mutation workloads.
- `workload.rs` and `mutation_query.rs` generate and measure shared forward/reverse query workloads.
- `mutation_runner.rs` generates the base once, writes/opens the shared packed base, measures overlay and rebuilt-packed mutation paths in alternating order, validates visible checksum equivalence, measures validated reopen, and compares query fingerprints.
- `mutation_files.rs` owns temporary benchmark artifact lifecycle.
- `report.rs` owns samples, human summaries, medians/throughput, and CSV serialization.
- `error.rs` owns benchmark failures and mismatch reporting.

Tests are in `arcana/src/synthetic/spec_tests.rs`, `arcana/src/synthetic/mutation/tests.rs`, inline synthetic module tests, `arcana/src/benchmark/cli_tests.rs`, `arcana/src/benchmark/mutation_runner_tests.rs`, and inline report tests in `arcana/src/benchmark/mod.rs`.

## Main runtime and data flows

### Lexicon snapshot to published Arcana state

1. `cli.rs` parses `sync`; `main.rs` calls `cli_sync::run_sync`.
2. `cli_sync_state.rs` acquires `.arcana/LOCK`.
3. `lexicon/snapshot.rs` resolves and verifies Lexicon `CURRENT`, its manifest, and every referenced object; `binary.rs` or `object.rs` decodes each object.
4. `lexicon/records.rs` converts records into normalized `RepositoryFacts` and compatibility warnings.
5. `cli_sync.rs` compares the prior Lexicon snapshot. A shared-object change forces rebuild. File-object changes attempt `repository/incremental.rs`; unchanged node identities permit an overlay, while unsupported/inconsistent incremental state falls back to rebuild.
6. Rebuild flows through `repository/compiler.rs` -> `storage/writer.rs` -> `snapshot/graph.rs` -> repository metadata/publication. Overlay flows through `repository/ownership.rs` and `incremental.rs` -> `snapshot/overlay_writer.rs` -> `snapshot/graph.rs` -> repository metadata/publication.
7. The immutable snapshot directory is renamed into place, `lexicon.snapshot` and any `compatibility.warnings` are retained, and `.arcana/CURRENT` is replaced atomically last.

### Repository snapshot to deterministic query response

1. `cli_protocol.rs` opens `protocol::ProtocolSnapshot`.
2. `repository/repository_snapshot.rs` verifies the complete generation; `snapshot/graph.rs` validates the packed base and overlay.
3. `protocol/server.rs` reads one request line; `session.rs` parses and routes it.
4. The owning query module reads catalogue indexes and visible graph adjacency. `snapshot/graph.rs` returns packed iterators directly without an overlay or merged visible neighbors with one.
5. `response.rs` and `session.rs` return one envelope containing the request ID and `arcana.query.v1`.

### Explicit vector build and search

1. `cli_vectors.rs` creates an `EmbeddingClient` and calls `vector/build.rs` or `vector/search.rs`.
2. Both resolve `.arcana/CURRENT` and open its validated repository snapshot.
3. Build renders eligible documents, reuses valid cache objects, embeds only misses, writes a temporary index, verifies it, checks `CURRENT`, and publishes it.
4. Search verifies snapshot/model/policy identity, embeds the query, scores the persisted normalized vectors, checks `CURRENT` again, and returns bounded hits. Exact graph traversal remains a separate protocol step.

## Where to begin for common changes

| Change | Begin here | Also inspect |
| --- | --- | --- |
| Add or change a CLI command/flag | `arcana/src/cli.rs` | `arcana/src/main.rs`, the narrow `cli_*.rs` owner, `cli_tests.rs` |
| Change Lexicon snapshot/object acceptance | `arcana/src/lexicon/snapshot.rs` | `format.rs`, `binary.rs`, `object.rs`, Lexicon tests |
| Change label/identity conversion or compatibility warnings | `arcana/src/lexicon/records.rs` | `repository/model.rs`, `cli_sync_tests.rs`, `LEXICON_CONTRACT.md` |
| Add a node or relation kind | `arcana/src/repository/model.rs` | `compiler.rs`, `protocol/traversal.rs`, `protocol/queries.rs`, `vector/documents.rs`, fact/contract tests |
| Change dense compilation/catalogue behavior | `arcana/src/repository/compiler.rs` | `catalogue.rs`, compiler/catalogue tests |
| Change file-scoped update rules | `arcana/src/repository/ownership.rs` | `incremental.rs`, `cli_update.rs`, sync/update tests |
| Change packed bytes or validation | `arcana/src/storage/format.rs` | `writer.rs`, `reader.rs`, corruption/round-trip tests |
| Change overlay semantics | `arcana/src/snapshot/overlay_validation.rs` | `overlay_writer.rs`, `overlay.rs`, `graph.rs`, overlay/graph tests |
| Change repository publication/validation | `arcana/src/repository/repository_snapshot.rs` | `repository_snapshot_validation.rs`, format/error files, snapshot tests |
| Add a protocol operation | `arcana/src/protocol/request.rs` | `session.rs`, the narrow query owner, `response.rs`, `protocol/tests.rs` |
| Change graph export | `arcana/src/protocol/graph_export.rs` | `request.rs`, `session.rs`, protocol tests |
| Change traversal depth/relation policy | `arcana/src/protocol/traversal.rs` | `path_queries.rs`, `analysis_queries.rs`, `architecture_queries.rs` |
| Change semantic eligibility/documents | `arcana/src/vector/documents.rs` | policy identity in `index.rs`, document/index tests, `vector-index.md` |
| Change vector caching/build publication | `arcana/src/vector/cache.rs` or `build.rs` | `index.rs`, `index_tests.rs` |
| Change semantic scoring/open behavior | `arcana/src/vector/search.rs` | `client.rs`, `index.rs`, `index_tests.rs` |
| Add a synthetic topology | `arcana/src/synthetic/spec.rs` | `generator.rs`, a topology owner, benchmark CLI/tests |
| Change benchmark methodology | `arcana/src/benchmark/mutation_runner.rs` | `mutation_plan.rs`, `mutation_query.rs`, `report.rs`, benchmark tests |

## Implemented versus planned or absent

Implemented now:

- Binary v1 and legacy canonical JSON Lexicon object ingestion, plus legacy/manual repository fact import paths.
- Deterministic dense compilation, packed forward/reverse storage, immutable graph and repository manifests, edge overlays, overlay-aware reads, and library-level compaction.
- One-shot and registered-consumer Lexicon sync with serialized writers and atomic Arcana `CURRENT` publication.
- The JSONL protocol operations present in `protocol/request.rs`, including graph export, traversal/analysis, architecture summary, statistics, unresolved queries, and snapshot diff.
- Explicit vector build/cache/search through Arcana’s CLI and library.
- Deterministic synthetic generation and overlay-versus-rebuild benchmarks.

Explicitly planned, transitional, or absent from the current implementation:

- Language adapters and parsing are outside Arcana; there is no Arcana adapter directory to extend.
- `update-facts` still takes a complete replacement fact file and selects declared file partitions from it. Direct adapter-produced file-scoped fact batches are described as a later boundary, not a current input mode.
- A provenance sidecar is mentioned as a later possibility in `LEXICON_CONTRACT.md`; no provenance sidecar writer/reader exists in the current Arcana source tree.
- Unresolved references are stored and queryable, but the “later resolver passes” mentioned in snapshot documentation are not implemented as an Arcana resolver command/module.
- Compaction exists as a library operation in `snapshot/compaction.rs`; there is no `arcana compact` CLI command in `cli.rs`.
- Overlays cannot add/remove dense nodes. `update-facts` returns a rebuild-required error for node-set changes; `sync` handles that case by rebuilding.
- Semantic indexing is not automatic during `sync`, is not required by the deterministic protocol, and is not automatically consumed by Grimoire’s active discovery interface.
- Arcana is a process/CLI and Rust-library boundary. No Go FFI/cgo integration is implemented.
