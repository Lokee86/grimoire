# Grimoire

Grimoire is a low-latency context compiler for software repositories. It maintains prepared local retrieval state, ranks repository evidence with inspectable signals, and emits bounded context packages without owning an agent or generation step.

This repository currently contains the first lexical baseline:

- incremental file indexing with unchanged-record reuse;
- a private go-git object repository with deterministic binary index shards;
- content-addressed snapshot publication through `refs/grimoire/state`;
- deterministic fallback chunking;
- lexical, filename, and path ranking;
- whole-chunk budget fitting;
- inspectable JSON context packages; and
- no request-time repository scanning.

Arcana, Demon Docs, semantic embeddings, language adapters, and daemon hosting are intentionally outside this first slice.

## Build

```bash
go build ./cmd/grimoire
```

## Index a repository

```bash
grimoire index --root /path/to/repository
```

The prepared index is stored as a private bare go-git object repository at `/path/to/repository/.grimoire` by default. Binary file records are distributed across 256 content-addressed shards, and `refs/grimoire/state` atomically publishes the current snapshot. Re-running the command reuses unchanged shard objects.

## Compile context

```bash
grimoire context \
  --root /path/to/repository \
  --query "where is player damage resolved" \
  --budget 2000
```

The request reads only the prepared index and emits a deterministic JSON package. The current budget is a conservative content-token estimate, not yet a model-specific tokenizer count.

## Benchmark the warm retrieval path

```bash
go test ./internal/retrieve -bench BenchmarkSearchTenThousandChunks -benchmem
```

The benchmark uses a prepared in-memory snapshot with 10,000 chunks. It intentionally excludes repository scanning and indexing work from the request path.

## Current boundary

Grimoire owns retrieval, ranking, context selection, and package manifests. It does not own repository relationship graphs, documentation health, agent orchestration, generative inference, or hosted vector infrastructure.
