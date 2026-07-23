# Lexicon application

Lexicon keeps the most recently observed relevant repository state. It does not follow the source repository's commits, index, staging area, or branches.

## Post-publication consumers

Lexicon can invoke deterministic one-shot consumers after a successful scan has published or confirmed the current immutable snapshot. Consumer definitions live under `.lexicon/consumers/*.json`:

```json
{
  "version": 1,
  "command": "/absolute/path/to/arcana",
  "args": ["sync", "--lexicon", "/repo/.lexicon", "--state", "/repo/.arcana"],
  "timeout": "30s"
}
```

Commands execute directly without a shell, in lexical filename order, with the repository as their working directory. `timeout` is optional; existing definitions without it remain unlimited. Lexicon provides `LEXICON_REPOSITORY`, `LEXICON_STATE_ROOT`, and `LEXICON_SNAPSHOT_ID`. Lexicon attempts every registered consumer, aggregates failures, and retries failed consumers on a later scan. After a successful invocation, its state file contains deterministic JSON such as `{"version":1,"snapshot_id":"sha256:..."}` and is replaced atomically; failed invocations leave their previous state unchanged. The already-published Lexicon snapshot remains valid.

The `lexicon consumer` commands and `internal/consumer` package expose list, add, remove, and one-shot execution operations. Definition names are simple `.json` filenames; path traversal and other extensions are rejected. Listing is lexical, adding replaces an existing definition atomically, removal deletes both the definition and its consumer state, and one-shot execution validates the requested immutable snapshot before invoking the selected consumer.

## Commands

```text
lexicon init [--repo PATH] [--adapters PATH] [--languages all|LIST]
lexicon scan [--repo PATH]
lexicon demon [--repo PATH] [--debounce 150ms] [--reconcile 30s]
lexicon rebuild [--repo PATH] [--languages LIST]
lexicon languages [list] [--repo PATH]
lexicon languages set [--repo PATH] --languages all|LIST
lexicon status [--repo PATH]
lexicon doctor [--repo PATH]
lexicon export [--repo PATH] --output PATH [--snapshot ID] [--languages LIST]
lexicon gc [--repo PATH] [--retain 20] [--dry-run]
lexicon consumer list [--repo PATH]
lexicon consumer add [--repo PATH] --name NAME --command PATH [--arg VALUE]... [--timeout DURATION]
lexicon consumer remove [--repo PATH] --name NAME
lexicon consumer run [--repo PATH] --name NAME [--snapshot ID]
lexicon version
```

`lexicon init` creates `.lexicon/repo`, performs complete analysis for the selected detected languages, creates the initial private state commit, and publishes an immutable analysis snapshot. Omitting `--languages` or using `all` enables every supported language.

`lexicon scan` replaces the private source mirror with the repository's current relevant files. The private Git repository supplies the diff from the last successful scan. For ordinary source-file modifications, Lexicon expands changed paths through the previous snapshot's reverse dependency graph, requests incremental JSONL records for that impacted file set, parses the stream once at the adapter boundary, writes replacement per-file objects, reuses every unaffected manifest entry, and publishes a new immutable snapshot. Structural changes and adapter fingerprint changes use complete-language analysis.

`lexicon demon` watches the repository recursively and invokes the same scan transaction after a debounce. A periodic complete reconciliation repairs missed filesystem events.

`lexicon rebuild` forces complete analysis of all enabled detected languages or an explicit enabled subset. `lexicon languages set` updates the configured selection, removes disabled language entries from the snapshot, scans immediately, and publishes the result.

`lexicon status` reports the repository, current snapshot, detected and enabled languages, and consumers. `lexicon doctor` verifies configuration, private Git state, immutable objects, adapter paths, runtime requirements, and consumer commands.

`lexicon export` reconstructs verified standalone JSONL libraries. `lexicon gc` deletes only unreachable snapshots and objects while preserving retention and consumer pins. Consumer commands manage and invoke deterministic post-publication hooks.

## Private state repository

```text
.lexicon/
    config.json
    CURRENT
    LOCK
    PENDING              # transient crash-recovery record
    consumers/
        arcana.json
    consumer-state/
        arcana.json
    objects/
        ab/cdef...
    snapshots/
        <snapshot-id>.json
    repo/
        .git/
        source/
```

The state commit is always amended and remains a parentless root commit, so only one commit is reachable. Reflogs are expired after replacement. The repository is an implementation detail used to answer one question: what source content changed since Lexicon last successfully published a snapshot?

Each snapshot manifest records the internal state commit, adapter and schema versions, configuration identity, source-content identity, and fact-object identity for every relevant file. Shared synthetic facts are stored in a separate language object. Objects and snapshot manifests are immutable; identical content reuses the existing object.

`CURRENT` contains the complete snapshot ID and is replaced atomically only after the internal state commit and every referenced object are durable. Consumers should resolve `CURRENT`, load the corresponding manifest, and then open its objects. They never need to observe the mutable mirror or adapter JSONL transport files.

