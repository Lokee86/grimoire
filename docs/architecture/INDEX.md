# Architecture

Architecture documentation describes implemented ownership, data flow, state transitions, and degradation behavior.

- [Component architecture](components.md) — monorepo layout, independent-use contract, dependency direction, state ownership, and release boundaries.
- [System overview](system-overview.md) — indexing, retrieval, query policy, assembly, and provider boundaries.
- [Prepared index](prepared-index.md) — immutable source identities, incremental rebuilds, and publication.

Related contracts:

- [Indexing](../reference/indexing.md)
- [Vector store](../reference/vector-store.md)
- [Query shape and assembly](../reference/query-shape-and-assembly.md)
- [Context package](../reference/context-package.md)
- [Lexicon contracts](../../lexicon/spec/README.md)
- [Arcana Lexicon contract](../../arcana/docs/LEXICON_CONTRACT.md)

Planned architecture changes belong under [Planning](../planning/INDEX.md), not here.
