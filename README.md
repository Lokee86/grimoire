# Lexicon

Lexicon is the shared language-analysis subsystem for the Warlock toolchain.

It provides one reusable adapter per programming language and a normalized fact contract for repository symbols and relationships. Arcana, Pitlord, Grimoire, Demon Docs, Homunculus, and other tools can consume the same language facts without maintaining duplicate parsers.

## Initial adapters

| Language | Implementation | Parser | Status |
| --- | --- | --- | --- |
| Ruby | Ruby | `Ripper` | Runnable slice |
| Python | Python | Standard-library `ast` | Runnable slice |
| GDScript | Go | Dedicated lexical/parser seam | Runnable slice |
| Rust | Rust | `syn` plus Cargo metadata | Runnable slice |
| TypeScript | TypeScript | TypeScript compiler API | Runnable slice |
| Go | Go | `go/parser`, `go/types`, SSA, and VTA | Complete semantic adapter |


## Contract

Adapters are independently executable and emit deterministic Lexicon JSONL. The fact contract is defined in [`spec/facts-v1.md`](spec/facts-v1.md), and the immutable application storage contract is defined in [`spec/snapshots-v1.md`](spec/snapshots-v1.md).

Every adapter must:

- accept a repository root and output path;
- emit exactly one header record followed by sorted fact records;
- normalize repository paths to forward-slash relative paths;
- use SHA-256 stable identities defined by the contract;
- distinguish definite relationships from unresolved or ambiguous references;
- exclude `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, and common language build/vendor directories;
- avoid embedding consumer-specific policy or storage assumptions.

## Application

Lexicon includes a standalone application that maintains the most recently observed relevant repository state independently of the source repository's Git state.

```text
lexicon init
lexicon scan
lexicon daemon
```

`init` performs the first complete scan. `scan` replaces Lexicon's private source mirror, uses its internal Git diff to update affected language libraries, and publishes an immutable content-addressed snapshot. `daemon` watches the filesystem, updates changed paths after a short debounce, and periodically reconciles the complete repository. See [`docs/APPLICATION.md`](docs/APPLICATION.md).

## Repository layout

```text
cmd/lexicon/
internal/
adapters/
    gdscript/
    go/
    python/
    ruby/
    rust/
    typescript/
docs/
spec/
tools/
```

Each adapter remains self-contained in its own directory. Shared behavior is specified by contract and fixtures rather than a cross-runtime helper library.

## Status

Lexicon now has a version-one fact contract, a complete Go semantic adapter, runnable adapters for Ruby, Python, GDScript, Rust, and TypeScript, and a transactional application layer with a private Git diff mirror, single-writer locking, immutable per-file fact objects, snapshot manifests, and atomic current-snapshot publication. Each adapter provides deterministic repository structure, declarations, imports, containment, and language-appropriate inheritance or implementation evidence. Some adapters also provide bounded direct call or reference evidence where the parser can resolve it soundly.

The non-Go adapters are functional foundations rather than complete semantic analyzers. Go includes type-aware internal and external calls, SSA/VTA possible dispatch, interfaces, closures, captures, conversions, and build-tag variants. Unsupported or ambiguous relationships remain unresolved rather than guessed.

## License

Apache License 2.0.
