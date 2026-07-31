# Lexicon Maintainer Map

Parent index: [Lexicon Documentation](README.md)

## Purpose

This document routes common Lexicon changes to their canonical documentation and implementation boundary. It intentionally avoids enumerating every source file.

## Overview

Use this page to select the owning Lexicon document or adapter README. Continue in that document's focused code map for exact implementation and test paths.

## Change routing

| Change area | Canonical documentation | Primary implementation boundary |
| --- | --- | --- |
| Commands and operator behavior | [Application](APPLICATION.md) | `cmd/lexicon/`, `internal/cli/` |
| Scan planning, publication, recovery, and concurrency | [Architecture](ARCHITECTURE.md) | `internal/scan/`, `internal/lock/`, `internal/watch/` |
| Immutable facts, objects, snapshots, export, and GC | [Architecture](ARCHITECTURE.md), specifications under `spec/` | `internal/objectstore/` |
| Adapter discovery and execution | [Application](APPLICATION.md), [Adapters](../adapters/README.md) | `internal/adapters/`, `internal/languages/` |
| Language semantics | Owning adapter README | `adapters/<language>/` |
| Dependency and incremental scope semantics | [Dependency semantics](DEPENDENCY_SEMANTICS.md) | adapter dependency emitters, `internal/objectstore/dependencies.go`, `internal/scan/plan.go` |
| Interstack contracts | [Architecture](ARCHITECTURE.md) | `internal/interstack/`, `internal/scan/interstack.go` |
| Post-publication consumers | [Application](APPLICATION.md) | `internal/consumer/`, `internal/cli/consumers.go` |
| Build, tests, corpora, and semantic validation | [Development](DEVELOPMENT.md) | `evaluation/`, adapter tests, `tools/` |
| Release bundles and installer verification | [Release packaging](RELEASE_PACKAGING.md) | `tools/package_release.py`, packaging smoke tests |

## Boundaries

- Adapters emit normalized facts; they do not publish snapshots.
- The scanner publishes immutable state; it does not implement language semantics.
- Lexicon does not own graph traversal, ranking, or packed graph storage.
- Focused implementation paths and tests belong in the `## Code map` section of the owning document or adapter README.

## Related docs

- [Architecture](ARCHITECTURE.md)
- [Application](APPLICATION.md)
- [Development](DEVELOPMENT.md)
- [Status](STATUS.md)

## Notes

Start here when ownership is unclear, then continue in the subject-specific document.
