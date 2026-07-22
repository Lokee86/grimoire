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

- accept a repository root and output path, plus optional repeated changed-file and removed-file scopes;
- emit exactly one header record followed by sorted fact records;
- normalize repository paths to forward-slash relative paths;
- use SHA-256 stable identities defined by the contract;
- distinguish definite relationships from unresolved or ambiguous references;
- exclude `.git/`, worktree metadata, common language build/vendor directories, and every Warlock state directory: `.ddocs/`, `.lexicon/`, `.arcana/`, `.grimoire/`, `.pitlord/`, `.cantrip/`, `.homunculus/`, `.incubus/`, `.ritual/`, and `.warlock/`;
- avoid embedding consumer-specific policy or storage assumptions.

## Application

Lexicon includes a standalone application that maintains the most recently observed relevant repository state independently of the source repository's Git state.

```text
lexicon init [--languages all|LIST]
lexicon scan
lexicon daemon
lexicon rebuild [--languages LIST]
lexicon languages [list|set]
lexicon status
lexicon doctor
lexicon export --output PATH
lexicon gc [--dry-run]
lexicon consumer <list|add|remove|run>
lexicon version
```

`init` performs the first complete scan. `scan` updates only impacted facts when safe. `daemon` triggers the same transaction from filesystem events. `rebuild` forces complete analysis, adapter fingerprint changes automatically rebuild affected languages, and language selection can disable unused runtimes. The remaining commands inspect health, manage downstream consumers, export verified JSONL libraries, and prune unreachable immutable storage. See [`docs/APPLICATION.md`](docs/APPLICATION.md).

### Ignore policy

An optional repository-root `.lexiconignore` controls which otherwise relevant files Lexicon mirrors and watches. It uses gitignore-compatible patterns, including comments, globs, `**`, path hierarchy, and `!` negation. The policy is applied consistently to complete mirror scans, path syncs, and daemon watch filtering; changing `.lexiconignore` causes the daemon to reload the policy and perform a complete scan.

The permanent exclusions are always enforced and cannot be re-included by `.lexiconignore`: Git and worktree metadata, Warlock state directories (`.ddocs/`, `.lexicon/`, `.arcana/`, `.grimoire/`, `.pitlord/`, `.cantrip/`, `.homunculus/`, `.incubus/`, `.ritual/`, `.warlock/`), dependency directories (`node_modules/`, `vendor/`), and common build or tool state directories (`target/`, `dist/`, `build/`, `.venv/`, `venv/`, `__pycache__/`, `.pytest_cache/`).

After every successful `scan`, Lexicon invokes registered one-shot consumers from `.lexicon/consumers/*.json`. This provides event-driven automation without requiring a consumer daemon: Arcana can register `arcana sync`, while the same registration also runs after daemon-triggered scans. Consumers receive `LEXICON_REPOSITORY`, `LEXICON_STATE_ROOT`, and `LEXICON_SNAPSHOT_ID` in their environment.

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

Each adapter remains self-contained in its own directory. Shared behavior is specified by contract and fixtures rather than a cross-runtime helper library. See [`docs/RELEASE_PACKAGING.md`](docs/RELEASE_PACKAGING.md) for distribution builds and runtime requirements.

## Status

Lexicon now has a version-one fact contract, a complete Go semantic adapter, runnable adapters for Ruby, Python, GDScript, Rust, and TypeScript, and a transactional application layer with a private Git diff mirror, dependency-aware scoped adapter repositories, relationship-topology fallback, validated incremental library merging, single-writer locking, immutable per-file fact objects, snapshot manifests, and atomic current-snapshot publication. Each adapter provides deterministic repository structure, declarations, imports, containment, and language-appropriate inheritance or implementation evidence. Some adapters also provide bounded direct call or reference evidence where the parser can resolve it soundly.

The non-Go adapters are functional foundations rather than complete semantic analyzers. Go includes type-aware internal and external calls, SSA/VTA possible dispatch, interfaces, closures, captures, conversions, and build-tag variants. Unsupported or ambiguous relationships remain unresolved rather than guessed.

## License

Apache License 2.0.
