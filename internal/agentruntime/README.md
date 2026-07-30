# Agent-runtime package

`internal/agentruntime` owns the flattened execution boundary used by the CLI and MCP discovery interfaces.

## Responsibilities

- Normalize unified discovery requests.
- Prepare or inspect aligned Grimoire, Lexicon, Arcana, and document state through `internal/repostate`.
- Execute exact, source, symbol, trace, impact, and source-inspection operations through `internal/agentquery`; broad search defers relationship expansion.
- Execute the independent document lane through `internal/knowledge`.
- Supplement document BM25 with `internal/knowledgevector` only when explicitly requested and current.
- Keep source and documentation separate in the public response and prevent either lane from consuming the other's result limit.
- Bind optional investigation sessions to repository and provider snapshots.
- Replace repeated nodes, source ranges, graph paths, and document sections with compact prior handles.
- Keep narrow search deltas handle-only so source ranges are materialized by `inspect`, while preserving the evidence-rich balanced path.
- Apply serialized byte and evidence ceilings only as emergency safeguards after structural deduplication and progressive expansion.
- Return one `grimoire.discovery.v1` response rather than nesting provider-specific responses.
- Surface conservative evidence assessment and next-action guidance without claiming exhaustive correctness.

## Boundary

The runtime coordinates existing packages. It does not own source ranking formulas, graph traversal semantics, document indexing, vector persistence, or investigation storage. Each evidence lane retains its own ranking and limit even though the runtime returns them together.
