# Lexicon LotusScript adapter

This directory contains Lexicon's deterministic LotusScript adapter. It scans dedicated `.ls` and `.lss` source files plus Domino ODP `.lsa` and `.lsdb` artifacts and emits facts-v1 JSONL.

## Current semantic coverage

The adapter models:

- repository, directory, file, and script-library modules;
- classes, user-defined types, functions, subs, methods, constructors, properties, fields, constants, local variables, and parameters;
- class and `Type` members declared without `Dim`, as required by LotusScript member syntax;
- `Use` and `UseLSX` declarations, including repository-local script-library resolution;
- transitive script-library visibility, `Option Public`, and explicit `Public`, `Private`, and `Protected` declarations;
- class inheritance, including private inherited members used through `Me` or a receiver of the current class;
- definite unqualified, `Me`, class-qualified, inherited, constructor, and declared-type receiver calls when one visible repository-local target exists;
- conservative `reads` and `writes` relationships for parameters, locals, module variables and constants, class fields, assignments, array indexing, and `ReDim`;
- explicit unresolved evidence for built-ins, external libraries, ambiguous names, and receivers without sufficient type evidence;
- line continuations, colon-separated statements, labels, apostrophe comments, `Rem` comments, `%REM` blocks, quoted strings, pipe-delimited strings, date literals, and `With ... End With` member shorthand;
- Domino ODP agents stored either as structured `<code><lotusscript>` DXL or base64 `$AssistAction` raw-item payloads.

Array and list indexing through declared variables is not emitted as a call. Constructor expressions are emitted once rather than as both a constructor and a plain function call. Cross-file calls and dataflow are emitted only when the target is visible through the transitive `Use` graph.

## Usage

```text
go run . --repo /path/to/repository --output -
```

The adapter accepts incremental scope flags so the application runner can invoke it uniformly, but currently emits a complete stream. Incremental narrowing will be added only after source ownership remains sound across script-library resolution.

## Conservative boundaries

Receiver resolution uses declared parameter, local-variable, module-variable, and class-field types. Assignment-only type inference, implicit undeclared-variable modeling, computed dispatch, and unconstrained runtime dispatch remain unresolved. Domino runtime classes such as `NotesSession` remain external unless their implementation is present in the repository.

Dataflow is deliberately symbol-based and local to declarations the adapter can prove. It does not attempt alias analysis, container element identity, interprocedural value flow, property side effects, or mutation performed by external runtime objects.

Content detection for LotusScript stored under generic `.txt`, `.bas`, or `.vb` extensions is not implemented. Base64 DXL extraction preserves ASCII source reliably; non-ASCII LMBCS payload decoding is not yet implemented. Empty `.lsdb` artifacts are ignored.

## Calibration

The pinned calibration set is recorded in `evaluation/corpus.json`:

- DUnit: calibration split;
- jsonparser-ls: validation split;
- HCL Volt MX LotusScript toolkit: holdout split.

The gates require deterministic calls, imports, inheritance, reads, and writes, including selected exemplar edges. The dated results and remaining limits are recorded in `docs/LOTUSSCRIPT_ADAPTER_VALIDATION.md`.

## Identities

Stable IDs use:

```text
lexicon:v1\0lotusscript\0<kind>\0<canonical identity>
```

Canonical identities are based on repository-relative file paths and case-normalized declaration ownership. Absolute checkout paths are never included.

## Tests

```text
go test ./...
```

## Code map

| Concern | Primary implementation | Related tests |
| --- | --- | --- |
| Entry and repository discovery | `main.go`, `discovery.go`, `dxl.go` | `main_test.go`, `calibration_test.go` |
| Parsing and syntax | `parser.go`, `syntax.go`, `model.go` | `main_test.go`, `scope_test.go` |
| Declarations, visibility, and types | `declaration_helpers.go`, `types.go` | `scope_test.go`, `dataflow_test.go` |
| Calls and semantic scope | `calls.go`, `analysis.go` | `scope_test.go`, `calibration_test.go` |
| Reads and writes | `dataflow.go` | `dataflow_test.go` |
| Fact emission | `facts.go` | main, scope, calibration, and dataflow tests |

The adapter preserves unresolved evidence when LotusScript or DXL behavior cannot be established statically.
