# Lexicon dependency semantics

All adapters emit dependency targets as facts-v1 `module` nodes. Synthetic targets use the canonical identity:

```text
dependency:<ecosystem>:<normalized-target>
```

Their paths are deterministic `.lexicon/dependencies/<ecosystem>/...` paths unless the target is a repository-local path. Manifest or module nodes emit `depends-on` edges. Existing `imports` edges remain unchanged; a local `depends-on` edge is added only when the local target resolves uniquely to a scanned module/file.

Dependency edge attributes are deterministic and always use these scalar fields:

- `category`: `runtime`, `development`, `test`, `build`, `peer`, `optional`, `plugin`, `autoload`, or `local`;
- `constraint`: the literal declared version/range, or an empty string;
- `source`: the normalized manifest section or literal loader/source text;
- `optional`, `dev`, `build`, `peer`, and `path`: boolean classification flags.

The parsers read manifest text or JSON/TOML data only. They never execute manifests, resolve packages by installing them, or evaluate dynamic expressions. Malformed, computed, URL/VCS, and otherwise non-literal entries are omitted unless an adapter's existing import analysis already emits its normal unresolved classification. Ordering follows facts-v1 node/edge ordering and all repeated runs are byte-identical.

## Adapter coverage

- Go: `go.mod` `require` entries, including blocks, `replace` entries, and repository-local Go imports.
- Python: literal `[project].dependencies`, `[project.optional-dependencies]`, `requirements*.txt` fallback, editable local requirements, and repository-local Python imports.
- Ruby: literal `Gemfile` `gem` calls, literal gemspec runtime/development dependency calls, `require_relative`, and `load`/`require` remain conservative for non-local targets.
- TypeScript/JavaScript/Svelte: `package.json` runtime/dev/peer/optional sections, `file:`/`link:` paths, relative imports, and unique `tsconfig`/`jsconfig` path-mapped imports.
- GDScript/Godot: enabled editor plugins, autoload entries, explicit `res://` resource/script paths in `project.godot`, and literal local `preload`/`load` references.
- Rust: Cargo normal/development/build/target/path dependencies through Cargo metadata, plus resolved local Rust module imports.

Unsupported forms include dynamic manifest construction, dependency execution, unresolved package-manager aliases, non-literal Ruby dependency calls, computed Godot paths, and unresolved Rust/Cargo metadata entries. These are not treated as proven dependency edges.
