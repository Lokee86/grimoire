# Java adapter

The Java adapter combines Lexicon's deterministic source/dependency model with a compiler-backed semantic pass built on the public `jdk.compiler` API. It parses repository source without running project code or annotation processors, then uses javac attribution to resolve repository-local calls, constructors, inheritance, overrides, type references, and field/parameter access. Release packages include a private minimized Java runtime and compiler helper.

## Build

Go 1.26 or newer and JDK 21 or newer are required for source-tree development. Install the repository-local verified Temurin JDK when needed:

```text
python tools/bootstrap_jdk.py
go build -o lexicon-java .
```

The adapter discovers `LEXICON_JDK_HOME`, `JAVA_HOME`, the repository-local `.tools/jdk`, or `java`/`javac` on `PATH`. Packaged releases use their bundled private runtime and do not require a system JDK. The analyzed project is not built and its dependencies are not downloaded.

## Usage

```text
lexicon-java --repo <repository-root> --output <facts.jsonl>
```

Use `--output -` (the default) for stdout. The adapter currently emits complete `mode: full` streams only. Output is UTF-8 JSONL with one header followed by canonically sorted nodes, edges, and unresolved records. JSON object keys are lexicographically ordered.

For source-tree execution:

```text
go -C adapters/java run . --repo /path/to/repository --output /tmp/java-facts.jsonl
```

## Modeled semantics

The foundation models:

- repository, directory, Java file, compilation-unit module, and Java package namespace nodes;
- source package membership and import declarations;
- repository-local exact type imports, package wildcards, static member imports, and static type wildcards when there is one safe target;
- class, interface, enum, record, and annotation declarations;
- nested member types;
- ordinary and compact constructors, methods, fields, record components, and callable parameters;
- file/module, package/module, declaration/member, and callable/parameter containment;
- declared `extends` and `implements` relationships from type headers;
- sealed-type `permits` references marked with `role: permitted-subtype`;
- annotation applications on modeled declarations;
- retained callable-body token ranges for modeled methods and constructors with bodies;
- javac-attributed repository-local method, inherited method, static method, constructor, explicit `this`/`super`, and overload-resolved `calls` edges;
- conservative `possible-calls` fallback evidence when source errors or missing classpaths prevent javac from selecting one repository-local target;
- compiler-verified `overrides` edges across modeled repository-local ancestry;
- compiler-attributed field and parameter `reads`/`writes`, including assignment, compound assignment, and increment/decrement access;
- repository-local type `references` from callable returns, parameters, fields, locals, generic arguments, arrays, and member references;
- literal direct Maven dependency coordinates from `pom.xml`, including version, scope, and optional metadata;
- literal supported Gradle Groovy/Kotlin dependency declarations, with configuration and source-span evidence;
- deterministic external dependency modules and `depends-on` edges, while computed, catalog-backed, project, platform, profile, dependency-management, plugin, and other unsupported forms remain unresolved;
- explicit unresolved relationship, `imports`, `defines`, `calls`, `reads`, `writes`, or `depends-on` evidence rather than guessed targets.

Classes, enums, and records use the common `type` node kind. Interfaces and annotation declarations use `interface`. Their exact Java surface is retained in `attributes.declaration_kind`. Record components use `field` with `declaration_kind: record-component`. Parameters preserve source order through their index and parent callable identity.

Import nodes are defined by their compilation-unit module. An import node emits `imports` to one repository-local declaration when that declaration is unique. Imports that require the JDK, third-party classpaths, generated code, or unavailable source remain `external-target`; duplicate safe candidates remain `ambiguous-target`. Package declarations are represented by package namespace nodes and package-to-module containment.

Declared type-header references resolve only to repository-local type declarations. The resolver considers an exact qualified name, the compilation unit's package, explicit non-static imports, non-static wildcard imports, and enclosing lexical type owners. It emits an edge only when the resulting target is unique; absent and ambiguous targets remain `external-target` and `ambiguous-target` unresolved records. `extends` clauses emit `extends`, `implements` clauses emit `implements`, and `permits` clauses emit `references` with `role: permitted-subtype` from the sealed declaration to the permitted declaration.

Annotations on modeled types, constructors, methods, fields, record components, and parameters use the same conservative lookup. A uniquely resolved repository-local annotation declaration is the target of an `annotates` edge from the annotated declaration. External annotations and ambiguous local annotation names remain unresolved `annotates` evidence, preserving the source expression and span.

Callable bodies are first scanned by the deterministic fallback resolver, then reconciled against javac attribution. Exact compiler facts replace competing heuristic `calls`, `possible-calls`, `reads`, `writes`, and unresolved evidence at the same source site. Compiler errors are indexed by source line; facts touching an erroneous region are not trusted, leaving conservative fallback candidates or unresolved evidence intact rather than accepting javac recovery guesses.

Only repository-local modeled declarations become graph endpoints. External JDK and dependency symbols still inform attribution but are not invented as Lexicon nodes. Implicit constructors, local variables, lambdas, anonymous classes, and local classes remain unmodeled nodes; their bodies are not incorrectly attributed to the enclosing callable. Local-variable access suppresses false field/parameter evidence without creating local-variable nodes.

