# ArcanaGraph

ArcanaGraph is an independent Rust engine for building and querying deterministic repository graphs.

It provides shared repository facts to tools such as Demon Docs and Context Grimoire without making either project own the graph substrate.

## Ownership boundaries

ArcanaGraph owns:

- Stable logical identities for repository entities.
- Typed, directed relationships with provenance.
- Deterministic graph snapshots.
- Content-addressed immutable storage.
- Forward, reverse, path, and bounded-neighbourhood queries.
- Provider ingestion for filesystem, Git, SCIP, and other factual evidence.

ArcanaGraph does not own:

- Documentation policy or Codemap decisions. Those remain in Demon Docs.
- Context ranking, token budgets, or agent rendering. Those remain in Context Grimoire.
- MCP workflow orchestration or model-specific behaviour.

Demon Docs may use ArcanaGraph evidence to generate and review Codemap suggestions. Context Grimoire may select task-relevant projections from the same graph.

## Status

ArcanaGraph is at the storage proof-of-concept stage. The repository currently contains only the project foundation and CLI shell.

## First milestone

The first implementation milestone is deliberately narrow:

1. Import a fixed JSON graph fixture deterministically.
2. Assign deterministic dense node identifiers.
3. Write a content-addressed snapshot.
4. Read the snapshot back.
5. Query forward and reverse neighbours.
6. Compare the packed representation with a SQLite reference implementation.

The packed format must earn its complexity through measured results rather than assumption.

## Development

Requirements:

- Stable Rust toolchain with Cargo.

Useful commands:

```sh
cargo fmt --all -- --check
cargo check --all-targets
cargo test --all-targets
cargo run -- --help
cargo run -- --version
```

Runtime graph data will live under `.arcana/` and is intentionally excluded from Git.

## Storage direction

The expected production shape is:

```text
content-addressed immutable graph objects
+ locality-aware packed adjacency
+ compact positional indexes
+ Merkle snapshot roots
+ append-only dirty-worktree overlays
```

The on-disk format is private to ArcanaGraph. Consumers will integrate through a versioned machine-readable command or local-service protocol.

## License

A license has not yet been selected.
