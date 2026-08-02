# Kotlin adapter

The Kotlin adapter is a self-contained Go executable that reads Kotlin source text and supported Gradle/Maven manifests and emits the Lexicon facts-v1 JSONL contract. It uses deterministic, non-executing parsers: it does not run application code, Gradle, Maven, Kotlin scripts, compiler plugins, annotation processors, dependency resolution, or repository build logic.

## Build

Go 1.22 or newer is required.

```text
cd adapters/kotlin
go build -o lexicon-kotlin-adapter .
```

No Kotlin, Java, Gradle, or external Go dependency is required.

## Usage

Write facts to a file:

```text
adapters/kotlin/lexicon-kotlin-adapter \
  --repo <repository-root> \
  --output <facts.jsonl>
```

Use `--output -` for stdout:

```text
cd adapters/kotlin
go run . --repo <repository-root> --output -
```

The adapter emits a complete `mode: "full"` repository view. Incremental owner-filtered emission is not currently implemented.

## Modeled semantics

The foundation emits:

- repository, source-ancestor directory, `.kt`/`.kts` file, source-module, and package namespace nodes;
- package declaration evidence from each source module to its namespace;
- import declaration nodes, including aliases and wildcard metadata, plus `imports` edges to deterministic syntactic target nodes;
- class, interface, object, companion object, enum class, data class, sealed class/interface, value class, and annotation-class declarations;
- declared class/interface/object supertypes, including constructor-form entries and `by` delegation entries;
- repository-local `extends` or `implements` edges selected from the resolved target declaration kind, with delegation expression metadata retained on delegated edges;
- repository-local annotation applications as `annotates` edges from the annotated declaration or parameter to the annotation-class declaration;
- explicit unresolved relationship records for external, ambiguous, or unsupported supertype and annotation targets rather than synthetic semantic targets;
- explicit primary constructors (including the implicit empty constructor) and secondary constructors;
- top-level functions and member methods;
- retained token ranges for modeled function, method, and secondary-constructor bodies, used only by the conservative runtime pass;
- top-level, member, extension, and primary-constructor properties, represented by the common facts-v1 `field` kind;
- callable and constructor parameters;
- `contains` and `defines` ownership edges for declarations and parameters;
- annotation spellings and sorted Kotlin modifiers when present;
- `suspend`, extension receiver, declared type, return type, default-parameter, mutability, and nullability metadata where syntactically available;
- literal supported Gradle Kotlin/Groovy dependencies, including `kapt` and `ksp`, plus direct literal Maven dependencies;
- deterministic external Maven-ecosystem module nodes and `depends-on` edges with manifest, configuration, coordinate, scope/optional, and source-span evidence;
- unresolved `depends-on` evidence for catalogs, interpolation, project dependencies, platform wrappers, properties, parent/profile/BOM/plugin-derived forms, and other declarations that cannot be proven without executing or resolving the build;
- explicit `unresolved` records with relation `defines`, reason `unsupported-form`, and parser diagnostics for malformed delimiters, malformed declarations, unsupported destructuring properties, unterminated literals/comments, and other declaration syntax the structural parser cannot interpret safely.

Enum entries, property accessors, local declarations, and initializer/body expressions are not emitted as declarations. Their text is skipped without execution.

## Conservative runtime semantics

The runtime pass consumes only tokens retained by the structural parser. It emits a definite `calls` edge only when one repository-local target remains after conservative name and accepted-arity matching. Covered forms are unqualified and `this` calls to methods on the containing declaration, unqualified same-package top-level functions, explicit repository-local object or companion calls (including a class name used to qualify its unique companion), direct constructor calls, explicit secondary-constructor `this(...)` and `super(...)` delegation, and repository-local extension calls through an explicitly typed identifier receiver. Extension matching requires a uniquely resolved repository-local receiver type from a callable parameter or scope-safe simple local/property declaration, plus a visible extension with the same resolved receiver type; aliases, explicit imports, wildcard imports, same-package declarations, and lexical member extensions follow the existing resolution order. Multiple sound overloads emit `possible-calls`; external names, inferred or dynamic receivers, safe calls, fluent chains, and recognized unsupported call forms remain unresolved `calls` evidence instead of being guessed.

Modeled methods emit `overrides` only to an exact declared-name, normalized extension-receiver, and normalized parameter-type signature on a uniquely resolved direct repository-local supertype. This is declaration matching, not compiler override validation.

The pass emits `reads` and `writes` only to modeled callable parameters and non-delegated properties when a bare identifier, `this` member, or explicit repository-local object/companion member proves ownership. Assignment writes; compound assignment and increment/decrement read and write. A local, destructured, lambda-bound, or delegated name suppresses property/parameter inference rather than creating or guessing a target.

## Canonical identities

Every node ID is the facts-v1 digest of:

```text
lexicon:v1\0kotlin\0<kind>\0<canonical identity>
```

Canonical identities never contain an absolute checkout path:

- repository: repository directory name;
- directory and file: normalized repository-relative path;
- source module: `source:<repository-relative path>`;
- package namespace: `package:<qualified package>`;
- import: source path, imported name, alias, and source-order occurrence;
- external syntactic import target: `external-import:<imported name>`;
- declaration: source-module identity followed by the declaration ownership chain and declaration kind/name;
- function/method: declaration owner, name, extension receiver, and declared parameter-type signature;
- constructor: containing declaration identity and declared parameter-type signature;
- property: declaration owner, name, and extension receiver;
- parameter: callable/constructor identity, zero-based position, and declared name.

