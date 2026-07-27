# Component architecture

Grimoire, Lexicon, and Arcana share one repository but retain separate ownership, state, and advanced command surfaces.

## Grimoire

The repository-root Go application is the primary discovery interface. It owns:

- exact and BM25 source discovery;
- independent documentation discovery;
- stable handles and progressive inspect, trace, and impact operations;
- repository-state preparation and provider routing;
- investigation-session deduplication;
- the CLI and MCP discovery contracts.

Grimoire does not own language parsing or graph semantics.

## Lexicon

`lexicon/` owns:

- one adapter per supported programming language;
- normalized symbols, spans, and relationships;
- immutable analysis objects and snapshots;
- incremental and Git-aware source analysis;
- standalone scan, export, and inspection commands.

Grimoire consumes Lexicon snapshots through immutable exports. Lexicon does not depend on Grimoire or Arcana.

## Arcana

`arcana/` owns:

- ingestion of one immutable Lexicon snapshot;
- packed forward and reverse repository graphs;
- neighbors, paths, impact, call chains, unresolved references, and graph inspection;
- optional semantic graph entry points;
- standalone graph commands and protocol behavior.

Arcana does not own language adapters or Grimoire's discovery response. Optional semantic indexing uses a compatible external embedding endpoint.

## Dependency direction

```text
repository source
    -> Lexicon snapshot
        -> Arcana graph snapshot

repository source and documentation
    -> Grimoire prepared state

Grimoire discovery
    -> reads Lexicon snapshot
    -> queries Arcana graph
    -> returns one provider-neutral response
```

Lexicon is upstream of Arcana. Grimoire may consume both but neither component calls back into Grimoire for deterministic analysis.

## Independent use

- Lexicon can analyze and export source facts without Arcana or Grimoire.
- Arcana can synchronize and answer graph queries without Grimoire.
- Grimoire can return exact, source, and document evidence without Lexicon or Arcana.
- Missing structural providers reduce available lanes and produce warnings; they do not invalidate unrelated evidence.

## State ownership

| State | Owner |
| --- | --- |
| `.grimoire/` source index | Grimoire |
| `.grimoire/knowledge/` documents and optional vectors | Grimoire |
| `.lexicon/` immutable language-analysis snapshots | Lexicon |
| `.arcana/` graph snapshots and optional graph vectors | Arcana |

Consumers interact through immutable manifests, exported facts, stable handles, and explicit protocols. No component mutates another component's state directly.

## Build and release boundaries

The repository-root workflow builds and packages the components together while preserving separate binaries:

- `grimoire`
- `lexicon`
- `arcana`

The workflow defaults to one build or test worker to avoid uncontrolled CPU fan-out. Higher concurrency requires an explicit `--jobs N`.

Each component may still be built and used independently from its owning source root.

## Product boundary

The active product path is Grimoire's progressive discovery interface. The former context-package compiler is not part of the CLI or MCP contract. Historical package evaluators and reports do not define current architecture.
