# Grimoire

Grimoire is a standalone local repository RAG and context compiler in the [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain). It prepares repository evidence, performs lexical and semantic retrieval, and emits exact-budget context packages without owning an agent or generation step.

The current implementation has incremental source preparation, a working local Qwen3 embedding provider, a custom content-addressed Rust vector engine, exact semantic retrieval, targeted literal recovery, deterministic candidate curation, lexical failure fallback, and exact package budgeting. Automatic maintenance and optional structural enrichment remain next.

## Current capabilities

- Incremental text-file indexing with unchanged-file reuse.
- Content-addressed prepared source state in a private go-git object repository.
- Deterministic language-agnostic fallback chunks and exact `o200k_base` counts.
- Managed installation of a verified local Qwen3 embedding runtime and model on Windows x64.
- Native 1024-dimensional output reduced and normalized to 512 dimensions.
- Content-addressed immutable vector objects keyed by embedding identity and source content.
- Reuse of unchanged embeddings across rebuilds and path changes.
- Versioned packed snapshots with sorted chunk IDs and aligned contiguous vectors.
- Memory-mapped Rust validation and concurrent exact dot-product search.
- A narrow C ABI with caller-owned buffers and no cross-runtime allocator ownership.
- Vector-backed context compilation with exact prepared-index identity validation and lexical failure fallback.
- Complete-query split-window embedding by default, using measured 64-token request batches with bounded concurrency.
- Conditional recovery for identifiers, paths, filenames, quoted phrases, configuration keys, error codes, and versions.
- Candidate deduplication, overlap removal, file/subsystem diversity, and bounded adjacent-chunk expansion.
- Selection-level retrieval source, rank, score, and inspectable reasons.
- Exact whole-chunk package fitting.

## Build

Grimoire targets Go 1.26.5 and Rust 1.90 or newer.

```bash
cd native/vector-engine
cargo build -p grimoire-vector-ffi --release
cd ../..
go build ./cmd/grimoire
```

For a packaged Windows build, place `grimoire_vector_ffi.dll` beside `grimoire.exe` or set `GRIMOIRE_VECTOR_ENGINE` to its path.

## Model setup

```bash
grimoire model setup
grimoire model serve
```

The blocking service defaults to `http://127.0.0.1:9876/v1`. From another shell:

```bash
grimoire model probe
```

The fixed provider is `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0`. See [Embedding model](docs/reference/embedding-model.md).

## Retrieval quick start

Prepare source state:

```bash
grimoire index --root /path/to/repository
```

With `grimoire model serve` running, build or incrementally refresh vector state:

```bash
grimoire vector build --root /path/to/repository
```

Run exact semantic search:

```bash
grimoire vector search \
  --root /path/to/repository \
  --query "where is player damage resolved" \
  --top-k 20
```

Compile a bounded semantic context package:

```bash
grimoire context \
  --root /path/to/repository \
  --query "where is player damage resolved" \
  --budget 2000
```

The default state location is `<repository>/.grimoire`. `context` uses the exact vector snapshot when available and emits a warning before falling back to the lexical baseline when semantic retrieval is unavailable or incompatible. Query embedding defaults to `fast`, which retains the complete query, divides it into deterministic 16-token windows, groups those windows into 64-token requests, and runs at most two requests concurrently. Use `--query-embedding-mode full` for one complete-query embedding or `--query-embedding-mode quality` for both approaches.

## Commands

```text
grimoire model setup    Install the managed embedding runtime and model.
grimoire model info     Report model, runtime, endpoint, and availability.
grimoire model serve    Run the local embeddings service.
grimoire model probe    Verify a query/document embedding pair.
grimoire index          Prepare or incrementally update source state.
grimoire vector build   Embed missing chunks and publish a packed snapshot.
grimoire vector search  Run exact semantic search over the snapshot.
grimoire vector info    Report native library and snapshot availability.
grimoire context        Retrieve semantic candidates and emit bounded JSON.
grimoire version        Print the development version.
```

## Architecture

```text
repository files
      │
      ├── prepared source objects ────────────────────────────────┐
      │                                                          │
      └── Qwen3 chunk embeddings                                 │
              │                                                   │
              ▼                                                   │
      content-addressed Rust vector objects                       │
              │                                                   │
              ▼                                                   │
      packed mmap snapshot ──► exact concurrent vector search ────┤
                                                                  │
query ──► configurable full/split batch embedding ─► concurrent searches ─┤
query literals ──► conditional exact recovery ────────────────────────────┤
                                                                           ▼
                                                    merge and candidate curation
                                                                  ▼
                                                     exact-budget package

If the semantic path is unavailable, the prepared source objects feed the deterministic lexical fallback before the same merge and curation stage.
```

Lexicon is optional structural enrichment. Grimoire remains independently usable without Lexicon, Arcana, Demon Docs, Warlock, remote embeddings, or hosted vector storage.

## Documentation

- [Documentation index](docs/INDEX.md)
- [System overview](docs/architecture/system-overview.md)
- [Vector store](docs/reference/vector-store.md)
- [Embedding model](docs/reference/embedding-model.md)
- [CLI reference](docs/reference/cli.md)
- [Current limitations](docs/limits/current-limitations.md)
- [Roadmap](docs/planning/roadmap.md)

## Development

```bash
cd native/vector-engine
cargo fmt --all --check
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cd ../..
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

## Status

The local embedding provider, persistent vector objects, packed snapshot, native ABI, exact semantic search, targeted exact recovery, deterministic candidate curation, and vector-backed context compilation are implemented and verified. Automatic maintenance, optional Lexicon enrichment, and stable external contracts remain unfinished.
