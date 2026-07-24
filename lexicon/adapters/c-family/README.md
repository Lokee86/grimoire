# Lexicon C and C++ adapter

This adapter uses the official Tree-sitter C and C++ grammars to emit deterministic Lexicon facts-v1 JSONL for mixed C/C++ repositories. C and C++ share one adapter, one cross-file symbol view, and one stable `c-family` identity namespace; each file records its actual language as `c` or `cpp`.

## Language surface

The adapter owns:

- C and C++ source files: `.c`, `.C`, `.cc`, `.cp`, `.cpp`, `.cxx`, and `.c++`;
- C and C++ headers and implementation includes: `.h`, `.hh`, `.hpp`, `.hxx`, `.h++`, `.inc`, `.inl`, `.ipp`, and `.tpp`;
- `compile_commands.json` as a per-file language hint;
- `CMakeLists.txt` as repository detection evidence.

C++-specific header extensions are parsed as C++. Ambiguous `.h` and `.inc` files use `compile_commands.json` when available, then conservative syntax markers; otherwise they are parsed as C.

## Usage

From this adapter directory:

```text
go run . --repo /path/to/repository --output /path/to/facts.jsonl
```

A packaged Lexicon distribution invokes `lexicon-c-family` automatically. `--output -` writes JSONL to stdout. Incremental calls accept repeated `--changed-file` and `--removed-file` arguments.

## Implemented semantics

The adapter emits:

- file and translation-unit module nodes;
- namespaces;
- classes, structs, unions, enums, typedefs, and aliases;
- functions, methods, constructors, prototypes, parameters, fields, variables, constants, enum members, and macros;
- repository-local and unresolved include evidence;
- class inheritance;
- definite calls when one repository-local callable resolves;
- `possible-calls` for multiple defensible overload or same-name targets;
- conservative reads and writes for parameters, locals, and fields;
- explicit unresolved call, include, inheritance, and parse evidence.

Repository-local declarations are resolved across C and C++ files. Function definitions are preferred over matching prototypes. Scope-chain resolution covers namespaces, types, methods, and local callable ownership without treating same-named global declarations as interchangeable when a narrower match exists.

## Identities

All nodes use the facts-v1 SHA-256 identity contract with language `c-family`. Canonical identities include repository-relative source ownership, semantic kind, qualified name, and callable signature where required. Absolute checkout paths are never included.

The stream header language is always `c-family`. File and declaration attributes include `language: c` or `language: cpp` so consumers can distinguish the parsed grammar without splitting the shared semantic graph.

## Includes and build context

Quoted includes resolve first relative to the including file, then by exact repository path, then by unique basename. System includes and missing local headers remain unresolved rather than becoming fabricated repository dependencies.

`compile_commands.json` currently selects the C or C++ grammar for listed files. Compiler defines, include search paths, target triples, generated headers, and conditional preprocessing are not yet replayed.

## Conservative boundaries

The adapter does not run a compiler or preprocessor. Consequently:

- inactive conditional branches may still be parsed;
- macro expansion and token-pasting semantics are not reconstructed;
- generated declarations and headers are unavailable unless present in the repository;
- template instantiation, overload ranking, ADL, implicit conversions, virtual dispatch, and function-pointer flow remain conservative;
- member calls without a uniquely provable repository target remain unresolved or possible;
- Objective-C and CUDA-specific semantics are outside this adapter.

Tree-sitter recovery permits partial facts from incomplete source. Files containing parse errors are marked and emit unresolved `unsupported-form` evidence.

## Incremental behavior

The adapter parses the complete C-family source set to preserve cross-file resolution, then emits only changed-file-owned records during incremental analysis. Removed paths are declared in the facts-v1 header. Shared synthetic replacement is not claimed, so incremental streams set `shared_complete: false`.

## Development

```text
go test ./...
go test -race ./...
```

The suite covers mixed C/C++ extraction, header language inference, repository-local includes, inheritance, calls, dataflow, deterministic repeated output, incremental ownership, and CLI output.
