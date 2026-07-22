# Lexicon application

Lexicon keeps the most recently observed relevant repository state. It does not follow the source repository's commits, index, staging area, or branches.

## Commands

```text
lexicon init [--repo PATH] [--adapters PATH]
lexicon scan [--repo PATH]
lexicon daemon [--repo PATH] [--debounce 150ms] [--reconcile 30s]
```

`lexicon init` creates `.lexicon/repo`, performs a complete relevant-file scan, generates language libraries, creates the initial private state commit, and publishes an immutable analysis snapshot.

`lexicon scan` replaces the private source mirror with the repository's current relevant files. The private Git repository supplies the diff from the last successful scan. Lexicon regenerates only the language libraries affected by that diff, amends the private root commit, stores content-addressed per-file fact objects, and atomically advances `CURRENT` to the completed snapshot.

`lexicon daemon` watches the repository recursively. Changed paths are debounced and copied into the private mirror, then the same internal Git diff and language-library update path runs. A periodic complete reconciliation repairs missed filesystem events.

## Private state repository

```text
.lexicon/
    config.json
    CURRENT
    LOCK
    objects/
        ab/cdef...
    snapshots/
        <snapshot-id>.json
    repo/
        .git/
        source/
        library/
            go.jsonl
            python.jsonl
            ruby.jsonl
            gdscript.jsonl
            rust.jsonl
            typescript.jsonl
```

The state commit is always amended and remains a parentless root commit, so only one commit is reachable. Reflogs are expired after replacement. The repository is an implementation detail used to answer one question: what changed since Lexicon last successfully updated its library?

Each snapshot manifest records the internal state commit, adapter and schema versions, configuration identity, source-content identity, and fact-object identity for every relevant file. Shared synthetic facts are stored in a separate language object. Objects and snapshot manifests are immutable; identical content reuses the existing object.

`CURRENT` contains the complete snapshot ID and is replaced atomically only after the internal state commit and every referenced object are durable. Consumers should resolve `CURRENT`, load the corresponding manifest, and then open its objects. They never need to observe the mutable mirror or materialized JSONL libraries.

## Transaction and recovery

Only one process may update a repository at a time. Manual scans and daemon updates acquire the same advisory lock; a competing writer receives an explicit busy error.

Before each scan, Lexicon restores the materialized library from the last internal commit. A crash before commit therefore leaves no accepted library changes. A crash after the internal commit but before snapshot publication is repaired by the next scan, which reconstructs and republishes the snapshot without rerunning adapters when the source diff is empty.

## Incremental boundary

The current adapters analyze repositories rather than individual files. The private diff therefore selects affected languages, and only those language libraries are regenerated. Each selected adapter sees the complete current source mirror, preserving cross-file resolution. The resulting stream is partitioned into immutable per-file objects using contract ownership. Unchanged fact objects are reused, but a changed file still causes its complete language adapter to run. A later adapter contract can narrow this to changed-file extraction plus dependent relationship repair without changing the application lifecycle or snapshot format.

## Watch behavior

The daemon ignores Git metadata, Lexicon state, linked worktrees, dependency directories, and build outputs. New directories are watched recursively. Deletes and renames remove their mirrored paths. Watcher errors trigger an immediate full reconciliation, and the configured reconciliation interval provides an additional recovery path.
