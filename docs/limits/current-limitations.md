# Current Limitations

## Status

Grimoire has a prepared lexical baseline, exact package budgeting, and an operational local embedding provider. It is not yet a complete hybrid RAG engine.

## Embeddings are not indexed yet

The Qwen3 model can be installed, served, queried, reduced to 512 dimensions, normalized, and probed. `grimoire index` does not yet embed chunks, and prepared snapshots do not persist vectors.

Planned removal condition: add model-versioned incremental vector records and atomic publication with the source snapshot.

## No vector retrieval

`grimoire context` does not perform cosine or approximate nearest-neighbour search. The current request path remains lexical-only.

Planned removal condition: retrieve semantic candidates from persisted normalized vectors, initially with the simplest measured implementation.

## Linear lexical search

A context request scans all prepared chunks and applies fixed substring boosts. It does not use postings, corpus statistics, or BM25.

Planned removal condition: maintain incremental postings and use BM25 while preserving inspectable metadata boosts.

## No hybrid fusion

Lexical and semantic result sets are not yet combined. There is no reciprocal-rank fusion or provider provenance in the context package.

Planned removal condition: independently retrieve bounded lexical and vector candidate sets, then fuse their ranks deterministically.

## Language-agnostic chunks

All supported files use the line-based fallback chunker. Function, method, type, Markdown-section, fenced-block, and structured-data boundaries are not preserved.

Planned removal condition: optionally consume Lexicon structural ranges while retaining fallback operation without Lexicon.

## Managed setup platform coverage

`grimoire model setup` currently installs the pinned `llama.cpp` runtime automatically only on Windows x64. Other platforms can use a manually installed runtime through `PATH` or `GRIMOIRE_LLAMA_SERVER`.

Planned removal condition: add verified pinned runtime assets for additional supported platforms.

## Whole-chunk selection only

The compiler greedily considers candidates in ranked order. It does not deduplicate overlapping evidence, diversify by subsystem, reserve evidence classes, or optimize globally.

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
- [Roadmap](../planning/roadmap.md)
- [System overview](../architecture/system-overview.md)
- [CLI](../reference/cli.md)
