# Agent-runtime package

`internal/agentruntime` owns the unified execution boundary used by the CLI and MCP agent interface.

## Responsibilities

- Validate and normalize progressive agent requests.
- Prepare or inspect deterministic repository state through `internal/repostate` according to the requested state mode.
- Execute code queries through `internal/agentquery`.
- Execute the independent documentation lane through `internal/knowledge`.
- Supplement knowledge BM25 with `internal/knowledgevector` only when a documentation snapshot is present and current.
- Bind persistent investigation sessions to repository and provider snapshots.
- Replace repeated nodes, source ranges, graph paths, and knowledge sections with compact prior handles.
- Return code evidence, knowledge evidence, warnings, state information, and investigation deltas through one response contract.

## Boundary

The runtime coordinates existing packages; it does not own source ranking formulas, graph traversal semantics, documentation indexing, vector persistence, or investigation storage. Code and knowledge remain separate retrieval lanes even when returned by one agent call.
