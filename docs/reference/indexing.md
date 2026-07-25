# Indexing

Grimoire separates source preparation from vector construction. This keeps repository scanning, embedding availability, and vector publication independently diagnosable.

## Prepared source index

```bash
grimoire index --root <repository>
```

The indexer resolves the repository and state roots, applies traversal rules, optionally resolves the current immutable Lexicon export, normalizes eligible text into semantic declaration chunks plus fallback gaps, computes immutable identities and exact token counts, reuses unchanged objects, and atomically publishes a prepared snapshot. That snapshot remains usable for lexical and exact retrieval without an embedding service.

## Permanent exclusions

These directory names are excluded at every traversal depth:

```text
.git
.grimoire
.ddocs
.lexicon
.arcana
.warlock
.worktrees
.workingtrees
```

The resolved `--state` path is also excluded, including a custom path with another name. These exclusions protect repository metadata, generated tool state, and nested worktree containers and cannot be re-included with ignore negation.

## Git-ignore behavior

Without `--ignore-file`, Grimoire loads the root `.gitignore` and nested `.gitignore` files as their directories are entered. Patterns use go-git's Git-ignore implementation and preserve normal scope and `!` negation behavior.

`--ignore-file` replaces the root and nested hierarchy with one explicit Git-ignore-syntax file. The control file itself is excluded. A missing explicit ignore file is an error.

By default, Grimoire also omits common dependency and generated-output directories (`node_modules`, `vendor`, `third_party`, `third-party`, `.next`, `.nuxt`, `coverage`, `dist`, `target`, and `out`). It skips common lockfiles, generated-code filename suffixes and headers, minified or bundled web assets, and large web/data files that are effectively one minified line. Small authored assets remain eligible. Pass `--include-generated` to bypass this policy; Git-ignore and explicit exclusions still apply.

## Supported files

Source and script extensions:

```text
.go .rs .py .rb .js .jsx .ts .tsx .java
.c .h .cc .cpp .hpp .cs .gd .sh .ps1
```

Documentation, configuration, and data extensions:

```text
.md .txt .toml .yaml .yml .json .xml
.html .css .scss .sql
```

Recognized extensionless names, matched case-insensitively:

```text
README LICENSE Makefile Dockerfile Gemfile Rakefile
```

An eligible entry must be a regular supported file, no larger than the configured maximum, and text-like. The current text check rejects files containing a NUL byte. Symlinks and other non-regular entries are not indexed. The default maximum is 2 MiB; a positive `--max-file-bytes` replaces it.

## Incremental identity and reuse

Grimoire computes SHA-256 over each eligible file and a separate preparation hash over the chunking contract plus that file's normalized Lexicon source spans. A prior file record is reused only when content hash, byte size, and preparation hash match. Reused records retain their existing chunks, semantic metadata, IDs, and token counts. New or changed files, or files whose Lexicon boundaries changed, are fully re-chunked.

A prior record is removed when its path is deleted, ignored, unsupported, oversized, binary, or otherwise absent from the eligible traversal result. Renames naturally reuse immutable content where the storage identity permits it while publishing the new path record.

Changing traversal, chunking, tokenizer, or schema behavior invalidates the relevant identity and forces affected work to be rebuilt. Semantic chunk metadata and preparation hashes use prepared-index format version 4, so older prepared state is rebuilt once on the next `grimoire index` run.

## Lexicon-aligned semantic chunking

When `<root>/.lexicon/CURRENT` exists, `grimoire index` resolves or creates the same immutable JSONL export used by structural retrieval. Declaration-level Lexicon spans become source boundaries for functions, methods, tests, and types. Nested closures and other callable-owned declarations do not fragment their owning declaration; smaller data-flow symbols remain inside the surrounding source chunk.

Nested declarations are represented by their narrowest non-overlapping leaves. For example, methods are prepared independently instead of duplicating the complete containing type. Source outside those semantic ranges remains indexed through the fallback chunker, so comments, package declarations, configuration, unsupported constructs, and files without Lexicon facts are not discarded.

Every semantic chunk stores the Lexicon kind and symbol name. A declaration above 1,536 tokens is split only for the hard token ceiling while retaining that semantic identity on each part.

`--lexicon-facts` selects an explicit JSONL export. `--lexicon-state` and `--lexicon-command` override automatic immutable-state discovery. An unreadable explicit facts export is an error; unavailable automatically discovered state warns and preserves fallback-only indexing.

## Fallback chunking

The language-agnostic fallback chunker:

- normalizes CRLF to LF;
- removes one final newline;
- skips empty or whitespace-only files;
- targets roughly 48 lines per chunk;
- prefers a recent blank-line boundary after at least eight useful lines;
- trims blank lines at chunk edges;
- enforces an exact 1,536-token ceiling after line-based chunking;
- recursively splits oversized chunks at line boundaries;
- falls back to token slices only when one source line alone exceeds the ceiling; and
- derives chunk identity from path, source range, exact text, and token-slice position when required.

Fallback boundaries are used for files with no valid Lexicon spans and for every uncovered region around semantic declarations.

## Token accounting

Changed chunks are counted with the embedded `o200k_base` tokenizer and store the exact count in prepared state. The manifest records tokenizer identity so counts cannot be reused under a different tokenizer.

Chunk counts cover chunk text only. Context compilation separately counts the complete serialized package, including paths, reasons, metadata, escaping, and formatting.

## Index statistics

The command reports:

- `scanned`: eligible files evaluated after filtering;
- `reused`: scanned files using prior records;
- `updated`: new or changed scanned files rebuilt; and
- `removed`: prior records absent from the new snapshot;
- `generated_skipped`: generated files or generated-directory roots omitted by the default policy;
- `semantic_files`: prepared files containing at least one Lexicon-aligned chunk; and
- `semantic_chunks`: chunks carrying a Lexicon declaration kind and name.

For a successful run:

```text
scanned = reused + updated
```

## Vector construction

Start the local model service, then run:

```bash
grimoire vector build --root <repository>
```

The builder validates prepared state, returns immediately when the current vector manifest already matches, deduplicates identical chunk text, reuses source identities recorded by the previous manifest, checks only newly introduced source hashes, embeds genuinely missing text in bounded concurrent request batches, ingests completed batches serially into the immutable native object store, writes the complete chunk-to-source manifest, and materializes a sorted packed snapshot.

The defaults are four documents per embedding request and one active request. Increase `--batch-concurrency` for a provider that benefits from independent requests. Object ingestion remains serialized, while content addresses and sorted materialization make publication deterministic regardless of embedding completion order.

The first embedding or ingestion error cancels outstanding request work and prevents publication of a new manifest. Immutable objects already written remain reusable by later builds.

## State compatibility

Query commands verify prepared snapshot identity, embedding identity, dimensions, and vector count. Missing, stale, or incompatible vector state causes `context` to warn and use lexical fallback. `vector search` requires valid semantic state and returns an error instead.

Run `grimoire index` after relevant source or indexing-rule changes and `grimoire vector build` after the prepared identity or embedding contract changes. Use `grimoire vector info` to inspect snapshot availability.

The `.grimoire/` directory is generated state and must not be treated as authored repository content.
