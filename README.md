# Grimoire

Grimoire builds deterministic, bounded context packages for codebase work. It combines a content-addressed source index, local embeddings, exact and lexical recovery, optional Lexicon and Arcana structure, query-shape analysis, evidence-aware assembly, and explicit token accounting.

The project is part of the Warlock toolchain, but Grimoire remains independently usable. Lexicon and Arcana are optional structural providers; source retrieval continues when either provider is unavailable.

## Current capabilities

- Incremental prepared indexing with immutable content identities.
- Local Qwen3 embeddings served by a managed `llama.cpp` runtime.
- CPU, Vulkan, and CUDA runtime selection on Windows x64.
- Packed native vector snapshots with deterministic exact search.
- Lexical fallback when semantic state is missing, stale, or unavailable.
- Exact recovery for concrete paths, symbols, and identifiers.
- Optional Lexicon symbol facts and Arcana graph evidence.
- Deterministic query-shape classification and automatic context budgets.
- Evidence-coverage assembly for automatic-budget requests.
- Versioned JSON context packages with exact `o200k_base` accounting.
- Repository-owned retrieval, ranking, structural, and adaptive-assembly evaluation.

## System flow

Index construction and context construction are separate pipelines:

```text
Repository
  -> prepared source index
  -> embedding batches
  -> content-addressed vector objects
  -> packed vector snapshot

Query
  -> semantic, lexical, exact, and structural retrieval
  -> candidate merge and deterministic ranking
  -> query-shape analysis
  -> automatic policy activation or explicit fixed budget
  -> evidence-aware assembly
  -> versioned context package
```

See [System overview](docs/architecture/system-overview.md) for ownership and failure boundaries.

## Build

Grimoire requires Go 1.26.5 and Rust 1.90 or newer.

```bash
cargo build --manifest-path native/vector-engine/Cargo.toml -p grimoire-vector-ffi --release
go build ./cmd/grimoire
```

On Windows, the native build produces `native/vector-engine/target/release/grimoire_vector_ffi.dll`. Grimoire discovers the DLL in the workspace, beside the executable, or through `GRIMOIRE_VECTOR_ENGINE`.

## Quick start

Install the managed embedding runtime and model on Windows x64:

```bash
grimoire model setup
```

Start the local embeddings service in one terminal:

```bash
grimoire model serve
```

Prepare and vectorize a repository in another terminal:

```bash
grimoire index --root .
grimoire vector build --root .
```

Compile an automatically sized context package:

```bash
grimoire context --root . --query "Where is context-package assembly implemented?"
```

With no positive `--budget`, Grimoire classifies the query as focused, bounded, or exploratory and selects a deterministic target. A positive budget retains fixed fit-to-budget behavior:

```bash
grimoire context --root . --query "Trace context assembly end to end" --budget 8000
```

Use `grimoire model probe` to verify the live embedding endpoint and `grimoire vector info` to inspect native snapshot availability.

## Context policy

Automatic requests currently use these target tiers:

| Scope | Minimum | Target | Maximum |
| --- | ---: | ---: | ---: |
| Focused | 2,000 | 3,000 | 6,000 |
| Bounded | 3,000 | 6,000 | 10,000 |
| Exploratory | 6,000 | 12,000 | 18,000 |

The target is a deterministic policy choice, not a promise that every package will fill the boundary. Assembly preserves ranked alternatives and stops when the scope-specific evidence requirements are satisfied or a hard cap is reached. The package records the profile, policy, coverage, and stopping decision.

See [Query shape and assembly](docs/reference/query-shape-and-assembly.md).

## Documentation

- [Documentation index](docs/INDEX.md)
- [Architecture](docs/architecture/INDEX.md)
- [CLI and data contracts](docs/reference/INDEX.md)
- [Development and evaluation](docs/development/INDEX.md)
- [Current limitations](docs/limits/INDEX.md)
- [Roadmap](docs/planning/INDEX.md)

Reference documentation describes implemented behavior. Unimplemented work belongs in the roadmap, and unresolved constraints belong in the limitations section.

## Development

Run the Go suite:

```bash
go test ./...
```

Run the native vector-engine suite:

```bash
cargo test --manifest-path native/vector-engine/Cargo.toml
```

Evaluation commands and checked-in report conventions are documented in [Testing and benchmarks](docs/development/testing-and-benchmarks.md).

## Current status

Grimoire has working prepared indexing, local embedding setup and service control, vector persistence and search, source and structural retrieval, adaptive context assembly, and judged evaluation. The main remaining work is retrieval-quality calibration across more repositories, better automatic package targets, operational hardening, and automatic maintenance of prepared/vector state.

Grimoire is not a source-of-truth language graph, a general version-control system, or an autonomous code editor. Lexicon owns normalized language facts; Arcana owns graph operations; Grimoire owns retrieval and context-package construction.
