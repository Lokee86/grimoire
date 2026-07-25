# Lexicon C and C++ adapter

This adapter uses the official Tree-sitter C and C++ grammars to emit deterministic Lexicon facts-v1 JSONL for mixed C/C++ repositories. C and C++ share one adapter, one cross-file symbol view, and one stable `c-family` identity namespace; each file records its actual language as `c` or `cpp`.

## Language surface

The adapter owns:

- C and C++ source files: `.c`, `.C`, `.cc`, `.cp`, `.cpp`, `.cxx`, and `.c++`;
- C and C++ headers and implementation includes: `.h`, `.hh`, `.hpp`, `.hxx`, `.h++`, `.inc`, `.inl`, `.ipp`, and `.tpp`;
- `compile_commands.json` as a per-file language hint;
- `CMakeLists.txt` as repository detection evidence.

C++-specific header extensions are parsed as C++. Ambiguous `.h` and `.inc` files use `compile_commands.json`, then language evidence propagated from repository includers, then conservative syntax markers. A header used by C remains attributed to C even when the C++ grammar provides better error recovery; file attributes record both the source language and parser language.

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
- `possible-calls` for multiple defensible overload, conditional macro, and indirect function-pointer targets;
- macro-invocation `references` plus bounded function-like macro expansion, nested expansion chains, argument substitution, and every provable callable in a macro body;
- call-resolution provenance including certainty, candidate count, direct/type/arity evidence, macro definition spans, expansion chains, substitutions, and expansion depth;
- direct `passes-to` value flow from caller parameters, locals, fields, and constants to callee parameters, including safely substituted macro arguments;
- `dynamic-target` evidence for function-pointer calls even when bounded possible targets are also known;
- conservative reads and writes for parameters, locals, and fields;
- explicit unresolved call, include, inheritance, macro-expansion, and parse evidence.

Repository-local declarations are resolved across C and C++ files. Function definitions are preferred over matching prototypes. Translation-unit ownership follows repository-local include closure, so source-file `static` declarations are visible to included `.c` fragments and headers without becoming globally visible. C++ call resolution conservatively narrows candidates by accepted argument count, explicit namespace or type qualification, proven enclosing-type ownership, and direct parameter/local/field receiver types. Every call edge records the evidence used and whether the result is definite or possible. Unknown callable shapes, unsupported receiver forms, inherited methods, and unresolved template semantics retain the previous candidate set rather than being guessed. C function-pointer members remain fields rather than becoming fabricated methods.

Function-pointer evidence is propagated through direct initializers, assignments, designated struct initializers, typedef aliases, callback arguments, and safely expanded macro bodies. These flows produce bounded `possible-calls`; they do not convert indirect dispatch into definite calls, and the accompanying `dynamic-target` evidence remains explicit.

Function-like macros are interpreted without invoking a compiler preprocessor. The adapter records macro parameters and callable expressions in each replacement body, binds invocation arguments, substitutes identifier tokens, follows nested repository-visible macros up to a fixed depth, and emits every provable body call at the invocation source span. Each emitted call retains the macro chain, definition span, original body callee, substituted arguments, and resolution evidence. Conditional or multiply visible macros remain possible. Cycles, arity mismatches, token pasting, stringification, and variadic substitution remain explicit unresolved evidence instead of being guessed.

## Identities

All nodes use the facts-v1 SHA-256 identity contract with language `c-family`. Canonical identities include repository-relative source ownership, semantic kind, qualified name, and callable signature where required. Absolute checkout paths are never included.

The stream header language is always `c-family`. File and declaration attributes include `language: c` or `language: cpp` so consumers can distinguish the parsed grammar without splitting the shared semantic graph.

## Includes and build context

Quoted includes resolve first relative to the including file, then by exact repository path, then by unique basename. System includes and missing local headers remain unresolved rather than becoming fabricated repository dependencies.

`compile_commands.json` currently selects the C or C++ grammar for listed files. Compiler defines, include search paths, target triples, generated headers, and conditional preprocessing are not yet replayed.

## Conservative boundaries

The adapter does not run a compiler or preprocessor. Consequently:

- inactive conditional branches may still be parsed;
- bounded repository-local macro calls and argument substitution are reconstructed, but token pasting, stringification, variadic substitution, full conditional-branch evaluation, compiler defines, and include-path configuration are not replayed;
- generated declarations and headers are unavailable unless present in the repository;
- template instantiation, overload ranking, ADL, implicit conversions, virtual dispatch, and non-local function-pointer flow remain conservative;
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

The suite covers mixed C/C++ extraction, includer-driven header language inference, parser fallback, include-closure translation units, function-pointer typedefs and flow, macro aliases, multiple body calls, nested expansion, argument substitution, unsupported expansion evidence, callable arity, explicit qualification, direct receiver types, repository-local includes, inheritance, calls, argument flow, reads/writes, deterministic repeated output, incremental ownership, and CLI output. `tools/call_resolution_metrics.py` reports call-site outcomes and possible-target fanout without confusing raw edge counts with call-site coverage. `tools/semantic_depth.py` reports relation depth, provenance labels, macro expansion depth, and direct argument flow.

## Calibration corpus

The pinned C calibration corpus includes Git at `9a0c4701dcd5725c4184599322b52933ff5005ca` and the Codebase Memory C backend at `97ce23f9827177fff3858831156e9795c6832b18`. Judged gates cover included-C translation units, header-to-source static calls, macro-mediated body calls, macro aliases, function-pointer typedefs and dispatch tables, definition selection, and include resolution.

On the CBM backend, adapter 0.5.0 emits 7,763 macro-body call occurrences across 43 unique source-target pairs, including 7,475 substitution-backed occurrences across 41 pairs. It retains the invocation span, macro definition span, expansion chain, body callee, substitutions, and ordinary call-resolution evidence for each edge. The same scan emits 42,472 direct `passes-to` occurrences across 8,554 unique caller-value-to-callee-parameter pairs. No CBM call remains classified as a missing repository target.

The pinned C++ corpus adds LevelDB `99b3c03b3284f5886f9ef9a4ef703d57373e61be` and fmt `407c905e45ad75fc29bf0f9bb7c5c2fd3475976f` as calibration cases, Catch2 `src/` at `191fa38c9b1596cd2576ab531d4ab4d5e8e05190` as validation, and nlohmann/json `include/nlohmann/` at `55f93686c01528224f448c19128836e7df245f72` as holdout. Exact judgments protect class and template nodes, cross-file calls, and inheritance edges without scanning Catch2's amalgamated duplicate or nlohmann/json's tests and vendored material.

On the accepted Git scan, 88,511 direct and expanded call-expression sites were observed. Of the 66,314 sites with repository-callable evidence, 64,456 had definite targets, 1,858 had possible targets only, and 3 retained both definite and possible evidence: 97.2% definite and 100% definite-or-possible. No call remained classified as a missing repository target. The remaining unresolved sites were external APIs, genuine dynamic dispatch, explicit ambiguity, or explicitly unsupported macro forms.

The C++ calibration raised definite call sites from 6,088 to 7,079 in fmt, 4,623 to 4,855 in LevelDB, and 1,890 to 2,184 in Catch2. The untouched nlohmann/json holdout improved from 1,737 to 1,857 definite sites. fmt's 90th-percentile possible-target fanout remains 13 after adding 1,106 macro-mediated call occurrences and 2,801 direct argument-flow occurrences. All four C++ cases require zero unresolved call `missing-target` records; remaining uncertainty is explicit overload, template, dynamic-dispatch, unsupported macro, or external-library evidence.
