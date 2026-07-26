# Arcana Semantic Graph Index

Arcana can build an optional semantic index over the current immutable repository graph. The index provides semantic entry points into Arcana's deterministic graph traversal without moving graph ownership into Grimoire Context.

## Ownership

- Arcana owns graph-document generation, vector persistence, index invalidation, and semantic graph search.
- Grimoire Context owns the existing embedding model runtime and endpoint.
- Arcana requests embeddings from that endpoint; it does not install or load a second model.
- Lexicon remains the authority for language facts and source identities.

The ordinary `arcana sync`, graph protocol, and packed snapshots remain embedding-free. A missing embedding server or vector index does not affect deterministic graph construction or graph queries.

## Build the index

Start the existing Grimoire embedding service, synchronize Arcana, then build the semantic graph index explicitly:

```text
grimoire model serve
arcana sync
arcana vectorize
```

Options:

```text
arcana vectorize \
  [--state <DIRECTORY>] \
  [--endpoint <URL>] \
  [--batch-size <N>] \
  [--batch-concurrency <N>]
```

Defaults:

- state: `.arcana`
- endpoint: `http://127.0.0.1:9876/v1`
- batch size: `32`
- batch concurrency: `1`
- model identity: `qwen3-embedding-0.6b-q8_0-512d`
- retained dimensions: `512`

The command first validates an exact existing snapshot, including its semantic-eligibility policy. Otherwise it resolves every eligible graph document to the immutable content-addressed cache, sends only missing documents to the endpoint, and materializes the snapshot files from cache objects. `--batch-concurrency` bounds simultaneous embedding requests; each successful batch writes and synchronizes its objects immediately, so a retry after an endpoint or process interruption resumes from completed work. Snapshot materialization remains serialized and deterministic.

Build output reports indexed-document and unique-vector counts, embedded and reused vectors, exact-snapshot reuse, embedding request count, snapshot bytes, elapsed milliseconds, and the published directory. The exact-snapshot path makes no embedding request. Cross-snapshot reuse occurs only when rendered graph-document bytes and the complete embedding contract are identical.

Build validation covers the manifest, semantic-eligibility policy, fixed data filenames, vector byte length, finite values, indexed-document record count and identities, and data checksums. Incomplete, corrupt, or policy-incompatible snapshots are rebuilt. Publication is serialized per snapshot and embedding identity, preserves the prior index if replacement fails, and aborts with a retry message rather than publishing under the wrong snapshot if `.arcana/CURRENT` changes during embedding.

## Indexed objects

Arcana constructs the complete repository graph, then applies semantic-eligibility policy version 1 only while generating index documents. The policy excludes `variable`, `parameter`, `field`, `import`, `export`, `directory`, and `repository` nodes. It includes `function`, `method`, `constructor`, `test`, `file`, `type`, `interface`, `module`, `namespace`, `message-channel`, `http-endpoint`, `config-key`, `signal`, and `constant` nodes. Existing `trait` and generic `symbol` declarations are also included because they are type-like or adapter-neutral declaration entry points. Every current node kind is matched explicitly, so adding a kind requires a deliberate policy decision.

These exclusions remove high-volume implementation detail that is usually reached more precisely by traversing from a declaration-level semantic match. They do not remove nodes or edges from Arcana snapshots, so exact traversal can still reach variables, parameters, fields, imports, exports, directories, and the repository root.

Arcana creates one deterministic document per eligible graph node, ordered by stable node key. Each document contains:

- node kind, name, path, and source span;
- bounded outgoing relationships and target identities;
- bounded incoming relationships and source identities; and
- bounded unresolved-reference evidence.

The embedding finds a relevant graph neighborhood. Arcana's exact graph protocol then resolves and expands the returned nodes. Vectors never replace authoritative graph relationships.

## Storage

Reusable vector objects are stored independently from indexes, which remain keyed by Arcana snapshot and model identity:

```text
.arcana/
  vector-cache/
    qwen3-embedding-0.6b-q8_0-512d/
      objects/
        <digest-prefix>/
          <document-and-embedding-digest>.avec
  vectors/
    <arcana-snapshot-digest>/
      qwen3-embedding-0.6b-q8_0-512d/
        manifest.json
        nodes.jsonl
        vectors.f32
```

