# Component architecture

Grimoire is one repository containing three independently usable applications. Repository consolidation reduces coordination and release friction; it does not erase technical ownership boundaries.

## Components

### Lexicon

Location: [`lexicon/`](../../lexicon/)

Lexicon owns language parsing, semantic analysis, normalized fact identities, source ownership, immutable analysis objects, snapshot publication, and language-adapter contracts.

Lexicon does not own graph-wide traversal, context ranking, package budgeting, or documentation policy.

### Arcana

Location: [`arcana/`](../../arcana/)

Arcana consumes verified Lexicon snapshots and owns graph ingestion, packed forward and reverse storage, overlays, compaction, graph snapshots, traversal, impact analysis, call-chain queries, unresolved-reference queries, and graph protocol compatibility.

Arcana does not own language adapters or Grimoire's context-selection policy.

### Grimoire Context

Location: repository root, primarily `cmd/grimoire`, `internal`, and `native/vector-engine`.

The context engine owns source preparation, embeddings, vector storage, exact and lexical retrieval, structural-provider orchestration, deterministic ranking, query-shape analysis, evidence assembly, token accounting, and context-package serialization.

It consumes Lexicon and Arcana through their application and state contracts rather than importing their domain internals.

## Dependency direction

```text
Lexicon
   ↓ immutable facts and snapshots
Arcana
   ↓ graph evidence
Grimoire Context
```

Grimoire Context may also consume Lexicon directly for symbol and source-span evidence. Neither Lexicon nor Arcana depends on the context engine.

The repository layout must not create reverse imports merely because the source now shares one Git root.

## Independent use

Each component must remain meaningful on its own:

- Lexicon can scan and export normalized facts without building Arcana or Grimoire Context.
- Arcana can synchronize and answer graph queries without running Grimoire Context.
- Grimoire Context can index and retrieve source without Lexicon or Arcana state.

Standalone operation is a product contract, not a requirement for separate repositories.

## Runtime and state boundaries

The components continue to publish separate repository-local state:

- `.lexicon/` — immutable language-analysis state.
- `.arcana/` — immutable graph state bound to a Lexicon snapshot.
- `.grimoire/` — prepared source and vector state.

Co-location does not permit one component to mutate another component's state format directly. Integration occurs through versioned manifests, exports, protocols, and explicit command boundaries.

## Build boundaries

The monorepo intentionally contains multiple build roots:

- the repository-root Go module for Grimoire Context;
- `native/vector-engine/Cargo.toml` for the context vector engine;
- `lexicon/go.mod` plus adapter-specific runtimes;
- `arcana/Cargo.toml` for the graph engine.

A root build does not imply all components were verified. Release and CI work must run the owning component's test matrix.

## Source history and former repositories

Arcana and Lexicon were imported with Git subtree history under `arcana/` and `lexicon/`. Their former repositories remain available as migration pointers and may later serve as release mirrors if that is useful for existing installation paths.

The canonical source of truth is now:

- `github.com/Lokee86/grimoire/arcana`
- `github.com/Lokee86/grimoire/lexicon`

## Release direction

The immediate consolidation changes source ownership, not every distribution surface. Current CLIs and state directories remain valid. Follow-up work may provide:

- one coordinated release manifest;
- root-level build and test orchestration;
- component-specific release artifacts;
- optional subtree mirrors for compatibility;
- one Grimoire installer that can install any subset of the components.

Those are release tasks, not reasons to weaken the component APIs now.
