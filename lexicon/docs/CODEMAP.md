# Lexicon codemap

This codemap points from runtime responsibilities to the source files that currently implement them. It is a navigation aid, not a replacement for the normative contracts under `spec/` or the operational descriptions in `docs/APPLICATION.md` and `docs/ARCHITECTURE.md`.

## Status convention

Everything in the component and flow sections below is implemented in the current source tree.

No roadmap document or queued implementation plan is present in the current Lexicon tree. The final section separately lists documented future possibilities and current non-claims; those items are not implemented behavior or compatibility commitments.

## Top-level ownership

| Path | Owns |
| --- | --- |
| `cmd/lexicon/` | Thin executable entry point for the application CLI |
| `internal/cli/` | Subcommand parsing, repository selection, operation wiring, and user-facing output |
| `internal/scan/` | Initialization, complete/scoped scan planning, adapter scheduling, interstack refresh, publication transactions, and consumer notification |
| `internal/adapters/` | Application-side adapter registry, fingerprinting, runtime selection, request construction, and process execution |
| `adapters/<language>/` | Independently executable language discovery, parsing, semantic analysis, and facts-v1 emission |
| `internal/objectstore/` | Adapter-stream parsing, ownership partitioning, binary objects, manifests, publication, recovery records, incremental dependency data, export, and garbage collection |
| `internal/state/` | Private source mirror and Git-backed change detection |
| `internal/files/` | Relevant-file classification and permanent plus `.lexiconignore` exclusions |
| `internal/scope/` | Temporary repositories for scoped adapter execution |
| `internal/interstack/` | Repository-wide cross-stack contract derivation after ordinary language analysis |
| `internal/consumer/` | Post-publication consumer definitions, invocation, and success state |
| `internal/config/` | `.lexicon/config.json`, state-root selection, adapter-root discovery, enabled languages, and analysis configuration identity |
| `internal/languages/` | Supported language/config-file registry and generic-extension mapping |
| `internal/watch/` | Filesystem events, debounce, ignore reload, and periodic reconciliation |
| `internal/lock/` | Single-writer repository lock |
| `spec/` | Normative facts, object, snapshot, and runtime-evidence contracts |
| `evaluation/` and `tools/` | Test-matrix orchestration, corpus validation, smoke tests, validators, reports, and release tooling |

## CLI entry points

The process boundary is deliberately thin:

- `cmd/lexicon/main.go` passes arguments and standard streams to `internal/cli.Run` and returns its exit code.
- `internal/cli/cli.go` dispatches `init`, `scan`, `demon`, `rebuild`, `export`, `gc`, `languages`, `consumer`, `status`, `doctor`, and `version`. It directly wires initialization, scans, and watch mode.
- `internal/cli/repository.go` resolves explicit repositories and upward discovery through `.lexicon/config.json`.
- `internal/cli/storage.go` owns `rebuild`, `export`, and `gc` flags and calls into `scan`, `objectstore`, and `lock`.
- `internal/cli/languages.go` owns language-list parsing, configuration updates, and the immediate scan after `languages set`.
- `internal/cli/consumers.go` owns `consumer list|add|remove|run` and exact-snapshot selection for one-shot execution.
- `internal/cli/status.go` and `internal/cli/doctor.go` format repository diagnostics; `internal/cli/doctor_checks.go` contains the checks.
- `internal/cli/version.go` owns version output.

Begin in `internal/cli/cli.go` for a new top-level command, then move the operation into its owning package rather than implementing policy in the dispatcher.

## Scan orchestration

### Construction and lifecycle

- `internal/scan/scanner_open.go` implements `Initialize`, `InitializeWithLanguages`, `Open`, and `New`. Initialization saves configuration, creates the private state repository, mirrors source, runs full analysis, derives interstack facts, and publishes the first snapshot.
- `internal/scan/scanner.go` defines `Scanner` and the normal `Scan`/`ScanPaths` transaction. It acquires the update lock, recovers `PENDING`, mirrors source, stages and reads changes, plans work, runs adapters, refreshes interstack, commits the manifest, then notifies consumers.
- `internal/scan/rebuild.go` implements forced complete analysis for all enabled or selected languages and uses the same publication and consumer path.
- `internal/scan/transaction.go` owns pending-publication recovery, current-manifest loading and legacy-library migration, private-state commit advancement, manifest publication, and `PENDING` cleanup.
- `internal/scan/snapshot.go` verifies no-op scans against the private state commit and detects adapter-fingerprint drift.

