# Documentation Policy

Parent index: [Grimoire Documentation](INDEX.md)

## Purpose

This document defines documentation ownership and minimum coverage for Grimoire, Lexicon, and Arcana.

## Overview

Grimoire uses the shared CLI-product profile with stateful, protocol, data-pipeline, and research capabilities. Lexicon uses the library-engine profile with stateful and data-pipeline capabilities; Arcana uses the library-engine profile with stateful, protocol, and data-pipeline capabilities. The root documentation tree owns the unified product and cross-component contracts, while Lexicon and Arcana retain independently indexed component documentation under their source roots.

## Canonical ownership

| Information | Canonical owner |
| --- | --- |
| Product introduction, installation, quick start, and component summary | `README.md` |
| Unified architecture, ownership, data flow, and state alignment | `docs/architecture/` |
| Exact Grimoire CLI, MCP, discovery, indexing, embedding, and storage contracts | `docs/reference/` |
| Tests, benchmarks, release workflow, retrieval quality, coverage, and behavioral contracts | `docs/development/` |
| Current product limitations | `docs/limits/` |
| Future and unresolved product work | `docs/planning/` |
| Lexicon application, architecture, adapters, semantics, development, packaging, status, and maintainer routing | `lexicon/docs/` |
| Arcana application, architecture, graph/storage contracts, development, status, and maintainer routing | `arcana/docs/` |
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
- Documentation baselines are not permitted; root, Lexicon, and Arcana checks must pass without suppressed findings.
- Pitlord owns executable repository architecture policy under `tools/pitlord/`; the shared documentation checker and `scripts/check_docs.py` own documentation structure and repository-specific documentation governance. All run in standards CI.

## Code map policy

Implementation-facing architecture, reference, development, contract, pipeline, and adapter documents include a focused `## Code map` section.

A focused code map identifies:

- the primary implementation files or folders for the document's subject;
- related generated or source artifacts when applicable;
- the tests or verification surface protecting that behavior; and
- important non-ownership boundaries.

The Grimoire, Lexicon, and Arcana maintainer maps are routing aids only. They help select the canonical subject document and must not replace subject-specific code maps or explanatory prose.

## Required change impact

Changes to CLI commands, schemas, defaults, diagnostics, evidence lanes, handles, sessions, repository preparation, snapshot alignment, embeddings, vectors, language facts, graph protocol, graph storage, or release layout require updates to the exact owner in the same change.

## Related docs

- [Documentation procedure](documentation-procedure.md)
- [Documentation coverage](development/documentation-coverage.md)
- [Component architecture](architecture/components.md)
- [Shared documentation standard](../.standards/docs/documentation-standard.md)

## Notes

The top-level product may depend on Lexicon and Arcana, but their independently usable surfaces remain documented and governed as components.
