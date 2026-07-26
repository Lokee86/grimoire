# Release workflow

The root workflow is a thin coordinator over the independent Grimoire Go,
Lexicon Go, Arcana Rust, and native vector-engine build roots. It does not add
workspace imports or generated source files to any component.

## Build and test

From the repository root:

```text
python scripts/workflow.py build --version 0.1.0-dev
python scripts/workflow.py test
python scripts/workflow.py smoke
```

On Windows, use `py -3` in place of `python` if needed. The build layout is:

```text
build/
  bin/
    grimoire(.exe)
    lexicon(.exe)
    arcana(.exe)
  native/
    grimoire_vector_ffi.dll | .so | .dylib
    grimoire-vector(.exe)
```

The Go CLIs receive the version through `-ldflags -X`. Arcana receives the same
value through `GRIMOIRE_RELEASE_VERSION` and a small Cargo build script; a
standalone `cargo build` still reports the manifest version.

## Local installation

Install the collected binaries into an explicit directory:

```text
python scripts/workflow.py install --source build --bin-dir PATH
python scripts/workflow.py install --source build --bin-dir PATH --component grimoire
python scripts/workflow.py install --source build --bin-dir PATH --component lexicon --component arcana
```

Omitting `--component` installs all three applications. Repeating it installs only the selected subset. The native library is copied beside `grimoire` whenever Grimoire is selected. This is important on Windows:
the `grimoire_vector_ffi.dll` is then discoverable by normal DLL lookup without
setting `GRIMOIRE_VECTOR_ENGINE`. The installer does not modify `PATH` or user
configuration.

## Release packaging

Release packaging runs the complete test matrix before building:

```text
python scripts/workflow.py release --version 1.2.3 --output dist
```

The command writes `dist/1.2.3/` with these independently usable archives:

```text
grimoire-1.2.3-<platform>-<arch>.zip
lexicon-1.2.3-<platform>-<arch>.zip
arcana-1.2.3-<platform>-<arch>.zip
vector-engine-1.2.3-<platform>-<arch>.zip
grimoire-bundle-1.2.3-<platform>-<arch>.zip
release-manifest.json
SHA256SUMS.txt
```

The Grimoire artifact includes its native library beside the executable. The
vector-engine artifact contains the library and diagnostic native CLI. The
combined bundle contains `bin/`, `native/`, `install.py`, and `VERSION`; run
the included installer with `python install.py --bin-dir PATH` after extracting it. The extracted bundle directory is the default source, and the same repeatable `--component` flag selects a subset. Archives use fixed timestamps and sorted entries for
deterministic packaging. The command only creates local files: it does not
publish, tag, or push a release.

Before publishing an artifact outside this repository, inspect
`release-manifest.json`, verify `SHA256SUMS.txt`, and exercise each CLI's
version command. The root build already checks that all three report the one
requested release version.