### Planning and execution

- `internal/scan/plan.go` maps source changes and drift to per-language complete or scoped plans. Structural changes force complete analysis; eligible modifications ask the object store for an incremental scope.
- `internal/objectstore/dependencies.go` computes reverse impacted-owner closure and forward context closure from the current immutable objects. It also decides when direct edits already require full analysis.
- `internal/scope/repository.go` materializes scoped repositories from selected context, adds language configuration files, expands Go files to complete packages, and expands Rust files to complete crates.
- `internal/scan/analysis.go` constructs adapter requests, runs language plans concurrently, validates each stream through `objectstore.ReadAnalysis`, applies full or incremental results, and retries scoped work as full analysis when required.
- `internal/objectstore/topology.go` compares scoped edge/unresolved topology with previous file objects and requests full analysis when a scoped result introduces unsafe relationships.
- `internal/scan/resource_scheduler.go` enforces one weighted process-wide CPU budget.
- `internal/scan/parallel_plan.go` inventories work and computes Go logical shards, active workers, merge fan-in, and reserved weight; `LEXICON_MAX_WORKERS` can lower the worker ceiling.
- `internal/scan/languages.go` detects languages in the mirror, applies enabled-language selection, and identifies snapshot language drift.
- `internal/scan/interstack.go` rebuilds or removes the synthetic `interstack` language after ordinary language entries form the candidate manifest.

### Source state and watch mode

- `internal/files/files.go` maps paths to languages and owns permanent ignored-directory names.
- `internal/files/ignore.go` loads and applies repository `.lexiconignore` rules without allowing permanent exclusions to be re-included.
- `internal/state/mirror.go` copies only relevant, non-symlink source files into `.lexicon/repo/source`, for a full tree or selected paths.
- `internal/state/git.go` initializes/opens the private Git repository, stages the mirror, parses source changes, and maintains one amended root state commit.
- `internal/watch/daemon.go` recursively watches non-ignored directories, debounces paths into `Scanner.ScanPaths`, reloads ignore policy on `.lexiconignore` changes, and runs full reconciliation after watcher errors or on the configured interval.
- `internal/lock/lock.go` is the shared writer guard used by scans, rebuilds, initialization, and garbage collection.

## Adapter boundary and language owners

### Application-side adapter execution

- `internal/languages/registry.go` is the source of truth for supported language names, source extensions, configuration files, and extension-qualified `generic-*` languages.
- `internal/adapters/registry.go` exposes that registry and hashes adapter source/runtime contents plus schema/config versions into an adapter fingerprint.
- `internal/adapters/runner.go` translates an `adapters.Request` into the appropriate Go, Python, Ruby, Cargo, Node, or generic process invocation. It passes changed/removed scopes and Go parallelism flags and rejects empty output.
- `internal/adapters/runner_packaged.go` locates packaged native adapter executables before source-tree fallbacks are used.
- `spec/facts-v1.md` is the normative adapter stream contract. `internal/objectstore/analysis.go`, `ingest_parse.go`, and `records.go` are the application-side parser, header/record validator, ownership reader, and typed-record boundary.

### Language adapter source

Each adapter owns its parser and semantic policy inside its directory; it does not publish snapshots.