Qualified names use the declared Kotlin package and nested declaration names. Callable qualified names include declared parameter types. Repository-relative source ownership and spans remain attached to every source-derived record. Relationship lookup recognizes exact qualified names, same-package declarations, explicit imports, aliases (including aliased nested ownership), wildcard imports, and uniquely resolved lexical nested ownership. Only declarations present in the analyzed repository are eligible semantic targets.

Duplicate malformed declarations that have the same canonical source identity receive a deterministic source-order suffix. Valid Kotlin overloads remain distinct through their declared receiver and parameter types.

## Discovery and permanent exclusions

Discovery recursively scans regular `.kt`, `.kts`, `build.gradle.kts`, `build.gradle`, and `pom.xml` files. Paths are normalized to repository-relative forward-slash form and sorted before parsing. Symlinks are not followed.

The following directory names are permanently excluded case-insensitively:

```text
.git .worktrees .workingtrees
.lexicon .arcana .grimoire .pitlord .cantrip .homunculus
.incubus .ritual .warlock .ddocs
.idea .vscode .vs
.gradle .kotlin
node_modules packages vendor .bundle
build dist out target artifacts generated bin obj coverage tmp temp
__pycache__ .pytest_cache .godot .import
```

These exclusions cover repository metadata, linked-worktree metadata, Warlock state, dependency/vendor trees, IDE state, language caches, and generated/build output. A source file under an excluded directory is never analyzed.

## Determinism

Files are discovered in canonical path order. Node identities use SHA-256, attributes use deterministic scalar values or sorted arrays, duplicate records are keyed canonically, and emission follows facts-v1 ordering: header, nodes, edges, then unresolved records. JSON object keys are lexicographically ordered by Go's JSON encoder. Identical source bytes and repository directory identity produce byte-identical output.

## Code map

| Concern | Primary implementation | Verification |
| --- | --- | --- |
| Discovery, parsing, and declarations | `discovery.go`, `parser*.go`, `analysis_declarations.go` | foundation adapter tests |
| Type, import, and relationship resolution | `analysis_resolution.go`, `analysis_relationships.go`, `analysis_imports.go` | relationship fixtures |
| Calls, extensions, overrides, reads, and writes | `analysis_runtime_*.go` | runtime and extension-call tests |
| Gradle and Maven dependency evidence | `dependency_analysis.go`, `gradle_dependencies.go`, `maven_dependencies.go` | dependency tests |

## Tests

```text
cd adapters/kotlin
go test -race ./...
```

The permanent foundation fixture covers package/import evidence, declarations, parameters, nullability, exclusions, malformed syntax, valid edge endpoints, repository-relative paths, CLI behavior, canonical IDs, and byte determinism. The focused relationship fixture covers exact, same-package, explicit-import, alias, aliased-nested, wildcard, and lexical-nested resolution; class and interface headers; delegation metadata; local annotation applications; ambiguous and external evidence; valid endpoints; and repeated full-stream byte determinism. The runtime fixture covers retained body/delegation ranges, top-level/member/object/companion/constructor calls, overload candidates, external/dynamic/unsupported calls, direct-super overrides, property/parameter reads, property writes, local/destructured/delegated exclusions, valid endpoints, and repeated byte determinism. The extension fixture covers typed parameters, locals, and properties; imported, aliased, wildcard, and lexical extension visibility; accepted arity; overload ambiguity; external extensions; fluent/safe/dynamic/inferred exclusions; scope and shadowing; valid endpoints; and repeated byte determinism. The dependency fixture covers Gradle Kotlin/Groovy, Maven, `kapt`/`ksp`, duplicates, unsupported computed forms, permanent exclusions, manifest evidence, valid endpoints, and repeated byte determinism.

## Current boundaries

This foundation is syntax-structural rather than compiler-semantic. It deliberately does not:

- evaluate Gradle settings, builds, source sets, dependencies, compiler plugins, or Kotlin scripts;
- invoke the Kotlin compiler or resolve targets from libraries, generated code, default imports, type aliases, expression types, or build configuration;
- perform compiler overload resolution beyond accepted argument counts for the bounded calls above, generic-call parsing, inherited or virtual dispatch, arbitrary imported callable binding, extension binding for inferred/external/chained/safe-call receivers, callable-value flow, or primary-constructor supertype-invocation analysis;
- infer reads or writes for locals, destructured/delegated symbols, arbitrary value receivers, accessors, initializers, aliases, or nested executable scopes;
- resolve computed, catalog-backed, interpolated, project, platform, parent/profile/BOM, plugin-generated, transitive, or installed dependency semantics; infer delegated-property behavior; or infer delegation behavior beyond retaining a declared supertype delegation expression and explicit secondary-constructor `this(...)`/`super(...)` calls;
- model enum entries, type parameters, local declarations, accessors, contracts, annotation arguments/use-site behavior, annotations as separate application nodes, or generated declarations;
- recover an unsafe declaration by guessing its name or shape.

Imports retain source evidence but target deterministic external syntactic symbols; they are not claims of successful compiler resolution. The relationship pass uses imports only to identify repository-local type and annotation-class targets. The runtime pass also uses imports while resolving explicit type/object/companion owners and constructors, and for the bounded repository-local extension visibility described above; it does not bind other imported top-level callables. External and ambiguous relationship targets remain unresolved. Runtime ambiguity becomes `possible-calls` only after ownership is proven; otherwise external, dynamic, ambiguous-owner, unsupported, or malformed evidence stays unresolved. These boundaries keep the adapter safe and deterministic while leaving compiler-backed resolution, broader runtime/dataflow analysis, and effective build-model resolution for later slices.