The cache key is SHA-256 over a domain separator, rendered graph-document bytes, model reference, stable embedding identity, and dimensions. Node keys are intentionally absent from rendered documents, so a node-key rename can reuse an object; a changed name, path, span, relationship neighborhood, unresolved reference, model, identity, or dimension cannot. Objects contain their content key, dimensions, vector checksum, and normalized little-endian `f32` values. Invalid objects are re-embedded and atomically repaired. Cache objects are immutable once valid.

`manifest.json` binds the index to:

- repository snapshot ID;
- graph snapshot ID;
- embedding model and stable model identity;
- semantic-eligibility policy version and a composite index identity containing that version;
- vector dimensions; and
- item count, fixed data filenames, and SHA-256 checksums of both data files.

`vectors.f32` contains normalized little-endian `f32` vectors materialized in deterministic node order. `nodes.jsonl` maps vector positions back to Arcana node keys, kinds, paths, and names. A snapshot manifest also records the unique-vector count when produced by the incremental builder.

The current index format is version 3 and semantic-eligibility policy version is 1. Its composite identity is `<embedding-identity>-arcana-semantic-v1`, while the containing directory remains keyed by embedding identity. Version 1 manifests lack data checksums; version 2 manifests predate semantic eligibility. Both are rebuilt rather than reused.

## Expected scope reduction

The policy is intentionally a substantial reduction, not a graph-size change. In the measured live states that motivated it:

- Grimoire had 30,260 graph nodes, including 22,295 variables. Excluding variables alone caps the semantic index at 7,965 documents, at least a 73.7% reduction; the other excluded kinds reduce it further.
- Space Rocks had 64,069 graph nodes, including 31,218 variables and 6,516 parameters. Excluding those two kinds alone caps the semantic index at 26,335 documents, at least a 58.9% reduction; fields, imports, exports, directories, and the repository node reduce it further.

Manifest `item_count`, vector byte-length validation, record-count validation, build summaries, and `grimoire status` all report and validate the indexed-document count. They do not compare that count with the complete graph node count.

## Query the index

Human-readable output:

```text
arcana semantic-query --query "where is profile persistence handled?"
```

Machine-readable output:

```text
arcana semantic-query \
  --query "where is profile persistence handled?" \
  --limit 10 \
  --json
```

The JSON response has this shape:

```json
{
  "matches": [
    {
      "score": 0.72,
      "node_key": "0123456789abcdef",
      "kind": "function",
      "path": "internal/profile/repository.go",
      "name": "InsertProfile"
    }
  ]
}
```

Semantic query performs cheap manifest and file-size checks when opening the pinned snapshot, then decodes and finite-checks each vector in the same single scoring pass. It does not checksum or pre-scan the complete vector file on every query. Full checksums and exhaustive structural validation remain build and `grimoire status` boundary work.

## Grimoire Context integration

Process integrations can pass `--expected-snapshot sha256:<digest>` to `semantic-query`. Arcana then rejects the query if `.arcana/CURRENT` no longer matches the graph snapshot that the caller already resolved. This prevents semantic seeds from one graph snapshot being expanded through another.

When Arcana is enabled for a context request, Grimoire checks for a vector index matching the exact Arcana snapshot already selected for deterministic traversal and the configured embedding identity.

If present, Grimoire:

1. asks Arcana for semantic graph matches using the same embedding endpoint supplied to Grimoire;
2. merges those matches with Lexicon-derived symbol seeds;
3. resolves the combined seeds through `arcana.query.v1`; and
4. requests deterministic operational-role, impact, unresolved-reference, and call-chain evidence.

Grimoire does not automatically build the Arcana vector index. If no matching index exists, it silently continues with Lexicon-seeded Arcana traversal. If semantic querying fails after a matching index is found, including a concurrent snapshot change, Grimoire warns and continues with the remaining structural and source retrieval paths. Provider discovery uses explicit command overrides, repository configuration, executables installed beside Grimoire, a discoverable Grimoire checkout, and only then `PATH`.
