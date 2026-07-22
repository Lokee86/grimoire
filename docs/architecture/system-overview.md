# System Overview

## Purpose

This document describes the implemented Grimoire lexical baseline: how repository files become prepared chunks, how queries are ranked, and how a bounded context package is produced.

## Product boundary

Grimoire owns prepared retrieval state, ranking, budgeted selection, and context-package output. It does not own language parsing, repository relationship graphs, documentation maintenance, agents, or generation.

The current implementation works without other Warlock tools. Future integrations may provide better evidence, but they do not replace Grimoire's selection and budgeting responsibilities.

## End-to-end flow

```text
index command
    │
    ├── resolve repository and state paths
    ├── load previous prepared snapshot when present
    ├── traverse eligible repository files
    ├── reuse unchanged file records
    ├── fallback-chunk changed text files
    ├── count each changed chunk with o200k_base
    └── atomically publish a new prepared snapshot

context command
    │
    ├── load the prepared snapshot
    ├── rank prepared chunks against the query
    ├── apply deterministic tie-breaking
    ├── select whole chunks under the serialized-package budget
    ├── count the exact indented JSON package with o200k_base
    └── emit the verified package
```

The context command does not traverse the repository or read source files. It operates only on the prepared state repository.

## Package ownership

| Package | Owns | Does not own |
| --- | --- | --- |
| `cmd/grimoire` | Process entry point and exit behavior | Command implementation |
| `internal/app` | CLI parsing and operation orchestration | Index formats, ranking rules, budget policy internals |
| `internal/ignore` | Git-ignore pattern loading and matching | Permanent tool-state exclusions |
| `internal/index` | Traversal, filtering, fallback chunking, incremental records, storage, publication | Query ranking and package selection |
| `internal/retrieve` | Query term extraction, candidate scoring, deterministic ordering | Token budgets and output packages |
| `internal/tokenizer` | Fixed `o200k_base` identity and exact token counting | Ranking, chunk boundaries, model selection, or wrapper overhead |
| `internal/compiler` | Whole-chunk budget selection, exact package accounting, and JSON package model | Candidate discovery and ranking |

## Code map

```text
cmd/grimoire/main.go
    └── internal/app.Run
            ├── index.Build
            │      └── ignore.Load / Policy.Ignored
            ├── index.Load / index.Save
            ├── retrieve.Search
            └── compiler.Compile
                   └── tokenizer.Count
```

Important implementation files:

| File | Responsibility |
| --- | --- |
| `internal/app/run.go` | Commands, flags, path resolution, and JSON output |
| `internal/index/build.go` | Repository traversal, file eligibility, incremental reuse, and changed-file chunking |
| `internal/index/chunk.go` | Deterministic line-based fallback chunking and exact chunk token counts |
| `internal/index/store.go` | Snapshot loading, validation, incremental shard writes, and atomic reference publication |
| `internal/index/codec.go` | Deterministic shard encoding and path validation |
| `internal/index/file_codec.go` | Deterministic file and chunk record encoding |
| `internal/ignore/policy.go` | Root, nested, and replacement Git-ignore behavior |
| `internal/retrieve/search.go` | Lexical scoring, reasons, limits, and tie-breaking |
| `internal/compiler/compiler.go` | Whole-chunk fitting, exact serialized-package accounting, and package construction |
| `internal/tokenizer/tokenizer.go` | Shared `o200k_base` codec and token-counting seam |

## Determinism

For the same prepared snapshot, query, candidate limit, and budget, the current context package is deterministic.

Deterministic behavior comes from:

- repository-relative slash-normalized paths;
- sorted file and shard records;
- content-derived chunk identities;
- fixed lexical scoring rules;
- score, path, and start-line ordering; and
- whole-chunk budget selection in ranked order.

Filesystem traversal order does not determine persisted file ordering.

## Current chunking

The fallback chunker normalizes CRLF to LF, removes one final newline, and divides non-empty text into blocks of approximately 48 lines. It prefers a recent blank-line boundary when that boundary would leave a useful block of more than eight lines.

Leading and trailing blank lines are removed from each chunk. Chunk identity is derived from the repository-relative path, source range, and exact chunk text.

This is deliberately language-agnostic. It does not understand functions, types, methods, Markdown sections, or structured configuration boundaries.

## Current retrieval

The lexical ranker lowercases the query and candidate text, extracts unique alphanumeric or underscore terms of at least two characters, and applies fixed boosts for:

- the complete query phrase in content;
- filename matches;
- path matches;
- leading-line matches; and
- repeated content matches, capped per term.

Each candidate records compact reasons for its score. Ties are resolved by path and then source start line.

The current implementation scans every prepared chunk in memory for each query. It avoids repository I/O but does not yet use a postings index or BM25.

## Current budget selection

Each prepared chunk stores its exact `o200k_base` count. The compiler considers ranked candidates in order and tentatively serializes the complete indented JSON package for each addition. A chunk is retained only when the resulting package fits the caller's budget. Rejected candidates are counted as omitted, and later smaller candidates may still fit.

The package-level `token_count` includes query text, metadata, paths, line ranges, scores, reasons, selected content, JSON syntax, indentation, escaping, and the trailing newline. The compiler stabilizes the self-referential `token_count` field, verifies the final bytes before output, and returns an error when the budget cannot fit even an empty package.

This budget is exact for `o200k_base`. It does not include any chat, tool, agent, or transport wrapper added after Grimoire emits the package.

## Related documentation

- [Prepared index](prepared-index.md)
- [Context package](../reference/context-package.md)
- [Current limitations](../limits/current-limitations.md)
- [Roadmap](../planning/roadmap.md)
