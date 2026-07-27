# Knowledge-vector package

`internal/knowledgevector` owns optional semantic ranking for the independent documentation knowledge lane.

## Responsibilities

- Derive a documentation-only identity from document paths, kinds, document hashes, section IDs, and section hashes.
- Build immutable content-addressed vectors for knowledge sections through the shared embedding client.
- Deduplicate identical section text and reuse existing native vector objects.
- Persist successful embedding batches immediately so interrupted builds can resume.
- Publish a packed snapshot and manifest only after every required section vector exists.
- Return an existing current snapshot immediately without object probes or rematerialization.
- Validate knowledge identity, embedding identity, dimensions, vector count, and native snapshot metadata before query embedding.
- Implement `knowledge.VectorRanker` and return section-ID scores only for the filtered candidate set.
- Report availability and freshness for `knowledge vector info`.

## State

State lives under:

```text
<knowledge-state>/vectors/<embedding-identity>/
```

The default knowledge state is `<root>/.grimoire/knowledge`. Repository location, Git commit metadata, and unrelated source fingerprints are intentionally excluded from the vector identity.

## Boundary

This package never embeds production source chunks and contributes only to the independent `document_matches` lane. BM25 ranking remains owned by `internal/knowledge`; native object and snapshot operations remain owned by `internal/vectorstore`; embedding requests remain owned by `internal/embedding`.
