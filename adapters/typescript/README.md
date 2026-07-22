# Lexicon JavaScript and TypeScript adapter

This adapter uses the TypeScript compiler API to emit deterministic Lexicon facts v1 JSONL for mixed JavaScript and TypeScript repositories.

## Supported source files

The scanner includes:

- `.ts`, `.tsx`, `.mts`, and `.cts`
- `.js`, `.jsx`, `.mjs`, and `.cjs`

It excludes `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, `node_modules/`, generated output directories, dependency/vendor trees, and common caches.

The stream language remains `typescript` because JavaScript and TypeScript share one compiler-backed semantic frontend and one stable node-ID namespace.

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

The adapter creates one TypeScript `Program` for all discovered JS/TS files with `allowJs` and `checkJs` enabled. Repository `tsconfig.json` or `jsconfig.json` options are preserved while analysis-required options remain enabled.

It emits declarations, imports, exports, inheritance, implementation, definite calls, possible calls, source spans, and explicit unresolved classifications.

Call resolution covers:

- functions, methods, constructors, overloads, and optional calls;
- compiler-backed receiver and signature resolution;
- class inheritance and interface dispatch candidates;
- arrow functions and function-valued declarations;
- higher-order parameters and callback arguments;
- JSX components and tagged templates;
- default exports and common transparent wrappers;
- JavaScript ESM imports and exports;
- CommonJS `require()`, `module.exports`, and `exports.name`;
- JSDoc-provided callable and structural types.

Relative imports may be extensionless or explicitly name any supported JS/TS extension. Exact and single-wildcard `baseUrl`/`paths` mappings from `tsconfig.json` or `jsconfig.json` resolve only unique scanned local targets.

## Canonical identities

All IDs use the contract `sha256:` form with language `typescript`.

- `repository`: repository root basename
- `directory`: normalized repository-relative directory path
- `file`: normalized repository-relative source path
- `module`: source path with its JS/TS extension removed
- declaration nodes: module key plus lexical declaration path
- import/export nodes: module key plus source position and names

Absolute checkout paths are never used in node identities or fact paths.

## Conservative boundaries

The adapter does not execute code or infer runtime mutation. Dynamic property installation, monkey-patching, computed exports, runtime-generated modules, and untyped polymorphic values remain dynamic or unresolved rather than guessed.

External packages are represented as `external-target` records unless their source is part of the scanned repository. JavaScript without types or JSDoc naturally provides less dispatch precision than typed TypeScript.

`.astro` files are not parsed by this adapter. Astro frontmatter and template/component relationships require a separate Astro-aware frontend if that support becomes necessary.

## Tests

```sh
npm test
```

The suite covers TypeScript semantics, JavaScript ESM, JSX, CommonJS, JSDoc flow, path mappings, exclusions, stable IDs, deterministic repeat runs, and the shared JSONL validator.
