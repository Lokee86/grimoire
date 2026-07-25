# MCP Server

`internal/mcpserver` owns Grimoire's agent transport boundary.

It exposes one structured `grimoire_query` tool over JSON-RPC 2.0 stdio. The
server accepts newline-delimited MCP messages and `Content-Length` framing,
preserves string or numeric request IDs, and returns both MCP text content and
`structuredContent`.

The package does not own repository analysis, ranking, investigation state, or
query semantics. Those are supplied through the `Handler` interface so the
agent query engine remains independently testable and usable from the CLI.
