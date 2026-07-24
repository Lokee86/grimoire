# Embedding package

`internal/embedding` owns Grimoire's fixed local embedding contract, managed `llama.cpp` runtime, HTTP client, and query batching.

## Model contract

- Model: `Qwen/Qwen3-Embedding-0.6B-GGUF:Q8_0`.
- Identity: `qwen3-embedding-0.6b-q8_0-512d`.
- Native output: 1,024 dimensions.
- Stored output: first 512 Matryoshka dimensions, L2-normalized by Grimoire.
- Query input: fixed repository source-code and documentation retrieval instruction.
- Document input: raw prepared chunk text.
- Similarity contract: inner product over normalized vectors.

## Managed runtime

The package owns pinned runtime/model metadata, verified download and rollback-safe atomic cache publication, runtime discovery, and Windows x64 backend selection for `auto`, `cuda`, `vulkan`, and `cpu`.

Automatic selection prefers supported CUDA, then Vulkan, then CPU. Explicit unavailable backends fail rather than silently changing the requested backend. Managed serving adds detached process supervision, readiness probing, backend-log verification, crash restart, atomic config/state publication, stale-process cleanup, bounded rotating logs, per-slot context contracts, preflight input-token rejection, and optional NVIDIA telemetry. Non-Windows systems may supply a compatible runtime and model through environment variables or `PATH`.

## Query plans

- `fast`: retain the complete query, split it into non-overlapping 16-token windows, group at most 64 query tokens per request, and run at most two requests concurrently.
- `full`: embed the complete query once.
- `quality`: embed the complete query plus all fast-mode windows.

Response processing validates indices and native dimensions, truncates to 512 dimensions, rejects malformed/non-finite vectors, and normalizes locally.

## Boundary

The package does not own source traversal, chunk boundaries, vector persistence, exact search, ranking, query-shape policy, evidence assembly, or context-package fitting. Chunking remains responsible for producing useful source regions; the embedding client independently enforces the active runtime's maximum input as the final safety boundary.
