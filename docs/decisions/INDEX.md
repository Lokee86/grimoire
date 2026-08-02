# Architecture decisions

These ADRs record major architecture decisions already implemented by Grimoire. They summarize why the current boundaries exist and link to canonical architecture and reference documents for detailed current behavior. Proposed work belongs under [Planning](../planning/INDEX.md), not here.

## Current decisions

- [ADR 0001: Monorepo with independent components](0001-monorepo-independent-components.md) — co-locate Grimoire, Lexicon, and Arcana while preserving executable, ownership, state, test, and release boundaries.
- [ADR 0002: Progressive discovery instead of context packages](0002-progressive-discovery.md) — preserve independent evidence lanes and expand stable handles instead of preassembling token-fitted packages.
- [ADR 0003: Immutable generated state and atomic publication](0003-immutable-generated-state.md) — publish validated identity-bearing generations and reject stale or mismatched state.
- [ADR 0004: Process and protocol boundaries between components](0004-process-protocol-boundaries.md) — integrate component-owned engines through immutable exports, explicit protocols, and thin command forwarding.
- [ADR 0005: Lodestone owns native vector storage and exact search](0005-lodestone-vector-boundary.md) — keep native vector persistence/search separate from Grimoire ranking and embedding policy.

## Canonical current architecture

- [System overview](../architecture/system-overview.md)
- [Component architecture](../architecture/components.md)
- [Analysis stack](../architecture/analysis-stack.md)
- [Prepared index](../architecture/prepared-index.md)
- [Unified discovery contract](../reference/agent-query.md)
- [Vector store](../reference/vector-store.md)

An ADR is superseded only by a later explicit decision. Until then, current implementation and canonical architecture/reference documents define operational detail; an ADR defines the durable rationale and boundary.
