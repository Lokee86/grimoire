# Lexicon GDScript adapter

This directory is a self-contained Go command that scans a repository's `.gd` files and emits deterministic Lexicon facts v1 JSONL.

## Usage

From this directory:

```sh
go run . --repo /path/to/repository --output /path/to/facts.jsonl
```

Or build it:

```sh
go build -o lexicon-gdscript .
./lexicon-gdscript --repo /path/to/repository --output /path/to/facts.jsonl
```

The first JSONL record is the v1 header. All subsequent records use the contract's node, edge, and unresolved ordering. Object keys are emitted by Go's sorted JSON map encoding.

## Canonical identities

All node IDs are `sha256:` digests of `lexicon:v1\0gdscript\0<kind>\0<canonical identity>`.

- `repository`: the repository directory's basename.
- `directory`: the normalized relative directory path; the root is `.`.
- `file`: the normalized relative `.gd` path.
- `module`: the normalized relative `.gd` path (one script/module per file).
- declarations: the source path followed by the containing declaration path and name. Duplicate declarations receive a deterministic source-order ordinal.
- `import`: source path, source-order ordinal, and normalized loader expression.

File `content_id` is the SHA-256 digest of the original file bytes. Absolute checkout paths never participate in IDs.

## Current slice

The lexical/parser seam recognizes:

- repository, source directories, `.gd` files, and script modules;
- `class_name` and inner `class` type declarations;
- `extends` targets on a script or named class;
- `func` declarations, including multiline parameter lists and `static`/`async` modifiers;
- `signal`, `const`, and `var` declarations, including a simple declared type;
- `preload()` and `load()` references with static `res://` paths;
- direct calls to uniquely defined functions in the same script.

It emits `contains` and `defines` containment/definition edges, `imports` and `references` edges for import references, `extends` edges for known local classes or scripts, and `calls` edges for uniquely resolved local functions. Dynamic, missing, ambiguous, builtin, and external targets are represented as unresolved records instead of speculative targets.

## Exclusions and limits

The scanner skips `.git`, `.worktrees`, `.workingtrees`, `.warlock`, `node_modules`, `target`, `__pycache__`, `.pytest_cache`, `.bundle`, `vendor`, `.godot`, `.import`, `build`, `dist`, `bin`, and `obj` directories. Only directories on the path to a `.gd` file become directory facts.

This is intentionally a first runnable lexical slice, not a complete GDScript parser. It does not evaluate expressions, follow generated paths, resolve project settings or autoloads, infer dynamic dispatch, parse every annotation, or model all Godot builtins. Unsupported syntax remains evidence-free or unresolved; the adapter does not guess.

## Tests

```sh
go test ./...
```
