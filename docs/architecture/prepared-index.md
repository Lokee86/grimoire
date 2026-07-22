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
manifest   format marker and prepared-index version
00         optional shard blob
01         optional shard blob
...
ff         optional shard blob
```

Only non-empty shards are present. A file path is assigned to one of 256 shards using the first byte of the SHA-256 digest of its repository-relative path.

## File records

Each shard stores path-keyed binary file records. A record contains:

- the source file's SHA-256 content hash;
- source file size;
- zero or more prepared chunks;
- each chunk's stable ID;
- source start and end lines;
- estimated token cost; and
- exact chunk text.

Paths are UTF-8, repository-relative, slash-separated, and validated to reject absolute paths, backslashes, NUL bytes, empty segments, `.` segments, and `..` segments.

## Binary formats

All current prepared-index formats are version 1.

| Record | Magic | Version field | Numeric byte order |
| --- | --- | --- | --- |
| Manifest | `GRIM` | 16-bit format version | Big-endian |
| Shard | `GRSH` | 8-bit shard version | Big-endian lengths and counts |
| File | `GRFL` | 8-bit file version | Big-endian lengths and metadata |

Shard paths are sorted before encoding. The root tree entries are also sorted. Equivalent logical snapshots therefore produce the same object identities.

These formats are internal and may change before a stable release. Readers reject unsupported versions, malformed lengths, invalid paths, duplicate paths, misplaced records, missing manifests, unexpected root entries, and trailing binary data.

## Incremental update behavior

Index construction loads the previous snapshot when one exists. Each eligible source file is hashed and compared with its prior file record.

- Matching content hash and size: reuse the previous record and its chunks.
- New or changed file: rebuild its chunks and mark its shard dirty.
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
- every other entry is a valid two-digit shard name;
- each shard decodes successfully;
- each file is stored in its expected shard;
- paths are valid and unique; and
- each file record and chunk range is well formed.

## Legacy cleanup

Successful saves remove the former `.grimoire/index.json` file when present. The active prepared state is object-backed; JSON is used only for command output and context packages.

## Code map

| File | Responsibility |
| --- | --- |
| `internal/index/repository.go` | Open/init repository, read current reference, compare bases, remove legacy JSON |
| `internal/index/store.go` | Load, validate, incrementally save, and publish snapshots |
| `internal/index/objects.go` | Git blob and tree construction plus manifest encoding |
| `internal/index/codec.go` | Shard names, shard binary format, and path validation |
| `internal/index/file_codec.go` | File/chunk binary format |
| `internal/index/model.go` | In-memory snapshot, file, and chunk models |

## Related documentation

- [System overview](system-overview.md)
- [Indexing reference](../reference/indexing.md)
- [Current limitations](../limits/current-limitations.md)
