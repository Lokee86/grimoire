# ADR 0005: Lodestone owns native vector storage and exact search

Parent index: [Architecture decisions](INDEX.md)

## Purpose

Record why Lodestone owns native vector persistence and exact search while Grimoire retains embedding, ranking, and product policy.

## Overview

The boundary isolates native storage and ABI concerns from discovery semantics and keeps vector capability optional for deterministic source discovery.

## Status

Accepted — this records the current vector boundary.

## Context

Optional semantic features need reusable vector objects and deterministic snapshot search without making embeddings part of required source discovery. Native storage, memory-mapped packed snapshots, and ABI safety are a distinct concern from document indexing, embedding production, ranking policy, and graph semantics.

Canonical current architecture: [Vector store](../reference/vector-store.md) and [System overview](../architecture/system-overview.md).

## Decision

Use Lodestone as the external native owner of immutable vector storage and exact vector search:

- Lodestone owns content-addressed vector objects, packed snapshots, validation, exact inner-product search, the stable C ABI, and the shared Go loader;
- Grimoire's `internal/vectorstore` is a compatibility facade over that boundary;
- `internal/knowledgevector` owns document-vector manifests, freshness, build orchestration, and ranking integration;
- `internal/embedding` owns embedding model/runtime configuration and requests;
- Arcana separately owns graph-neighbourhood vector indexes and their graph semantics;
- repository-wide source-code embeddings are not part of production source retrieval.

Vector features remain optional supplements, not authorities over exact, lexical, symbol, or graph evidence.

## Ownership and dependency consequences

- Lodestone does not own document sectioning, ranking, knowledge identities, embedding policy, or Arcana graph behavior.
- Grimoire does not duplicate Lodestone's native object and packed-snapshot implementation.
- Grimoire binds document vectors to the current knowledge-index and embedding identities before search.
- Arcana's vector state remains under `.arcana/`; Grimoire document-vector state remains under `.grimoire/knowledge/`.
- Callers use Grimoire or Arcana product seams rather than the native ABI directly.

## State, lifecycle, and failure consequences

- Successful batches persist immutable objects immediately; a complete manifest and packed snapshot publish only after all required vectors exist.
- Failed builds leave completed objects reusable but do not publish an incomplete current manifest.
- Snapshot handles are opaque and serialized per handle because the native handle is not re-entrant.
- Missing, stale, incompatible, or unavailable document vectors leave document retrieval on BM25 and emit a warning before query embedding.
- Vector failures never disable exact source, BM25 source, symbol, relationship, trace, or impact operations.

## Alternatives considered

- Reimplement native vector persistence and search in Grimoire. Rejected because it duplicates a dedicated storage owner and ABI-tested implementation.
- Let Lodestone own document manifests, freshness, embedding, and ranking. Rejected because those policies belong to Grimoire's knowledge lane.
- Require vectors for all source retrieval. Rejected because deterministic exact and BM25 source discovery must work without embeddings.
- Use approximate nearest-neighbour search by default. Rejected for the current corpus and contract in favor of deterministic exact search.

## Compatibility and migration impact

The ABI uses UTF-8 pointer/length inputs, borrowed foreign buffers, numeric handle registry keys, and error conversion across the native boundary. Snapshot acceptance requires matching knowledge, embedding, dimension, and count identities. The production Go dynamic loader is currently Windows-only; unsupported platforms retain BM25 document retrieval. ABI, loader, or packed-format changes require coordinated Lodestone and Grimoire compatibility updates.

## Verification

This decision is protected by `internal/vectorstore` integration tests, knowledge-vector build/resume/freshness/ranking tests, embedding tests, CLI vector tests, and no-vector discovery workflows. Lodestone internals are verified in their owning repository; Grimoire verifies the compatibility boundary and degradation behavior.

See [Knowledge retrieval](../reference/knowledge.md), [Embedding model](../reference/embedding-model.md), and the [Behavioral contract matrix](../development/behavioral-contract-matrix.md).

## Risks and debt

- The Go loader currently limits native document vectors to Windows.
- Exact float32 scanning may become expensive for very large document corpora.
- Immutable vector objects currently lack reachability-based garbage collection.
- Native ABI and executable/library discovery add platform and version-skew complexity.
- Ingestion persistence remains serialized even when embedding requests overlap.

## Superseding conditions

Supersede this ADR only if another storage/search owner or algorithm demonstrates a required capability while preserving optional degradation, deterministic identity checks, resumable publication, ABI safety, and the separation of storage from embedding and ranking policy. Repository-wide source embeddings require a separate explicit decision.

## Related docs

- [Vector store](../reference/vector-store.md)
- [Operations and trust boundaries](../architecture/operations-and-trust.md)
- [Component architecture](../architecture/components.md)

## Notes

The pinned local source checkout is a bounded development exception until Lodestone publishes a tagged consumable release.
