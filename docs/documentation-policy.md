# Documentation Policy

Parent index: [Grimoire Documentation](INDEX.md)

## Purpose

This document defines documentation ownership and minimum coverage for Grimoire, Lexicon, and Arcana.

## Overview

Grimoire uses the shared CLI-product profile with stateful and research capabilities. The root documentation tree owns the unified product and cross-component contracts. Lexicon and Arcana retain independently indexed component documentation under their source roots.

## Canonical ownership

| Information | Canonical owner |
| --- | --- |
| Product introduction, installation, quick start, and component summary | `README.md` |
| Unified architecture, ownership, data flow, and state alignment | `docs/architecture/` |
| Exact Grimoire CLI, MCP, discovery, indexing, embedding, and storage contracts | `docs/reference/` |
| Tests, benchmarks, release workflow, retrieval quality, coverage, and behavioral contracts | `docs/development/` |
| Current product limitations | `docs/limits/` |
| Future and unresolved product work | `docs/planning/` |
| Lexicon application, architecture, adapters, semantics, development, packaging, status, and codemap | `lexicon/docs/` |
| Arcana application, architecture, graph/storage contracts, development, status, and codemap | `arcana/docs/` |
| Stable agent operating rules | `AGENTS.md` |

## Rules

- Root reference pages describe implemented Grimoire behavior, defaults, failure modes, and public contracts.
- Architecture pages identify implemented ownership, non-ownership, state, lifecycle, failure, and recovery boundaries.
- Component-specific behavior belongs in the owning component documentation and is summarized, not duplicated, at the root.
- Source, document, symbol, and relationship evidence lanes remain distinct in documentation.
- Benchmark and research claims name their corpus, method, artifacts, limitations, and task scope.
- Planning pages never serve as current implementation reference.
- Limitations remain explicit until resolved.
- Every production package, command family, component boundary, stateful flow, and machine-readable contract maps to canonical current documentation.
- Every documentation tree is independently indexed and checked.

## Required change impact

Changes to CLI commands, schemas, defaults, diagnostics, evidence lanes, handles, sessions, repository preparation, snapshot alignment, embeddings, vectors, language facts, graph protocol, graph storage, or release layout require updates to the exact owner in the same change.

## Related docs

- [Documentation procedure](documentation-procedure.md)
- [Documentation coverage](development/documentation-coverage.md)
- [Component architecture](architecture/components.md)
- [Shared documentation standard](../.standards/docs/documentation-standard.md)

## Notes

The top-level product may depend on Lexicon and Arcana, but their independently usable surfaces remain documented and governed as components.
