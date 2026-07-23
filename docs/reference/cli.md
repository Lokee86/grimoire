# CLI Reference

## Invocation

```text
grimoire <command> [flags]
```

Current top-level commands are `index`, `context`, `eval`, `model`, `vector`, and `version`.

## `grimoire model setup`

Install Grimoire's pinned local embedding runtime and model:

```bash
grimoire model setup [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cache <path>` | operating-system user cache plus `grimoire` | Managed runtime and model directory |
| `--backend <name>` | `auto` | `auto`, `cuda`, `vulkan`, or `cpu` llama.cpp runtime |
| `--timeout <duration>` | `45m` | Complete download and installation timeout |

On Windows x64 the command downloads a pinned `llama.cpp` runtime and `Qwen3-Embedding-0.6B-Q8_0.gguf`, verifies fixed SHA-256 digests, and publishes them atomically into the cache. `auto` selects CUDA when a compatible NVIDIA driver is present, otherwise Vulkan when available, then CPU. Set `GRIMOIRE_LLAMA_BACKEND` or pass `--backend` to override detection. Repeated setup reuses verified files.

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
| `--port <n>` | `9876` | Bind port |
| `--context-size <n>` | `8192` | Runtime context size |
| `--ubatch-size <n>` | `2048` | Runtime physical batch size |
| `--parallel <n>` | `4` | Concurrent llama.cpp server slots |

The command enables embedding mode and last-token pooling. Grimoire performs final 512-dimensional truncation and L2 normalization in its client.

## `grimoire model probe`

Verify the running embeddings endpoint with a real query/document pair:

```bash
grimoire model probe [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | OpenAI-compatible embeddings base URL |
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
| `--exclude <path>` | none | Root-relative or absolute path to exclude; repeatable |

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
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | OpenAI-compatible embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--batch-size <n>` | `4` | Documents per embedding request; matches the default four server slots |
| `--batch-concurrency <n>` | `1` | Concurrent embedding requests; ingestion remains serialized |
| `--timeout <duration>` | `30m` | Complete build timeout |

The command returns the existing snapshot immediately when the prepared-index identity and vector manifest already match. Changed builds deduplicate identical chunk text, reuse source identities recorded by the previous manifest without probing the object store, check only newly introduced source hashes, embed only genuinely missing vectors, persist completed batches immediately, and materialize a sorted memory-mapped snapshot. Progress and throughput are written to stderr while the final JSON result remains on stdout.

The local llama.cpp server already distributes one request across its four slots, so the defaults send four documents in one request and avoid competing request queues. Higher request concurrency remains available for remote providers; object-store ingestion stays serialized and deterministic. The result reports chunk and unique-vector counts, embedded and reused counts, object checks, cache-hit status, duration, snapshot size, and peak memory.

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
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | Embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--timeout <duration>` | `2m` | Query embedding timeout |
| `--query-embedding-mode <mode>` | `fast` | `fast`, `full`, or `quality` query plan |
| `--query-window-tokens <n>` | `16` | Tokens per split-query window |
| `--query-batch-tokens <n>` | `64` | Maximum split-query tokens per embedding request |
| `--query-batch-concurrency <n>` | `2` | Maximum concurrent query embedding requests |
| `--query-max-tokens <n>` | `0` | Optional query-token limit; zero keeps the complete query |

`fast` mechanically partitions the complete query into non-overlapping windows, groups those windows into requests capped at 64 query tokens, runs at most two requests concurrently, searches the returned vectors concurrently, and merges duplicate results. `full` embeds the complete query once. `quality` embeds both the full query and every split window. `--query-max-tokens` remains an optional explicit safety limit; its default of zero does not truncate the query.

Results contain chunk identity, source path, line range, and the best exact dot-product score across the query vectors that matched each chunk.

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
| `--budget <n>` | `0` | Maximum `o200k_base` tokens; zero selects a deterministic automatic target |
| `--candidate-limit <n>` | `200` | Maximum merged exact plus semantic/fallback primary candidates before curation |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | OpenAI-compatible embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--structure <bool>` | `true` | Include available Lexicon and Arcana structural evidence |
| `--structure-timeout <duration>` | `30s` | Complete structural-provider timeout |
| `--lexicon-facts <path>` | automatic snapshot export | Explicit standalone Lexicon JSONL directory override |
| `--lexicon-state <path>` | `<root>/.lexicon` | Lexicon immutable state directory |
| `--lexicon-command <path>` | `lexicon` | Executable used for immutable snapshot export |
| `--arcana-state <path>` | `<root>/.arcana` | Arcana immutable graph-state directory |
| `--arcana-command <path>` | `arcana` | Executable used for graph synchronization and protocol queries |
| `--timeout <duration>` | `2s` | Complete semantic retrieval timeout |
| `--query-embedding-mode <mode>` | `fast` | `fast`, `full`, or `quality` query plan |
| `--query-window-tokens <n>` | `16` | Tokens per split-query window |
| `--query-batch-tokens <n>` | `64` | Maximum split-query tokens per embedding request |
| `--query-batch-concurrency <n>` | `2` | Maximum concurrent query embedding requests |
| `--query-max-tokens <n>` | `0` | Optional query-token limit; zero keeps the complete query |