## Object garbage collection

Objectstore garbage collection retains the snapshot named by `CURRENT` and the configured number of newest snapshot manifests. It also retains every snapshot named by a `snapshot_id` field in `.lexicon/consumer-state/*.json`; those pins protect consumer work that still refers to an older immutable snapshot. A pinned snapshot that is missing, or a pin file that is malformed or lacks a valid `snapshot_id`, is a hard error and aborts collection.

The planner follows every preserved manifest's file and shared-object references. Execution deletes only unreferenced snapshot manifests and fact objects. Planning and deletion are deterministic, and execution rejects a plan if `CURRENT` changed after planning. `Store.GarbageCollect(options, dryRun)` performs the bounded plan-and-execute transaction; `dryRun` returns the same deletion lists without removing any files. The explicit planning and execution methods remain available when callers need to inspect a plan before applying it.

## Snapshot export

The `lexicon export` command and objectstore export API write standalone language libraries without changing the current snapshot:

```text
lexicon export --output /tmp/lexicon-export --languages python,go
```

```go
err := (objectstore.Store{Root: "/repo/.lexicon"}).Export(
    "CURRENT", "/tmp/lexicon-export", []string{"python", "go"},
)
```

The snapshot argument may be `CURRENT`, an exact snapshot ID, or empty (which
also resolves `CURRENT`). An empty language list exports every language;
requested languages must exist in the manifest. Each selected language becomes
`<language>.jsonl` in the destination. Export reconstructs the full header from
manifest metadata, verifies every referenced object, combines shared and
per-file records, sorts records deterministically, and atomically replaces each
destination file.

## Transaction and recovery

Only one process may update a repository at a time. Manual scans and demon updates acquire the same advisory lock; a competing writer receives an explicit busy error.

Before advancing the private state commit, Lexicon writes a durable `PENDING` record containing the complete candidate manifest and whether the source commit must advance. If a process stops before the commit, the next scan discards that candidate and recomputes it. If the commit succeeded but publication did not, the next scan attaches the committed state ID to the pending manifest and atomically publishes it without rerunning adapters. Publications that do not require a source commit, such as adapter-only rebuilds, are also recoverable from `PENDING`.

Existing installations that still contain committed `.lexicon/repo/library/*.jsonl` files are migrated once into snapshot objects when necessary, after which the materialized directory is removed from the private state commit. New scans never create it.

## Incremental boundary

A modified source file starts an impacted-file object update rather than a complete language rewrite. Lexicon reads the previous immutable snapshot, follows cross-file edges in reverse, and includes every transitive dependent. Files that previously contained unresolved relationships are also included conservatively because a newly introduced declaration may resolve them.

Lexicon builds a temporary scoped repository containing the impacted files, their transitive forward dependencies from the previous snapshot, and the language's configuration files. Go scopes expand to complete packages and Rust scopes expand to complete crates because those are their minimum sound semantic units. The adapter emits only records owned by the impacted files. Shared synthetic records from the scoped view are marked partial and are not allowed to replace the previous complete shared set.

A directly edited file that already owns cross-file or unresolved relationships uses complete-language analysis immediately, because a partial candidate universe could preserve the same edge identity while changing its true resolution. Leaf and local-only files use the scoped path. Before applying a scoped result, Lexicon compares emitted edge and unresolved topology with the previous file objects. A new relationship, a new unresolved reference, or a scoped adapter failure automatically retries the complete language repository. When the scoped result is accepted, unaffected file objects remain byte-identical and are reused.

Additions, deletions, renames, copies, language configuration changes, missing prior dependency data, and invalid prior snapshot state trigger a complete rebuild of the affected language. This is a correctness boundary, not a permanent protocol limitation. The incremental contract already carries removed-file scopes, so future dependency metadata can narrow those cases without changing consumers or snapshot storage.

## Watch behavior

The demon ignores Git metadata, Lexicon state, linked worktrees, dependency directories, and build outputs. An optional repository-root `.lexiconignore` adds gitignore-compatible patterns, including comments, globs, `**`, path hierarchy, and `!` negation, on top of those permanent exclusions. Ignored files are omitted from complete mirror scans, path syncs, and demon watch filtering. Permanent exclusions cannot be re-included by `.lexiconignore`; they include `.git/`, `.worktrees/`, `.workingtrees/`, the Warlock state directories, `node_modules/`, `vendor/`, `target/`, `dist/`, `build/`, `.venv/`, `venv/`, `__pycache__/`, and `.pytest_cache/`.

The demon keeps the loaded ignore policy in memory while processing filesystem events. A change to `.lexiconignore` reloads the policy, refreshes recursive watches, and triggers a complete scan. New directories are watched recursively. Deletes and renames remove their mirrored paths. Watcher errors trigger an immediate full reconciliation, and the configured reconciliation interval provides an additional recovery path.
