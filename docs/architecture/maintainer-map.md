# Grimoire Maintainer Map

Parent index: [Architecture Documentation](INDEX.md)

## Purpose

This document routes maintainers to the canonical document and implementation boundary for common Grimoire changes. It is not a repository inventory and does not replace the focused code maps inside implementation-facing documents.

## Overview

Use this page only when the owning document is unclear. Once routed, use that document's focused code map for exact implementation and verification paths.

## Component routing

| Change area | Canonical documentation | Primary implementation boundary |
| --- | --- | --- |
| CLI, MCP, and command dispatch | [CLI](../reference/cli.md), [Agent MCP](../reference/agent-mcp.md) | `cmd/grimoire/`, `internal/app/`, `internal/mcpserver/` |
| Discovery modes, lanes, handles, and response shaping | [Unified discovery](../reference/agent-query.md) | `internal/agentruntime/`, `internal/agentquery/`, `internal/evidence/` |
| Prepared source state and lexical retrieval | [Prepared index](prepared-index.md), [Indexing](../reference/indexing.md) | `internal/index/`, `internal/lexical/`, `internal/retrieve/` |
| Documentation retrieval and vectors | [Knowledge](../reference/knowledge.md), [Embedding model](../reference/embedding-model.md) | `internal/knowledge/`, `internal/knowledgevector/`, `internal/embedding/` |
| Lexicon facts and symbol evidence | [Lexicon reference](../reference/lexicon.md) | `internal/lexiconfacts/`, `internal/structure/`, `lexicon/` |
| Arcana graph evidence | [Arcana reference](../reference/arcana.md) | `internal/arcanagraph/`, `internal/structure/`, `arcana/` |
| Repository freshness and aligned snapshots | [Analysis stack](analysis-stack.md) | `internal/repostate/` |
| Investigation sessions and stable handles | [Agent MCP](../reference/agent-mcp.md) | `internal/investigation/` |
| Testing, evaluation, packaging, and release | [Testing and benchmarks](../development/testing-and-benchmarks.md), [Release workflow](../development/release-workflow.md) | `evaluation/`, `scripts/`, `.github/workflows/` |

## Component maintainer maps

- [Lexicon maintainer map](../../lexicon/docs/MAINTAINER_MAP.md)
- [Arcana maintainer map](../../arcana/docs/MAINTAINER_MAP.md)

## Boundaries

- Grimoire owns provider-neutral discovery orchestration, not language parsing or graph storage.
- Lexicon owns normalized language facts and immutable fact snapshots.
- Arcana owns graph compilation, packed graph state, and deterministic graph queries.
- Focused implementation paths and tests belong in each subject document's `## Code map` section.

## Related docs

- [Component architecture](components.md)
- [Analysis stack](analysis-stack.md)
- [Documentation coverage](../development/documentation-coverage.md)

## Notes

Use this map to choose the owning document first. Use that document's code map to locate implementation and tests.
