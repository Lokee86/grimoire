# Lexicon application

Lexicon keeps the most recently observed relevant repository state. It does not follow the source repository's commits, index, staging area, or branches.

## Commands

```text
lexicon init [--repo PATH] [--adapters PATH]
lexicon scan [--repo PATH]
lexicon daemon [--repo PATH] [--debounce 150ms] [--reconcile 30s]
```

`lexicon init` creates `.lexicon/repo`, performs a complete relevant-file scan, generates language libraries, and creates the initial private state commit.

`lexicon scan` replaces the private source mirror with the repository's current relevant files. The private Git repository supplies the diff from the last successful scan. Lexicon regenerates only the language libraries affected by that diff and amends the private root commit.

`lexicon daemon` watches the repository recursively. Changed paths are debounced and copied into the private mirror, then the same internal Git diff and language-library update path runs. A periodic complete reconciliation repairs missed filesystem events.

## Private state repository

```text
.lexicon/
    config.json
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

## Incremental boundary

The current adapters analyze repositories rather than individual files. The private diff therefore selects affected languages, and only those language libraries are regenerated. Each selected adapter sees the complete current source mirror, preserving cross-file resolution. A later adapter contract can narrow this to changed-file extraction plus dependent relationship repair without changing the application lifecycle.

## Watch behavior

The daemon ignores Git metadata, Lexicon state, linked worktrees, dependency directories, and build outputs. New directories are watched recursively. Deletes and renames remove their mirrored paths. Watcher errors trigger an immediate full reconciliation, and the configured reconciliation interval provides an additional recovery path.
