# ArcanaGraph

ArcanaGraph is an independent repository-graph project. Its role is to model
repositories as a queryable graph and to provide the storage and snapshot
foundations needed to inspect repository structure and relationships.

## Ownership boundaries

- **ArcanaGraph** owns the repository graph model, graph storage, snapshots,
  forward/reverse queries, and measurements of storage representations.
- **Demon Docs** owns documentation content and documentation-specific
  authoring, maintenance, and publishing workflows. It may consume graph data,
  but it does not own ArcanaGraph's graph persistence or query implementation.
- **Context Grimoire** owns context composition and retrieval workflows for its
  consumers. It may query ArcanaGraph, but it does not own the repository graph
  model or its storage lifecycle.

Keeping these boundaries explicit lets ArcanaGraph remain useful as a small,
standalone repository service rather than becoming a subsystem of either
neighboring project.

## First storage proof of concept

The first implementation milestone will establish a measurable storage seam
with:

1. deterministic fixture import;
2. dense IDs for imported repository entities;
3. a content-addressed snapshot;
4. forward and reverse queries; and
5. packed-versus-SQLite measurement.

This repository currently contains only the Rust package and its reusable
library boundary. No graph implementation is included yet.

## Development

The package uses Rust edition 2024 and has no third-party dependencies.

```text
cargo fmt -- --check
cargo check
cargo test
cargo run -- --help
cargo run -- --version
```
