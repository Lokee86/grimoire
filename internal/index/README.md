# Prepared Index

`internal/index` owns repository traversal, file eligibility, fallback chunking, incremental file records, and private prepared-state persistence.

## Owns

- supported-file filtering and size limits;
- permanent metadata and tool-state exclusions;
- default generated, vendored, lockfile, bundle, and minified-content filtering;
- application of `internal/ignore` policy;
- SHA-256 content identity and unchanged-record reuse;
- deterministic fallback chunks and chunk IDs;
- exact 1,536-token post-chunk safety enforcement with line-first and token-slice fallback splitting;
- exact `o200k_base` counts stored with changed chunks;
- prepared-index tokenizer identity and incompatible-index rebuild bases;
- deterministic published snapshot identity from the content-addressed root;
- prepared snapshot, file, and chunk models;
- binary shard and file codecs;
- go-git object storage and validation; and
- compare-and-swap publication through `refs/grimoire/state`.

## Does not own

- query interpretation or ranking;
- context-package budget selection;
- language adapters or syntax-aware ranges;
- semantic embeddings; or
- daemon lifecycle and file watching.

## Main files

- `build.go` - traversal, filtering, reuse, update, and removal detection.
- `exclusions.go` - permanent directory and explicit state-path exclusions.
- `generated.go` - default generated/dependency path, filename, header, and minification policy.
- `chunk.go` - fallback chunking, chunk identity, exact token counting, and hard token-ceiling splitting.
- `model.go` - snapshot models.
- `store.go` - load, validate, save, and publish.
- `repository.go` - private repository lifecycle and state-reference helpers.
- `objects.go` - Git blobs, root tree, and manifest.
- `codec.go` - shard encoding and path validation.
- `file_codec.go` - file/chunk record encoding.

## Related documentation

- [Prepared-index architecture](../../docs/architecture/prepared-index.md)
- [Indexing reference](../../docs/reference/indexing.md)
- [Current limitations](../../docs/limits/current-limitations.md)
