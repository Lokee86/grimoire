# Grimoire

Grimoire is a unified repository-discovery system for agents and humans. It combines exact source search, BM25 source retrieval, separately ranked documentation, Lexicon symbols, and Arcana relationships behind one progressive interface.

Use Grimoire when an investigation crosses files, packages, languages, generated contracts, ownership boundaries, or documentation. Keep using ordinary shell search and direct file reads when a precise identifier makes them cheaper.

Grimoire is part of the [Warlock toolchain](https://github.com/Lokee86/warlock-toolchain). Lexicon and Arcana remain independently usable components, but `grimoire` is the normal product entry point.

## Install a release

Download and extract the combined bundle for your platform:

```text
grimoire-bundle-<version>-<platform>-<arch>.zip
```

Run the included installer from the extracted directory:

```bash
python install.py --bin-dir /path/on/your/PATH
```

On Windows, use `py -3` instead of `python` when that is the configured launcher:

```powershell
py -3 install.py --bin-dir "$HOME\bin"
```

The installer copies:

- `grimoire`, `lexicon`, and `arcana` into the selected binary directory;
- the Lodestone native library beside `grimoire`;
- Lexicon runtime adapters with the Lexicon installation;
- `grimoire/SKILL.md` into `~/.agents/skills` and `~/.hermes/skills` when Grimoire is selected.

The installer does **not** modify `PATH`. Add the selected binary directory yourself, then verify the installation:

```bash
grimoire version
grimoire lexicon check
grimoire arcana check
```

Install only selected applications by repeating `--component`:

```bash
python install.py --bin-dir /path/on/your/PATH --component grimoire
python install.py --bin-dir /path/on/your/PATH --component lexicon --component arcana
```

Override skill destinations or skip skill installation:

```bash
python install.py --bin-dir /path/on/your/PATH --skills-dir /custom/skills
python install.py --bin-dir /path/on/your/PATH --skip-skills
```

See [Installation and agent setup](docs/reference/installation.md) for source builds, release layout, PATH guidance, MCP configuration, and troubleshooting.

## First repository investigation

From the repository you want to inspect:

```bash
grimoire status --root . --refresh
grimoire search --root . --query "Where is session creation handled?" --breadth narrow --code-only --session session-flow
```

Narrow search returns handle-only discovery by default. Follow the returned stable handles instead of repeating searches or requesting inline source:

```bash
grimoire inspect --root . --handle '<handle>' --session session-flow
grimoire trace --root . --anchor '<handle>' --depth 3 --limit 6 --session session-flow
grimoire impact --root . --anchor '<handle>' --direction incoming --depth 3 --limit 6 --session session-flow
```

Use `orient` only when the repository is unfamiliar and the task has no useful search terms:

```bash
grimoire orient --root . --limit 6 --session unfamiliar-repo
```

Source and structural discovery do not require repository-wide code embeddings. Documentation vectors are optional and affect only the separate document lane.

## Agent setup

A normal agent should have all three of these capabilities:

1. ordinary shell, Git, search, and file-reading tools;
2. the installed Grimoire skill;
3. the `grimoire_discover` MCP tool.

Configure the agent host to launch this stdio server for the repository:

```text
command: grimoire
args: ["mcp", "--root", "/absolute/path/to/repository"]
```

The exact host configuration format varies, but the process command is always:

```bash
grimoire mcp --root /absolute/path/to/repository
```

The installer places the canonical skill at:

```text
~/.agents/skills/grimoire/SKILL.md
~/.hermes/skills/grimoire/SKILL.md
```

Start a new agent session after installation so the host can discover the skill. The skill teaches agents to:

- use direct search and file reads first when exact paths or symbols are known;
- escalate localized uncertainty to `breadth: "narrow"` and distributed investigations to `breadth: "balanced"`;
- reuse one investigation `session`;
- treat narrow search as handle-only discovery and use `inspect` for exact evidence;
- use `trace` or `impact` only for a named unresolved relationship;
- avoid unnecessary `force-refresh` operations;
- verify material conclusions against exact source;
- switch back to shell search or direct reads when they are cheaper.

A good first MCP request is:

```json
{
  "schema": "grimoire.discovery.v1",
  "mode": "search",
  "root": "/absolute/path/to/repository",
  "query": "camera visibility network interest realtime snapshot",
  "breadth": "balanced",
  "limit": 6,
  "code_only": true,
  "state_mode": "refresh-if-needed",
  "session": "network-interest"
}
```

After state is known to be current, use `state_mode: "current-only"` for follow-ups.

See [Agent operating guide](docs/reference/agent-mcp.md) and the canonical [`SKILL.md`](skills/grimoire/SKILL.md).

## When Grimoire helps

Grimoire is most useful for:

- architectural and ownership questions;
- cross-language or generated-contract traces;
- call paths and dependency impact;
- source-plus-document investigations;
- unfamiliar repositories where likely entry points are not known;
- investigations that would otherwise require many rounds of broad search and reading.

Plain shell inspection can remain faster for a simple literal lookup or short call chain with obvious lexical anchors. The intended workflow is not “Grimoire instead of shell”; it is “use the cheapest reliable discovery path, then verify against source.”

Repository-owned agent benchmarks currently show both outcomes:

- a narrow packet-trace task favored plain shell inspection;
- a broad network-interest architecture task was completed by Grimoire in about 9 minutes versus about 17 minutes for plain and CBM-assisted agents, with roughly half the model calls and noncached tokens while preserving technical quality.

These are task-specific measurements, not universal performance guarantees. See [Agent benchmark findings](docs/development/agent-benchmark-findings.md).

## Discovery contract

`grimoire search` returns independent evidence lanes:

| Lane | Meaning |
| --- | --- |
| `exact_matches` | Literal identifiers, paths, routes, configuration keys, and other concrete source matches |
| `source_matches` | BM25-ranked implementation ranges |
| `document_matches` | Separately indexed documentation, rationale, plans, and architecture notes |
| `symbol_matches` | Lexicon-grounded declarations and definitions |
| `relationship_matches` | Direct Arcana relationships, with Lexicon fallback when Arcana is unavailable |

Each result carries provenance and a stable handle. `breadth: "balanced"` preserves independent lane limits so one evidence class cannot consume another. `breadth: "narrow"` uses one combined code-evidence budget, defaults to four handle-only results, and defers exact source ranges until `inspect`.

Responses also include a conservative `assessment` describing observed and missing owner, control-flow, public-boundary, and test dimensions plus the smallest justified next action. It is workflow guidance, not proof of correctness.

Source and documentation remain deliberately separate:

- source describes current repository behavior;
- documentation describes intent, rationale, constraints, plans, or history and may be stale.

The response schema is `grimoire.discovery.v1`. See [Unified discovery contract](docs/reference/agent-query.md).

## Component architecture

| Component | Path | Responsibility |
| --- | --- | --- |
| **Grimoire** | repository root | Unified discovery API, source and documentation retrieval, stable handles, progressive investigation, state orchestration, and agent sessions |
| **Lexicon** | [`lexicon/`](lexicon/) | Polyglot language analysis, normalized symbols and relationships, immutable analysis objects, and snapshots |
| **Arcana** | [`arcana/`](arcana/) | Packed repository graphs, traversal, impact analysis, paths, and direct relationship queries |
| **Lodestone** | sibling repository | Native vector storage and exact vector search used by optional document and graph vector features |

The normal product flow is:

```text
repository
  -> prepared source and documentation indexes
  -> Lexicon symbols
  -> Arcana relationships
  -> one Grimoire discovery response
```

Grimoire also exposes specialist engine commands without requiring callers to locate provider binaries:

```bash
grimoire lexicon status --repo .
grimoire lexicon doctor --repo .
grimoire lexicon scan --repo .
grimoire arcana sync --lexicon .lexicon --state .arcana
grimoire arcana query --graph <graph> --catalogue <catalogue> --name <symbol>
```

`grimoire lexicon check` and `grimoire arcana check` report the resolved command and version. Other arguments are forwarded to the owning engine with stdin, stdout, stderr, and exit status preserved.

The former context-package compiler is retired. Grimoire now supports progressive evidence discovery rather than attempting to predict and token-fit an entire investigation before the agent begins.

See [Component architecture](docs/architecture/components.md).

## Current capabilities

### Grimoire

- Incremental source indexing with immutable content identities.
- Deterministic exact and BM25 source discovery.
- Independent documentation indexing with BM25 and optional vectors.
- Lexicon-grounded symbol discovery.
- Arcana-backed relationships, trace, paths, and impact analysis.
- Stable snapshot-qualified handles for exact follow-up inspection.
- Balanced independent per-lane limits plus a four-result combined narrow mode.
- Handle-only narrow discovery with source expansion deferred to inspection.
- Conservative evidence assessment and next-action guidance.
- One stdio MCP tool for search, orient, trace, impact, and inspect.
- Automatic repository-state preparation and alignment across Grimoire, Lexicon, Arcana, and documentation state.
- Persistent investigation sessions that replace repeated evidence with compact prior handles.
- Namespaced Lexicon and Arcana administrative commands.
- Repository-owned retrieval and end-to-end agent benchmarks.

### Lexicon

- Go, GDScript, Python, Ruby, Rust, JavaScript, TypeScript, Svelte, C, C++, C#, Java, Kotlin, and generic adapters.
- Normalized facts output and compact immutable binary objects.
- Atomic content-addressed snapshots, incremental analysis, deterministic merges, and consumer hooks.

### Arcana

- Lexicon snapshot ingestion without rebuilding language parsers.
- Packed forward and reverse graph storage, immutable snapshots, overlays, and compaction.
- Deterministic graph operations for neighbors, paths, impact, call chains, unresolved references, roles, and snapshot differences.
- Optional graph-neighborhood vector indexes using Grimoire's shared embedding service.

## Build from source

Requirements:

- Python 3.12 or newer;
- Go 1.26.5;
- Rust 1.90 or newer;
- Node.js 22 for the TypeScript adapter;
- a sibling Lodestone checkout at `../lodestone`, or `LODESTONE_ROOT` pointing to it.

Build the combined layout:

```bash
python scripts/workflow.py build --version 0.1.0-dev
```

The default `build/` directory contains:

```text
build/
  bin/grimoire
  bin/lexicon
  bin/arcana
  native/<lodestone library>
  adapters/<Lexicon runtime adapters>
  skills/grimoire/SKILL.md
```

Install the build:

```bash
python scripts/workflow.py install --source build --bin-dir /path/on/your/PATH
```

Run the complete bounded test matrix:

```bash
python scripts/workflow.py test
```

The workflow defaults to one worker to avoid uncontrolled Go and Cargo process fan-out. Increase concurrency only deliberately:

```bash
python scripts/workflow.py test --jobs 2
```

Run the packaging/install smoke suite without compiling the whole product:

```bash
python scripts/workflow.py smoke
```

Component build roots remain independently usable. See [Release workflow](docs/development/release-workflow.md) for exact commands and package layout.

## Repository layout

```text
arcana/                 Rust graph engine and CLI
lexicon/                Go orchestration and polyglot language adapters
cmd/grimoire/           Unified Grimoire CLI
internal/agentquery/    Source, symbol, and relationship discovery
internal/agentruntime/  Documentation lane, state preparation, and sessions
internal/knowledge/     Documentation indexing and retrieval
evaluation/             Judged corpora, benchmark harnesses, and reports
skills/grimoire/        Canonical agent skill
docs/                   Architecture, reference, development, limits, and plans
../lodestone/           External vector storage and exact-search engine
```

## Documentation

- [Documentation index](docs/INDEX.md)
- [Installation and agent setup](docs/reference/installation.md)
- [System overview](docs/architecture/system-overview.md)
- [Component architecture](docs/architecture/components.md)
- [Lexicon–Arcana–Grimoire analysis stack](docs/architecture/analysis-stack.md)
- [Grimoire maintainer map](docs/architecture/maintainer-map.md)
- [Discovery CLI](docs/reference/cli.md)
- [Lexicon reference](docs/reference/lexicon.md)
- [Arcana reference](docs/reference/arcana.md)
- [Unified discovery contract](docs/reference/agent-query.md)
- [Agent and MCP guide](docs/reference/agent-mcp.md)
- [Testing and benchmarks](docs/development/testing-and-benchmarks.md)
- [Agent benchmark findings](docs/development/agent-benchmark-findings.md)
- [Recent changes — July 2026](docs/development/recent-changes-2026-07.md)
- [Current limitations](docs/limits/current-limitations.md)
- [Roadmap](docs/planning/roadmap.md)
- [Lexicon documentation and codemap](lexicon/docs/README.md)
- [Arcana documentation and codemap](arcana/docs/README.md)

Reference documentation describes implemented behavior. Unimplemented work belongs in the roadmap, and unresolved constraints belong in limitations.

## Current status

Lexicon and Arcana are consolidated into this repository while retaining explicit process, state, and release boundaries. Grimoire owns the normal repository-investigation workflow. The active product path is progressive evidence discovery, not preassembled context packages.
