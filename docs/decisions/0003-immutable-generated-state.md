# ADR 0003: Immutable generated state and atomic publication

Parent index: [Architecture decisions](INDEX.md)

## Purpose

Record why generated repository state is immutable, identity-bearing, and atomically published by its owning component.

## Overview

Generation identities and atomic publication prevent partial state, stale-handle reuse, and ambiguous authority across Grimoire, Lexicon, Arcana, and vector storage.

## Status

Accepted — this records the current state model.

## Context

Grimoire, Lexicon, Arcana, and optional vector facilities derive reusable state from repository content and configuration. Queries and cross-component consumers need deterministic identities, safe concurrent publication, explicit freshness, and protection from partially written generations.

Canonical current architecture: [Prepared index](../architecture/prepared-index.md) and [Analysis stack](../architecture/analysis-stack.md).

## Decision

Treat generated analysis and retrieval state as immutable, identity-bearing generations:

- content-address reusable objects where the owning format supports them;
- publish complete snapshots through an atomic current reference or manifest boundary;
- bind consumers and handles to explicit source, Lexicon, Arcana, knowledge, and vector identities;
- validate complete state before accepting it;
- rebuild or publish a new generation instead of mutating a visible generation in place;
- retain separate owner-specific state directories and format versions.

Incremental work may reuse immutable objects or publish Arcana overlays, but the visible generation remains a validated snapshot.

## Ownership and dependency consequences

- Grimoire owns prepared source and document state under `.grimoire/`.
- Lexicon owns immutable facts, objects, manifests, and `.lexicon/CURRENT`.
- Arcana owns graph bases, overlays, manifests, and `.arcana/CURRENT`, bound to the consumed Lexicon snapshot.
- Lodestone owns immutable vector objects and packed exact-search snapshots; Grimoire owns document-vector manifests and freshness.
- Consumers may read published state but may not mutate another owner's objects or current reference.

## State, lifecycle, and failure consequences

- Writers serialize or use compare-and-swap publication so a completed newer generation is not silently overwritten.
- Current references advance only after required objects and manifests are durable and validated.
- Failed builds may leave reusable unreachable objects, but must not expose incomplete current state.
- Stale or mismatched handles and provider identities produce explicit errors or warnings.
- Unsupported prepared-state versions rebuild according to state mode; they are not interpreted optimistically.
- Arcana overlays validate their base identity, and compaction publishes a new equivalent base without mutating the source snapshot.

## Alternatives considered

- Update one mutable database in place. Rejected because partial writes and concurrent readers blur generation identity.
- Use timestamps alone for freshness. Rejected because timestamps do not establish content or configuration equivalence.
- Let Grimoire rewrite Lexicon or Arcana state. Rejected because it violates component ownership and format boundaries.
- Silently resolve stale handles against current content. Rejected because it can return evidence from the wrong generation.

## Compatibility and migration impact

Generated formats are independently versioned and may require rebuilds before a stable release. Prepared-index incompatibility triggers a current-format rebuild rather than destructive repository reset. Legacy readers or consumers must reject unsupported schemas and identities. Stable source and structural handles remain valid only while their qualifying snapshots remain active and aligned.

## Verification

This decision is protected by prepared-index codec, validation, compare-and-swap, incremental-reuse, and exclusion tests; Lexicon object, transaction, lock, and publication tests; Arcana ingestion, manifest, overlay, compaction, and publication-failure tests; and vector build/resume/freshness tests.

See the [Behavioral contract matrix](../development/behavioral-contract-matrix.md).

## Risks and debt

- Immutable objects can accumulate; documentation-vector objects currently lack reachability-based garbage collection.
- Fully materialized prepared state can consume substantial memory on large repositories.
- Multiple independently versioned formats increase compatibility and diagnostic work.
- Initial preparation and forced refreshes can be expensive.

## Superseding conditions

Supersede this ADR only if a replacement state model provides equivalent generation identity, atomic visibility, deterministic reuse, owner isolation, stale-handle safety, and recoverable failure behavior. Any mutable-state replacement must define concurrency, rollback, migration, and cross-component alignment explicitly.

## Related docs

- [Prepared index](../architecture/prepared-index.md)
- [Analysis stack](../architecture/analysis-stack.md)
- [Operations and trust boundaries](../architecture/operations-and-trust.md)

## Notes

Private generated roots remain reconstructable state, not user-authored authorities.