## Canonical identities and paths

All IDs are lowercase SHA-256 values over the facts-v1 payload:

```text
lexicon:v1\0java\0<kind>\0<canonical-identity>
```

Canonical identities are:

- repository: repository directory name;
- directory, file, and compilation-unit module: normalized repository-relative path;
- package namespace: dotted package name, or `<default>`;
- type/interface: dotted package plus lexical containing types and simple name;
- method: containing type plus method name and normalized ordered parameter-type signature;
- constructor: containing type plus `<init>` and normalized ordered parameter-type signature;
- field and record component: containing type plus field name;
- parameter: callable identity plus source-order index and parameter name;
- import: repository-relative file, source line/column, and imported target.

Content IDs hash the unmodified file bytes. Absolute checkout paths never enter identities or emitted paths. Paths and owners are normalized repository-relative forward-slash paths; `.` is used only for synthetic repository/default-package placement.

## Discovery and permanent exclusions

Discovery walks deterministically without following directory symlinks and reads `.java`, `pom.xml`, `build.gradle`, and `build.gradle.kts` files. Directory, source, and manifest lists are sorted before analysis. Matching is case-insensitive for the permanent excluded directory names.

The adapter excludes Git/worktree metadata (`.git`, `.worktrees`, `.workingtrees`), Lexicon and Warlock state (`.ddocs`, `.lexicon`, `.arcana`, `.grimoire`, `.pitlord`, `.cantrip`, `.homunculus`, `.incubus`, `.ritual`, `.warlock`), dependency/cache trees (`.bundle`, `.cache`, `.gradle`, `.idea`, `.mvn`, `.next`, `node_modules`, `packages`, `vendor`), and build/output trees (`bin`, `build`, `coverage`, `dist`, `generated`, `generated-sources`, `obj`, `out`, `target`, `temp`, `tmp`).

## Current boundaries

The compiler pass groups source files by Java source root and provides all eligible roots as a deterministic source path. Annotation processing is disabled. Embedded fixture projects beneath `src/test/resources` and `src/test/projects` remain available to the deterministic parser but are not sent through javac as part of the host repository. It does not evaluate Maven or Gradle builds, download dependencies, execute plugins, load generated sources, or execute analyzed code. Missing third-party classpaths can therefore leave external and some downstream expressions unresolved, but repository-local symbols that javac can attribute still produce exact edges. A compiler failure is isolated by recursively splitting the affected source-root batch; only irreducible failing files fall back completely to the deterministic parser.

The adapter does not create nodes for implicit/default constructors, initializer bodies, module descriptors, local/anonymous classes, lambdas, or local variables. It does not model reflection, runtime dependency injection, framework-generated dispatch, alias/control-flow dataflow, array-element ownership, effective Maven models, Gradle build models, profiles, BOMs, version catalogs, transitive dependencies, or generated code. Annotation values and repeatable/inherited semantics are not interpreted. Incremental compiler attribution is not yet implemented; each changed Java analysis currently reprocesses the discovered source-root batches.

Malformed literals/comments, unclosed type/member bodies, malformed package/import syntax, and member forms that cannot be classified safely remain explicit `unsupported-form` unresolved records. Valid declarations already isolated before a later malformed region may still be emitted; the unresolved record preserves the unsupported boundary. Invalid UTF-8 or NUL-containing files retain file/module evidence and emit an unresolved `defines` record.

## Code map

| Concern | Primary implementation | Verification |
| --- | --- | --- |
| Discovery, parsing, declarations, and dependencies | `discovery.go`, `parser*.go`, `dependencies.go`, `maven_dependencies.go`, `gradle_dependencies.go` | package-local fixture tests |
| Conservative fallback semantics | `relationships.go`, `runtime_*.go` | relationship and runtime tests |
| Compiler-backed attribution | `compiler_runner.go`, `compiler_facts.go`, `compiler/src/` | compiler semantic and evidence-replacement tests |
| Runtime discovery and release packaging | `compiler_environment.go`, `compiler_runtime.go`, `tools/java_release.py` | adapter tests and packaging smoke checks |

## Test

```text
go -C adapters/java test ./...
```

The permanent fixtures cover packages, local and external imports, static imports, every modeled type form, ordinary and compact constructors, methods, multi-field declarations, record components, parameters, nested types, malformed syntax, permanent exclusions, declared inheritance and implementation, sealed permits, local and external annotations, exact/package/explicit/wildcard/lexical resolution, definite and possible calls, typed parameter and declaration-before-use/scoped local receivers, literal and explicitly typed argument narrowing, null/reference ambiguity, multi-argument tradeoffs, unknown argument expressions, external and ambiguous receiver types, unsupported and dynamic receiver evidence, constructor calls, direct overrides, field/parameter reads and writes, local shadowing, retained body ranges, literal Maven and Gradle dependencies, duplicate dependency evidence, unsupported computed dependency forms, manifest exclusions, ambiguous targets, valid endpoints, canonical identities and ordering, byte determinism, checkout-path independence, stdout, and file output. A generated stream can also be checked with:

```text
python tools/validate_jsonl.py /path/to/facts.jsonl
python tools/semantic_report.py /path/to/facts.jsonl
```
