# Architecture

Architecture documentation describes implemented ownership, data flow, state transitions, and degradation behavior.

- [Component architecture](components.md) — monorepo layout, independent-use contract, dependency direction, state ownership, and release boundaries.
- [Analysis stack](analysis-stack.md) — the implemented Lexicon publication, Arcana synchronization, Grimoire preparation, snapshot alignment, and degradation lifecycle.
- [Grimoire codemap](codemap.md) — file-level ownership for CLI/MCP, indexing, retrieval, documents, providers, sessions, and evaluation.
- [System overview](system-overview.md) — unified discovery lanes, provider routing, progressive expansion, and fallback boundaries.
- [Prepared index](prepared-index.md) — immutable source identities, incremental rebuilds, and publication.

Related contracts:

- [Unified discovery contract](../reference/agent-query.md)
- [Grimoire MCP interface](../reference/agent-mcp.md)
- [Indexing](../reference/indexing.md)
- [Vector store](../reference/vector-store.md)
- [Lexicon contracts](../../lexicon/spec/README.md)
- [Arcana Lexicon contract](../../arcana/docs/LEXICON_CONTRACT.md)

The former context-package architecture is retired from the product interface. Historical evaluation artifacts remain historical evidence, not active architecture.

Planned architecture changes belong under [Planning](../planning/INDEX.md), not here.
