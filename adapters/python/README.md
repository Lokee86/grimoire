# Lexicon Python adapter

A self-contained Python package that emits Lexicon facts v1 JSONL using only the Python standard library and `ast`.

## Usage

From the repository root:

```sh
PYTHONPATH=adapters/python python -m lexicon_python \
  --repo /path/to/repository \
  --output /path/to/facts.jsonl
```

`--output -` writes JSONL to stdout. The output directory is created when needed.

The stream starts with the v1 header, then contains nodes, edges, and unresolved records. Keys are serialized lexicographically and records use the contract's canonical ordering. The generated stream can be checked with `python tools/validate_jsonl.py facts.jsonl`.

## Node identities

All node IDs are `sha256:` digests of the v1 identity form. The canonical identities for this slice are:

- `repository`: repository root basename;
- `directory`: normalized repository-relative directory path (`.` is the root);
- `file`: normalized repository-relative Python file path;
- `module`: dotted module name (`pkg/__init__.py` is `pkg`, and a root `__init__.py` uses the repository basename);
- `type`: dotted module and lexical class name;
- `function`: dotted module and lexical function name;
- `method`: dotted module, class, and method name;
- `import`: module name plus source-line occurrence and bound name.

File nodes also carry the SHA-256 content identity of the original file bytes. Absolute checkout paths are never included in identities or paths.

## Supported facts

The adapter emits repository, directory, file, module, type, function, method, and import nodes. It emits `contains` edges for repository structure and file/module ownership, `defines` edges for declarations, `imports` edges for resolved in-repository modules/symbols, `extends` edges for simple resolved inheritance, and direct `calls` edges when a callable is statically identified. It also tracks simple function-local assignments such as `worker = Worker()` and resolves a later `worker.run()` only when the constructor and method are uniquely known; conflicting or control-flow-dependent assignments invalidate that inference. Unsupported, dynamic, external, builtin, ambiguous, or missing targets are represented as `unresolved` records instead of guessed edges.

Python files are scanned in deterministic order while excluding `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, `__pycache__/`, `.pytest_cache/`, `.bundle/`, `node_modules/`, `target/`, build/dist/virtual-environment directories, and vendor directories.

## Current limits

- Resolution is repository-local and syntax-based; it does not execute imports or inspect installed packages.
- Local constructor-flow resolution is deliberately intraprocedural and linear; branch-dependent, conflicting, attribute-based, factory-produced, and reassigned values remain unresolved.
- Calls through computed attributes, factories, dynamic imports, and other non-dotted callable expressions remain unresolved.
- Builtin classification uses the running Python interpreter's authoritative builtin namespace.
- Inheritance resolution covers simple names and dotted names that map to scanned modules or declarations; metaclasses, generated bases, and runtime mutation are not inferred.
- Python files that cannot be decoded or parsed still produce file/module facts plus an unresolved parse record, but no declarations from the file.
- Imports inside function bodies are represented as import facts, but only module-level bindings participate in cross-file symbol resolution.
