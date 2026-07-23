# Vector-store bridge

`internal/vectorstore` owns the Go side of the Rust vector-engine ABI.

## Responsibilities

- Locate and load the native vector library.
- Validate the ABI version.
- Check immutable object existence.
- Invoke JSONL object ingestion and packed-snapshot materialization.
- Open, inspect, search, and close snapshot handles.
- Manage caller-owned result and ID buffers, including retry on insufficient capacity.
- Convert native hits into Go values.
- Protect handle lifetime with Go synchronization.
- Apply `runtime.KeepAlive` to every borrowed Go buffer.

The production bridge currently uses the Windows DLL ABI. Non-Windows builds return `ErrUnavailable` until equivalent platform loaders exist.

## Concurrency

Snapshot `Info` and `Search` calls hold the engine read lock. `Close` holds the write lock and is idempotent. The native engine clones shared snapshot ownership before a scan, so an active search remains valid while a handle is being removed from the registry.

Object ingestion is called serially by the current application workflow. The bridge does not define a multi-writer transaction API.

## Boundary

This package does not own embedding generation, chunk identity, storage formats, similarity algorithms, ranking, or context selection. Those remain in `embedding`, `index`, the Rust engine, `retrieve`, and the package pipeline.
