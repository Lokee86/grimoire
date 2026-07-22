# Current Limitations

## Status

Grimoire is a pre-release lexical baseline. The limitations below describe current behavior, not hypothetical concerns.

## Language-agnostic chunks

Grimoire currently chunks all supported files with the same line-based fallback algorithm. It does not preserve function, method, type, Markdown section, fenced-block, or structured-data boundaries.

Planned removal condition: consume structural chunks from Lexicon while retaining fallback chunking when no adapter is available or parsing fails.

## Linear prepared-chunk search

A context request does not scan the repository, but it does scan all chunks loaded from prepared state and counts string occurrences for each query term.

Planned removal condition: prepare lexical postings during indexing and use an established scorer such as BM25 while preserving deterministic metadata boosts and inspectable reasons.

## Single tokenizer

Grimoire counts chunks and emitted packages only with `o200k_base`. The package budget is exact under that encoding, but a consumer model using another tokenizer may count the same JSON differently.

Grimoire also cannot account for system prompts, chat framing, tool schemas, or other wrapper content added after it emits the package.

Planned removal condition: add another tokenizer only when a concrete consumer demonstrates that the single shared approximation is insufficient. Wrapper overhead remains the consumer's responsibility.

## Whole-chunk selection only

The compiler greedily considers candidates in ranked order. It does not deduplicate overlapping evidence, diversify by file or subsystem, reserve budget for specific evidence classes, or optimize the package globally.

Planned removal condition: add measured selection improvements without hiding why each chunk was selected.

## Fixed lexical scoring

Current ranking uses fixed substring boosts rather than corpus statistics, stemming, fuzzy matching, symbol identity, or semantic similarity. Short one-character query terms are discarded.

Planned removal condition: prepared lexical scoring first; optional semantic retrieval only after its latency and quality can be measured against the lexical baseline.

## Narrow file eligibility

Only the built-in extension and extensionless-name allowlist is indexed. There is no configuration for adding file types, assigning language identities, or marking generated content beyond ignore rules.

Planned removal condition: add explicit indexing configuration when a concrete repository need establishes the contract.

## Full snapshot materialization

`index.Load` decodes all file records and chunks into memory. Context requests do not yet lazily read relevant shards or use a resident daemon-owned index.

Planned removal condition: measure real repositories before selecting lazy loading, memory mapping, daemon residency, or another storage access strategy.

## Manual index refresh

There is no file watcher or Grimoire daemon. Callers must run `grimoire index` after repository changes.

Planned removal condition: add standalone incremental maintenance, then allow the Warlock runtime to provide shared change events and supervision when installed.

## No optional evidence providers

The package currently records only `lexical` as a retrieval source. Lexicon, Arcana, Demon Docs, Git-change evidence, and semantic embeddings are not connected.

Planned removal condition: add providers behind explicit bounded interfaces without making the base lexical mode dependent on every tool.

## Pre-release compatibility

The CLI, prepared-index binary formats, scoring reasons, and context-package schema are versioned where needed but are not yet stable public compatibility promises.

Planned removal condition: declare compatibility policy at the first stable release.

## No stable diagnostic protocol

Errors are human-readable Go errors. There are no stable diagnostic codes, JSON error envelopes, or documented exit-code classes.

Planned removal condition: define a machine-readable diagnostic contract when Grimoire gains external integrations that require it.

## Related documentation

- [Roadmap](../planning/roadmap.md)
- [System overview](../architecture/system-overview.md)
- [CLI](../reference/cli.md)
