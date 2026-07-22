# Lexicon TypeScript adapter

The first runnable TypeScript adapter slice uses the TypeScript compiler API. It is self-contained under this directory and emits deterministic Lexicon facts v1 JSONL.

## Setup and usage

From this directory:

```sh
npm install
npm run build
node dist/cli.js --repo /path/to/repository --output /path/to/facts.jsonl
```

`--output -` writes UTF-8 JSONL to stdout. The output directory is created when needed. From the Lexicon repository root, the same flow is:

```sh
npm --prefix adapters/typescript install
npm --prefix adapters/typescript run build
node adapters/typescript/dist/cli.js --repo /path/to/repository --output /path/to/facts.jsonl
```

Validate a stream with:

```sh
python tools/validate_jsonl.py /path/to/facts.jsonl
```

## Canonical identities

All node IDs use the contract's `sha256:` identity form with language `typescript`.

- `repository`: repository root basename.
- `directory`: normalized repository-relative directory path; the root is `.`.
- `file`: normalized repository-relative `.ts` or `.tsx` path.
- `module`: normalized repository-relative source path without its `.ts`/`.tsx` extension. Import resolution also recognizes a matching `/index` module.
- `type`: module key plus lexical class or type-alias name.
- `interface`: module key plus lexical interface name.
- `function`: module key plus lexical function name.
- `method`: module key plus lexical class/interface and method name.
- `constructor`: module key plus lexical class/interface and `constructor`.
- `field`: module key plus lexical owner and field name.
- `variable`/`constant`: module key plus lexical scope and binding name; `const` declarations are constants.
- `import`/`export`: module key, source position, and imported/exported name set.

File nodes carry the SHA-256 digest of the original file bytes. No absolute checkout path is used in a node identity or fact path.

## Supported facts

The adapter emits repository, directory, file, module, class/type, interface, function, method, constructor, field, variable/constant, import, and export nodes. It emits `contains` edges for repository structure and file/module ownership, `defines` edges for declarations and import/export facts, `imports` edges for resolved local modules and named symbols, `extends` edges for statically resolved class/interface heritage, `implements` edges for statically resolved class contracts, and conservative `calls` edges for direct identifier function calls and identifier constructors that resolve to one scanned target.

It scans `.ts` and `.tsx` files in sorted repository-relative path order. It excludes `.git/`, `.worktrees/`, `.workingtrees/`, `.warlock/`, `node_modules/`, `build/`, `dist/`, coverage/cache directories, vendor directories, and other common generated output directories.

## Current limits

- Resolution is repository-local and syntax-based. It does not execute TypeScript, inspect package exports, or inspect installed packages.
- Relative imports resolve literal module paths and `/index` modules. Exact and single-wildcard `baseUrl`/`paths` mappings from a repository `tsconfig.json` or `jsconfig.json` resolve only unique scanned local targets.
- External packages are represented by unresolved `external-target` records rather than guessed nodes.
- Named imports and simple dotted heritage expressions resolve when the target declaration is scanned. Default and namespace imports resolve to their module node.
- Dynamic `import(...)`, computed heritage targets, computed names, malformed source, generated declarations, and unsupported relationships produce unresolved records rather than guessed edges.
- Export declarations are represented as `export` nodes. Static re-export sources also produce an `imports` edge when the local module is available.
- Direct identifier calls and constructors resolve only to one scanned function or type. Property/element calls, callable variables, optional chains, overloads, method dispatch, JSX, references, decorators, type-checker-only symbols, and runtime module loading remain unresolved or unmodeled.

## Tests

```sh
npm test
```

The focused tests cover modules, classes/interfaces, methods, variables/constants, imports/exports, inheritance and implementation, exclusions, stable IDs/content IDs, canonical ordering, repeat runs, and the root JSONL validator.
