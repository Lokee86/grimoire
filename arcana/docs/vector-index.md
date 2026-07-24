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
  [--batch-size <N>]
```

Defaults:

- state: `.arcana`
- endpoint: `http://127.0.0.1:9876/v1`
- batch size: `32`
- model identity: `qwen3-embedding-0.6b-q8_0-512d`
- retained dimensions: `512`

The command reuses a complete matching index. It rebuilds when the current Arcana graph snapshot, graph identity, model, model identity, or dimensions differ.

## Indexed objects

Arcana creates one deterministic document per graph node. Each document contains:

- node kind, name, path, and source span;
- bounded outgoing relationships and target identities;
- bounded incoming relationships and source identities; and
- bounded unresolved-reference evidence.

The embedding finds a relevant graph neighborhood. Arcana's exact graph protocol then resolves and expands the returned nodes. Vectors never replace authoritative graph relationships.

## Storage

Indexes are stored by Arcana snapshot and model identity:

```text
.arcana/
  vectors/
    <arcana-snapshot-digest>/
      qwen3-embedding-0.6b-q8_0-512d/
        manifest.json
        nodes.jsonl
        vectors.f32
```

`manifest.json` binds the index to:

- repository snapshot ID;
- graph snapshot ID;
- embedding model and stable model identity;
- vector dimensions; and
- item count and data filenames.

`vectors.f32` contains normalized little-endian `f32` vectors. `nodes.jsonl` maps vector positions back to Arcana node keys, kinds, paths, and names.

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

## Grimoire Context integration

When Arcana is enabled for a context request, Grimoire checks for a vector index matching Arcana's current snapshot and the configured embedding identity.

If present, Grimoire:

1. asks Arcana for semantic graph matches using the same embedding endpoint supplied to Grimoire;
2. merges those matches with Lexicon-derived symbol seeds;
3. resolves the combined seeds through `arcana.query.v1`; and
4. requests deterministic operational-role, impact, unresolved-reference, and call-chain evidence.

Grimoire does not automatically build the Arcana vector index. If no matching index exists, it silently continues with Lexicon-seeded Arcana traversal. If semantic querying fails after a matching index is found, Grimoire warns and continues with the remaining structural and source retrieval paths.
