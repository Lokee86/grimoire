# Go adapter

The Go adapter scans one Go module and emits deterministic Lexicon facts v1 JSONL. It combines repository-wide AST extraction with `golang.org/x/tools/go/packages`, Go type information, SSA, and variable-type analysis.

## Run

```bash
go run . -repo /path/to/module -output facts.jsonl
python ../../tools/validate_jsonl.py facts.jsonl
```

## Modeled semantics

The adapter models:

- packages, files, types, functions, methods, tests, imports, and containment;
- internal, standard-library, external, and built-in callable contracts;
- definite calls through `calls`;
- conservative dynamic targets through `possible-calls`;
- interfaces and implementation relationships;
- function values, callbacks, method values, and returned function values;
- closures, nested calls, and captured variables;
- type conversions through `converts-to`;
- mutually exclusive build-tag declarations under one logical symbol identity;
- AST-only callable contracts for files excluded from the active host build.

The scanner excludes `.git`, `.worktrees`, `.workingtrees`, `.ddocs`, `.lexicon`, `.arcana`, `.grimoire`, `.pitlord`, `.cantrip`, `.homunculus`, `.incubus`, `.ritual`, `.warlock`, and `vendor` directories. Every node ID follows the Lexicon SHA-256 identity contract. File content IDs hash the unmodified file bytes.

`calls` indicates one definite callable contract. Multiple sound runtime targets remain explicit `possible-calls` relationships rather than being promoted to certainty.

## Dispatch and relationship boundaries

The semantic pass emits `extends` for named embedded/base relationships, `implements` for repository-local interface satisfaction, `uses-trait`/`includes` for embedded implementation relationships, and `overrides` from concrete methods to inherited or interface contract methods. Interface declarations are contracts only and are never runtime call targets. A proven concrete target emits `calls`; multiple concrete implementations emit `possible-calls`.

Reflection, `reflect`-derived calls, external packages, generated methods without repository evidence, unsafe runtime mutation, and otherwise dynamic function values remain unresolved or externally classified. Build-tag variants are merged only where their callable identity is stable.

## Canonical identities

The SHA-256 payload defined by the shared contract uses these Go identity strings:

| Kind | Canonical identity |
| --- | --- |
| repository | `repository:<module path>` |
| directory | `directory:<repository-relative path>` |
| file | `file:<repository-relative path>` |
| module | `package:<import path>:<package name>` |
| namespace | `namespace:<import path or synthetic namespace>` |
| import | `import:<internal-or-external>:<import path>` |
| type | `type:<import path>:<type name>` |
| function | `function:<import path>:<function name>` |
| method | `method:<import path>:<receiver>.<method name>` |
| interface method | `interface-method:<import path>:<interface>.<method name>` |
| test | `test:<import path>:<test name>` |
| closure | `closure:<import path>:<file>:<line>:<column>` |
| captured variable | `variable:<module>:<file>:<line>:<column>:<name>` |

Compiler-generated wrappers and external closures use deterministic `ssa-function:` identities. Synthetic built-in and type-expression nodes use stable language namespaces such as `go:builtins` and `go:types`. Absolute checkout paths are never part of an identity.

## Dependency semantics

The adapter emits repository `depends-on` edges for literal `go.mod` `require` directives, both single-line and parenthesized forms, and for literal `replace` directives. Each target is a facts-v1 `module` node using `dependency:go:<normalized-target>` identity; its synthetic path is `.lexicon/dependencies/go/...`. Edges carry deterministic `category`, `constraint`, `source`, `optional`, `dev`, `build`, `peer`, and `path` attributes. Repository-local Go imports additionally emit module-to-module `depends-on` edges when the imported package is uniquely scanned, while preserving `imports`.

Malformed directives, dynamic module construction, and unresolved external package contents are not inferred. The adapter does not execute `go.mod` or install dependencies.
## Dataflow facts

The adapter emits conservative `reads` and `writes` edges from the containing callable to repository-local parameters, variables, constants, and fields. Assignments write, compound assignments and increment/decrement read and write, and initializer, argument, and return expressions contribute reads. Lexical shadowing is respected. Unresolved selectors, built-ins, external package values, reflection, and unsafe aliasing are omitted rather than guessed.
