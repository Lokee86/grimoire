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
| Go | Pending migration | Existing Arcana adapter | Deferred |

The Go adapter remains in Arcana temporarily while other Arcana work is active.

## Contract

Adapters are independently executable and emit deterministic Lexicon JSONL. The contract is defined in [`spec/facts-v1.md`](spec/facts-v1.md).

Every adapter must:

- accept a repository root and output path;
- emit exactly one header record followed by sorted fact records;
- normalize repository paths to forward-slash relative paths;
- use SHA-256 stable identities defined by the contract;
- distinguish definite relationships from unresolved or ambiguous references;
- exclude `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, and common language build/vendor directories;
- avoid embedding consumer-specific policy or storage assumptions.

## Repository layout

```text
adapters/
    gdscript/
    python/
    ruby/
    rust/
    typescript/
spec/
tools/
```

Each adapter remains self-contained in its own directory. Shared behavior is specified by contract and fixtures rather than a cross-runtime helper library.

## Status

Lexicon now has a version-one fact contract and runnable first slices for Ruby, Python, GDScript, Rust, and TypeScript. Each adapter provides deterministic repository structure, declarations, imports, containment, and language-appropriate inheritance or implementation evidence. Some adapters also provide bounded direct call or reference evidence where the first parser slice can resolve it soundly.

These are functional foundations rather than complete semantic analyzers. Dynamic dispatch, generated code, framework-specific behavior, external-package resolution, and deeper type or call analysis remain explicit later work. Unsupported or ambiguous relationships are emitted as unresolved facts rather than guessed edges.

## License

Apache License 2.0.
