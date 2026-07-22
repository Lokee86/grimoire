# Lexicon TypeScript adapter

The TypeScript adapter uses the TypeScript compiler API and type checker to emit deterministic Lexicon facts v1 JSONL for `.ts` and `.tsx` repositories.

## Setup and usage

From this directory:

```sh
npm ci
npm run build
node dist/cli.js --repo /path/to/repository --output /path/to/facts.jsonl
```

`--output -` writes UTF-8 JSONL to stdout. Validate a generated stream from the Lexicon repository root with:

```sh
python tools/validate_jsonl.py /path/to/facts.jsonl
```

## Static call graph

The adapter builds a real TypeScript `Program` from the repository's `tsconfig.json` or `jsconfig.json`, then maps compiler symbols and resolved signatures back to Lexicon declarations.

A single proven repository target emits `calls`. Multiple legitimate static targets emit `possible-calls`. Calls into installed packages, browser APIs, runtime-created values, or unresolved dynamic values remain explicit unresolved records rather than guessed local edges.

Current resolution covers:

- direct, imported, overloaded, optional, property, element, constructor, and tagged-template calls;
- arrow functions, function expressions, function-valued variables and properties;
- local methods and constructors through compiler-resolved signatures;
- nominal class dispatch and possible implementation sets for interface/base-typed receivers;
- local higher-order parameters through fixed-point argument propagation;
- callback arguments when the selected parameter is callable;
- JSX components, including local default exports;
- transparent callable wrappers such as `forwardRef`, `memo`, and `Object.assign`;
- exact and wildcard `baseUrl`/`paths` mappings;
- named re-exports and package-relative module resolution.

## Emitted graph

The adapter emits repository, directory, file, module, type, interface, function, method, constructor, field, variable, constant, import, and export nodes. Relationships include:

- `contains` and `defines` for ownership;
- `imports` for resolved local modules and symbols;
- `extends` and `implements` for heritage;
- `calls` for definite call targets;
- `possible-calls` for virtual dispatch, callbacks, and other exact static candidate sets.

All IDs use the contract's `sha256:` identity form. File nodes include a SHA-256 content identity, and no absolute checkout path is included in emitted facts.

## Deliberate boundaries

- Installed package implementations are not scanned as repository nodes.
- Values produced by framework hooks, browser globals, dependency injection, reflection, proxies, and runtime mutation remain external or dynamic unless ordinary TypeScript flow proves a local implementation.
- Computed property names without a statically known key, unresolved `any`/`unknown` receivers, dynamic imports, and runtime module loading remain unresolved.
- Interface calls include scanned class implementations compatible with the receiver type; runtime implementations outside the repository remain outside the local candidate set.
- Generated output and dependency directories are excluded, including `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, `node_modules/`, `build/`, `dist/`, coverage/cache directories, and vendor directories.

## Tests

```sh
npm test
```

The suite covers imports and re-exports, alias mappings, inheritance, nominal and interface dispatch, overloads, callbacks, wrappers, JSX/default exports, tagged templates, optional calls, deterministic repeat runs, stable identities, and the shared JSONL validator.
