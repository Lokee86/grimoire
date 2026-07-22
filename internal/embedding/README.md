# Embedding package

`internal/embedding` owns Grimoire's fixed local embedding-model contract and the client/runtime boundary used to reach it.

Current contract:

- model: `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0`;
- runtime: a local `llama.cpp` embeddings server;
- native output: 1024 dimensions;
- stored output: the first 512 Matryoshka dimensions, normalized with L2 inside Grimoire;
- query prompting: a repository source-code and documentation retrieval instruction;
- default query plan: retain the complete query and mechanically split it into 16-token windows;
- split request policy: at most 64 query tokens per HTTP request and two concurrent requests by default;
- optional query plans: one complete full query, or full query plus the bounded split-window requests;
- document prompting: raw chunk text with no instruction; and
- similarity: inner product over normalized vectors.

The package does not own vector persistence, nearest-neighbour search, hybrid ranking, chunk boundaries, or context-package selection.
