# Vector store

Grimoire stores normalized embeddings through a Rust native engine exposed to Go by `internal/vectorstore`.

## State location

For the fixed embedding identity, vector state is stored below:

```text
<state>/vectors/qwen3-embedding-0.6b-q8_0-512d/
```

The directory contains immutable objects, the current packed snapshot, and `snapshot.manifest.json`. Temporary JSONL ingestion and record-list files are removed after use.

## Layout and identities

The engine maintains:

- immutable content-addressed vector objects;
- a manifest binding prepared chunks to vector source identities; and
- a sorted packed snapshot used for memory-mapped exact search.

An object address includes the storage format identity, embedding identity, and source-content hash. Reusing the same model/source address with different vector bytes is rejected. Identical chunk text can therefore reuse one stored vector across multiple chunks.

## Build and publication

`grimoire vector build` sends missing embeddings to the native `IngestJSONL` boundary in serialized batches. Each batch writes immutable objects. After all required vectors exist, Grimoire writes the complete chunk manifest and asks the engine to materialize the packed snapshot.

The current manifest is published only after successful materialization. A failed build does not make an incomplete manifest current, while immutable vector objects completed before the failure remain reusable.

## Snapshot reads

A snapshot contains a versioned header, embedding identity, dimensions and vector count, sorted chunk entries and object hashes, a compact UTF-8 chunk-ID table, and one 64-byte-aligned `float32` matrix. Before use, the engine validates section bounds, integer overflow, UTF-8 IDs, duplicate IDs, alignment, exact matrix length, and finite vector values, then memory-maps the snapshot.

Snapshot handles are opaque native values. Search clones native shared ownership before scanning, so closing a handle cannot invalidate an active search. The Go bridge protects handle lifetime with a read/write lock and keeps borrowed Go buffers alive for every ABI call.

## Search

```bash
grimoire vector search --query "where is damage resolved"
```

The engine performs exact inner-product search over normalized 512-dimensional vectors. Small snapshots scan serially; larger snapshots partition work through Rayon. Query plans may produce multiple vectors; Grimoire searches them concurrently, keeps the best score per chunk, and returns deterministic score/path ordering.

The packed format is an exact-search representation, not an approximate nearest-neighbour index.

## ABI contract

- Strings cross as UTF-8 pointer-and-length pairs.
- Go owns all query, result, ID, and metadata buffers.
- Rust borrows foreign buffers only for one call and never retains Go pointers.
- Rust allocations are not returned for Go to free.
- Snapshot handles are numeric registry keys rather than raw pointers.
- Panics are converted to ABI errors.

## Compatibility and discovery

A snapshot is accepted only when its manifest agrees with the prepared source identity, embedding identity, dimensions, and vector count. `grimoire vector info` reports native-library and snapshot availability.

On Windows, Grimoire checks `GRIMOIRE_VECTOR_ENGINE`, the executable directory, and `native/vector-engine/target/{release,debug}` beneath workspace ancestors. The Rust core is portable, but equivalent non-Windows Go loaders are not yet implemented.

`grimoire context` degrades to lexical retrieval when semantic state cannot be used. Direct vector commands fail because they have no lexical substitute.

## Ownership boundary

`native/vector-engine` owns immutable objects, packed snapshots, validation, and exact vector search. `internal/vectorstore` owns library discovery, ABI validation, Go handle lifetimes, caller-owned buffers, and conversion to Grimoire types. Embedding, ranking, and package assembly remain outside this boundary.
