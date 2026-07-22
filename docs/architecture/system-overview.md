# System Overview

## Purpose

Grimoire is a standalone repository RAG and context-compilation tool. It owns prepared retrieval state, exact semantic candidate retrieval, lexical failure fallback, exact budgeted selection, and context-package output.

The implemented foundation includes source preparation, a lexical fallback, exact output budgeting, an operational local embedding provider, persistent content-addressed vectors, packed memory-mapped snapshots, exact semantic search, and semantic context compilation.

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

vector build / search
    │
    ├── reuse vectors by embedding identity and source-content hash
    ├── batch-embed only missing chunks
    ├── publish a sorted aligned float32 snapshot
    ├── memory-map and validate the snapshot in Rust
    └── run serial or concurrent exact dot-product search

context command
    │
    ├── load prepared source state and the packed vector snapshot
    ├── validate model identity, dimensions, and chunk count
    ├── embed the query and run exact vector retrieval
    ├── fall back to deterministic lexical ranking on semantic failure
    ├── fit whole chunks under the package budget
    └── emit verified JSON with source/rank/score provenance
```

The embedding path is independently probeable and used by `vector build`, `vector search`, and `context`. It remains separate from source indexing so explicit one-shot preparation and vector refresh stay independently controllable.

## Retrieval flow

```text
prepared chunks ──► incremental model vectors
                              │
query ──► instructed query embedding
                              │
                              ▼
                 exact full-vector scan
                              │
                              ▼
                    context selection
                              │
                              ▼
               exact o200k_base package

semantic failure ──► deterministic lexical fallback
```

Targeted exact lookup may later recover paths, identifiers, quoted phrases, configuration keys, and error codes. Lexicon may enrich chunks with authoritative symbols and structural ranges. Neither is required for the baseline semantic path.

## Package ownership

| Package | Owns |
| --- | --- |
| `internal/app` | CLI parsing and operation orchestration |
| `internal/ignore` | Git-ignore pattern loading and matching |
| `internal/index` | Traversal, fallback chunking, source records, storage, and atomic publication |
| `internal/retrieve` | Shared candidate provenance plus lexical fallback scoring and deterministic ordering |
| `internal/embedding` | Fixed model identity, verified setup, runtime launch, query formatting, HTTP client, reduction, normalization, and probing |
| `internal/vectorstore` | Native-library discovery, ABI validation, caller-owned buffers, and snapshot-handle lifecycle |
| `native/vector-engine` | Immutable vector objects, packed snapshot format, mmap validation, and exact concurrent search |
| `internal/tokenizer` | Fixed `o200k_base` counting |
| `internal/compiler` | Whole-chunk package selection and exact serialized-package accounting |

Vector storage has its own Rust engine and Go bridge. Future exact lookup, structural enrichment, and selection policies should receive concrete ownership seams rather than being folded into the model client or compiler.

## Code map

```text
cmd/grimoire/main.go
    └── app.Run
        ├── index.Build / index.Save
        ├── context command
        │   ├── embedding.Client
        │   ├── vectorstore.Library / vectorstore.Engine
        │   ├── retrieve.Search fallback
        │   └── compiler.Compile / compiler.Marshal
        ├── vector commands
        │   ├── embedding.Client
        │   └── vectorstore.Library / vectorstore.Engine
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

Source preparation, vector-object addressing, packed snapshot materialization, exact semantic result ordering, lexical fallback ranking, and package compilation are deterministic for the same inputs.

Embedding inference is locally controlled and uses a fixed model artifact, prompt format, dimension reduction, and normalization. Exact floating-point values may still vary with runtime build and hardware backend; future vector compatibility must record enough identity to prevent silent mixing.

## Product boundaries

Grimoire does not own language parsing, repository relationship graphs, documentation maintenance, agent orchestration, or generative inference. Those may supply optional evidence, but they are not prerequisites for baseline semantic RAG.

Grimoire does own its local embedding provider contract and vector retrieval state. It does not require hosted embedding APIs or hosted vector infrastructure.

## Related documentation

- [Embedding model](../reference/embedding-model.md)
- [Vector store](../reference/vector-store.md)
- [Prepared index](prepared-index.md)
- [Context package](../reference/context-package.md)
- [Current limitations](../limits/current-limitations.md)
- [Roadmap](../planning/roadmap.md)
