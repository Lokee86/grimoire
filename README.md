# Grimoire

Grimoire is a low-latency repository context compiler in the [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain). It maintains prepared local retrieval state, ranks repository evidence with inspectable signals, and emits bounded context packages without owning an agent or generation step.

The current implementation is the first lexical baseline. It is usable independently and does not require Lexicon, Arcana, Demon Docs, a model service, or a vector database.

## Current capabilities

- Incremental text-file indexing with unchanged-file reuse.
- A private bare go-git object repository for prepared state.
- Deterministic binary file records distributed across 256 content-addressed shards.
- Atomic snapshot publication through `refs/grimoire/state`.
- Root and nested `.gitignore` semantics, including negation.
- A configurable replacement ignore file using Git-ignore syntax.
- Deterministic fallback chunking for supported text files.
- Exact `o200k_base` token counts stored with prepared chunks.
- Inspectable lexical, filename, path, and leading-line ranking signals.
- Whole-chunk fitting under an exact serialized-package token budget.
- Deterministic, agent-independent JSON context packages.
- No request-time repository traversal or source-file reads.

## Build

Grimoire currently targets Go 1.26.5.

```bash
go build ./cmd/grimoire
```

## Quick start

Prepare an index:

```bash
grimoire index --root /path/to/repository
```

Compile context from the prepared state:

```bash
grimoire context \
  --root /path/to/repository \
  --query "where is player damage resolved" \
  --budget 2000
```

The default prepared-state location is `/path/to/repository/.grimoire`. Context requests read that state rather than rescanning the repository.

## Command summary

```text
grimoire index   Prepare or incrementally update repository retrieval state.
grimoire context Rank prepared chunks and emit a bounded JSON context package.
grimoire version Print the development version.
```

See the [CLI reference](docs/reference/cli.md) for every current flag and output contract.

## Architecture

```text
repository files
      │
      ▼
index.Build ──► fallback chunks ──► private go-git object repository
                                           │
                                           ▼
query ──► retrieve.Search ──► compiler.Compile ──► o200k_base count ──► JSON context package
```

The request path loads prepared state, performs deterministic lexical ranking, and selects whole chunks while counting the exact indented JSON package with `o200k_base`. Paths, reasons, query text, metadata, and the trailing newline all consume budget. The request path does not traverse or read source files. The current lexical search still scans the prepared chunks in memory; a postings index is planned but not implemented.

See the [system overview](docs/architecture/system-overview.md) and [prepared-index design](docs/architecture/prepared-index.md) for the implemented architecture.

## Indexing and ignore behavior

By default, Grimoire follows the repository's root and nested `.gitignore` files. Pass `--ignore-file path/to/file` to replace that hierarchy with another Git-ignore-syntax file. Relative paths are resolved from the indexed repository root.

The following directories remain permanently excluded because they contain repository metadata, generated tool state, or nested worktrees:

- `.git/`
- `.grimoire/`
- `.ddocs/`
- `.arcana/`
- `.warlock/`
- `.worktrees/`
- `.workingtrees/`

A custom `--state` path is also excluded from traversal. Grimoire does not hard-code conventional dependency or build directories such as `vendor/`, `node_modules/`, `target/`, or `dist/`; repository ignore rules decide whether those paths are indexed.

See [Indexing reference](docs/reference/indexing.md) for supported file types, size limits, binary detection, and incremental behavior.

## Product boundary

Grimoire owns:

- retrieval-state maintenance;
- candidate retrieval and ranking;
- context selection and budgeting; and
- context-package manifests.

Grimoire does not own:

- language adapters or normalized source facts, which belong to Lexicon;
- repository relationship graphs, which belong to Arcana;
- documentation health and maintenance, which belong to Demon Docs;
- agent orchestration or generative inference; or
- hosted vector infrastructure.

Lexicon will eventually provide structural source chunks. Grimoire will retain its current fallback chunker so its basic lexical mode remains independently usable.

## Documentation

- [Documentation index](docs/INDEX.md)
- [System overview](docs/architecture/system-overview.md)
- [Prepared index](docs/architecture/prepared-index.md)
- [CLI reference](docs/reference/cli.md)
- [Indexing reference](docs/reference/indexing.md)
- [Context package format](docs/reference/context-package.md)
- [Current limitations](docs/limits/current-limitations.md)
- [Roadmap](docs/planning/roadmap.md)
- [Testing and benchmarks](docs/development/testing-and-benchmarks.md)

## Development

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
go test ./internal/retrieve -bench BenchmarkSearchTenThousandChunks -benchmem
```

The benchmark measures retrieval over 10,000 prepared chunks. It intentionally excludes repository scanning and index construction from the request path.

## Status

Grimoire is in active development. The prepared lexical baseline and exact `o200k_base` budgeting are implemented and tested; language-aware chunking, prepared lexical postings, semantic retrieval, daemon maintenance, and optional Warlock-toolchain evidence providers remain future work.
