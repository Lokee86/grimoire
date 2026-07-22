# Embedding package

`internal/embedding` owns Grimoire's fixed local embedding-model contract and the client/runtime boundary used to reach it.

Current contract:

- model: `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0`;
- runtime: a local `llama.cpp` embeddings server;
- native output: 1024 dimensions;
- stored output: the first 512 Matryoshka dimensions, normalized with L2 inside Grimoire;
- query prompting: a repository source-code and documentation retrieval instruction;
- default query plan: mechanically split 16-token windows for measured minimum uncached-query latency;
- optional query plans: one capped full query, or full query plus windows;
- split inputs: one batched HTTP embedding request followed by concurrent vector searches;
- document prompting: raw chunk text with no instruction; and
- similarity: inner product over normalized vectors.

The package does not own vector persistence, nearest-neighbour search, hybrid ranking, chunk boundaries, or context-package selection.