| Adapter | Entry point and primary implementation seams |
| --- | --- |
| C / C++ | `adapters/c-family/main.go`; discovery and Tree-sitter parsing in `discovery.go`, `parser.go`, and `syntax.go`; declarations/includes/macros in `declarations*.go`, `include_resolution.go`, and `macro*.go`; calls, pointers, arguments, and relationships in `call*.go`, `pointer*.go`, `argument_resolution.go`, `resolution.go`, and `relationship_resolution.go`; facts in `facts.go` and `emission.go` |
| Go | `adapters/go/main.go`; repository/module loading in `adapter.go` and `modules.go`; AST facts in `ast_*.go`; typed semantics in `semantic*.go`; deterministic partitioning in `parallel.go` and `semantic_parallel.go`; dependencies and output in `dependencies.go`, `facts.go`, and `facts_json.go` |
| GDScript | `adapters/gdscript/main.go`; repository/project loading in `repository.go` and `project_config.go`; syntax/declarations in `lexer.go`, `declaration*.go`, and `expressions.go`; inheritance, bindings, calls, and dataflow in `extends.go`, `semantic_*.go`, `resolution_bindings.go`, `call*.go`, `relationships.go`, and `dataflow.go`; output in `jsonl.go` |
| Generic fallback | `adapters/generic/main.go`; file selection in `files.go`; conservative parsing/masking in `parse.go` and `mask.go`; analysis and output in `analyze.go` and `facts.go` |
| LotusScript | `adapters/lotusscript/main.go`; discovery and ODP/DXL extraction in `discovery.go` and `dxl.go`; parsing/declarations/types/calls in `parser.go`, `syntax.go`, `declaration_helpers.go`, `types.go`, and `calls.go`; repository analysis and facts in `analysis.go` and `facts.go` |
| Python | `adapters/python/lexicon_python/__main__.py`; orchestration in `adapter.py`; discovery in `discovery.py`; AST extraction in `extraction*.py`; bindings, calls, resolution, and dependencies in `bindings.py`, `callgraph*.py`, `resolution.py`, and `dependencies.py`; contract/output in `model.py`, `contract.py`, and `emission.py` |
| Ruby | `adapters/ruby/lexicon_ruby.rb`; flags in `cli.rb`; repository loading in `repository.rb`; Ripper extraction in `ripper_*.rb`; relationships, call resolution/flow, and dependencies in `relationships.rb`, `call_*.rb`, and `dependencies.rb`; model/contract/output in `model.rb`, `semantic_model.rb`, `contract.rb`, and `emitter.rb` |
| Rust | `adapters/rust/src/main.rs` delegates to `cli.rs`; `orchestrator.rs` coordinates discovery, parsing, extraction, semantic resolution, relationships, dataflow, and dependencies across the other `src/*.rs` modules; `contract.rs` and `emit.rs` own the stream boundary |
| JavaScript / TypeScript / Svelte | `adapters/typescript/src/cli.ts`; `orchestration.ts` coordinates `discovery.ts`, compiler/Svelte AST handling in `ast*.ts` and `svelte.ts`, call resolution/flow in `call-*.ts` and `resolution.ts`, `dataflow.ts`, and `dependencies.ts`; `model.ts`, `contract.ts`, and `emission.ts` own output. `dist/` is built output, not the source owner |

The owning adapter README records each language's implemented semantic coverage and conservative limits.

## Object storage, snapshots, and state

### Data model and object creation

- `internal/objectstore/model.go` defines fact-object, file-entry, language-entry, snapshot-manifest, and adapter-header models.
- `internal/objectstore/analysis.go`, `ingest_parse.go`, and `records.go` parse one facts-v1 stream into typed records and determine source ownership.
- `internal/objectstore/update.go` builds complete, incremental, or synthetic shared language entries. It writes one object per owned source file, preserves unaffected file entries during scoped updates, and stores unowned synthetic records in a shared object.
- `internal/objectstore/store.go` computes content identities, writes/loads immutable objects, writes/loads manifests, verifies hashes, and atomically updates `CURRENT`.
- `internal/objectstore/store_write.go` implements durable temporary writes and immutable/atomic replacement; `replace_{unix,windows}.go` and `sync_{unix,windows}.go` own platform-specific replacement and directory-sync behavior.

### Binary format, manifests, and recovery

