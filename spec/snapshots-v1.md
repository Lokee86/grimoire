# Lexicon snapshot contract v1

Lexicon snapshots expose one complete, immutable analysis state. The mutable source mirror and materialized language JSONL files are implementation details and are not consumer consistency boundaries.

## Layout

```text
.lexicon/
    CURRENT
    LOCK
    objects/<first-two-hex>/<remaining-hex>
    snapshots/<64-hex>.json
    repo/
```

`CURRENT` contains one `sha256:<64-hex>` snapshot ID followed by a newline.

## Fact objects

A fact object contains all records owned by one source file, or the shared synthetic records for one language:

```json
{
  "version": 1,
  "language": "python",
  "owner": "src/example.py",
  "source_content_id": "sha256:...",
  "adapter_version": "0.1.0",
  "schema_version": 1,
  "analysis_config_id": "sha256:...",
  "records": []
}
```

`owner` and `source_content_id` are omitted for a shared language object. Record ownership follows `facts-v1.md`: explicit `owner`, span path, file-node path, then the owning source node for edge and unresolved records.

The object ID is SHA-256 over:

```text
lexicon:fact-object:v1\0<canonical object JSON>
```

The object file path omits the `sha256:` prefix and splits the first two hexadecimal characters into a directory. Existing objects are immutable. Writing different bytes under an existing ID is an error.

## Snapshot manifest

```json
{
  "version": 1,
  "state_commit": "<private Git commit>",
  "languages": [
    {
      "language": "python",
      "adapter_version": "0.1.0",
      "schema_version": 1,
      "repository": "example",
      "analysis_config_id": "sha256:...",
      "shared_object_id": "sha256:...",
      "files": [
        {
          "path": "src/example.py",
          "language": "python",
          "content_id": "sha256:...",
          "object_id": "sha256:..."
        }
      ]
    }
  ]
}
```

Languages and files are sorted lexicographically. `shared_object_id` is omitted when the adapter emitted no unowned records.

The snapshot ID is SHA-256 over:

```text
lexicon:snapshot:v1\0<canonical manifest JSON>
```

The manifest filename omits the `sha256:` prefix. Readers must verify the manifest bytes against the requested snapshot ID before accepting it.

## Publication

A successful update follows this order:

1. acquire the repository update lock;
2. replace the relevant private source mirror;
3. calculate the private Git diff;
4. compute the impacted file closure or select the complete-language fallback;
5. request and validate full or incremental language streams;
6. merge incremental records into complete materialized language files and atomically replace them;
7. amend the single private state commit;
8. write all missing immutable fact objects;
9. write the immutable snapshot manifest;
10. atomically replace `CURRENT`;
11. release the update lock.

Consumers resolve `CURRENT` once and then read only the referenced immutable manifest and objects. They therefore observe either the previous complete snapshot or the new complete snapshot, never a partially published state.

## Recovery

Before a scan, uncommitted materialized language output is restored from the private state commit. This rolls back a process that failed before step 6.

If the private commit completed but `CURRENT` was not published, a no-change scan rebuilds the manifest from the committed language streams and atomically republishes it without rerunning adapters.

## Incremental analysis

Ordinary source modifications update only the changed files and their transitive dependents. The dependency closure is calculated from cross-file relationships in the previous snapshot; owners with unresolved relationships are included conservatively. The adapter executes against a temporary repository containing that emission set, its transitive forward dependencies, and required language configuration. Go expands scopes to packages and Rust expands scopes to crates.

A directly edited file with prior cross-file or unresolved relationships selects complete-language analysis before a scope is built. Scoped streams are reserved for leaf and local-only direct edits; they contain selected file-owned records and declare their shared synthetic set partial, so previous complete shared records remain authoritative. Before merge, new edge or unresolved topology causes a complete-language retry. A scoped adapter failure also retries the complete language repository. Additions, deletions, renames, copies, configuration changes, missing dependency state, and corrupt libraries use the same full fallback. More precise structural invalidation can be added without changing this snapshot contract.
