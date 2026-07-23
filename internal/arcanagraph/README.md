# Arcana Graph Provider

`internal/arcanagraph` owns Grimoire's optional process boundary to Arcana.

## Owns

- resolving the Arcana snapshot matching the Lexicon snapshot used by a query;
- one-shot `arcana sync` when graph state is missing or stale;
- JSONL request/response handling for `arcana.query.v1`;
- bounded symbol resolution, operational-role, impact, unresolved-reference, and shortest-call-chain queries; and
- conversion from Arcana protocol results into Grimoire structural evidence.

## Does not own

- graph storage or traversal algorithms;
- Lexicon fact generation;
- task relevance scoring;
- source-candidate curation; or
- final package budgeting.

Arcana remains a standalone process boundary. Grimoire does not link it through FFI.
