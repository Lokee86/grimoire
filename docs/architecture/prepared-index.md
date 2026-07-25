# Prepared Index

## Purpose

Grimoire stores prepared retrieval state in a private bare Git object repository. The repository is an implementation detail used for content addressing, immutable object reuse, deterministic snapshot roots, and atomic publication. It is not the source repository's normal Git history.

## Default location

The default state path is:

```text
<repository-root>/.grimoire
```

`grimoire index --state <path>` and `grimoire context --state <path>` select another location. Relative state paths are resolved from the repository root. The active state path is excluded from indexing.

## Snapshot layout

The stable state reference is:

```text
refs/grimoire/state
```

It points directly to a Git tree object. The tree contains:

```text
manifest   format version and tokenizer identity
00         optional shard blob
01         optional shard blob
...
ff         optional shard blob
```

Only non-empty shards are present. A file path is assigned to one of 256 shards using the first byte of the SHA-256 digest of its repository-relative path.

## File records

Each shard stores path-keyed binary file records. A record contains:

- the source file's SHA-256 content hash;
- a SHA-256 preparation hash over the chunking contract and normalized Lexicon spans;
- source file size;
- zero or more prepared chunks;
- each chunk's stable ID;
- source start and end lines;
- exact `o200k_base` token count;
- exact chunk text; and
- optional semantic kind and symbol name for Lexicon-aligned chunks.

Paths are UTF-8, repository-relative, slash-separated, and validated to reject absolute paths, backslashes, NUL bytes, empty segments, `.` segments, and `..` segments.

## Binary formats

The current prepared-index format is version 4. Shard encoding remains version 1 because its path-to-record container did not change. File records are version 3 because they store preparation identity and semantic chunk metadata.

| Record | Magic | Current version | Numeric byte order |
| --- | --- | --- | --- |
| Manifest | `GRIM` | Prepared-index format `4` | Big-endian version and tokenizer-name length |
| Shard | `GRSH` | Shard format `1` | Big-endian lengths and counts |
| File | `GRFL` | File format `3` | Big-endian lengths and metadata |

The manifest records `o200k_base` as the tokenizer identity. Readers reject a different identity rather than interpreting its stored counts under the wrong tokenizer.

Shard paths are sorted before encoding. The root tree entries are also sorted. Equivalent logical snapshots therefore produce the same object identities.

These formats are internal and may change before a stable release. Readers reject unsupported versions, malformed lengths, invalid paths, duplicate paths, misplaced records, missing manifests, unexpected root entries, and trailing binary data.

## Incremental update behavior

Index construction loads the previous snapshot when one exists. Each eligible source file is hashed and compared with its prior file record.

- Matching content hash, size, and preparation hash: reuse the previous record, chunks, semantic metadata, and exact token counts.
- New or changed file, or changed Lexicon source boundaries: rebuild its semantic and fallback chunks, count them with `o200k_base`, and mark its shard dirty.
- Removed or newly ignored file: remove its record and mark its prior shard dirty.
- Unaffected shard: retain the previous shard object unchanged.

Only dirty shards are re-encoded and written.

## Publication

Publication uses compare-and-swap semantics on `refs/grimoire/state`:

1. Open or initialize the private bare repository.
2. Confirm the state reference still matches the snapshot used as the update base.
3. Write changed shard blobs and the manifest blob.
4. Write the deterministic root tree.
5. Atomically replace the state reference only if it has not changed.

If another writer publishes first, Grimoire returns `ErrConflict` rather than silently overwriting the newer snapshot.

If the resulting root tree already matches the current state, no new reference is published.

## Validation on load

Loading verifies:

- the private Git repository exists;
- the state reference exists;
- the reference resolves to a tree;
- every root entry is a regular file;
- exactly one valid manifest is present;
- the manifest declares prepared-index version 4 and tokenizer `o200k_base`;
- every other entry is a valid two-digit shard name;
- each shard decodes successfully;
- each file is stored in its expected shard;
- paths are valid and unique; and
- each file record and chunk range is well formed.

## Migration and legacy cleanup

Older prepared state does not contain the current preparation hashes and semantic chunk metadata. `grimoire index` recognizes an incompatible manifest, uses the current state root as its compare-and-swap base, rebuilds all eligible file records, and publishes a version-4 snapshot without deleting the state repository first. `grimoire context` requires a compatible index and reports the incompatibility until indexing is run.

Successful saves also remove the former `.grimoire/index.json` file when present. The active prepared state is object-backed; JSON is used only for command output and context packages.

## Code map

| File | Responsibility |
| --- | --- |
| `internal/index/repository.go` | Open/init repository, read current reference, compare bases, remove legacy JSON |
| `internal/index/store.go` | Load, validate, incrementally save, and publish snapshots |
| `internal/index/objects.go` | Git blob and tree construction plus manifest encoding |
| `internal/index/codec.go` | Shard names, shard binary format, and path validation |
| `internal/index/file_codec.go` | File/chunk binary format |
| `internal/index/semantic.go` | Lexicon span normalization, semantic boundaries, fallback gaps, and preparation hashes |
| `internal/index/model.go` | In-memory snapshot, file, and chunk models |

## Related documentation

- [System overview](system-overview.md)
- [Indexing reference](../reference/indexing.md)
- [Current limitations](../limits/current-limitations.md)