The command validates the vector snapshot manifest against the exact content-addressed prepared-index identity before query embedding, then validates model identity, dimensions, and vector count and performs exact vector retrieval. `fast` embeds the complete query as fixed non-overlapping windows grouped into bounded 64-token requests, with at most two requests active concurrently. `full` embeds the complete query once. `quality` adds the full-query vector to the split windows. Concrete literal signals also activate targeted exact recovery. Provider candidates are merged before deterministic query-shape analysis. When `--budget` is omitted or zero, focused queries select 3,000 tokens, bounded queries 6,000, and exploratory queries 12,000. A positive explicit budget bypasses automatic selection. Candidates are then deduplicated, diversified, and expanded with bounded prepared neighbours. Automatic assembly stops after deterministic evidence coverage is reached: focused requests remain around one anchor region, bounded requests require two represented regions, and exploratory requests require three. The emitted package records the assembly decision. Explicit-budget requests retain the existing fit-to-budget behavior.

Structural enrichment is enabled by default. When Lexicon state exists, Grimoire resolves `.lexicon/CURRENT`, creates or reuses a cached `lexicon export`, and emits matched symbols, source spans, and immediate relationships as first-class package evidence. It then resolves the Arcana snapshot for the same Lexicon ID, invokes one-shot `arcana sync` when necessary, and queries Arcana's JSONL protocol for operational roles, impact, unresolved references, and shortest call chains. Structural failures warn and preserve standalone source retrieval. Use `--structure=false` to skip both providers or the explicit state, command, and facts flags to override discovery.

If the vector path is missing, stale, incompatible, or unavailable, the command writes a warning to stderr and substitutes the deterministic lexical fallback before the same exact-recovery and curation stages. Structural evidence can still be emitted during semantic fallback.

## `grimoire eval retrieval`

Run a repository-owned judged retrieval corpus against one or more query modes:

```bash
grimoire eval retrieval --cases <path> --root <repository> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cases <path>` | none | Required judged corpus JSON |
| `--root <path>` | `.` | Repository being evaluated |
| `--state <path>` | `<root>/.grimoire` | Prepared and vector state |
| `--modes <list>` | `fast,full,quality,lexical` | Comma-separated modes to execute |
| `--variant <name>` | `standalone` | Result label for paired comparisons |
| `--budget <n>` | case budget | Optional fixed budget override for every case |
| `--adaptive` | `false` | Replace case budgets with query-shape targets and evidence-coverage assembly |
| `--candidate-limit <n>` | `200` | Normal ranked candidate limit |
| `--probe-limit <n>` | `800` | Broader diagnostic ranking probe used only for failure attribution |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | Embeddings endpoint |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--structural-providers <list>` | `none` | `none`, `lexicon`, or `lexicon,arcana` |
| `--structure-timeout <duration>` | `30s` | Per-case structural-provider timeout |
| `--lexicon-facts <path>` | automatic snapshot export | Explicit standalone Lexicon JSONL directory override |
| `--lexicon-state <path>` | `<root>/.lexicon` | Lexicon immutable state directory |
| `--lexicon-command <path>` | `lexicon` | Executable used for immutable snapshot export |
| `--arcana-state <path>` | `<root>/.arcana` | Arcana immutable graph-state directory |
| `--arcana-command <path>` | `arcana` | Executable used for graph synchronization and protocol queries |
| `--timeout <duration>` | `10s` | Per-case source-retrieval timeout |
| `--output-dir <path>` | `evaluation/results` | JSON and Markdown result directory |
| `--output-prefix <name>` | generated | Shared result filename prefix |
| query planning flags | context defaults | Window, batch, concurrency, and optional max-token settings |

The corpus is separate from deterministic unit-test fixtures. A case may require source evidence, structural evidence, or both. Source expectations use `required`, `supporting`, and `forbidden`. Structural expectations use `required_structural`, `supporting_structural`, and `forbidden_structural`.

Structural expectations require `provider` and `kind`. Optional assertions include subject `symbol` and `path`, relationship `relation`, `direction`, and `certainty`, related `target_symbol` and `target_path`, an ordered `chain` subsequence, and unresolved-reference `expression`. Before retrieval, the runner verifies every referenced source path and any symbol paired with a path.

`--structural-providers none` runs the source-only baseline. `lexicon` executes immutable Lexicon export and symbol matching. `lexicon,arcana` additionally synchronizes and queries Arcana against the same snapshot. Arcana cannot be enabled without Lexicon because Lexicon-matched symbols are its bounded graph-query seeds.

For each case and mode the runner records source and structural timings, provider warnings, selected source chunks, retained structural facts, immutable provider snapshots, final serialized package tokens, separate source and structural recall, separate irrelevant-evidence rates, and failure attribution. `--adaptive` also records the selected automatic budget, curated and assembled candidate counts, represented evidence coverage, and the assembly stop reason. Source and structural failures distinguish adaptive assembly loss from later budget-fitting loss. `--adaptive` cannot be combined with a fixed `--budget` override. The broad source-ranking probe does not contribute to reported context latency.

Outputs are a machine-readable JSON report and a concise Markdown comparison grouped by mode and category. Package comparison includes median and p95 tokens, median selected chunks, and median budget utilization. A case passes only when every required source and structural expectation survives into the final context package.

## `grimoire version`

```bash
grimoire version
```

Current value: `0.1.0-dev`.

## Environment variables

| Variable | Meaning |
| --- | --- |
| `GRIMOIRE_LLAMA_BACKEND` | Managed setup backend: `auto`, `cuda`, `vulkan`, or `cpu` |
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
