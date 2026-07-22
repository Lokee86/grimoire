# Grimoire

Grimoire is a standalone local repository RAG and context compiler in the [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain). It prepares repository evidence, retrieves it through standard lexical and semantic paths, and emits exact-budget context packages without owning an agent or generation step.

The current implementation has the source-storage and package-compilation foundation, a lexical retrieval baseline, and a working local Qwen3 embedding provider. Persistent vectors, BM25 postings, and hybrid rank fusion are the next core work.

## Current capabilities

- Incremental text-file indexing with unchanged-file reuse.
- Content-addressed prepared state in a private go-git object repository.
- Atomic snapshot publication.
- Git-ignore traversal with permanent tool-state exclusions.
- Deterministic language-agnostic fallback chunks.
- Exact `o200k_base` chunk and package token counts.
- Inspectable deterministic lexical ranking.
- Exact whole-chunk fitting under the serialized JSON budget.
- Managed installation of a verified local embedding runtime and model on Windows x64.
- `Qwen3-Embedding-0.6B` Q8 inference through local `llama.cpp`.
- Repository-query instruction formatting and raw document embedding.
- Native 1024-dimensional output reduced and normalized to 512 dimensions.
- A live endpoint probe that reports semantic similarity.

`grimoire index` and `grimoire context` remain lexical-only until vector persistence and hybrid retrieval are connected.

## Build

Grimoire currently targets Go 1.26.5.

```bash
go build ./cmd/grimoire
```

## Model setup

Install the managed local runtime and model:

```bash
grimoire model setup
```

Start the blocking embeddings service:

```bash
grimoire model serve
```

From another shell, verify it:

```bash
grimoire model probe
```

The fixed provider is `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0`. See [Embedding model](docs/reference/embedding-model.md) for cache paths, environment overrides, and the exact vector contract.

## Source retrieval quick start

Prepare source state:

```bash
grimoire index --root /path/to/repository
```

Compile a bounded lexical context package:

```bash
grimoire context \
  --root /path/to/repository \
  --query "where is player damage resolved" \
  --budget 2000
```

The default state location is `<repository>/.grimoire`. Context requests read prepared state rather than rescanning source files.

## Commands

```text
grimoire model setup   Install the managed embedding runtime and model.
grimoire model info    Report model, runtime, endpoint, and availability.
grimoire model serve   Run the local embeddings service.
grimoire model probe   Verify a real query/document embedding pair.
grimoire index         Prepare or incrementally update source state.
grimoire context       Rank current lexical candidates and emit bounded JSON.
grimoire version       Print the development version.
```

## Architecture

```text
repository files
      │
      ▼
prepared chunks ────────────────┬── current lexical scan ──┐
                                │                           │
                                └── Qwen3 embeddings ───────┤  vector persistence next
                                                            ▼
                                                   hybrid fusion next
                                                            ▼
                                               exact-budget JSON package
```

Lexicon is optional structural enrichment. Grimoire retains fallback chunking and must function as a complete hybrid RAG tool without Lexicon, Arcana, Demon Docs, Warlock, remote embeddings, or hosted vector storage.

## Product boundary

Grimoire owns:

- prepared source, lexical, and vector retrieval state;
- local embedding-provider integration;
- lexical and semantic candidate retrieval;
- hybrid ranking;
- context selection and exact budgeting; and
- context-package manifests.

It does not own language adapters, relationship graphs, documentation maintenance, agents, or generative inference.

## Documentation

- [Documentation index](docs/INDEX.md)
- [System overview](docs/architecture/system-overview.md)
- [Embedding model](docs/reference/embedding-model.md)
- [CLI reference](docs/reference/cli.md)
- [Prepared index](docs/architecture/prepared-index.md)
- [Context package](docs/reference/context-package.md)
- [Current limitations](docs/limits/current-limitations.md)
- [Roadmap](docs/planning/roadmap.md)

## Development

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

## Status

Grimoire is in active development. The local model provider is installed and verified; incremental vector records, BM25 postings, hybrid fusion, and stronger selection remain unfinished.
