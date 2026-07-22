# Lexicon

Lexicon is the shared language-analysis subsystem for the Warlock toolchain.

It provides one reusable adapter per programming language and a normalized fact contract for repository symbols and relationships. Arcana, Pitlord, Grimoire, Demon Docs, Homunculus, and other tools can consume the same language facts without maintaining duplicate parsers.

## Initial adapters

| Language | Implementation | Initial parser |
| --- | --- | --- |
| Ruby | Ruby | Prism when available, Ripper-compatible fallback boundary |
| Python | Python | Standard-library `ast` |
| GDScript | Go | Lexical parser with an explicit upgrade seam for a full parser |
| Rust | Rust | `syn` plus Cargo metadata |
| TypeScript | TypeScript | TypeScript compiler API |
| Go | Pending migration | Existing Arcana adapter |

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

The repository and version-one fact contract are being established alongside initial adapters for Ruby, Python, GDScript, Rust, and TypeScript. The first implementations focus on repository structure, declarations, imports, containment, inheritance/implementation where practical, and direct call/reference evidence.

## License

Apache License 2.0.
