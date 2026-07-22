# System Overview

## Purpose

Grimoire is a standalone repository RAG and context-compilation tool. It owns prepared retrieval state, lexical and semantic candidate retrieval, hybrid ranking, exact budgeted selection, and context-package output.

The implemented foundation currently includes source preparation, a lexical baseline, exact output budgeting, and an operational local embedding provider. Vector persistence and hybrid retrieval remain under development.

## Current flow

```text
index command
    │
    ├── traverse eligible repository files
    ├── reuse unchanged file records
    ├── fallback-chunk changed text files
    ├── count chunks with o200k_base
    └── atomically publish prepared source state

model setup / serve
    │
    ├── install verified llama.cpp runtime and Q8 GGUF model
    ├── expose a local OpenAI-compatible embeddings endpoint
    ├── instruct repository queries
    ├── embed raw document chunks
    └── reduce native 1024 dimensions to normalized 512 dimensions

context command
    │
    ├── load prepared source state
    ├── perform current linear lexical ranking
    ├── fit whole chunks under the package budget
    └── emit verified JSON
```

The embedding path is real and independently probeable, but it is not yet called by `index` or `context`.

## Target retrieval flow

```text
prepared chunks
    ├── incremental BM25 postings ───────► lexical candidates
    └── incremental model vectors ──────► semantic candidates
                                                │
query ──► instructed query embedding ───────────┤
                                                ▼
                                      deterministic rank fusion
                                                ▼
                                       context selection
                                                ▼
                                  exact o200k_base package
```

Lexicon may later replace fallback boundaries with structural ranges, but the entire retrieval path must work without it.

## Package ownership

| Package | Owns |
| --- | --- |
| `internal/app` | CLI parsing and operation orchestration |
| `internal/ignore` | Git-ignore pattern loading and matching |
| `internal/index` | Traversal, fallback chunking, source records, storage, and atomic publication |
| `internal/retrieve` | Current lexical candidate scoring and deterministic ordering |
| `internal/embedding` | Fixed model identity, verified setup, runtime launch, query formatting, HTTP client, reduction, normalization, and probing |
| `internal/tokenizer` | Fixed `o200k_base` counting |
| `internal/compiler` | Whole-chunk package selection and exact serialized-package accounting |

Future vector storage and hybrid fusion should receive their own concrete ownership seams rather than being folded into the model client.

## Code map

```text
cmd/grimoire/main.go
    └── app.Run
        ├── index.Build / index.Save
        ├── retrieve.Search
        ├── compiler.Compile / compiler.Marshal
        └── model commands
            ├── embedding.Setup
            ├── embedding.Serve
            └── embedding.Client.Probe
```

## Embedding contract

The fixed provider is `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0`, served locally through `llama.cpp`.

Queries receive the fixed repository retrieval instruction. Documents remain raw. Native 1024-dimensional output is truncated to the first 512 Matryoshka dimensions and L2-normalized inside Grimoire. Inner product is therefore cosine similarity.

Model identity, dimensions, preprocessing, runtime compatibility, and future vector schema must collectively determine whether persisted vectors can be reused.

## Determinism

Source preparation, current lexical ranking, and package compilation are deterministic for the same snapshot, query, limits, and budget.

Embedding inference is locally controlled and uses a fixed model artifact, prompt format, dimension reduction, and normalization. Exact floating-point values may still vary with runtime build and hardware backend; future vector compatibility must record enough identity to prevent silent mixing.

## Product boundaries

Grimoire does not own language parsing, repository relationship graphs, documentation maintenance, agent orchestration, or generative inference. Those may supply optional evidence, but they are not prerequisites for baseline hybrid RAG.

Grimoire does own its local embedding provider contract and vector retrieval state. It does not require hosted embedding APIs or hosted vector infrastructure.

## Related documentation

- [Embedding model](../reference/embedding-model.md)
- [Prepared index](prepared-index.md)
- [Context package](../reference/context-package.md)
- [Current limitations](../limits/current-limitations.md)
- [Roadmap](../planning/roadmap.md)
