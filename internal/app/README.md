# Application package

`internal/app` owns Grimoire's CLI and MCP surfaces plus cross-package orchestration. It converts commands and flags into typed calls without absorbing provider domain ownership.

## Commands

- `orient`, `search`, `trace`, `impact`, and `inspect` — the unified discovery interface.
- `query` — compatibility entry point for the same discovery modes.
- `mcp` — stdio server exposing one `grimoire_discover` tool.
- `status` — repository and provider state inspection or refresh.
- `index` — prepared source-state construction.
- `knowledge index|search|inspect` — standalone document-lane diagnostics.
- `vector build|info` — optional document-vector workflows.
- `model setup|info|serve|start|stop|probe` — managed embedding runtime operations.
- `investigation create|status|close` — persistent discovery-ledger lifecycle.
- `eval knowledge` and `eval arcana` — component evaluation.
- `version` — build identity.

The former `context` command is retired.

## Discovery flow

1. Normalize the repository and requested state mode.
2. Align Grimoire source state with available Lexicon and Arcana snapshots.
3. Execute provider-neutral source and structural discovery through `internal/agentquery`.
4. Execute the independent document lane through `internal/knowledge`.
5. Return exact, source, document, symbol, and relationship evidence as separate lanes.
6. Expand returned handles through inspect, trace, or impact.
7. Optionally record evidence in one investigation session.

No package compiler or token-fitting stage sits between discovered evidence and the agent.

## State preparation

`discovery_prepare.go` runs Grimoire-owned source and document preparation in process while preserving Lexicon and Arcana as explicit executable boundaries. Provider failures become warnings when other discovery lanes can continue.

## Document lane

Document BM25 is deterministic and independent. Optional vectors supplement only document ranking. Missing or stale vectors preserve BM25 results and produce a warning.

## File map

- `run.go` — top-level dispatch and help.
- `query.go` — direct discovery CLI parsing.
- `mcp.go` — unified MCP schema and server.
- `discovery_prepare.go` — in-process Grimoire preparation plus external provider execution.
- `index.go` and related files — source preparation.
- `knowledge.go` and `knowledge_vectors.go` — document diagnostics and optional vectors.
- `model*.go` — embedding runtime setup and lifecycle.
- `investigation.go` — investigation-ledger lifecycle.
- `eval_knowledge.go` and `eval_arcana.go` — component evaluation.

The retired context command and its package assembly, compiler, curation, query-shape, diff-context, graph-ranking, and source-evaluation implementations have been removed.

## Boundary

`internal/app` coordinates packages and translates errors. Source ranking belongs to `retrieve`; document ranking to `knowledge`; language facts to Lexicon; graph semantics to Arcana; vector storage to Lodestone; and investigation persistence to `investigation`.
