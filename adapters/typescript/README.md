# Lexicon JavaScript, TypeScript, and Svelte adapter

This adapter uses the TypeScript compiler API plus an offset-preserving Svelte script frontend to emit deterministic Lexicon facts v1 JSONL for mixed JavaScript, TypeScript, and Svelte repositories.

## Supported source files

The scanner includes:

- `.ts`, `.tsx`, `.mts`, and `.cts`
- `.js`, `.jsx`, `.mjs`, and `.cjs`
- `.svelte` component files with JavaScript or TypeScript `<script>` blocks

It excludes `.git/`, `.worktrees/`, `.workingtrees/`, `.ddocs/`, `.lexicon/`, `.arcana/`, `.grimoire/`, `.pitlord/`, `.cantrip/`, `.homunculus/`, `.incubus/`, `.ritual/`, `.warlock/`, `node_modules/`, generated output directories, dependency/vendor trees, and common caches.

The stream language remains `typescript` because JavaScript, TypeScript, and Svelte scripts share one compiler-backed semantic frontend and one stable node-ID namespace.

## Setup and usage

From this directory:

```sh
npm install
npm run build
node dist/cli.js --repo /path/to/repository --output /path/to/facts.jsonl
```

`--output -` writes UTF-8 JSONL to stdout. From the Lexicon repository root:

```sh
npm --prefix adapters/typescript install
npm --prefix adapters/typescript run build
node adapters/typescript/dist/cli.js --repo /path/to/repository --output /path/to/facts.jsonl
```

Validate a stream with:

```sh
python tools/validate_jsonl.py /path/to/facts.jsonl
```

## Analysis model

The adapter creates one TypeScript `Program` for all discovered JS, TS, and Svelte files with `allowJs` and `checkJs` enabled. Svelte markup and styles are replaced with same-length whitespace before parsing, so declarations and relationships emitted from `<script>` blocks retain their original component line and column spans. Repository `tsconfig.json` or `jsconfig.json` options are preserved while analysis-required options remain enabled.

It emits declarations, imports, exports, inheritance, implementation, definite calls, possible calls, source spans, and explicit unresolved classifications.

Class `extends` and `implements` relationships are compiler-backed. Mixin-style `extends factory(Base)` heritage is emitted as `uses-trait` when the factory is repository-local, and concrete overrides target inherited class methods rather than interface declarations. Interface signatures are contracts only: they are removed from runtime call targets, with one concrete implementation emitted as `calls` and multiple candidates as `possible-calls`.

Call resolution covers:

- functions, methods, constructors, overloads, and optional calls;
- compiler-backed receiver and signature resolution;
- class inheritance and interface dispatch candidates;
- arrow functions and function-valued declarations;
- higher-order parameters and callback arguments;
- JSX components and tagged templates;
- Svelte module and instance script content;
- default exports and common transparent wrappers;
- JavaScript ESM imports and exports;
- CommonJS `require()`, `module.exports`, and `exports.name`;
- JSDoc-provided callable and structural types.

Relative imports may be extensionless or explicitly name any supported JS/TS extension or `.svelte`. Exact and single-wildcard `baseUrl`/`paths` mappings from `tsconfig.json` or `jsconfig.json` resolve only unique scanned local targets.

## Canonical identities

All IDs use the contract `sha256:` form with language `typescript`.

- `repository`: repository root basename
- `directory`: normalized repository-relative directory path
- `file`: normalized repository-relative source path
- `module`: source path with its JS/TS or `.svelte` extension removed
- declaration nodes: module key plus lexical declaration path
- import/export nodes: module key plus source position and names

Absolute checkout paths are never used in node identities or fact paths.

## Conservative boundaries

The adapter does not execute code or infer runtime mutation. Dynamic property installation, monkey-patching, computed exports, runtime-generated modules, reflection, prototype mutation, and untyped polymorphic values remain dynamic or unresolved rather than guessed.

Svelte markup is not yet converted into component-use, event-handler, binding, slot, or template-expression relationships. Module and instance script blocks are analyzed together as one virtual TypeScript module, so duplicate names across those two Svelte scopes remain a conservative limitation.

External packages are represented as `external-target` records unless their source is part of the scanned repository. JavaScript without types or JSDoc naturally provides less dispatch precision than typed TypeScript.

`.astro` files are not parsed by this adapter. Astro frontmatter and template/component relationships require a separate Astro-aware frontend if that support becomes necessary.

## Tests

```sh
npm test
```

The suite covers TypeScript semantics, JavaScript ESM, JSX, CommonJS, JSDoc flow, Svelte script extraction and module resolution, path mappings, exclusions, stable IDs, deterministic repeat runs, and the shared JSONL validator.

## Dependency semantics

Literal `package.json` `dependencies`, `devDependencies`, `peerDependencies`, and `optionalDependencies` entries emit repository `depends-on` facts. `file:` and `link:` values are marked with `path: true`; targets use facts-v1 `module` nodes with `dependency:typescript:<normalized-target>` identity and `.lexicon/dependencies/typescript/...` paths. Relative imports and unique exact/single-wildcard `tsconfig.json` or `jsconfig.json` path mappings emit local module `depends-on` edges while preserving `imports`. The same rules apply to JavaScript and Svelte script modules.

Malformed package JSON, computed or runtime-generated dependency entries, unresolved package aliases, and dynamic module expressions are unsupported. The adapter reads package metadata as data and never executes package manifests.
