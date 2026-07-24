# Lexicon language adapters

This directory owns Lexicon's language-specific semantic extraction implementations. Each adapter is independently executable and emits the shared facts-v1 JSONL contract.

## Purpose

Adapters translate language syntax and static semantic evidence into normalized repository facts:

- source structure and declarations;
- imports and dependencies;
- containment and definition ownership;
- inheritance, implementation, trait, mixin, and override relationships;
- definite calls and conservative possible calls;
- reads and writes where local dataflow is sound;
- explicit unresolved evidence for unsupported, ambiguous, dynamic, external, or built-in targets.

## Does not own

Adapters do not own:

- Lexicon repository initialization or CLI behavior;
- incremental plan selection or scoped-repository construction;
- snapshot manifests, object storage, publication, recovery, or garbage collection;
- Arcana graph storage or query policy;
- Grimoire ranking or context-package construction;
- consumer-specific interpretation of facts.

Those boundaries keep one language implementation reusable by every Warlock consumer.

## Direct folders

| Folder | Language surface | Implementation | Primary semantic frontend |
| --- | --- | --- | --- |
| [go/](go/README.md) | Go | Go | `go/parser`, `go/types`, packages, SSA, and VTA |
| [gdscript/](gdscript/README.md) | GDScript | Go | Dedicated parser and bounded type-flow model |
| [generic/](generic/README.md) | Curated unsupported source extensions | Go | Conservative line-oriented fallback |
| [python/](python/README.md) | Python | Python | Standard-library `ast` |
| [ruby/](ruby/README.md) | Ruby | Ruby | Standard-library `Ripper` |
| [rust/](rust/README.md) | Rust | Rust | `syn` and Cargo metadata |
| [typescript/](typescript/README.md) | JavaScript, TypeScript, and Svelte script blocks | TypeScript | TypeScript compiler API and Svelte script extraction |

## Shared execution contract

Every adapter must:

- accept a repository root and output destination;
- support `-` or the documented equivalent for stdout when practical;
- emit exactly one facts-v1 header followed by canonical nodes, edges, and unresolved records;
- use normalized repository-relative forward-slash paths;
- exclude Git/worktree metadata, Warlock state directories, dependencies, build outputs, and language caches;
- use stable SHA-256 identities without absolute checkout paths;
- preserve definite versus possible relationship semantics;
- retain source ownership for records that may be incrementally replaced;
- produce byte-identical output for identical input and configuration;
- avoid executing analyzed application code or manifests;
- emit a full stream when file ownership cannot be determined safely.

Incremental-capable adapters additionally accept changed and removed file scopes defined by [facts-v1](../spec/facts-v1.md). Shared synthetic records may replace the complete shared set only when the stream declares `shared_complete: true`.

## Relationship policy

A `calls` edge means one definite statically identified callable contract. Multiple defensible concrete runtime targets use `possible-calls`. Interface, trait, protocol, and mixin declarations remain contracts unless the language provides separate runtime evidence.

Adapters must not select an arbitrary same-named declaration to avoid an unresolved record. Unsupported or ambiguous evidence remains explicit.

`imports` records source-level import evidence. `depends-on` records package, module, plugin, resource, or manifest dependency evidence. One does not replace the other.

## Parallelism

The application may execute independent adapters concurrently. An adapter must therefore keep all mutable execution state process-local and must not write shared repository files outside its requested output.

The Go adapter also accepts worker, logical-shard, and merge-fan-in parameters. Its shard-local semantic work must merge deterministically and produce the same facts for every valid execution plan.

Other adapters may add safe process or project partitioning later, but partition boundaries must preserve language semantics and deterministic ownership. Arbitrary file sharding is not acceptable when it changes the available type, import, or dispatch context.

## Adding an adapter

A new adapter requires:

1. a self-contained executable entry point;
2. documented runtime and build requirements;
3. canonical identity definitions;
4. permanent exclusion behavior;
5. declaration, relationship, unresolved, dependency, dataflow, and determinism tests;
6. application registry and runner integration;
7. full-stream support before incremental narrowing;
8. an adapter README covering usage, modeled semantics, conservative boundaries, identities, tests, dependencies, and dataflow;
9. semantic acceptance evidence;
10. status and documentation index updates.

See [Development and verification](../docs/DEVELOPMENT.md) and [Semantic acceptance gates](../docs/SEMANTIC_ACCEPTANCE.md).

## Placement rules

Language-specific parser, resolver, model, emitter, fixtures, and tests belong inside the owning adapter folder.

Cross-language record meaning belongs in `spec/`. Application orchestration belongs in `internal/scan` and `internal/adapters`. Do not create a cross-runtime helper dependency merely to share implementation convenience; share behavior through the versioned contract and acceptance fixtures.
