# Lexicon Ruby adapter

The Ruby adapter uses the standard-library `Ripper` parser and emits deterministic Lexicon facts v1 JSONL without loading the analyzed application.

## Usage

From the repository root:

```sh
ruby adapters/ruby/lexicon_ruby.rb \
  --repo /path/to/ruby/repository \
  --output /path/to/facts.jsonl
```

The adapter scans `*.rb` files in lexical path order. It excludes `.git/`, `.worktrees/`, `.workingtrees/`, `.ddocs/`, `.lexicon/`, `.arcana/`, `.grimoire/`, `.pitlord/`, `.cantrip/`, `.homunculus/`, `.incubus/`, `.ritual/`, `.warlock/`, `.bundle/`, `vendor/`, `node_modules/`, `target/`, `build/`, `dist/`, `tmp/`, `log/`, and `coverage/`. Paths in facts are repository-relative and use forward slashes.

Validate output with:

```sh
python tools/validate_jsonl.py /path/to/facts.jsonl
```

## Static call graph

A single proven repository-local target emits `calls`. Multiple valid targets emit `possible-calls` plus an `ambiguous-target` record. Calls that depend on framework behavior, external libraries, reflection, duck typing, or runtime mutation remain explicitly classified instead of being guessed.

The static resolver covers:

- bare, explicit-receiver, indexed, operator, `super`, `yield`, and constructor calls;
- instance methods, `def self.name`, and `class << self` singleton methods;
- lexical constants and nested namespaces;
- inheritance, `include`, `prepend`, `extend`, and explicit mixin installation such as `Host.include(Concern)`;
- Ruby method lookup order across prepends, owners, includes, and superclasses;
- `module_function`, method aliases, `undef`, and synthetic constructors;
- local variables, instance variables, class variables, parameters, optional and keyword defaults, assignments, and branch unions;
- callsite argument propagation and factory return propagation;
- blocks, lambdas, block parameters, explicit block parameters, and `yield` relationships;
- chained calls through statically recovered return values;
- `Struct.new`, `Data.define`, `Class.new`, and `Module.new` constant factories;
- generated Struct/Data accessors;
- common Active Record model and relation return shapes, restricted to types descending from `ActiveRecord::Base`;
- Rails concern-style `included` and `class_methods` blocks;
- literal `require`, `require_relative`, and `load` imports.

Callsite spans are retained, so repeated calls from the same source method remain separate graph evidence. Reopened Ruby classes and modules contribute to the same semantic owner while preserving source-specific declarations.

## Emitted facts

The adapter emits repository, directory, file, module, type, method, constructor, function, constant, and import nodes. Relationships include `contains`, `defines`, `imports`, `extends`, `includes`, `calls`, and `possible-calls`.

File content IDs are SHA-256 identities of the original bytes. Node IDs use the Lexicon v1 canonical identity contract. Output records and JSON object keys are deterministically sorted.

## Deliberate unresolved boundaries

- `send`, `public_send`, `eval`, `class_eval`, `module_eval`, `define_method`, and similar metaprogramming remain dynamic.
- Rails-generated model accessors, controller helpers, callbacks, routes, validations, migrations, and test DSL methods are classified as external framework behavior unless ordinary Ruby declarations prove a local target.
- Duck-typed parameters and injected objects remain dynamic when their complete runtime type set cannot be established.
- Monkey patching, refinements, autoload behavior, dynamic constant lookup, and non-literal reflection cannot be made complete through static Ripper analysis alone.
- Core-library calls and literal receiver operations are classified as builtin rather than linked to repository nodes.
- Parse failures produce file facts and an unresolved parse record, but no declarations from the failed source.

## Tests

```sh
ruby adapters/ruby/test/test_adapter.rb
```
