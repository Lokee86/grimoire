# Application package

`internal/app` owns Grimoire's CLI surface and cross-package orchestration. It converts commands and flags into typed package calls, but it does not absorb the domain ownership of those packages.

## Commands

- `model setup`, `info`, `serve`, and `probe` — managed embedding runtime and endpoint operations.
- `index` — prepared source-state construction.
- `vector build`, `search`, and `info` — persistent semantic-state workflows.
- `context` — source, structural, policy, assembly, and package orchestration.
- `eval retrieval` — judged corpus execution and report publication.
- `version` — build identity.

## Context pipeline

The application layer:

1. resolves repository and state paths;
2. validates prepared and vector compatibility;
3. runs semantic retrieval or lexical fallback;
4. performs concrete exact recovery;
5. schedules Lexicon and Arcana work under bounded timeouts;
6. merges provider candidates;
7. asks `queryshape` for a profile and retrieval policy;
8. curates source candidates;
9. activates `assembly` only when no positive fixed budget was supplied; and
10. invokes `compiler` with source and structural evidence.

Provider failures become warnings when source retrieval can continue. Explicit invalid command options and required-state failures return errors.

## File map

- `root.go` — top-level command dispatch.
- `model.go` — runtime setup, discovery, serving, and endpoint probes.
- `indexing.go` — prepared-index command.
- `vector.go` and vector build files — native snapshot workflows and ingestion coordination.
- `context.go` — public context command and automatic-versus-fixed policy switch.
- `context_evaluation.go` — production pipeline execution for judged cases.
- `structural_context.go` — Lexicon/Arcana discovery, scheduling, and composition.
- `eval_retrieval.go` — corpus flags, run matrix, and report output.
- `version.go` — version response.

## Boundary

`internal/app` may coordinate packages and translate errors, but it must not become the home for ranking formulas, query classification, evidence coverage rules, graph semantics, token fitting, vector storage, or corpus scoring. Those belong to `retrieve`, `queryshape`, `assembly`, `structure`, `compiler`, `vectorstore`, and `evaluation` respectively.
