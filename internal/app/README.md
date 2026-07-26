# Application package

`internal/app` owns Grimoire's CLI surface and cross-package orchestration. It converts commands and flags into typed package calls without absorbing domain ownership.

## Commands

- `model setup`, `info`, `serve`, and lifecycle commands — managed embedding runtime operations.
- `index` — prepared source-state construction.
- `knowledge index|search|inspect|vector` — independent documentation and rationale retrieval.
- `vector build|info` — aliases for documentation-vector workflows.
- `query` and `mcp` — progressive code and knowledge retrieval.
- `context` — deterministic source, structural, policy, assembly, and package orchestration.
- `eval retrieval` — judged source-corpus execution and report publication.
- `investigation create|status|close` — persistent discovery-ledger lifecycle.
- `version` — build identity.

## Context pipeline

The production context path:

1. resolves repository and prepared source state;
2. performs deterministic BM25 source retrieval;
3. performs concrete exact recovery;
4. schedules Lexicon and Arcana work under bounded timeouts;
5. merges exact, lexical, and structural source candidates;
6. asks `queryshape` for a profile and retrieval policy;
7. curates source candidates;
8. activates `assembly` only when no positive fixed budget was supplied; and
9. invokes `compiler` with source and structural evidence.

Repository-wide source vectors are not part of this path. Provider failures become warnings when deterministic source retrieval can continue.

## Knowledge pipeline

`knowledge search` always uses BM25. When a current documentation vector snapshot exists, `internal/knowledgevector` supplies supplemental scores through the existing `knowledge.VectorRanker` seam. Missing, stale, or unavailable vectors leave BM25 results intact and expose a vector warning in the response.

## File map

- `run.go` — top-level dispatch and shared source-index command.
- `model*.go` — runtime setup, discovery, serving, and endpoint probes.
- `knowledge.go` and `knowledge_vectors.go` — knowledge CLI and documentation vectors.
- `context.go` — public deterministic context command.
- `context_evaluation.go` and `context_semantic.go` — judged experimental retrieval modes, including historical source-vector comparisons.
- `context_structure.go` — Lexicon/Arcana discovery, scheduling, and composition.
- `eval_retrieval.go` — corpus flags, run matrix, and report output.
- `investigation.go` — investigation-ledger lifecycle commands.

## Boundary

`internal/app` may coordinate packages and translate errors, but ranking formulas, query classification, evidence coverage, graph semantics, token fitting, vector storage, knowledge-vector identity, and corpus scoring belong to `retrieve`, `queryshape`, `assembly`, `structure`, `compiler`, `vectorstore`, `knowledgevector`, and `evaluation` respectively.
