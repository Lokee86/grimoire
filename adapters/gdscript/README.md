# Lexicon GDScript adapter

This directory contains Lexicon's deterministic GDScript semantic adapter. It scans `.gd` source files and emits Lexicon facts v1 JSONL.

## Usage

```sh
go run . --repo /path/to/project --output /path/to/facts.jsonl
```

The repository may be a Godot project root or any directory containing GDScript. `--output -` writes JSONL to stdout.

## Declarations and relationships

The adapter emits:

- repositories, directories, files, and script modules;
- `class_name` and inner class types;
- functions, signals, constants, and variables;
- lexical `contains` and `defines` relationships;
- `preload()` and `load()` import/reference facts;
- local and path-based inheritance;
- definite `calls` relationships;
- conservative `possible-calls` relationships for ambiguous dispatch and callbacks;
- explicit unresolved classifications for builtin, external, dynamic, missing, and ambiguous targets.

## Static call resolution

Version 0.3 resolves statically defensible repository-local calls through:

- same-script functions and methods;
- correctly owned inner-class methods;
- `self` and `super` dispatch;
- local inheritance and method overriding;
- class constructors and static methods;
- literal preload aliases and direct `preload(...).new()` expressions;
- explicitly typed parameters, locals, members, and return values;
- constructor assignment flow such as `service = Service.new()`;
- untyped assignment flow through resolved parameters;
- argument propagation between resolved callers and callees;
- factory return propagation and chained calls;
- member/property flow when the receiver type is known;
- `Callable(receiver, "method")` references;
- common callback arguments such as `signal.connect(handler)` and `values.map(handler)`;
- literal dynamic invocation names used by `call`, `call_deferred`, `rpc`, and `rpc_id`.

The type-flow analysis is a bounded deterministic fixed point. It combines only concrete local evidence; it does not execute code or guess runtime types.

## Conservative boundaries

GDScript and Godot permit substantial runtime behavior. The adapter intentionally leaves these unresolved rather than inventing graph edges:

- values whose type never becomes statically recoverable;
- computed method names and reflective dispatch;
- scene-tree node types inferred only from `.tscn` structure;
- runtime script replacement and dynamically generated resources;
- computed preload/load paths;
- engine methods, engine classes, and external addons outside the scanned source set;
- autoload names unless they are otherwise represented by local static evidence;
- signal connections or callbacks whose callable target is computed dynamically.

Builtin and engine-owned calls are classified separately from missing repository-local targets. Multiple defensible local targets become `possible-calls` rather than an arbitrary definite edge.

## Source exclusions

The scanner ignores generated state and dependency/build trees including:

- `.git`, `.worktrees`, and `.workingtrees`;
- Warlock tool state directories such as `.lexicon`, `.arcana`, `.grimoire`, and `.warlock`;
- `.godot`, `.import`, `node_modules`, `vendor`, `target`, `build`, `dist`, `bin`, and `obj`;
- common language and test caches.

## Determinism and identity

All IDs use the contract form:

```text
sha256(lexicon:v1\0gdscript\0<kind>\0<canonical identity>)
```

Absolute checkout paths never participate in IDs. Records are sorted deterministically, and repeated scans of unchanged input produce byte-identical JSONL.

## Tests

```sh
go test ./...
```

The suite covers declarations, imports, exclusions, stable IDs, class/static calls, typed receivers, preload aliases, inheritance, `self`/`super`, inner classes, constructor and parameter flow, factory returns, callbacks, contract ordering, and repeat-run determinism.
