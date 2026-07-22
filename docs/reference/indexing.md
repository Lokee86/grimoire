# Indexing Reference

## Scope

`grimoire index` traverses the selected repository root, applies structural and ignore exclusions, accepts supported regular text files, reuses unchanged records, chunks changed files, and publishes a prepared snapshot.

## Permanent exclusions

These directory names are always excluded at any traversal depth:

```text
.git
.grimoire
.ddocs
.arcana
.warlock
.worktrees
.workingtrees
```

The resolved `--state` path is also excluded, including custom paths outside the default `.grimoire` name.

Permanent exclusions cannot be re-included with an ignore negation because they protect repository metadata, generated tool state, and nested worktree containers.

## Git-ignore behavior

Without `--ignore-file`, Grimoire loads:

1. the root `.gitignore`; and
2. nested `.gitignore` files as their directories are entered.

Patterns use go-git's Git-ignore implementation and support standard pattern scope and `!` negation.

With `--ignore-file`, Grimoire uses only the selected Git-ignore-syntax file. It replaces the root and nested `.gitignore` hierarchy rather than layering a Grimoire-specific second ignore language. The configured control file itself is not indexed.

A missing explicitly configured ignore file is an error.

Grimoire does not automatically exclude dependency, generated, coverage, or build directories beyond the permanent tool-state list. Add those paths to repository ignore rules when they should not be indexed.

## Supported files

Files are selected by lowercase extension or recognized extensionless name.

### Source and script extensions

```text
.go .rs .py .rb .js .jsx .ts .tsx .java
.c .h .cc .cpp .hpp .cs .gd .sh .ps1
```

### Documentation, configuration, and data extensions

```text
.md .txt .toml .yaml .yml .json .xml
.html .css .scss .sql
```

### Recognized extensionless names

```text
README LICENSE Makefile Dockerfile Gemfile Rakefile
```

Name matching is case-insensitive.

## File eligibility

An eligible file must be:

- a regular file;
- supported by name or extension;
- no larger than the configured maximum; and
- text-like, defined currently as containing no NUL byte.

The default maximum is 2 MiB. `--max-file-bytes` replaces it when positive.

Symlinks and other non-regular directory entries are not indexed.

## Content identity and reuse

Grimoire computes a SHA-256 hash of each eligible file. A previous file record is reused only when both its content hash and byte size match.

Reused records retain their existing chunks and chunk IDs. Changed records are fully re-chunked by the current fallback chunker.

A prior record is removed when its path is deleted, becomes ignored, becomes unsupported, exceeds the size limit, becomes binary, or otherwise no longer appears in the eligible traversal result.

## Fallback chunking

The current chunker:

- normalizes CRLF line endings to LF;
- removes one final newline;
- skips empty or whitespace-only files;
- targets approximately 48 lines per chunk;
- prefers a recent blank-line boundary after at least eight useful lines;
- trims blank lines at chunk edges; and
- derives chunk identity from path, source range, and exact text.

It does not understand language syntax. Lexicon-provided structural chunking is planned.

## Token estimate

Each chunk stores:

```text
max(1, (byte_length + 2) / 3)
```

This is a deterministic heuristic used by the current budget fitter. It is not a model tokenizer and is not guaranteed to match or conservatively bound every model's token count.

## Statistics

The index command reports:

- `scanned`: eligible files evaluated after filtering;
- `reused`: scanned files with reused prior records;
- `updated`: scanned files rebuilt as new or changed; and
- `removed`: prior records absent from the resulting snapshot.

For a successful run:

```text
scanned = reused + updated
```

## Related documentation

- [CLI](cli.md)
- [Prepared index](../architecture/prepared-index.md)
- [Current limitations](../limits/current-limitations.md)
