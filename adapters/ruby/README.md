# Lexicon Ruby adapter

The first runnable Ruby adapter slice uses Ruby's standard-library `Ripper` parser boundary and emits Lexicon facts v1 JSONL.

## Usage

From the repository root:

```sh
ruby adapters/ruby/lexicon_ruby.rb \
  --repo /path/to/ruby/repository \
  --output /path/to/facts.jsonl
```

The adapter scans `*.rb` files in lexical path order. It excludes `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, `.bundle/`, `vendor/`, `node_modules/`, `target/`, `build/`, `dist/`, `tmp/`, `log/`, and `coverage/` directories. Paths in facts are repository-relative and use forward slashes.

Validate the output against the root contract with:

```sh
python tools/validate_jsonl.py /path/to/facts.jsonl
```

## Canonical identities

- `repository`: repository directory basename.
- `directory`: normalized repository-relative directory path; the root is `.`.
- `file`: normalized repository-relative file path.
- `module` and `type`: source path plus qualified Ruby constant name.
- `method`: source path plus qualified method name (`Owner#name` for instance methods and `Owner.name` for singleton methods).
- `constant`: source path plus qualified Ruby constant name.
- `import`: source path, token position, and required target.
- External superclass placeholders use an `external` canonical prefix so local declarations can be resolved without inventing a source path.

## Supported slice

The adapter emits repository, directory, file, module, type, method, constant, and import nodes; `contains` and `defines` declaration edges; `imports` edges for literal `require`, `require_relative`, and `load`; `extends` edges for simple constant superclasses; and conservative `calls` edges for bare calls that resolve to exactly one method defined on the same class or module owner. Call evidence retains repository-relative source spans and source-derived expressions. File content IDs are SHA-256 identities of the original bytes, and all records are deterministically sorted with lexicographically ordered JSON object keys.

Explicit receivers, inherited/framework dispatch, blocks, `send`, common Ruby metaprogramming calls, dynamic require targets, singleton classes, alias/undef forms, and syntax that Ripper cannot represent as a complete AST remain unresolved or unmodeled instead of being guessed.

## Tests

```sh
ruby adapters/ruby/test/test_adapter.rb
```
