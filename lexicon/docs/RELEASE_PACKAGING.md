# Lexicon release packaging

`tools/package_release.py` builds a clean distribution containing the Lexicon application and the runtime files required by every supported adapter.

## Build

From the repository root:

```text
python tools/package_release.py --output release --version <version>
```

Omitting `--version` leaves the application version as `dev`, which is appropriate only for local packaging tests.

Initialize a repository with the packaged application:

```text
release/lexicon init --repo /path/to/repository
```

On Windows, packaged executables use the `.exe` suffix.

## Distribution layout

The release directory contains:

- the `lexicon` application executable;
- `adapters/go/lexicon-go`;
- `adapters/gdscript/lexicon-gdscript`;
- `adapters/generic/lexicon-generic`;
- `adapters/rust/lexicon-rust`;
- the compiled TypeScript `dist/cli.js`;
- TypeScript production package metadata and runtime dependencies;
- the Python adapter source package;
- the Ruby adapter source files.

Tests, fixtures, caches, generated corpus output, source build trees, and development-only dependencies are not copied.

The packaged executable discovers the adjacent `adapters/` directory automatically. `lexicon init --adapters PATH` or `LEXICON_ADAPTERS` remains available when the executable and adapter directory are installed separately.

## Build requirements

Creating a complete distribution requires:

- Go for the application, Go adapter, GDScript adapter, and generic adapter;
- Rust and Cargo for the Rust adapter;
- Node.js and npm for TypeScript compilation and production dependency installation;
- Python to run the packaging script.

The packaging process must build from a verified source tree. It does not replace the repository test matrix.

## Runtime requirements

A packaged distribution does not require Go, Cargo, npm, or the TypeScript compiler.

Runtime requirements are:

- operating-system libraries required by the compiled Go and Rust binaries;
- Node.js for the compiled JavaScript, TypeScript, and Svelte adapter;
- Python for the Python adapter;
- Ruby for the Ruby adapter.

`lexicon doctor` validates configured adapter paths, runtime availability, storage, and consumer commands for an initialized repository.

## Verification

Before publishing a release:

1. run `python evaluation/run_tests.py`;
2. run the root and Go-adapter race suites when concurrency changed;
3. build the release into a clean output directory;
4. initialize a temporary mixed-language repository with the packaged executable;
5. run `status`, `doctor`, `scan`, and `export` from the package;
6. verify exported JSONL with `tools/validate_jsonl.py`;
7. confirm the package contains no tests, fixtures, caches, generated evaluation output, or build trees;
8. confirm the application reports the intended release version rather than `dev`.

The smoke utilities in `tools/` cover application operations, but the final packaged path must also be exercised because adapter discovery and runtime contents differ from source execution.

## Source-development fallback

The adapter runner prefers packaged binaries and the compiled TypeScript entry point. When those packaged paths are absent in a source checkout, it can use source-development execution paths such as `go run`, `cargo run`, Python module execution, Ruby source execution, and the locally built TypeScript output.

This fallback is for development. Releases should contain the packaged forms described above.
