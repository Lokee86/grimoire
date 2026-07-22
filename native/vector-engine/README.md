# Grimoire Vector Engine

This Rust workspace owns Grimoire's persistent vector objects, packed search snapshots, memory-mapped validation, and exact nearest-neighbour search.

## Workspace

```text
crates/core   immutable objects, snapshot format, mmap validation, parallel search
crates/ffi    narrow C ABI with caller-owned buffers and opaque numeric handles
crates/cli    diagnostic ingest, build, inspect, and search commands
```

The Go application remains responsible for repository chunking, embedding requests, source-hash calculation, command orchestration, and presentation of results.

## Build

```bash
cargo build -p grimoire-vector-ffi --release
cargo build -p grimoire-vector-cli --release
```

On Windows the library is emitted as `target/release/grimoire_vector_ffi.dll`. Grimoire discovers it in the workspace, beside the Go executable, or through `GRIMOIRE_VECTOR_ENGINE`.

## Storage model

Each immutable vector object is addressed by BLAKE3 over:

```text
format identity + embedding identity + source-content hash
```

The object stores the embedding identity, source identity, dimensions, and little-endian `float32` vector. Reusing the same model/source address with different vector bytes is rejected.

A materialized snapshot contains:

- a fixed versioned header;
- embedding identity;
- sorted chunk-ID entries and object hashes;
- a compact UTF-8 chunk-ID table; and
- one 64-byte-aligned contiguous `float32` vector matrix.

Snapshots are validated before use and then memory-mapped. Search performs exact dot products, uses a bounded per-worker top-K heap, scans small snapshots serially, and partitions larger snapshots through Rayon.

## ABI ownership

The ABI intentionally avoids cross-runtime allocation ownership:

- Go owns every input and output buffer.
- Rust borrows those buffers only for the duration of one call.
- Rust never retains a Go pointer.
- Rust-owned snapshots live behind numeric handles in a synchronized registry.
- Search clones an `Arc` before scanning, so closing a handle cannot invalidate an active search.
- Closing removes the handle; later calls return an invalid-handle error.
- Panics are caught before crossing the ABI.

See `crates/ffi/include/grimoire_vector.h` for the exact exported contract.

## Verification

```bash
cargo fmt --all --check
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo build -p grimoire-vector-ffi --release
```

The Go suite includes a real DLL integration test and an application-level index/embed/reuse/search test. Sanitizer builds should also be run on supported nightly toolchains before release; the ABI is deliberately small enough for exhaustive open/search/close stress and malformed-snapshot testing.

## Current encoding

Version 1 uses aligned `float32` vectors. Quantized `float16`, `int8`, or narrower encodings are intentionally deferred until benchmarks measure both retrieval quality and total search latency. The format has an explicit version boundary so a later encoding does not silently mix with version 1 data.