- `internal/objectstore/binary_format.go`, `binary_encode.go`, and `binary_decode.go` implement deterministic objects-v1 bytes, shared string tables, independent node/edge/unresolved sections, bounds checks, and materialization. `store.go` retains compatible reads for legacy JSON objects.
- `internal/objectstore/manifest_update.go` provides deterministic language replacement/removal in candidate manifests.
- `internal/objectstore/manifest.go` exists for migration from the former materialized `.lexicon/repo/library/*.jsonl` layout.
- `internal/objectstore/pending.go` persists and loads the durable `PENDING` candidate used by `internal/scan/transaction.go` for crash recovery.
- `spec/objects-v1.md` and `spec/snapshots-v1.md` are authoritative for binary bytes, identities, manifest meaning, publication order, and consumer consistency.

### Export and garbage collection

- `internal/objectstore/export.go` resolves a snapshot and atomically writes selected JSONL libraries.
- `export_languages.go`, `export_objects.go`, and `export_records.go` validate requested languages, verify object metadata, merge shared and file records, sort them, and reconstruct a full facts-v1 stream.
- `gc.go` plans retention from `CURRENT`, newest snapshots, and consumer pins.
- `gc_storage.go` enumerates manifests/objects and reads `.lexicon/consumer-state/*.json` pins.
- `gc_validate.go` validates plans and prevents preserved/deleted overlap; `gc_execute.go` rechecks `CURRENT` before dry-run reporting or deletion.

### On-disk state

`internal/config.StateRoot` selects `.lexicon/` by default or `LEXICON_STATE_DIR` when configured. Current owners of the layout are:

- `config.json`: `internal/config/config.go`;
- `LOCK`: `internal/lock/lock.go`;
- `PENDING`: `internal/objectstore/pending.go` plus `internal/scan/transaction.go`;
- `CURRENT`, `snapshots/`, and `objects/`: `internal/objectstore/`;
- `repo/.git/` and `repo/source/`: `internal/state/`;
- `consumers/` and `consumer-state/`: `internal/consumer/`;
- transient adapter/interstack output and scoped repositories under `tmp/`: `internal/scan/analysis.go`, `internal/scan/interstack.go`, and `internal/scope/repository.go`.

## Consumers

- `internal/consumer/registry.go` lists, validates, atomically adds, removes, and runs named `.json` definitions.
- `internal/consumer/definition.go` encodes/decodes the versioned definition and duration-compatible timeout.
- `internal/consumer/runner.go` invokes definitions in lexical order without a shell, supplies `LEXICON_REPOSITORY`, `LEXICON_STATE_ROOT`, and `LEXICON_SNAPSHOT_ID`, applies optional timeouts, and aggregates failures.
- `internal/consumer/state.go` atomically records the last successfully processed snapshot. Those state files also pin snapshots during garbage collection.
- `internal/cli/consumers.go` is the user-facing management surface; `internal/scan/scanner.go` invokes all consumers only after a scan has published or confirmed a snapshot.

A consumer failure does not roll back the already-published snapshot; it prevents that consumer's success state from advancing.

## Interstack resolution

`internal/scan/interstack.go` calls `internal/interstack.Build` after ordinary language entries are ready. The implemented post-pass is organized as follows:

- `build.go` exports each ordinary immutable language entry, parses it, runs resolution over the mirrored source, writes a full synthetic stream, and re-enters through `objectstore.ReadAnalysis`.
- `load.go` reads language-owned nodes from exported facts; `model.go` builds indexes by path, name, qualified name, and enclosing callable.
- `resolve.go` collects eligible Go, Ruby, GDScript, TypeScript/JavaScript, Python, and Rust source and sequences all detectors. Test/fixture directories, generated/state directories, binary files, and files over the implemented size limit are skipped.
- `http_consumers.go` detects Rails and Go route registrations and maps endpoints to handlers; `http.go` detects HTTP producers/helpers, normalizes route shapes, resolves unique targets, and preserves missing/ambiguous outcomes.
- `messages.go` derives packet/message channel nodes and publish/consume edges only from dispatch context plus message-like values.
- `config.go` detects supported environment-read forms; `boundary_config.go` handles the explicit Lexicon/Grimoire boundary keys.
- `processes.go` maps recognized Lexicon/Arcana process and CLI-command invocations and command ownership.
- `boundary_nodes.go` owns process, CLI-command, and `arcana.query.v1` protocol nodes and producer/consumer links.
- `state_contracts.go` maps implemented Lexicon/Arcana state handoffs to synthetic state-path nodes.
- `paths.go` owns synthetic `@interstack/...` paths and identifier helpers; `emit.go` owns stable IDs, deterministic ordering, and facts-v1 encoding.

