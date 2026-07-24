# Current limitations

These constraints apply to the merged system. They are not descriptions of future work.

## Retrieval quality remains corpus-bound

Grimoire has judged source and structural evaluation, but the primary corpora remain small relative to the variety of languages, repository layouts, and development tasks the product may encounter. A passing Grimoire or Gum corpus does not establish equivalent recall elsewhere.

Semantic, lexical, exact, structural, ranking, curation, assembly, and fitting stages can fail independently. Use per-case attribution rather than treating low final recall as one undifferentiated search problem.

## Automatic policy is deterministic heuristic policy

Query-shape analysis does not use an LLM or learned classifier. It uses prompt features plus retrieval confidence and dispersion. Focused, bounded, and exploratory tiers are concrete calibration choices, not proof of evidence sufficiency.

The public CLI supports automatic selection or one exact positive budget. It does not expose caller-supplied minimum/maximum ranges.

## Evidence coverage is not semantic proof

Automatic assembly measures represented regions, roles, candidate reserves, exact anchors, and provider evidence. It cannot prove that retained evidence answers the question. Poor ranking or an unsuitable target can still omit useful evidence during final fitting.

## Managed model setup is Windows x64 only

`grimoire model setup` installs pinned CPU, Vulkan, or CUDA `llama.cpp` artifacts only on Windows x64. Other platforms require a compatible runtime through `GRIMOIRE_LLAMA_SERVER` or `PATH`, plus a local model through `GRIMOIRE_EMBEDDING_MODEL` when managed state is unavailable.

Backend detection is capability-based and cannot guarantee that a detected GPU backend is fastest or most stable for every driver and device.

## The model service is external process state

`grimoire model serve` is blocking. Grimoire does not supervise it as a persistent daemon or automatically restart it. Source indexing remains available without the service, but vector builds and semantic queries require a live compatible endpoint.

## The Go native loader is Windows-only

The Rust vector engine is portable, but the production Go dynamic-library loader currently targets a Windows DLL. Non-Windows Go builds return `ErrUnavailable`; `context` can fall back to lexical retrieval, while direct vector commands cannot.

## Vector search is exact float32 scanning

Snapshot format version 1 stores aligned `float32` vectors and performs exact inner-product scanning. It does not use float16, int8, specialized quantized kernels, or approximate-nearest-neighbour indexes. Exact search is deterministic but may become material for very large corpora.

## Immutable vector objects are not garbage-collected

Deleted or replaced chunks disappear from the current manifest and snapshot, but immutable vector objects remain in the object store for possible reuse. There is no reachability-based cleanup across retained snapshots.

## Object ingestion is serialized

Embedding requests may execute concurrently, but completed batches enter the native object store through a serialized JSONL ingestion boundary. Increasing request concurrency cannot remove that persistence cost and can instead increase endpoint and memory pressure.

## Lexical fallback is linear

When semantic retrieval is unavailable, the fallback scans all prepared chunks and applies deterministic lexical scoring. It is a resilience path, not a full BM25 or postings-based search engine.

## Exact recovery scans prepared chunks

Concrete path, identifier, phrase, key, code, and version recovery is conditional, but an activated exact query still scans prepared paths and text. There is no persistent compact literal index.

## Source chunks are language-agnostic

All source files currently use the line-based fallback chunker. Lexicon symbols and spans enrich retrieval and package evidence but do not replace source chunk boundaries.

## Prepared snapshots are fully materialized

`index.Load` decodes the prepared snapshot into memory. There is no lazy shard reader or resident retrieval process for very large repositories.

## File eligibility is fixed

Supported extensions and extensionless names are compiled into Grimoire. There is no repository configuration for additional file classes or generated-content classification beyond ignore rules and explicit exclusions.

## State maintenance is explicit

Grimoire does not continuously watch repositories or automatically rebuild prepared and vector state. Callers must run `grimoire index` and `grimoire vector build` after relevant changes. Compatibility checks prevent silently using mismatched vector state.

## Structural components remain optional runtime dependencies

Lexicon and Arcana source now lives in this repository, but their executables, state formats, and publication lifecycles remain independently owned. Grimoire Context does not yet build, install, start, or maintain them automatically. Missing, stale, timed-out, or incompatible structural components produce warnings and preserve source retrieval, but structural evidence is incomplete.

Arcana queries use Lexicon matches as bounded graph seeds. A Lexicon miss can prevent otherwise relevant Arcana evidence from being requested. Current provider breadth is deliberately bounded rather than exhaustive.

## Selection and fitting remain whole-item heuristics

Curation removes duplicates and overlaps, promotes diversity, and adds bounded neighbours. Assembly preserves scope-specific reserves. The compiler fits complete structural facts and complete source chunks in deterministic order; it does not trim items or solve a global optimization problem.

## One output tokenizer

Context packages are measured with `o200k_base`. Consumers using another tokenizer may count the same JSON differently. Chat framing, tool schemas, and wrapper overhead remain the consumer's responsibility.

## Diagnostics are not a stable API

Errors are human-readable, but diagnostic codes, JSON error envelopes, and exit-code classes are not stable. Stderr wording is not a compatibility contract.

## Package compatibility is pre-release

The current context package version is 5. Consumers must reject unsupported versions rather than infer compatibility from field presence. CLI, prepared-state, vector-state, and package migration policy are not yet stable release promises.
