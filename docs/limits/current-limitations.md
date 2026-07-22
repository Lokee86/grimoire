# Current Limitations

## Status

Grimoire has incremental prepared state, exact package budgeting, an operational local embedding provider, persistent vector state, exact semantic search, vector-backed context compilation, and deterministic lexical fallback. It is not yet a complete context-selection and evidence-enrichment system.

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

## No targeted exact recovery

The normal context path relies on semantic ranking. It does not yet perform cheap exact lookup for paths, filenames, raw identifiers, quoted phrases, configuration keys, error codes, or version strings.

Planned removal condition: add compact conditional exact indexes and merge their candidates with semantic results while retaining provenance.

## Language-agnostic chunks

All supported files use the line-based fallback chunker. Function, method, type, Markdown-section, fenced-block, and structured-data boundaries are not preserved.

Planned removal condition: optionally consume Lexicon structural ranges while retaining fallback operation without Lexicon.

## Managed setup platform coverage

`grimoire model setup` currently installs the pinned `llama.cpp` runtime automatically only on Windows x64. Other platforms can use a manually installed runtime through `PATH` or `GRIMOIRE_LLAMA_SERVER`.

Planned removal condition: add verified pinned runtime assets for additional supported platforms.

## Whole-chunk selection only

The compiler greedily considers candidates in ranked order. It does not deduplicate overlapping evidence, expand useful neighbours, diversify by subsystem, reserve evidence classes, or optimize globally.

Planned removal condition: add measured selection improvements without obscuring why evidence was selected.

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
