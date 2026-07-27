# MCP server package

`internal/mcpserver` is the transport-only JSON-RPC 2.0 stdio server used by Grimoire.

It exposes one structured `grimoire_discover` tool. The owning application supplies the discovery schema, handler, description, and instructions.

## Responsibilities

- MCP initialization and capability negotiation.
- Tool listing and one tool-call dispatch path.
- Bounded message decoding.
- Structured-content and text-content responses.
- Protocol and handler error translation.

## Boundary

This package does not prepare repository state, interpret discovery requests, query source or documentation, or own provider routing. `internal/app` and `internal/agentruntime` own those responsibilities.
