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
docs/                   Platform and context-engine documentation
../lodestone/           External vector storage and exact-search engine
```

The former standalone Arcana and Lexicon repositories are retained as migration pointers. Current development happens in this repository. Their histories were imported as Git subtrees rather than flattened copies.

See [Component architecture](docs/architecture/components.md) for ownership, dependency, release, and standalone-use rules.

## Current capabilities

### Grimoire

- Incremental prepared indexing with immutable content identities.
- Local Qwen3 embeddings served by a managed `llama.cpp` runtime.
- CPU, Vulkan, and CUDA runtime selection on Windows x64.
- Packed native documentation-vector snapshots with deterministic exact search.
- Deterministic BM25 source retrieval, exact recovery, Lexicon facts, and Arcana graph evidence.
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
Repository source
  -> Grimoire prepared source index
  -> lexical and exact retrieval
  -> Lexicon facts and Arcana graph evidence
  -> deterministic context package

Repository documentation
  -> independent knowledge index
  -> BM25 plus optional documentation vectors
  -> cited knowledge sections

Arcana graph snapshot
  -> optional semantic graph index through the shared embedding server
```

See [System overview](docs/architecture/system-overview.md).

## Build

The components keep separate build boundaries inside the monorepo. The root
workflow delegates to each owning Go or Cargo project and collects the outputs
without turning the repository into one build system. It requires Python,
Go 1.26.5, and Rust 1.90 or newer:

```bash
python scripts/workflow.py build --version 0.1.0-dev
```

The default output is `build/`, containing `bin/grimoire`, `bin/lexicon`,
`bin/arcana`, and the Lodestone native library under `native/`.
On Windows the executable and library names have `.exe` and `.dll` suffixes.
The component build roots remain independently usable:

```bash
go build ./cmd/grimoire
cd lexicon && go build -o bin/lexicon ./cmd/lexicon
cd ../arcana && cargo build --release
cargo build --manifest-path ../../lodestone/Cargo.toml -p lodestone-ffi --release
```

Run all owning test suites from the root:

```bash
python scripts/workflow.py test
```

For a local install, select the destination explicitly. The native vector
library is copied beside `grimoire`, including the Windows DLL, so the existing
discovery rules work without setting `LODESTONE_LIBRARY`:

```bash
python scripts/workflow.py install --source build --bin-dir /path/to/bin
python scripts/workflow.py install --source build --bin-dir /path/to/bin --component grimoire
python scripts/workflow.py install --source build --bin-dir /path/to/bin --component lexicon --component arcana
# Windows example:
python scripts/workflow.py install --source build --bin-dir C:/Users/<user>/bin
```

Use `py -3` instead of `python` when that is the Windows launcher configured on
the machine.

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

Prepare source and documentation state, then optionally vectorize documentation:

```bash
grimoire index --root .
grimoire knowledge index --root .
grimoire vector build --root .
```

Source retrieval does not require or consume repository-wide code embeddings. `grimoire vector build` owns only the independent documentation knowledge lane.

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

The root smoke check validates release layout, deterministic archives, version
validation, and Windows-style installation paths without requiring a compiler:

```bash
python scripts/workflow.py smoke
```

The complete component matrix is run with `python scripts/workflow.py test`.
Evaluation commands and checked-in report conventions for the context engine
are documented in [Testing and benchmarks](docs/development/testing-and-benchmarks.md).
Release packaging and artifact verification are documented in
[Release workflow](docs/development/release-workflow.md).

## Current status

The source trees and histories of Lexicon and Arcana are consolidated into Grimoire. The components still publish separate state, expose separate advanced CLIs, and retain explicit ownership boundaries, while the normal Grimoire workflow discovers and consumes their repository-local state automatically.

Grimoire has working prepared indexing, managed local embedding setup and service control, documentation-vector persistence, deterministic source and structural retrieval, adaptive context assembly, and judged evaluation. Lexicon and Arcana retain their existing standalone behavior inside `lexicon/` and `arcana/`.
