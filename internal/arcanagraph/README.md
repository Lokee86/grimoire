# Arcana Graph Provider

`internal/arcanagraph` owns Grimoire's optional process boundary to Arcana.

## Owns

- resolving the Arcana snapshot matching the Lexicon snapshot used by a query;
- one-shot `arcana sync` when graph state is missing or stale;
- JSONL request/response handling for `arcana.query.v1`;
- optional semantic seed retrieval from an existing Arcana-owned vector index;
- bounded symbol resolution, operational-role, impact, unresolved-reference, and shortest-call-chain queries; and
- conversion from Arcana protocol results into Grimoire structural evidence.

## Does not own

- graph, graph-vector, or traversal storage;
- Arcana vector-index creation;
- Lexicon fact generation;
- task relevance scoring;
- source-candidate curation; or
- final package budgeting.

Arcana remains a standalone process boundary. Grimoire does not link it through FFI. When a vector manifest matching the current Arcana snapshot and embedding identity exists, this provider asks `arcana semantic-query --json` for semantic graph seeds, merges them with Lexicon seeds, and expands the combined set through the deterministic Arcana protocol. It does not create vector state during a context request.
