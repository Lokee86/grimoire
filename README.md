# Grimoire

Grimoire is a unified repository-discovery system. It presents source retrieval, documentation retrieval, Lexicon symbols, and Arcana relationships through one progressive interface while preserving the internal ownership boundaries of each engine.

| Component | Path | Responsibility |
| --- | --- | --- |
| **Grimoire** | repository root | Unified discovery API, source and documentation retrieval, stable handles, progressive investigation, and repository-state orchestration |
| **Lexicon** | [`lexicon/`](lexicon/) | Polyglot language analysis, normalized symbols and relationships, immutable analysis objects, and snapshots |
| **Arcana** | [`arcana/`](arcana/) | Packed repository graphs, traversal, impact analysis, paths, and direct relationship queries |

The normal product flow is:

```text
repository
  -> prepared source and documentation indexes
  -> Lexicon symbols
  -> Arcana relationships
  -> one Grimoire discovery response
```

Agents and users do not choose a provider. Grimoire routes each operation internally and returns independent evidence lanes so one kind of evidence cannot suppress another.

Grimoire is part of the [Warlock toolchain](https://github.com/Lokee86/warlock-toolchain). Lexicon and Arcana remain independently usable advanced components, but Grimoire is the primary interface for repository investigation and the normal product entry point for their specialist commands.

## Discovery contract

`grimoire search` returns separately ranked lanes:

- **Exact matches** — literal identifiers, paths, configuration keys, routes, and other concrete source matches.
- **Source matches** — BM25-ranked implementation ranges.
- **Document matches** — separately indexed documentation, rationale, plans, and architecture notes.
- **Symbol matches** — Lexicon-grounded declarations and definitions.
- **Relationship matches** — direct Arcana relationships, with Lexicon relationship fallback when Arcana is unavailable.

Each result carries provenance and a stable handle. Follow-up operations consume those handles directly:

```bash
grimoire inspect --handle <handle>
grimoire trace --anchor <handle>
grimoire impact --anchor <handle> --direction incoming
```

Source and documentation are intentionally separate. Source describes current repository behavior. Documentation describes intent, rationale, constraints, or historical decisions and may be stale.

## Engine command namespaces

Grimoire exposes direct administrative and specialist operations without making users locate or invoke provider binaries themselves:

```bash
grimoire lexicon status --repo .
grimoire lexicon doctor --repo .
grimoire lexicon scan --repo .
grimoire arcana sync --lexicon .lexicon --state .arcana
grimoire arcana query --graph <graph> --catalogue <catalogue> --name <symbol>
```

`grimoire lexicon check` and `grimoire arcana check` report the resolved provider command and version. All other arguments are forwarded to the owning engine with stdin, stdout, stderr, and exit status preserved.

The standalone `lexicon` and `arcana` binaries remain available for component development and independent use. Product documentation and ordinary installations can treat `grimoire` as the single entry point.

The former context-package command is retired. Grimoire no longer attempts to predict and deterministically compress an entire investigation into one answer-shaped bundle. Agents discover evidence progressively instead.

## Repository layout

```text
arcana/                 Rust graph engine and CLI
lexicon/                Go orchestration and polyglot language adapters
cmd/grimoire/           Unified Grimoire CLI
internal/agentquery/    Source, symbol, and relationship discovery
internal/agentruntime/  Documentation lane, state preparation, and sessions
internal/knowledge/     Documentation indexing and retrieval
docs/                   Architecture and interface documentation
../lodestone/           External vector storage and exact-search engine
```

See [Component architecture](docs/architecture/components.md) for ownership, dependency, release, and standalone-use rules.

## Current capabilities

### Grimoire

- Incremental source indexing with immutable content identities.
- Deterministic exact and BM25 source discovery.
- Independent documentation indexing with BM25 and optional vectors.
- Lexicon-grounded symbol discovery.
- Arcana-backed direct relationship discovery, trace, paths, and impact analysis.
- Stable snapshot-qualified handles for exact follow-up inspection.
- Independent per-lane limits; exact, source, document, symbol, and relationship evidence do not compete for one shared quota.
- One stdio MCP tool for search, orient, trace, impact, and inspect.
- Automatic repository-state preparation and alignment across Grimoire, Lexicon, and Arcana.
- Persistent investigation sessions that replace repeated evidence with compact prior handles.
- Repository-owned discovery and agent-outcome benchmarks.

### Lexicon

- Go, GDScript, Python, Ruby, Rust, JavaScript, TypeScript, Svelte, C, C++, C#, Java, Kotlin, and generic adapters.
- Normalized facts output and compact immutable binary objects.
- Atomic content-addressed snapshots, incremental analysis, deterministic merges, and consumer hooks.

### Arcana

- Lexicon snapshot ingestion without rebuilding language parsers.
- Packed forward and reverse graph storage, immutable snapshots, overlays, and compaction.
- Deterministic graph operations for neighbors, paths, impact, call chains, unresolved references, roles, and snapshot differences.
- Optional graph-neighborhood vector indexes using Grimoire's shared embedding service.

## Build

The root workflow delegates to the owning Go and Cargo projects. It requires Python, Go 1.26.5, and Rust 1.90 or newer:

```bash
python scripts/workflow.py build --version 0.1.0-dev
```

The default output is `build/`, containing `bin/grimoire`, `bin/lexicon`, `bin/arcana`, and the Lodestone native library under `native/`.

Component build roots remain independently usable:

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

Install selected components:

```bash
python scripts/workflow.py install --source build --bin-dir /path/to/bin
python scripts/workflow.py install --source build --bin-dir /path/to/bin --component grimoire
python scripts/workflow.py install --source build --bin-dir /path/to/bin --component lexicon --component arcana
```

Use `py -3` instead of `python` when that is the configured Windows launcher.

## Quick start

Show the current workflow:

```bash
grimoire help
```

Search a repository. State is refreshed when needed:

```bash
grimoire search --root . --query "Where is session creation handled?"
```

Orient in an unfamiliar repository:

```bash
grimoire orient --root .
```

Inspect and expand returned handles:

```bash
grimoire inspect --root . --handle <handle>
grimoire trace --root . --anchor <handle> --depth 4
grimoire impact --root . --anchor <handle> --direction incoming
```

Omit documentation when only implementation evidence is wanted:

```bash
grimoire search --root . --query "SPACE_ROCKS_LOCAL_SERVER_PORT" --code-only
```

Documentation vectors are optional and affect only the document lane:

```bash
grimoire model setup
grimoire model start
grimoire vector build --root .
grimoire search --root . --query "Why is match state authoritative?" --document-vectors
```

Source and structural discovery do not require repository-wide code embeddings.

Serve the same interface over MCP stdio:

```bash
grimoire mcp --root .
```

The server exposes one `grimoire_discover` tool. Start with `search`, then use returned handles with `inspect`, `trace`, or `impact`. Reuse one investigation `session` across a task to avoid replaying evidence already returned.

Lexicon and Arcana executables are discovered beside Grimoire, through repository configuration, in the consolidated checkout, or on `PATH`. `GRIMOIRE_LEXICON_COMMAND` and `GRIMOIRE_ARCANA_COMMAND` provide explicit command overrides for the namespaced command surface. Discovery-provider failures are reported as warnings; Grimoire continues with the evidence lanes that remain available.

## Documentation

- [Documentation index](docs/INDEX.md)
- [System overview](docs/architecture/system-overview.md)
- [Component architecture](docs/architecture/components.md)
- [Discovery CLI](docs/reference/cli.md)
- [Discovery contract](docs/reference/agent-query.md)
- [MCP interface](docs/reference/agent-mcp.md)
- [Development and benchmarks](docs/development/INDEX.md)
- [Current limitations](docs/limits/INDEX.md)
- [Roadmap](docs/planning/INDEX.md)
- [Lexicon documentation](lexicon/docs/README.md)
- [Arcana documentation](arcana/docs/)

Reference documentation describes implemented behavior. Unimplemented work belongs in the roadmap, and unresolved constraints belong in the limitations section.

## Current status

Lexicon and Arcana are consolidated into this repository while retaining explicit technical boundaries. Grimoire now owns the normal discovery workflow and exposes their information through one interface. The active product path is progressive evidence discovery, not preassembled context packages.
