# Current limitations

These constraints describe the active unified discovery system.

## Discovery quality remains corpus-bound

Current judged corpora cover only a fraction of possible languages, repository layouts, and development tasks. Passing Grimoire, Lexicon, Arcana, or agent-discovery cases does not establish equivalent recall elsewhere.

Measured benefit is task-shaped. Broad architectural investigations can improve substantially, while exact lookups and short named call chains may remain faster with direct shell inspection. Grimoire is an optional discovery aid inside a normal repository workflow, not a requirement for every query.

Exact source, BM25 source, document, symbol, and relationship lanes can fail independently. Evaluate them separately before attributing a missed investigation to the interface as a whole.

## Independent lanes are not a global answer ranking

Grimoire deliberately does not merge heterogeneous evidence into one score. The consumer must decide whether implementation, documentation, symbols, or graph relationships are most useful for the current task.

Per-lane rank is meaningful only inside that lane. Scores from different providers are not calibrated against one another.

## Source excerpts are bounded

Search results include compact source excerpts to reduce unnecessary inspection calls. Excerpts are capped and may omit relevant surrounding context. Full evidence still requires `inspect` on the returned handle.

## Documentation can be stale

Document results include freshness and provenance metadata where available, but relevance does not prove that a document still matches implementation. Current source remains the authority for executable behavior.

## Relationship discovery is intentionally local

The search relationship lane returns bounded direct relationships around discovered symbols. It is not an exhaustive transitive graph query. Use `trace` for paths and `impact` for bounded dependents.

When Arcana is unavailable, Lexicon relationship facts provide a narrower fallback. Unsupported or unresolved language constructs can still omit edges.

## Structural components remain optional runtime dependencies

Lexicon and Arcana retain independent executables, state formats, and publication lifecycles. Missing, stale, timed-out, or incompatible structural state produces warnings while exact, source, and document discovery continue when possible.

Grimoire prepares and aligns available state but does not supervise provider daemons.

## Semantic source boundaries depend on Lexicon coverage

Prepared source uses Lexicon declaration spans when current facts are available. Unsupported files, omitted constructs, stale state, ambiguous overlaps, and source outside declarations retain fallback line-window chunks.

Nested declarations are reduced to non-overlapping leaf spans. Oversized semantic declarations are split at the hard token ceiling.

## Prepared state is fully materialized

Prepared source snapshots and lexical postings are decoded into memory. There is no lazy shard reader or long-lived resident retrieval service for very large repositories.

## File eligibility is fixed

Supported extensions and extensionless names are compiled into Grimoire. Repositories can add ignore rules and explicit exclusions but cannot currently register new file classes or generated-content classifiers.

## Exact recovery has a scanning fallback

Identifier-aware postings localize most exact searches. Queries with no lexical token, such as punctuation-only literals, still fall back to scanning prepared chunks to preserve exact behavior.

## Managed model setup is Windows x64 only

`grimoire model setup` installs pinned CPU, Vulkan, or CUDA `llama.cpp` artifacts only on Windows x64. Other platforms require a compatible runtime and local model configured externally.

Backend detection cannot guarantee the fastest or most stable backend for every device and driver.

## The embedding service is external process state

`grimoire model serve` is blocking. Grimoire does not supervise or restart it as a daemon. Exact, source, symbol, and deterministic graph discovery remain available without embeddings. Document vectors and optional semantic graph entry points require a compatible live endpoint.

## The Go native vector loader is Windows-only

The Rust vector engine is portable, but the production Go dynamic-library loader currently targets a Windows DLL. On unsupported platforms the document lane remains BM25-only.

## Document vector search is exact float32 scanning

The current snapshot stores aligned `float32` vectors and performs exact inner-product search. It does not use quantized or approximate-nearest-neighbour indexes. This is deterministic but may become material for very large document corpora.

## Immutable vector objects are not garbage-collected

Replaced document sections disappear from current manifests and snapshots, but immutable vector objects remain available for reuse. There is no reachability-based cleanup across retained snapshots.

## Object ingestion is serialized

Embedding requests may overlap, but native object ingestion is serialized. Increasing endpoint concurrency can increase CPU, GPU, and memory pressure without removing persistence cost.

## State maintenance is request-driven

Grimoire does not continuously watch repositories. Discovery defaults to refresh-if-needed preparation. Documentation vectors remain explicit build artifacts, and freshness checks prevent silently using mismatched snapshots.

Initial preparation can dominate a first query because source, Lexicon, Arcana, and documentation state may all need alignment. Repeated `force-refresh` requests can erase the efficiency benefit of progressive discovery.

## Investigation sessions store evidence, not reasoning

Sessions deduplicate returned nodes, ranges, documents, relationships, and paths. They do not preserve an agent's private reasoning or guarantee that two semantically equivalent queries map to identical evidence handles.

## Release workflow is deliberately conservative

The root test, build, and release workflow defaults to one worker across Go and Cargo to prevent uncontrolled CPU fan-out. `--jobs N` is an explicit operator choice and can still overload a machine when set too high.

## Diagnostics and compatibility are pre-release

Human-readable errors, diagnostic codes, JSON error envelopes, exit-code classes, CLI spelling, and prepared/vector state migration policy are not yet stable release promises.

The current public discovery schema is `grimoire.discovery.v1`. Consumers must reject unsupported schemas rather than infer compatibility from field presence.
