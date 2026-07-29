# Lexicon LotusScript adapter

This directory contains Lexicon's deterministic LotusScript adapter. It scans dedicated `.ls` and `.lss` source files plus Domino ODP `.lsa` and `.lsdb` artifacts and emits facts-v1 JSONL.

## Current semantic coverage

The adapter models:

- repository, directory, file, and script-library modules;
- classes, user-defined types, functions, subs, methods, constructors, properties, fields, constants, local variables, and parameters;
- `Use` and `UseLSX` declarations, including repository-local script-library resolution;
- class inheritance;
- definite unqualified, `Me`, class-qualified, inherited, constructor, and declared-type receiver calls when one repository-local target exists;
- explicit unresolved evidence for built-ins, external libraries, ambiguous names, and receivers without sufficient type evidence;
- line continuations, apostrophe comments, `Rem` comments, `%REM` blocks, quoted strings, and pipe-delimited strings;
- Domino ODP agents stored either as structured `<code><lotusscript>` DXL or base64 `$AssistAction` raw-item payloads.

Array and list indexing through declared variables is not emitted as a call. Constructor expressions are emitted once rather than as both a constructor and a plain function call.

## Usage

```text
go run . --repo /path/to/repository --output -
```

The adapter accepts incremental scope flags so the application runner can invoke it uniformly, but currently emits a complete stream. Incremental narrowing will be added only after source ownership remains sound across script-library resolution.

## Conservative boundaries

Receiver resolution uses declared parameter, local-variable, and class-field types. Assignment-only inference and unconstrained runtime dispatch remain unresolved. Domino runtime classes such as `NotesSession` remain external unless their implementation is present in the repository.

Content detection for LotusScript stored under generic `.txt`, `.bas`, or `.vb` extensions is not implemented. Base64 DXL extraction preserves ASCII source reliably; non-ASCII LMBCS payload decoding is not yet implemented. Empty `.lsdb` artifacts are ignored.

## Calibration

The pinned calibration set is recorded in `evaluation/corpus.json`:

- DUnit: calibration split;
- jsonparser-ls: validation split;
- HCL Volt MX LotusScript toolkit: holdout split.

The dated results and remaining limits are recorded in `docs/LOTUSSCRIPT_ADAPTER_VALIDATION.md`.

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