The resulting `interstack` language is stored as one shared object by `objectstore.BuildSharedLanguage`; it does not claim ownership of ordinary source files or replace language-owned symbols.

## Main runtime and data flows

### `lexicon init`

1. `cmd/lexicon/main.go` -> `internal/cli/cli.go:runInit`.
2. `internal/cli/repository.go` selects the repository; `internal/config/config.go` locates adapters and writes configuration.
3. `internal/scan/scanner_open.go` creates `.lexicon/repo`, mirrors relevant source through `internal/state/mirror.go`, and detects enabled languages.
4. `internal/scan/analysis.go` runs complete adapter plans through `internal/adapters/runner.go`.
5. `internal/objectstore` validates each JSONL stream, partitions records, and writes immutable objects.
6. `internal/scan/interstack.go` derives the synthetic repository-wide library.
7. `internal/scan/transaction.go` writes `PENDING`, advances private source state, publishes the manifest and `CURRENT`, then clears `PENDING`.

### `lexicon scan` or watched path update

1. `internal/scan/scanner.go` acquires the lock, recovers pending work, loads `CURRENT`, and synchronizes the mirror.
2. `internal/state/git.go` reports staged source changes.
3. `internal/scan/plan.go` combines changes, missing/disabled languages, and adapter drift into deterministic language plans.
4. Eligible modifications use dependency scope from `internal/objectstore/dependencies.go` and a temporary repository from `internal/scope/repository.go`; structural or unsafe changes use complete analysis.
5. `internal/scan/analysis.go` schedules adapter processes, validates output, retries unsafe scoped analysis as full, and reuses unchanged file objects.
6. Interstack is rebuilt from the complete candidate language manifest.
7. The publication transaction advances atomically; `internal/consumer/runner.go` then invokes registered consumers.

`lexicon demon` enters this same flow through `internal/watch/daemon.go`; it does not use a separate storage or publication path.

### Snapshot read/export/collection

- Consumers resolve `CURRENT` once, load the referenced immutable manifest, then read only referenced objects.
- `lexicon export` follows `internal/cli/storage.go` -> `objectstore.Export`, verifies immutable inputs, and reconstructs deterministic full JSONL.
- `lexicon gc` acquires the writer lock, plans reachability from retained manifests and consumer pins, rechecks `CURRENT`, then reports or deletes unreachable manifests and objects.

## Where to begin for common changes

| Change | Begin here | Also check |
| --- | --- | --- |
| Add/change a CLI subcommand | `internal/cli/cli.go` and the narrow `internal/cli/*.go` owner | `docs/APPLICATION.md`, `internal/cli/*_test.go`; route policy into an owning internal package |
| Change repository discovery or adapter-root lookup | `internal/cli/repository.go` or `internal/config/config.go` | `repository_test.go`, `config_test.go`, status/doctor behavior |
| Add a supported language or extension | `internal/languages/registry.go` | `internal/adapters/runner.go`, the adapter directory/README, registry/runner tests, `docs/STATUS.md` |
| Change language semantics | The owning `adapters/<language>/` parser/resolver/emitter | Adapter fixtures, facts-v1 contract, adapter README, focused corpus case |
| Change complete/scoped selection | `internal/scan/plan.go` and `internal/objectstore/dependencies.go` | `internal/scope/repository.go`, topology checks, incremental/plan tests |
| Change scan concurrency | `internal/scan/analysis.go`, `resource_scheduler.go`, or `parallel_plan.go` | parallel-plan/analysis tests, root race suite; Go adapter parallel files and race suite when applicable |
| Change object ownership or incremental replacement | `internal/objectstore/analysis.go`, `ingest_parse.go`, and `update.go` | dependency/topology tests, `spec/facts-v1.md`, `spec/objects-v1.md` |
| Change binary object bytes | `binary_format.go`, `binary_encode.go`, and `binary_decode.go` | binary codec/golden tests, `spec/objects-v1.md`, legacy-read compatibility |
| Change publication or recovery | `internal/scan/transaction.go`, `internal/objectstore/store.go`, and `pending.go` | transaction/store tests, platform atomic-write files, `spec/snapshots-v1.md` |
| Change export or GC | `internal/objectstore/export*.go` or `gc*.go` | `export_test.go`, `gc_test.go`, CLI operation tests, smoke tools |
| Change ignore or watch behavior | `internal/files/ignore.go`, `internal/state/mirror.go`, or `internal/watch/daemon.go` | ignore/mirror/watch tests and `docs/APPLICATION.md` |
| Change consumer behavior | `internal/consumer/` | `internal/cli/consumers.go`, consumer tests, GC pin handling |
| Add/change an interstack contract | The specific detector in `internal/interstack/` | `resolve_test.go`, `boundary_*_test.go`, `build_test.go`, `internal/scan/interstack_test.go`; keep ambiguity unresolved |
| Change a public record/format contract | The relevant file under `spec/` first | Every writer, reader, validator, exporter, golden fixture, and external consumer |

