# Lexicon ingestion boundary

Arcana consumes Lexicon snapshot contract v1 and compacts its durable identities into a packed repository graph. The legacy complete JSONL importer remains available for migration and diagnostics, but snapshot synchronization is the primary integration boundary.

## Identity boundary

Lexicon owns cross-tool SHA-256 node identities. Arcana stores each full Lexicon identity in the catalogue, hashes it into an internal 64-bit `NodeKey`, checks for compaction collisions during import, and assigns dense packed `NodeId` values during compilation. Dense IDs are snapshot-local and must never escape as durable cross-tool identities. Lexicon file content IDs are compacted for Arcana's internal change detection.

Arcana continues to read its legacy TSV facts during migration, but no language adapter is owned by this repository.

## Preserved semantics

Arcana accepts the common Lexicon node and relation vocabulary, including:

- interfaces, traits, constructors, and parameters;
- definite `calls` and conservative `possible-calls` as separate relations;
- conversions, implementations, inheritance, trait use, overrides, reads, writes, and annotations;
- unresolved references with source spans and candidate metadata.

Source spans are preserved in the catalogue and unresolved-reference store. Explicit file ownership drives Arcana's file-scoped replacement model. Arbitrary Lexicon `attributes` are currently ignored rather than persisted; adding a provenance sidecar later will not require changing graph identity.

## Snapshot synchronization

`arcana sync` resolves Lexicon's atomic `CURRENT` pointer, verifies the content-addressed snapshot manifest and every referenced fact object, and compares file object identities with the Lexicon snapshot consumed by the previous Arcana state. Added, changed, and removed file-object paths become Arcana's file-scoped replacement set. Any language-level shared-object change conservatively forces a packed rebuild, including when file objects changed in the same snapshot.

Arcana stores immutable graph states under `.arcana/snapshots/<lexicon-snapshot-digest>/`. All sync writers share `.arcana/LOCK`, and `.arcana/CURRENT` is replaced atomically only after the new state verifies. The repository manifest identifies Lexicon as the adapter and records the consumed Lexicon snapshot ID as its adapter version, while a `lexicon.snapshot` sidecar makes the relationship explicit.

When node identities remain unchanged, Arcana emits one cumulative overlay against the packed base. Node additions or removals, unusable prior state, unsupported incremental ownership, or any incremental planning failure fall back to a complete packed rebuild. This choice is internal; callers invoke the same `sync` operation in every case.

`arcana sync --register` writes a versioned command definition under `.lexicon/consumers/`. Lexicon invokes that one-shot command after every successful manual or daemon-triggered scan. The event only reduces latency: immutable snapshots remain the durable handoff, so Arcana can also catch up later through an explicit `arcana sync`.

Scoped Lexicon `mode=incremental` JSONL streams remain invalid as complete import input. Arcana derives incremental scope from verified snapshot manifests rather than accepting a partial stream without its surrounding snapshot state.
