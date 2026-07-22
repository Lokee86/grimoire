# Go repository adapter

The adapter scans a Go module and writes canonical Arcana repository facts. It
combines repository-wide AST extraction with Go type information, SSA, and
variable-type analysis (VTA).

Requirements:

- Go 1.22 or newer;
- a repository root containing `go.mod`.

From this directory:

```text
go run . -repo /path/to/repository -output /path/to/repository.facts.tsv
go test ./...
```

The scanner skips `.git`, `.worktrees`, `.workingtrees`, and `vendor`
directories. Deterministic UTF-8 TSV output uses fact format version 2:

- `N` records for nodes;
- `E` records for resolved relationships;
- `U` records only when no sound target or callable contract can be represented.

## Semantic model

The adapter models:

- internal functions, methods, recursion, tests, and packages;
- standard-library and external API symbols without indexing dependency source;
- Go built-ins as callable symbol nodes;
- type conversions through `converts-to` edges rather than call edges;
- interfaces, interface methods, embedded interfaces, and `implements` edges;
- definite dispatch through `calls` edges;
- conservative multi-target dispatch through `possible-calls` edges;
- function variables, callbacks, method values, and returned function values via SSA/VTA;
- closures as independent function nodes, including calls inside nested closures;
- captured locals and parameters as variable nodes reached through `references` edges;
- mutually exclusive build-tag declarations as one canonical package-level symbol;
- AST-only callable contracts for files excluded from the active host build.

`calls` means the analysis has one definite callable contract. `possible-calls`
means more than one runtime target remains sound. Interface calls retain both the
interface-method contract and any repository implementations discovered by VTA.
External closures or wrappers that have no source declaration receive stable
synthetic symbol nodes rather than disappearing from the graph.

Node identities and file content IDs use FNV-1a 64-bit with the same offset basis
and prime as Arcana's Rust `StableHasher`. Identity strings include a kind prefix
to keep categories distinct.

The command replaces an existing fact output file. Arcana's importer converts
the facts into a verified repository snapshot containing the packed graph,
catalogue, unresolved records, source facts, and manifests.
