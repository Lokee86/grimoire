# Lexicon Rust adapter

This directory contains a self-contained Rust CLI that emits Lexicon facts v1 JSONL for a Cargo repository.

## Usage

From this directory:

```text
cargo run -- --repo /path/to/repository --output /path/to/facts.jsonl
```

The adapter uses `cargo_metadata` to identify workspace packages and Cargo targets, and `syn` to parse Rust source. `--repo` must point to a Cargo workspace or package containing `Cargo.toml`. The output parent directory is created when needed.

## Emitted facts

The current slice emits:

- one repository node, source directory and Rust file nodes with content identities;
- Cargo target crate nodes represented as `module` nodes;
- inline and external module declarations;
- structs and enums as `type` nodes;
- traits and trait methods;
- free functions and impl methods;
- `use` declarations as `import` nodes;
- containment and definition edges for declarations and files;
- statically resolved local import edges;
- simple local `impl Trait for Type` `implements` edges;
- unresolved records for macro-generated declarations, missing module files, unsupported imports, and external or missing import/implementation targets.

Node identity uses the contract form `lexicon:v1\\0rust\\0<kind>\\0<canonical identity>`. Declaration identities are workspace-relative Cargo package/target/module qualified names; file and directory identities are normalized repository-relative paths. No absolute checkout path is included in an identity or emitted path.

## Current limits

- The scanner parses `.rs` files reachable from Cargo targets and then processes remaining package-local Rust files as crate-level fallback files.
- Import resolution is intentionally conservative: simple paths are resolved only when a local declaration is known. Globs, grouped imports, aliases, and external symbols become unresolved records.
- Trait implementation extraction is limited to local, syntactically simple `impl Trait for Type` relationships. Generic, macro-generated, dynamic, and external relationships are not guessed.
- Macro bodies, fields, expressions, calls, lifetimes, and type references are not yet modeled as separate facts.
- Source spans use parser locations; synthetic or unavailable spans are omitted by the parser boundary.

The adapter writes records in the contract order: header, canonically sorted nodes, edges, and unresolved records. JSON object keys are lexicographically ordered.
