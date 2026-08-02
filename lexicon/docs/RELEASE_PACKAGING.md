# Lexicon release packaging

Parent index: [Lexicon Documentation](README.md)

## Purpose

This document defines Lexicon's release build, distribution layout, runtime requirements, verification, installer behavior, and source-development fallbacks.

## Overview

A release packages the Lexicon application together with independently executable language adapters and required runtime assets without making source-development fallbacks part of the installed contract.

`tools/package_release.py` builds a clean distribution containing the Lexicon application and the runtime files required by every supported adapter.

## Build

From the repository root:

```text
python tools/package_release.py --output release --version <version>
```

Omitting `--version` leaves the application version as `dev`, which is appropriate only for local packaging tests.

Install an extracted package for the current user:

```text
# Windows PowerShell
.\install.ps1

# Linux and other Unix-like systems
./install.sh
```

The PowerShell installer defaults to `%LOCALAPPDATA%\Programs\Lexicon` and adds that directory to the user `PATH`. Use `-InstallDir PATH` to choose another location or `-NoPath` to leave `PATH` unchanged.

The shell installer defaults to `${XDG_DATA_HOME:-$HOME/.local/share}/lexicon` and creates `$HOME/.local/bin/lexicon`. Pass an installation directory as its first argument, or set `LEXICON_INSTALL_DIR` and `LEXICON_BIN_DIR`.

Initialize a repository with the installed application:

```text
lexicon init --repo /path/to/repository
```

On Windows, packaged executables use the `.exe` suffix.

## Distribution layout

The release directory contains:

- the `lexicon` application executable;
- the platform current-user installer: `install.ps1` on Windows or `install.sh` on Unix-like systems;
- `adapters/c-family/lexicon-c-family`;
- `adapters/go/lexicon-go`;
- `adapters/gdscript/lexicon-gdscript`;
- the self-contained C# adapter under `adapters/csharp/`;
- the Java adapter executable, compiler JAR, and bundled Java runtime under `adapters/java/`;
- `adapters/kotlin/lexicon-kotlin`;
- `adapters/lotusscript/lexicon-lotusscript`;
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

- Go plus a working CGO C compiler for the application, C/C++ adapter, Go adapter, GDScript adapter, LotusScript adapter, and generic adapter;
- Rust and Cargo for the Rust adapter;
- JDK 21 or newer for the Java compiler engine and bundled Java runtime;
- a .NET SDK capable of self-contained publishing for the C# adapter;
- Node.js and npm for TypeScript compilation and production dependency installation;
- Python to run the packaging script.

The packaging process must build from a verified source tree. It does not replace the repository test matrix.

## Runtime requirements

A packaged distribution does not require Go, Cargo, npm, or the TypeScript compiler.

Runtime requirements are:

- operating-system libraries required by the compiled Go, Tree-sitter/CGO, Rust, and self-contained C# binaries;
- no separately installed Java runtime for the packaged Java adapter, because its runtime and compiler engine are bundled;
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
6. run `python tools/smoke_installers.py --distribution PATH --version VERSION` to verify the platform installer;
7. verify exported JSONL with `tools/validate_jsonl.py`;
8. confirm the package contains no tests, fixtures, caches, generated evaluation output, or build trees;
9. confirm the application reports the intended release version rather than `dev`.

The smoke utilities in `tools/` cover application operations, but the final packaged path must also be exercised because adapter discovery and runtime contents differ from source execution.

## Source-development fallback

The adapter runner prefers packaged binaries, the packaged C# and Java runtimes, and the compiled TypeScript entry point. When those packaged paths are absent in a source checkout, it can use source-development execution paths such as `go run`, `dotnet run`, embedded Java compiler construction from an installed JDK, `cargo run`, Python module execution, Ruby source execution, and the locally built TypeScript output.

This fallback is for development. Releases should contain the packaged forms described above.

## Code map

| Packaging concern | Primary implementation | Related verification |
| --- | --- | --- |
| Release assembly | `tools/package_release.py` | `tools/test_package_release.py` |
| Runtime reconciliation | `tools/reconcile_runtime.py` | `tools/test_reconcile_runtime.py` |
| Application smoke test | `tools/smoke_app.py` | packaged executable checks |
| Operations smoke test | `tools/smoke_operations.py` | init/scan/export/consumer/GC checks |
| Installer smoke test | `tools/smoke_installers.py` | release-layout installation checks |
| Adapter executables and runtime lookup | `internal/adapters/runner_packaged.go`, `runner.go` | adapter runner tests and smoke checks |
| Version/build identity | `internal/cli/version.go`, build metadata | CLI tests and release workflow |

The package contains runtime artifacts; adapter source directories and generated validation output remain development inputs rather than installed authority.

## Related docs

- [Development and verification](DEVELOPMENT.md)
- [Application and operations](APPLICATION.md)
- [Adapter documentation](../adapters/README.md)
- [Root release workflow](../../docs/development/release-workflow.md)

## Notes

Packaged executables and compiled adapter outputs are release artifacts; adapter source remains the implementation owner.