## Tests and verification owners

- Root application/package tests live beside implementation as `internal/**/*_test.go`. The densest lifecycle coverage is in `internal/scan/`, `internal/objectstore/`, `internal/interstack/`, `internal/consumer/`, `internal/cli/`, `internal/state/`, `internal/files/`, `internal/scope/`, `internal/watch/`, and `internal/adapters/`.
- Go-based adapters keep fixtures and `*_test.go` in `adapters/c-family/`, `adapters/go/`, `adapters/gdscript/`, `adapters/generic/`, and `adapters/lotusscript/`.
- Python adapter coverage is `adapters/python/tests/test_adapter.py`.
- Ruby adapter coverage is `adapters/ruby/test/test_adapter.rb` plus its fixtures.
- Rust adapter coverage is wired by `adapters/rust/src/main.rs` to `src/tests.rs`, with repository fixtures under `adapters/rust/tests/fixtures/`.
- TypeScript/JavaScript/Svelte coverage is `adapters/typescript/tests/*.test.js`, run through the package scripts in `adapters/typescript/package.json`.
- `evaluation/run_tests.py` is the canonical application-plus-all-adapters test matrix.
- `evaluation/run_validation.py`, `corpus.json`, and `validation/baseline.json` own repeat-run real-repository gates and dated evidence; generated validation outputs are not source owners.
- `tools/smoke_app.py`, `tools/smoke_operations.py`, and focused `tools/test_*.py` scripts cover packaged operations, publication/export/consumer/GC paths, and reporting utilities.
- Focused commands and full/race expectations are documented in `docs/DEVELOPMENT.md`.

## Implemented versus future or out-of-scope work

Implemented behavior is the code mapped above and the current capability list in `docs/STATUS.md`. In particular, adapters, immutable objects, atomic snapshots, scoped-analysis fallback, interstack derivation, deterministic consumers, export, GC, and optional watch mode all have current source owners.

Documented future possibilities, not implemented commitments:

- `docs/APPLICATION.md` notes that richer dependency metadata could narrow additions, deletions, renames, copies, configuration changes, and other cases that currently force complete-language analysis.
- `docs/GO_ADAPTER_VALIDATION.md` notes that exact retained call-site coverage would require a future call-site fact layer rather than reconstruction from deduplicated graph edges.
- `adapters/README.md` allows other adapters to add safe partitioning later, but no shared arbitrary-file sharding implementation exists.

Current non-claims are not planned behavior unless a separate proposal says otherwise:

- no perfect semantic precision or recall and no complete dynamic runtime dispatch recovery;
- no runtime instrumentation provider implementation;
- no graph query, reachability, impact-analysis, ranking, or packed-graph traversal API in Lexicon;
- no complete external-package implementation graph when source is not scanned;
- no unimplemented language/framework semantics listed in `docs/STATUS.md` or the owning adapter README.

Do not infer a feature from a contract placeholder, limitation, generated artifact, ignored validation output, or permissive comment. New planned work should be labeled in a proposal and must not be described in this codemap as current behavior until its source owner exists.
