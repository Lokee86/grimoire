# Lexicon current status

Status date: July 24, 2026.

This document describes the implementation currently present on `main`. Dated validation reports record evidence from specific runs and should not be treated as permanent performance guarantees.

## Application

Implemented:

- repository initialization and upward root discovery through `.lexicon/config.json`;
- configurable adapter discovery through `--adapters`, `LEXICON_ADAPTERS`, packaged adjacency, repository-local adapters, or the current working directory;
- detected and explicitly enabled language selection;
- complete scans and dependency-aware scoped scans;
- private source mirroring and Git-backed change detection;
- non-Git correctness through source content identities and snapshot comparison;
- immutable per-file and shared-language fact objects;
- deterministic binary object encoding with legacy JSON-object reads;
- atomic snapshot manifests and `CURRENT` publication;
- durable `PENDING` recovery and single-writer locking;
- deterministic JSONL export;
- retention-aware object garbage collection with consumer pins;
- status and doctor diagnostics;
- deterministic post-publication consumers;
- optional debounced watch mode with periodic full reconciliation;
- concurrent language analysis under a weighted CPU budget.

The primary execution model remains one-shot CLI operations. Watch mode invokes the same bounded scan transaction and is not required by snapshot consumers.

## Adapter status

| Adapter | Version | Implemented semantic scope | Principal limits |
| --- | ---: | --- | --- |
| C / C++ | 0.2.0 | Shared C-family view, includer-aware header attribution, parser fallback, translation-unit linkage, types, functions, includes, inheritance, calls, macro references, function-pointer evidence, dataflow | No compiler/preprocessor replay, template instantiation, overload ranking, ADL, virtual dispatch proof, generated headers, Objective-C, or CUDA semantics |
| Go | 0.1.0 | Multi-module discovery, packages, types, calls, closures, interfaces, implementations, overrides, dataflow, dependencies, SSA/VTA possible dispatch | Reflection, plugins, cgo/assembly, generated runtime behavior, and exact call-site graph retention |
| GDScript | 0.3.0 | Godot project scoping, classes, inheritance, autoloads, callbacks, bounded type flow, calls, possible calls, dataflow, dependencies | Scene-tree-only type evidence, engine internals, runtime script replacement, computed dispatch and resource paths |
| Python | 0.3.0 | Imports, lexical scopes, inheritance, protocols, callbacks, callable flow, C3 lookup, dataflow, dependencies | Monkey patching, metaclasses, dynamic imports/reflection, framework injection without ordinary value-flow evidence |
| Ruby | 0.3.0 | Reopened owners, inheritance, mixins, blocks, callbacks, bounded Rails-aware flow, dataflow, dependencies | Open runtime mutation, `send`/`eval`, refinements, dynamic constants, framework-generated behavior without declarations |
| Rust | 0.3.0 | Cargo workspaces, modules, traits, implementations, UFCS, callbacks, dataflow, dependencies | Procedural macro expansion, build-script-generated behavior, unsafe aliasing, unconstrained runtime registration |
| JavaScript / TypeScript / Svelte | 0.4.0 | Compiler-backed imports, inheritance, interfaces, calls, callbacks, CommonJS, JSDoc, Svelte script blocks, dataflow, dependencies | Svelte template semantics, Astro, prototype/runtime mutation, computed exports and properties, untyped external behavior |
| Generic fallback | 0.1.0 | Curated source extensions, file/module facts, high-confidence type and function declarations, static import evidence | No resolved calls, inheritance, dataflow, dispatch, or language-specific project semantics |

All adapters emit the same facts-v1 contract and preserve definite, possible, and unresolved relationships as distinct evidence.

## Incremental analysis

Implemented safe narrowing includes:

- reverse dependency closure from the previous snapshot;
- conservative inclusion of owners with unresolved relationships;
- forward dependency context in temporary scoped repositories;
- package expansion for Go and crate expansion for Rust;
- replacement of changed-file-owned objects only;
- unchanged object reuse;
- preservation of complete shared synthetic facts during partial scoped analysis;
- complete-language retry on unsafe topology, wrong stream mode, or scoped failure.

Current full-analysis triggers include additions, deletions, renames, copies, language configuration changes, adapter fingerprint changes, invalid prior state, and direct edits whose previous cross-file or unresolved relationships make partial replacement unsafe.

These fallbacks protect correctness. They are optimization limits, not contract failures.

## Parallel analysis

Implemented application-level behavior:

- independent language plans start concurrently;
- each plan reserves a weighted share of the process-wide CPU budget;
- results merge into the manifest in deterministic language-plan order.

Implemented Go-adapter behavior:

- repository-size-dependent logical shards;
- bounded active workers;
- weighted work partitioning by semantic file size;
- shard-local nodes, edges, callsites, semantic identities, and unresolved state;
- deterministic fan-in merge;
- final repository-wide SSA/VTA pass;
- `LEXICON_MAX_WORKERS` as a lower worker ceiling.

Logical shard count does not equal simultaneous worker count. Deterministic output is required across worker and merge configurations.

## Go multi-module repositories

Implemented:

- discovery of every `go.mod` beneath the repository root;
- nearest-module ownership for each source file;
- independent package loading and semantic analysis per module;
- deterministic merge into one repository fact stream;
- module-qualified package and symbol identities;
- repository-directory identity when no root-level module exists.

This removes the previous assumption that a scanned repository contains exactly one root `go.mod`.

## Validation evidence

Current acceptance mechanisms include:

- application and adapter fixture suites;
- root and Go-adapter race suites for concurrent code;
- facts-v1 validation;
- semantic relation reports;
- byte-for-byte repeat-run comparison;
- positive and expected-negative relation gates;
- real-repository corpus cases across C, GDScript, Python, Ruby, Rust, TypeScript, JavaScript, and Svelte;
- pinned Git and Codebase Memory C-backend judgments for the C/C++ shared adapter;
- fixture and application smoke coverage for the C/C++ shared adapter;
- separate dated Go real-repository validation.

See:

- [Semantic acceptance gates](SEMANTIC_ACCEPTANCE.md)
- [Cross-adapter corpus validation](SEMANTIC_CORPUS_VALIDATION.md)
- [Go adapter validation](GO_ADAPTER_VALIDATION.md)
- [Evaluation harness](../evaluation/README.md)

## Explicit non-claims

Lexicon does not currently claim:

- perfect semantic precision or recall;
- complete runtime dispatch recovery for dynamic languages;
- runtime instrumentation;
- graph-query, reachability, impact-analysis, or ranking APIs;
- a complete dependency implementation graph for external packages whose source is not scanned;
- Svelte template, Astro template, procedural-macro expansion, Godot scene-tree inference, or framework-generated code semantics;
- that dated corpus counts or elapsed times remain unchanged after implementation updates.

High unresolved counts are expected where the language or framework permits runtime behavior that static evidence cannot prove. An unresolved record is preserved information, not automatically a defect.

## Documentation status

The root README is the entry point. `docs/README.md`, `adapters/README.md`, `spec/README.md`, `evaluation/README.md`, and `tools/README.md` index their owning documentation surfaces. Current behavior, normative contracts, adapter-specific semantics, and dated validation evidence are intentionally separated.
