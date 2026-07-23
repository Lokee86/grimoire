# Lexicon internal packages

This directory contains the private Go implementation of the Lexicon application. Packages under `internal/` are not public Go APIs.

## Purpose

The internal packages own application orchestration, repository state, snapshot storage, consumer execution, and watch behavior around the independently executable language adapters.

## Direct folders

| Folder | Responsibility |
| --- | --- |
| `adapters/` | Locate adapter runtimes, calculate fingerprints, construct requests, and execute adapter processes |
| `cli/` | Parse commands and flags, resolve repositories, and format user-facing operation results |
| `config/` | Read and write `.lexicon/config.json`, adapter roots, enabled languages, and analysis configuration identity |
| `consumer/` | Manage deterministic post-publication consumer definitions, execution, and state |
| `files/` | Discover relevant repository files and enforce permanent plus `.lexiconignore` exclusions |
| `languages/` | Define the supported application language registry |
| `lock/` | Enforce one repository writer at a time |
| `objectstore/` | Parse adapter analysis, write/read immutable objects, manage manifests, export JSONL, recover publication, and collect garbage |
| `scan/` | Plan complete or scoped analysis, schedule adapters, apply results, and publish transactions |
| `scope/` | Build temporary language-scoped repositories with required context |
| `state/` | Maintain the private source mirror and detect changes between successful publications |
| `watch/` | Convert filesystem events and reconciliation intervals into scan transactions |

## Does not own

These packages do not own language parsing or semantic extraction; that belongs in `adapters/<language>/`.

They also do not define the public facts, object, snapshot, or runtime-evidence contracts; those live in [`spec/`](../spec/README.md).

## Related documentation

- [Architecture](../docs/ARCHITECTURE.md)
- [Application and operations](../docs/APPLICATION.md)
- [Development and verification](../docs/DEVELOPMENT.md)
- [Current status](../docs/STATUS.md)

## Placement rules

Put behavior in the narrowest owning package. Do not create vague shared utility packages or consumer-specific policy inside application orchestration.

Cross-package types should represent durable application concepts such as adapter requests, scan plans, manifests, or consumer definitions. Keep parser-specific models inside their adapter runtime.
