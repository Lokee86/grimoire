# Embedding Model

## Fixed model contract

Grimoire's first semantic provider is:

| Field | Value |
| --- | --- |
| Model | `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0` |
| Runtime | Local `llama.cpp` embeddings server |
| Native dimensions | 1024 |
| Stored dimensions | 512 |
| Similarity | Inner product over L2-normalized vectors |
| Default endpoint | `http://127.0.0.1:8080/v1` |
| Grimoire identity | `qwen3-embedding-0.6b-q8_0-512d` |

Grimoire keeps the first 512 Matryoshka dimensions from the model's native vector and normalizes that reduced vector itself. This identity must be part of future vector-record compatibility checks.

## Query and document inputs

Queries are formatted as:

```text
Instruct: Given a software development query, retrieve relevant source code and documentation from a repository
Query:<query text>
```

Repository chunks are embedded as their raw prepared text. The query instruction is not added to documents.

## Managed setup

On Windows x64:

```bash
grimoire model setup
```

The command installs a pinned CPU build of `llama.cpp` and the official Q8 GGUF model beneath the user's cache directory. Both downloads are checked against fixed SHA-256 digests before publication.

The default cache is returned by the command and can be overridden:

```bash
grimoire model setup --cache D:/models/grimoire
```

Automatic runtime setup currently targets Windows x64. On another platform, install `llama.cpp` separately and expose `llama-server` or `llama` on `PATH`, or set `GRIMOIRE_LLAMA_SERVER`.

## Runtime lookup

Grimoire checks runtime locations in this order:

1. `--runtime` supplied to `model serve` or `model info`;
2. `GRIMOIRE_LLAMA_SERVER`;
3. Grimoire's managed runtime;
4. `llama-server` on `PATH`;
5. the modern `llama` multicall binary on `PATH`.

Model lookup checks:

1. `--model-file` supplied to `model serve`;
2. `GRIMOIRE_EMBEDDING_MODEL`;
3. Grimoire's managed model.

When no local model is found, `model serve` allows `llama.cpp` to resolve the fixed Hugging Face model reference.

## Running and probing

Start the blocking local service:

```bash
grimoire model serve
```

In another shell, verify a real query/document embedding pair:

```bash
grimoire model probe
```

`model probe` reports the model identity, endpoint, final dimension count, and similarity. It does not write vectors into prepared repository state.

## Current boundary

The provider, managed setup, server launcher, query formatting, vector reduction, normalization, incremental chunk embedding, persistent vector storage, exact nearest-neighbour retrieval, and vector-backed `grimoire context` integration are implemented. Source indexing and vector refresh remain separate explicit operations, and context selection still uses whole ranked chunks.
