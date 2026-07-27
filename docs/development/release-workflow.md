# Release workflow

The root workflow coordinates the independent Grimoire Go, Lexicon Go, Arcana Rust, and Lodestone native build roots. It does not merge their build systems or generate source into component trees.

## CPU bounds

Build, test, and release commands default to one worker:

```bash
python scripts/workflow.py test
python scripts/workflow.py build --version 0.1.0-dev
python scripts/workflow.py release --version 1.2.3 --output dist
```

The bound applies to:

- Go package build concurrency;
- Go test parallelism;
- Cargo build jobs;
- Rust test threads.

This default prevents the workflow from spawning a large package-test swarm. Increase concurrency only deliberately:

```bash
python scripts/workflow.py test --jobs 2
python scripts/workflow.py release --version 1.2.3 --jobs 2
```

`--jobs` must be at least 1. A high value can still overload the machine.

Use `py -3` instead of `python` when that is the configured Windows launcher.

## Smoke checks

The deterministic packaging/install smoke suite does not compile every component:

```bash
python scripts/workflow.py smoke
python scripts/test_workflow.py
```

It validates archive layout, fixed metadata, checksums, version validation, selected-component installation, and concurrency defaults.

## Build layout

```text
build/
  bin/
    grimoire(.exe)
    lexicon(.exe)
    arcana(.exe)
  native/
    lodestone_ffi.dll | liblodestone_ffi.so | liblodestone_ffi.dylib
```

The Go CLIs receive the version through linker flags. Arcana receives the same value through `GRIMOIRE_RELEASE_VERSION`; a standalone Cargo build still reports its manifest version.

## Local installation

```bash
python scripts/workflow.py install --source build --bin-dir PATH
python scripts/workflow.py install --source build --bin-dir PATH --component grimoire
python scripts/workflow.py install --source build --bin-dir PATH --component lexicon --component arcana
```

Omitting `--component` installs all three applications. Repeating it installs only the selected subset. When Grimoire is selected, the Lodestone native library is copied beside the executable so ordinary dynamic-library discovery works without extra configuration.

The installer does not modify `PATH` or user configuration.

## Release packaging

```bash
python scripts/workflow.py release --version 1.2.3 --output dist
```

The command first runs the bounded complete test matrix, then builds and writes:

```text
dist/1.2.3/
  grimoire-1.2.3-<platform>-<arch>.zip
  lexicon-1.2.3-<platform>-<arch>.zip
  arcana-1.2.3-<platform>-<arch>.zip
  grimoire-bundle-1.2.3-<platform>-<arch>.zip
  release-manifest.json
  SHA256SUMS.txt
```

The Grimoire archive contains its required native library beside the executable. The combined bundle contains `bin/`, `native/`, `install.py`, and `VERSION`.

Archives use sorted entries and fixed timestamps for deterministic packaging. The command creates local files only; it does not publish, tag, or push a release.

Before external publication:

1. Inspect `release-manifest.json`.
2. Verify `SHA256SUMS.txt`.
3. Exercise all component version commands.
4. Run a bounded discovery smoke test against a representative repository.
