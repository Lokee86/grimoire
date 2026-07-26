# Grimoire

Grimoire is a deterministic repository-intelligence platform. One repository now contains the complete source-to-context stack while preserving three independently usable component boundaries:

| Component | Path | Responsibility |
| --- | --- | --- |
| **Grimoire** | repository root | Retrieval, ranking, task analysis, token budgeting, and context-package construction |
| **Lexicon** | [`lexicon/`](lexicon/) | Polyglot language analysis, normalized source facts, immutable analysis objects, and snapshots |
| **Arcana** | [`arcana/`](arcana/) | Repository-graph construction, packed graph storage, semantic graph indexing, traversal, impact analysis, and graph queries |

The components form one natural pipeline:

```text
source repository
  -> Lexicon facts and snapshots
  -> Arcana graph snapshots
  -> Grimoire retrieval and context packages
```

They remain separate applications and technical boundaries. Lexicon can be used without Arcana or the context engine. Arcana can be used directly by graph consumers. Grimoire Context continues to provide source retrieval when structural state or executables are unavailable.

Grimoire is part of the [Warlock toolchain](https://github.com/Lokee86/warlock-toolchain), but the repository and each component remain independently usable.

## Repository layout

```text
arcana/                 Rust graph engine and CLI
lexicon/                Go orchestration plus polyglot language adapters
cmd/grimoire/           Grimoire Context CLI
internal/               Context retrieval, ranking, assembly, and integration
native/vector-engine/   Rust vector storage and exact-search engine
docs/                   Platform and context-engine documentation
```

The former standalone Arcana and Lexicon repositories are retained as migration pointers. Current development happens in this repository. Their histories were imported as Git subtrees rather than flattened copies.

See [Component architecture](docs/architecture/components.md) for ownership, dependency, release, and standalone-use rules.

## Current capabilities

### Grimoire

- Incremental prepared indexing with immutable content identities.
- Local Qwen3 embeddings served by a managed `llama.cpp` runtime.
- CPU, Vulkan, and CUDA runtime selection on Windows x64.
- Packed native vector snapshots with deterministic exact search.
- Lexical fallback when semantic state is missing, stale, or unavailable.
- Exact recovery for concrete paths, symbols, and identifiers.
- Progressive agent queries for orientation, code-first search, graph trace, impact, and exact handle inspection.
- A single stdio MCP tool that automatically refreshes deterministic state, separates code evidence from repository knowledge, and returns stable expansion handles.
- Persistent investigation sessions that replace repeated evidence with compact prior handles across multi-step agent work.
- Lexicon symbol facts plus lexical- and semantic-seeded Arcana graph evidence when structural state is available.
- Deterministic query-shape classification and automatic context budgets.
- Evidence-coverage assembly for automatic-budget requests.
- Versioned JSON context packages with exact `o200k_base` accounting.
- Repository-owned retrieval, ranking, structural, and adaptive-assembly evaluation.

### Lexicon

- Go, GDScript, Python, Ruby, Rust, JavaScript, TypeScript, Svelte, and generic adapters.
- Normalized facts-v1 adapter output and compact immutable binary objects.
- Atomic content-addressed snapshots, incremental analysis, deterministic merges, and consumer hooks.

### Arcana

- Lexicon snapshot ingestion without rebuilding language parsers.
- Packed forward and reverse graph storage, immutable snapshots, overlays, and compaction.
- Deterministic graph protocol operations for paths, impact, call chains, unresolved references, roles, and snapshot differences.
- Optional graph-neighborhood vector indexes built through Grimoire's existing embedding server, without a second model runtime.

## System flow

Repository preparation and context construction remain explicit stages:

```text
Repository
  -> Grimoire prepared source index
  -> embedding batches
  -> content-addressed vector objects
  -> packed vector snapshot

Repository
  -> Lexicon immutable analysis snapshot
  -> Arcana immutable graph snapshot
  -> optional Arcana semantic graph index through the shared embedding server

Query
  -> semantic, lexical, exact, and structural retrieval
  -> candidate merge and deterministic ranking
  -> query-shape analysis
  -> automatic policy activation or explicit fixed budget
  -> evidence-aware assembly
  -> versioned context package
```

See [System overview](docs/architecture/system-overview.md).

## Build

The components keep separate build boundaries inside the monorepo.

Grimoire Context requires Go 1.26.5 and Rust 1.90 or newer:

```bash
cargo build --manifest-path native/vector-engine/Cargo.toml -p grimoire-vector-ffi --release
go build ./cmd/grimoire
```

Lexicon:

```bash
cd lexicon
go build -o bin/lexicon ./cmd/lexicon
```

Arcana:

```bash
cd arcana
cargo build --release
```

On Windows, the native vector build produces `native/vector-engine/target/release/grimoire_vector_ffi.dll`. Grimoire discovers the DLL in the workspace, beside the executable, or through `GRIMOIRE_VECTOR_ENGINE`.

## Quick start

The root command prints the complete normal workflow:

```bash
grimoire help
```

Install and start the managed embedding runtime:

```bash
grimoire model setup
grimoire model start
```

Prepare and vectorize a repository:

```bash
grimoire index --root .
grimoire vector build --root .
```

Inspect or automatically prepare deterministic analysis state without building vectors:

```bash
grimoire status --root .
grimoire status --root . --refresh
```

Query progressively without requiring vectors:

```bash
grimoire query orient --root .
grimoire query search --root . --query "Where is session creation handled?"
```

Serve the unified agent interface over MCP stdio:

```bash
grimoire mcp --root .
```

The MCP server exposes one `grimoire_query` tool. Use one investigation `session` across an agent task so repeated nodes, source ranges, graph paths, and documents return as prior handles instead of replayed content. `orient` and `search` keep production code in the query lane while documentation and design rationale are returned through the independent knowledge lane.

Compile an automatically sized context package:

```bash
grimoire context --root . --query "Where is context-package assembly implemented?"
```

Lexicon and Arcana state is prepared automatically when their executables are available beside Grimoire or on `PATH`. Provider failures are reported as warnings and Grimoire continues with deterministic source retrieval; the source index is still refreshed when required.

A positive budget retains fixed fit-to-budget behavior:

```bash
grimoire context --root . --query "Trace context assembly end to end" --budget 8000
```

Structural enrichment uses repository-local `.lexicon/` and `.arcana/` state when available. Build or install the component executables, then initialize their state with the commands documented in [`lexicon/README.md`](lexicon/README.md) and [`arcana/README.md`](arcana/README.md). Run `arcana vectorize` after `arcana sync` to add semantic graph entry points through the same embedding server used by Grimoire. Grimoire uses a matching existing index automatically but never builds one during a context request. Missing structural providers warn and fall back to source retrieval.

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
- [Component architecture](docs/architecture/components.md)
- [Architecture](docs/architecture/INDEX.md)
- [CLI and data contracts](docs/reference/INDEX.md)
- [Development and evaluation](docs/development/INDEX.md)
- [Current limitations](docs/limits/INDEX.md)
- [Roadmap](docs/planning/INDEX.md)
- [Lexicon documentation](lexicon/docs/README.md)
- [Arcana documentation](arcana/docs/)

Reference documentation describes implemented behavior. Unimplemented work belongs in the roadmap, and unresolved constraints belong in the limitations section.

## Development

Run each component from its own build root:

```bash
go test ./...
cargo test --manifest-path native/vector-engine/Cargo.toml

cd lexicon && python evaluation/run_tests.py
cd ../arcana && cargo test --all-targets
```

Evaluation commands and checked-in report conventions for the context engine are documented in [Testing and benchmarks](docs/development/testing-and-benchmarks.md).

## Current status

The source trees and histories of Lexicon and Arcana are consolidated into Grimoire. The components still publish separate state, expose separate advanced CLIs, and retain explicit ownership boundaries, while the normal Grimoire workflow discovers and consumes their repository-local state automatically.

Grimoire has working prepared indexing, managed local embedding setup and service control, vector persistence and search, source and structural retrieval, adaptive context assembly, and judged evaluation. Lexicon and Arcana retain their existing standalone behavior inside `lexicon/` and `arcana/`.
