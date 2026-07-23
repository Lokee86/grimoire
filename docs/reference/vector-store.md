# Vector Store

## Purpose

Grimoire persists normalized chunk embeddings in a custom Rust storage engine rather than a hosted or general-purpose vector database.

## State location

For the fixed embedding identity, vector state is stored below:

```text
<state>/vectors/qwen3-embedding-0.6b-q8_0-512d/
```

The directory contains an immutable object store, the current packed snapshot, and a persistent snapshot manifest. Embedding batches are ingested into immutable objects as they complete, allowing interrupted builds to reuse completed work. Temporary JSONL ingest and record-list files are removed after use.

## Object identity and reuse

Go computes SHA-256 over the exact chunk text. Rust addresses the vector object using BLAKE3 over the vector format identity, embedding identity, and that source-content hash.

Consequences:

- unchanged text reuses its existing embedding even if its path or line range changes;
- identical text can share one vector object;
- changing the model identity or vector contract produces a different object address;
- an existing address cannot be overwritten with different vector bytes; and
- deleted chunks disappear from the next snapshot without deleting reusable immutable objects.

## Snapshot format

Snapshot format version 1 stores:

- fixed magic and version fields;
- embedding identity;
- dimensions and vector count;
- sorted chunk IDs and source-object hashes; and
- a 64-byte-aligned contiguous `float32` vector matrix.

The engine validates section bounds, integer overflow, UTF-8 IDs, duplicate IDs, alignment, exact matrix length, and finite vector values before exposing a snapshot.

## Snapshot manifest and freshness

`snapshot.manifest.json` binds the packed vector snapshot to:

- the exact content-addressed prepared-index root;
- the packed snapshot identity;
- the embedding identity;
- dimensions; and
- vector count.

Both `vector search` and `context` validate the prepared identity before embedding the query. Any repository reindex that changes prepared content produces a different identity, even when the total chunk count is unchanged. `vector search` rejects the stale snapshot; `context` warns and uses its lexical failure fallback until `vector build` publishes a matching snapshot and manifest.

## Search

Vectors and query embeddings are L2-normalized by Grimoire. The engine therefore uses exact dot product as cosine similarity.

Small snapshots scan serially. Larger snapshots are divided into coarse contiguous partitions. Each worker keeps a bounded local top-K heap, followed by a deterministic merge ordered by descending score and ascending snapshot index.

No approximate-nearest-neighbour index is used. That remains deferred until repository-scale benchmarks show an exact scan is inadequate.

## C ABI memory contract

- All strings are UTF-8 pointer-and-length pairs.
- Go owns query, result, ID, and metadata buffers.
- Rust does not retain foreign pointers after a call.
- Rust allocations are never returned for Go to free.
- Snapshot handles are numeric registry keys rather than raw pointers.
- Active searches hold an internal reference even if another caller closes the handle.
- Panics are converted to ABI errors.

## Commands

```bash
grimoire vector build --root /path/to/repository
grimoire vector search --root /path/to/repository --query "where is damage resolved"
grimoire vector info --root /path/to/repository
```

`vector build` requires a prepared source index, a running embedding endpoint, and the Rust library. Repeated builds embed only missing source-content identities. Grimoire first checks for the DLL beside its executable, then searches `native/vector-engine/target/{release,debug}` below ancestors of both the executable and current working directory.

## Platform coverage

The Rust core is portable. The current Go dynamic-library loader is implemented for Windows. `GRIMOIRE_VECTOR_ENGINE` can point to an explicit DLL; source builds also discover debug and release DLLs under `native/vector-engine/target`.
