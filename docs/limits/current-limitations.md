# Current Limitations

## Status

Grimoire has incremental prepared state, exact package budgeting, an operational local embedding provider, persistent vector state, exact semantic search, vector-backed context compilation, deterministic lexical fallback, immutable Lexicon symbol evidence, and bounded Arcana graph evidence. It is not yet a complete context-selection and evidence-enrichment system.

## Manual vector refresh

Vector state is refreshed explicitly with `grimoire vector build`; source indexing does not automatically invoke model inference. Repeated builds reuse unchanged vector objects.

Planned removal condition: add automatic maintenance while retaining explicit one-shot commands.

## Windows-only Go native loader

The Rust engine is portable, but the current Go dynamic-library bridge loads a DLL only on Windows. Other platform builds return an unavailable error and `context` uses the lexical fallback.

Planned removal condition: add equivalent loaders and release packaging for supported non-Windows targets.

## Float32 exact scan only

Snapshot format version 1 stores aligned `float32` vectors and performs exact dot-product scanning. It does not yet use float16, int8, narrower quantization, specialized SIMD kernels, or approximate-nearest-neighbour indexes.

Planned removal condition: benchmark alternative encodings and kernels, adding complexity only when speed and retrieval-quality evidence justify it.

## Immutable vector objects are not garbage-collected

Deleted or replaced chunks disappear from the current snapshot, but reusable immutable objects remain in the object store.

Planned removal condition: add safe reachability-based cleanup across retained snapshots.

## Lexical fallback is linear

When semantic retrieval is unavailable, the fallback scans all prepared chunks and applies fixed substring boosts. It does not use postings, corpus statistics, or BM25.

This is intentionally a failure path rather than the normal retrieval path. A larger lexical engine should only be added if measured fallback use or retrieval failures justify its runtime and maintenance cost.

## Targeted exact recovery still scans prepared chunks

Exact recovery is conditional and skipped for ordinary natural-language queries, but an activated query currently scans prepared chunk paths and text. It does not yet use a persistent compact identifier/path index.

Planned removal condition: add a compact index only when repository-scale benchmarks show the conditional scan is material.

## Language-agnostic source chunks

All supported files still use the line-based fallback chunker. Lexicon symbols and source spans are now emitted as separate structural evidence and can steer source retrieval, but Grimoire does not yet use those ranges to replace fallback chunk boundaries.

Planned removal condition: optionally prepare Lexicon-aligned source chunks while retaining fallback operation without Lexicon.

## Structural evidence depends on local tool executables

Automatic enrichment discovers repository state through `.lexicon/CURRENT` and `.arcana/CURRENT`, but creating a missing cached export requires the `lexicon` executable and catching up a missing or stale graph requires the `arcana` executable. Explicit executable and state paths are supported. Provider failure emits a warning and preserves standalone source retrieval.

Planned removal condition: add Warlock-supervised discovery or a stable shared invocation registry without coupling Grimoire to either implementation.

## Structural retrieval policy is intentionally bounded

Lexicon matching currently selects direct query-matched symbols and immediate relationships. Arcana queries operational roles, bounded impact, unresolved references, and shortest call chains among a small seed set. Grimoire does not yet infer every graph operation that may be useful for arbitrary tasks, and the judged retrieval evaluator does not yet score structural evidence.

Planned removal condition: add task-shaped structural query planning and judged structural-evidence cases before expanding query breadth.

## Managed setup platform coverage

`grimoire model setup` currently installs the pinned `llama.cpp` runtime automatically only on Windows x64. Other platforms can use a manually installed runtime through `PATH` or `GRIMOIRE_LLAMA_SERVER`.

Planned removal condition: add verified pinned runtime assets for additional supported platforms.

## Selection remains heuristic and whole-item based

Candidate curation removes duplicates and overlapping ranges, promotes file/subsystem diversity, and adds bounded prepared neighbours. The compiler fits one leading structural fact and one leading source selection before the remaining structural and source evidence, but it still uses deterministic greedy fitting rather than global optimization and never trims an individual fact or chunk.

Planned removal condition: add stronger evidence-class allocation only when deterministic quality fixtures demonstrate a concrete failure.

## Single output tokenizer

Context packages are measured only with `o200k_base`. A consumer using another tokenizer may count the same package differently. Grimoire also cannot count chat framing or tool schemas added later.

Planned removal condition: add another tokenizer only for a concrete consumer. External wrapper overhead remains the consumer's responsibility.

## Narrow file eligibility

The built-in extension and extensionless-name allowlist is fixed. There is no configuration for additional file types or generated-content classification beyond ignore rules.

## Full snapshot materialization

`index.Load` decodes all source records and chunks into memory. There is no lazy shard access or resident retrieval process.

## Manual index refresh

There is no watcher or Grimoire daemon. Callers must run `grimoire index` after repository changes.

## Pre-release compatibility

CLI behavior, prepared binary formats, model identity, ranking reasons, and context-package schemas are versioned where needed but are not stable public promises.

## No stable diagnostic protocol

Errors are human-readable Go errors. Stable diagnostic codes, JSON error envelopes, and documented exit-code classes do not yet exist.

## Related documentation

- [Embedding model](../reference/embedding-model.md)
- [Vector store](../reference/vector-store.md)
- [Roadmap](../planning/roadmap.md)
- [System overview](../architecture/system-overview.md)
- [CLI](../reference/cli.md)
