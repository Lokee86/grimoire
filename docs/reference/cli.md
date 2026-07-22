# CLI Reference

## Invocation

```text
grimoire <command> [flags]
```

Current top-level commands are `index`, `context`, `model`, `vector`, and `version`.

## `grimoire model setup`

Install Grimoire's pinned local embedding runtime and model:

```bash
grimoire model setup [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cache <path>` | operating-system user cache plus `grimoire` | Managed runtime and model directory |
| `--timeout <duration>` | `45m` | Complete download and installation timeout |

On Windows x64 the command downloads a pinned CPU `llama.cpp` runtime and `Qwen3-Embedding-0.6B-Q8_0.gguf`, verifies fixed SHA-256 digests, and publishes them atomically into the cache. Repeated setup reuses verified files.

The JSON result contains the cache, runtime, and model paths plus their identities.

## `grimoire model info`

Report the fixed model contract and whether a runtime and local model are discoverable:

```bash
grimoire model info [--runtime <path>] [--endpoint <url>]
```

This command does not start a server or send an embedding request.

## `grimoire model serve`

Start a blocking local `llama.cpp` embeddings service:

```bash
grimoire model serve [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--runtime <path>` | discovered runtime | `llama-server` or `llama` executable |
| `--model-file <path>` | managed model, then fixed remote model reference | Local GGUF file |
| `--host <address>` | `127.0.0.1` | Bind address |
| `--port <n>` | `8080` | Bind port |
| `--context-size <n>` | `8192` | Runtime context size |
| `--ubatch-size <n>` | `2048` | Runtime physical batch size |

The command enables embedding mode and last-token pooling. Grimoire performs final 512-dimensional truncation and L2 normalization in its client.

## `grimoire model probe`

Verify the running embeddings endpoint with a real query/document pair:

```bash
grimoire model probe [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--endpoint <url>` | `http://127.0.0.1:8080/v1` | OpenAI-compatible embeddings base URL |
| `--query <text>` | sample code-retrieval query | Query to instruct and embed |
| `--document <text>` | sample source passage | Raw document to embed |
| `--timeout <duration>` | `2m` | Request timeout |

The result reports the fixed identity, endpoint, 512 dimensions, and inner-product similarity.

## `grimoire index`

Prepare or incrementally update source retrieval state:

```bash
grimoire index [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root |
| `--state <path>` | `<root>/.grimoire` | Prepared-state repository |
| `--ignore-file <path>` | root and nested `.gitignore` files | Replacement Git-ignore file |
| `--max-file-bytes <n>` | 2 MiB | Maximum eligible source file size |

The command prepares source chunks and exact token counts. Persistent embeddings are built separately by `grimoire vector build`.

Output:

```json
{
  "state": "/absolute/path/to/repository/.grimoire",
  "files": 21,
  "stats": {
    "scanned": 21,
    "reused": 20,
    "updated": 1,
    "removed": 0
  }
}
```

## `grimoire vector build`

Embed missing prepared chunks and publish the current packed vector snapshot:

```bash
grimoire vector build [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root |
| `--state <path>` | `<root>/.grimoire` | Prepared-state repository |
| `--endpoint <url>` | `http://127.0.0.1:8080/v1` | OpenAI-compatible embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--batch-size <n>` | `16` | Documents per embedding request |
| `--timeout <duration>` | `30m` | Complete build timeout |

The command reuses immutable vectors for unchanged chunk text, embeds only missing source identities, and materializes a sorted memory-mapped snapshot.

## `grimoire vector search`

Run exact semantic search over the packed snapshot:

```bash
grimoire vector search --query <text> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root used to resolve state |
| `--state <path>` | `<root>/.grimoire` | Prepared-state repository |
| `--query <text>` | none | Required semantic query |
| `--top-k <n>` | `20` | Maximum results |
| `--endpoint <url>` | `http://127.0.0.1:8080/v1` | Embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--timeout <duration>` | `2m` | Query embedding timeout |

Results contain chunk identity, source path, line range, and exact dot-product score.

## `grimoire vector info`

Report native-library and snapshot availability:

```bash
grimoire vector info [--root <path>] [--state <path>] [--engine <path>]
```

When a snapshot exists, the result includes its embedding identity, dimensions, and vector count.

## `grimoire context`

Compile a bounded semantic context package from prepared and vector state:

```bash
grimoire context [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root used to resolve state |
| `--state <path>` | `<root>/.grimoire` | Prepared-state repository |
| `--query <text>` | none | Required retrieval query |
| `--budget <n>` | `2000` | Maximum `o200k_base` tokens in emitted JSON |
| `--candidate-limit <n>` | `200` | Maximum ranked semantic candidates |
| `--endpoint <url>` | `http://127.0.0.1:8080/v1` | OpenAI-compatible embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--timeout <duration>` | `2s` | Complete semantic retrieval timeout |

The command validates the vector snapshot manifest against the exact content-addressed prepared-index identity before query embedding, then validates model identity, dimensions, and vector count, performs exact vector retrieval, and records selection-level source, rank, score, and reasons. If the vector path is missing, stale, incompatible, or unavailable, it writes a warning to stderr and uses the deterministic lexical fallback.

## `grimoire version`

```bash
grimoire version
```

Current value: `0.1.0-dev`.

## Environment variables

| Variable | Meaning |
| --- | --- |
| `GRIMOIRE_LLAMA_SERVER` | Explicit `llama.cpp` runtime executable |
| `GRIMOIRE_EMBEDDING_MODEL` | Explicit local GGUF model file |
| `GRIMOIRE_VECTOR_ENGINE` | Explicit Rust vector-engine DLL |

## Error behavior

Errors remain human-readable and do not yet have stable diagnostic codes or exit-code classes.

## Related documentation

- [Embedding model](embedding-model.md)
- [Vector store](vector-store.md)
- [Indexing](indexing.md)
- [Context package](context-package.md)
- [Prepared index](../architecture/prepared-index.md)
