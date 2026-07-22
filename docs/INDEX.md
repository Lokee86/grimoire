# Grimoire Documentation

The root [README](../README.md) is the product introduction and quick start. This documentation tree separates implemented architecture and exact reference material from current limitations and future plans.

## Documentation rules

- Current behavior belongs in architecture, reference, development, or limits documentation.
- Future, unresolved, or dependency-blocked work belongs under `planning/`.
- A planned capability must not be described as current behavior.
- Architecture documents describe ownership boundaries and include code maps where useful.
- Reference documents describe exact user-visible contracts.
- Package READMEs describe code-level ownership and non-ownership.

## Architecture

- [Architecture index](architecture/INDEX.md)
- [System overview](architecture/system-overview.md)
- [Prepared index](architecture/prepared-index.md)

## Reference

- [Reference index](reference/INDEX.md)
- [CLI](reference/cli.md)
- [Indexing](reference/indexing.md)
- [Context package](reference/context-package.md)
- [Embedding model](reference/embedding-model.md)
- [Vector store](reference/vector-store.md)

## Development

- [Development index](development/INDEX.md)
- [Testing and benchmarks](development/testing-and-benchmarks.md)
- [Retrieval quality and latency baselines](development/retrieval-quality.md)

## Limits

- [Limits index](limits/INDEX.md)
- [Current limitations](limits/current-limitations.md)

## Planning

- [Planning index](planning/INDEX.md)
- [Roadmap](planning/roadmap.md)

## Code seam documentation

- [`internal/app`](../internal/app/README.md)
- [`internal/index`](../internal/index/README.md)
- [`internal/ignore`](../internal/ignore/README.md)
- [`internal/retrieve`](../internal/retrieve/README.md)
- [`internal/selection`](../internal/selection/README.md)
- [`internal/compiler`](../internal/compiler/README.md)
- [`internal/embedding`](../internal/embedding/README.md)
- [`internal/vectorstore`](../internal/vectorstore/README.md)
- [`native/vector-engine`](../native/vector-engine/README.md)
