# Lexicon tooling scripts

This directory contains repository-maintenance, validation, reporting, smoke-test, runtime-evidence, and release-packaging utilities.

## Purpose

The scripts support development and release verification. They are not part of the adapter facts contract or the snapshot consumer API.

## Direct files

| File | Responsibility |
| --- | --- |
| `validate_jsonl.py` | Validate facts-v1 JSONL structure, identities, ownership, references, and canonical ordering |
| `semantic_report.py` | Summarize emitted node, relation, unresolved, and ownership counts |
| `reconcile_runtime.py` | Validate runtime-evidence streams and classify observations against a static facts stream |
| `package_release.py` | Build a clean application distribution containing the executable and required adapter runtimes |
| `smoke_app.py` | Exercise initialization and core application behavior against temporary repositories |
| `smoke_operations.py` | Exercise operational commands such as export, garbage collection, language selection, and consumers |
| `test_reconcile_runtime.py` | Regression tests for runtime-evidence reconciliation |
| `test_semantic_report.py` | Regression tests for semantic reporting |

## Common commands

Validate and summarize an adapter stream:

```text
python tools/validate_jsonl.py /path/to/facts.jsonl
python tools/semantic_report.py /path/to/facts.jsonl
```

Run tooling regressions and application smoke coverage:

```text
python tools/test_reconcile_runtime.py
python tools/test_semantic_report.py
python tools/smoke_app.py
python tools/smoke_operations.py
```

Build a release directory:

```text
python tools/package_release.py --output release --version <version>
```

Omit `--version` only for a local packaging test where a `dev` application version is acceptable.

## Does not own

These scripts do not define facts-v1, objects-v1, snapshots-v1, or runtime-evidence-v1 semantics. The normative contracts live under [`spec/`](../spec/README.md).

The scripts also do not replace adapter fixtures or the real-repository corpus. The complete test matrix and semantic corpus live under [`evaluation/`](../evaluation/README.md).

## Placement rules

Put bounded developer or release utilities here when they operate across application components or contracts.

Language-specific test helpers belong in the owning adapter. Repeatable corpus orchestration belongs in `evaluation/`. Production application behavior belongs in Go packages under `internal/` or the owning adapter runtime, not in a Python maintenance script.
