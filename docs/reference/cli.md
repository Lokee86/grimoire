# CLI Reference

## Invocation

```text
grimoire <command> [flags]
```

Current top-level commands are `status`, `index`, `knowledge`, `query`, `mcp`, `context`, `eval`, `model`, `vector`, `investigation`, `version`, and `help`. Running `grimoire` with no arguments, or using `grimoire help`, `grimoire -h`, or `grimoire --help`, prints the normal workflow and exits successfully.

## `grimoire status`

Inspect repository identity and the prepared Lexicon, Arcana, and Grimoire state:

```bash
grimoire status --root <repository>
grimoire status --root <repository> --refresh
grimoire status --root <repository> --refresh --force
```

The command emits status schema version 2 with Git/source identity, Lexicon, Arcana, prepared-source, knowledge-index, Arcana-vector, and documentation-vector freshness; it also reports stale reasons, performed actions, elapsed times, warnings, and deterministic-query readiness. `status` is read-only by default. `--refresh` incrementally prepares Lexicon, Arcana, Grimoire source state, and the documentation knowledge index only when needed; `--force` refreshes all deterministic state. Vector indexes are inspected but never built by this command.

## `grimoire knowledge`

`knowledge index` builds the independent documentation/rationale index under `<root>/.grimoire/knowledge`; `knowledge search` returns exact cited sections as JSON; `knowledge inspect` reports state or one document/section; and `knowledge vector build|info` owns optional documentation embeddings. Search uses deterministic BM25 by default. Pass `--vectors=true` to opt into the current documentation vector snapshot as a supplemental ranker. See [Knowledge retrieval](knowledge.md).

## `grimoire investigation`

The investigation ledger stores deduplicated agent-facing discovery evidence under `<grimoire-state>/investigations/<session-id>/`. Sessions are bound to one repository snapshot and optional provider snapshot identities.

```bash
grimoire investigation create --session <id> --snapshot <repository-id> [--provider name=identity]
grimoire investigation status --session <id>
grimoire investigation close --session <id>
```

All commands accept `--root <path>` and `--state <path>`; state defaults to `<root>/.grimoire`. Evidence recording is owned by `internal/investigation` for the query and MCP layers. See [`internal/investigation`](../../internal/investigation/README.md) for the package contract.

## `grimoire mcp`

Serve the unified `grimoire_query` agent tool over MCP stdio:

```bash
grimoire mcp --root <repository>
grimoire mcp --root <repository> --state-mode current-only
```

The default `--state-mode` is `refresh-if-needed`. The server checks and incrementally prepares repository state before each call, retrieves code and repository knowledge through separate lanes, and supports persistent `session` names that replace repeated evidence with prior handles. See [Agent MCP runtime](agent-mcp.md).

## `grimoire model setup`

Install Grimoire's pinned local embedding runtime and model:

```bash
grimoire model setup [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cache <path>` | operating-system user cache plus `grimoire` | Managed runtime and model directory |
| `--backend <name>` | `auto` | `auto`, `cuda`, `vulkan`, or `cpu` llama.cpp runtime |
| `--force` | `false` | Revalidate and atomically reinstall the selected runtime |
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
| `--backend <name>` | `auto` | Runtime backend contract |
| `--host <address>` | `127.0.0.1` | Bind address |
| `--port <n>` | `9876` | Bind port |
| `--context-size <n>` | `8192` | Runtime context size |
| `--ubatch-size <n>` | `2048` | Runtime physical batch size |
| `--parallel <n>` | `4` | Concurrent llama.cpp server slots |
| `--gpu-layers <n>` | `-1` | Automatic all-GPU placement for CUDA/Vulkan, zero for CPU, or explicit layer count |

The command enables embedding mode and last-token pooling. Grimoire performs final 512-dimensional truncation and L2 normalization in its client.

## Managed runtime lifecycle

Start a detached supervised service:

```bash
grimoire model start [flags]
```

`model start` accepts the same runtime, model, backend, host, port, context, ubatch, parallel, and GPU-layer settings as `model serve`, plus:

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cache <path>` | user cache plus `grimoire` | Runtime config, state, stop marker, and default log root |
| `--startup-timeout <duration>` | `2m` | Time allowed for a verified embedding probe |
| `--restart-limit <n>` | `5` | Child crashes restarted before failure; zero disables restart |
| `--restart-delay <duration>` | `2s` | Delay before a crash restart |
| `--health-interval <duration>` | `15s` | Health-probe interval |
| `--log <path>` | managed log path | Combined supervisor and llama.cpp log |
| `--log-max-bytes <n>` | `16777216` | Rotation threshold |
| `--log-backups <n>` | `3` | Rotated log files retained |

The supervisor rejects duplicate live instances, verifies CUDA/Vulkan initialization from the runtime log, and writes an atomic state file containing process IDs, backend, model/runtime paths, context values, maximum accepted input tokens, readiness, restart count, and last error.

```bash
grimoire model status [--cache <path>] [--timeout 10s]
grimoire model stop [--cache <path>] [--timeout 30s]
grimoire model restart [start flags]
```

`model status` performs a real embedding probe and includes NVIDIA utilization, memory, temperature, power, graphics clock, and available thermal/power slowdown reasons. `model stop` uses the supervisor stop marker first and forcibly terminates stale managed processes only after the timeout.

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

## Arcana semantic graph index

Arcana uses the same running embedding endpoint; no second model installation or model process is required.

Build an index for the current `.arcana/CURRENT` graph snapshot:

```bash
arcana vectorize [--state .arcana] [--endpoint http://127.0.0.1:9876/v1] [--batch-size 32]
```

Search the graph index:

```bash
arcana semantic-query --query "where is profile persistence handled?" [--limit 10] [--json]
```

The index is stored under `.arcana/vectors/<snapshot-digest>/<embedding-identity>/`. Building is explicit. `grimoire context` uses a validated index matching the exact resolved Arcana snapshot when Arcana structural retrieval is enabled, but never builds the index as a query side effect. Missing, stale, corrupt, or concurrently invalidated vector state falls back to Lexicon-seeded deterministic Arcana traversal. `grimoire status` reports the matching default-model index as `current` only after validating its graph identity, embedding contract, data sizes, record count, and data checksums.

See [`../../arcana/docs/vector-index.md`](../../arcana/docs/vector-index.md).

## `grimoire query`

Run bounded, progressive repository queries without building a one-shot context
package or requiring vectors:

```bash
grimoire query <orient|search|trace|impact|inspect> [flags]
```

The common flags are `--root`, `--state`, `--query`, `--anchor`, `--target`,
`--limit`, `--depth`, `--direction`, repeatable `--relation`, repeatable
`--handle`, `--adjacent-context`, and the existing Lexicon/Arcana state and
command overrides. `--request <json>` accepts one complete
`grimoire.query.v1` request object. Results are JSON and every node or source
range carries a snapshot-qualified handle accepted by later `trace`, `impact`,
or `inspect` calls.

See [Agent query API](agent-query.md) for the schema and mode behavior.

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
| `--include-generated` | `false` | Include generated, vendored, lock, bundled, and minified content |
| `--exclude <path>` | none | Root-relative or absolute path to exclude; repeatable |
| `--lexicon-facts <path>` | automatic snapshot export | Explicit Lexicon JSONL export directory for semantic source spans |
| `--lexicon-state <path>` | `<root>/.lexicon` | Lexicon immutable state directory |
| `--lexicon-command <path>` | `lexicon` | Executable used for immutable snapshot export |

The command prepares Lexicon-aligned declaration chunks when current facts are available, fills every uncovered source region with deterministic fallback chunks, and stores exact token counts. Source chunks are not embedded; source retrieval is lexical, exact, and structural.

Output:

```json
{
  "state": "/absolute/path/to/repository/.grimoire",
  "files": 21,
  "chunking": "lexicon",
  "lexicon_snapshot": "sha256:...",
  "stats": {
    "scanned": 21,
    "reused": 20,
    "updated": 1,
    "removed": 0,
    "generated_skipped": 4,
    "semantic_files": 16,
    "semantic_chunks": 93
  }
}
```

## `grimoire vector build`

Build or refresh the optional documentation vector snapshot for the independent knowledge index:

```bash
grimoire knowledge index --root <repository>
grimoire vector build --root <repository> [flags]
```

`grimoire vector build` is an alias for `grimoire knowledge vector build`. It never embeds source-code chunks.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root |
| `--state <path>` | `<root>/.grimoire/knowledge` | Knowledge-state directory |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | OpenAI-compatible embeddings base URL |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--batch-size <n>` | `8` | Documentation sections per embedding request |
| `--batch-concurrency <n>` | `1` | Concurrent embedding requests; object ingestion remains serialized |
| `--timeout <duration>` | `30m` | Complete build timeout |

The builder deduplicates identical section text, reuses immutable content-addressed vector objects, persists successful batches immediately, and publishes a packed snapshot bound to the exact knowledge-index identity. A changed knowledge index makes the old snapshot stale; knowledge search then falls back to BM25 without embedding the query.

## `grimoire vector info`

Inspect documentation-vector availability and freshness:

```bash
grimoire vector info [--root <path>] [--state <path>] [--engine <path>]
```

The result reports the knowledge identity, expected identity, snapshot identity, model, dimensions, count, size, and whether the snapshot is current. Semantic documentation search is performed through `grimoire knowledge search`; there is no separate source-vector search command.

## `grimoire context`

Compile a bounded deterministic context package from prepared source and structural state:

```bash
grimoire context [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root used to resolve state |
| `--state <path>` | `<root>/.grimoire` | Prepared-state repository |
| `--query <text>` | none | Retrieval task; optional when `--diff` is set |
| `--diff <scope>` | none | `working-tree`, `staged`, `unstaged`, or one Git revision/range |
| `--diff-timeout <duration>` | `10s` | Complete Git diff and untracked-file discovery timeout |
| `--budget <n>` | `0` | Maximum `o200k_base` tokens; zero selects a deterministic automatic target |
| `--candidate-limit <n>` | `200` | Maximum merged exact, lexical, and structural candidates before curation |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | OpenAI-compatible embeddings base URL |
| `--structure <bool>` | `true` | Include available Lexicon and Arcana structural evidence |
| `--structure-timeout <duration>` | `30s` | Complete structural-provider timeout |
| `--lexicon-facts <path>` | automatic snapshot export | Explicit Lexicon JSONL export directory override |
| `--lexicon-state <path>` | `<root>/.lexicon` | Lexicon immutable state directory |
| `--lexicon-command <path>` | discovered | Executable override used for immutable snapshot export |
| `--arcana-state <path>` | `<root>/.arcana` | Arcana immutable graph-state directory |
| `--arcana-command <path>` | discovered | Executable override used for graph synchronization, semantic search, and protocol queries |

The command retrieves source evidence with deterministic BM25, targeted exact recovery, Lexicon facts, and Arcana graph evidence. Repository-wide source embeddings are not built or queried. Provider commands are resolved from explicit overrides, repository `.grimoire/providers.json`, adjacent installed executables, a discoverable Grimoire checkout, and finally `PATH`. Provider candidates are merged before deterministic query-shape analysis. When `--budget` is omitted or zero, focused queries select 3,000 tokens, bounded queries 6,000, and exploratory queries 12,000. A positive explicit budget bypasses automatic selection. Candidates are then deduplicated, diversified, and expanded with bounded prepared neighbours. Automatic assembly stops after deterministic evidence coverage is reached; the emitted package records the assembly decision. Explicit-budget requests retain the existing fit-to-budget behavior.

Diff-aware context treats changed prepared chunks as primary candidates and emits bounded `git-diff` structural evidence for every changed span. Changed paths, hunk headings, and declaration lines are added only to the internal retrieval query so Lexicon and Arcana can locate callers, dependencies, contracts, and tests; the package retains the human-facing query. `working-tree` compares tracked files with `HEAD` and also includes untracked, non-ignored files. `staged` compares `HEAD` with the index, `unstaged` compares the index with tracked working files, and any other value is passed to Git as one revision/range argument. A diff with no changed spans is an error rather than silently producing ordinary query context.

```bash
grimoire context --diff working-tree
grimoire context --diff HEAD~1 --query "review these changes for regressions"
grimoire context --diff main...HEAD --query "identify affected callers and missing tests"
```

Structural enrichment is enabled by default. When Lexicon state exists, Grimoire resolves `.lexicon/CURRENT`, creates or reuses a cached `lexicon export`, and emits matched symbols, source spans, and immediate relationships as first-class package evidence. It then resolves the Arcana snapshot for the same Lexicon ID, invokes one-shot `arcana sync` when necessary, and queries Arcana's JSONL protocol for operational roles, impact, unresolved references, and shortest call chains. The component executables are independently built from `lexicon/` and `arcana/` in this repository or discovered through the configured command paths. Structural failures warn and preserve source-only retrieval. Use `--structure=false` to skip both components or the explicit state, command, and facts flags to override discovery.

Documentation vectors are intentionally independent of context assembly. Missing or stale documentation vectors affect only the knowledge lane and never produce context warnings.

## `grimoire eval knowledge`

Run the checked-in documentation corpus against the production `internal/knowledge` search seam:

```bash
grimoire eval knowledge \
  --root . \
  --cases evaluation/knowledge/grimoire.json \
  --vectors=false
```

The command loads an existing `grimoire knowledge index`, executes every frozen case in corpus order, and writes a JSON report plus a Markdown review under `evaluation/results/`. Knowledge search always runs BM25. Vectors are disabled by default; pass `--vectors=true` for a paired supplemental-vector run. Vector failures are recorded per case while BM25 results remain scoreable.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cases <path>` | none | Frozen documentation corpus JSON |
| `--root <path>` | `.` | Repository being evaluated |
| `--state <path>` | `<root>/.grimoire/knowledge` | Prepared knowledge state |
| `--vectors` | `false` | Attempt optional documentation vectors as a BM25 supplement |
| `--top-k <n>` | corpus value | Override the corpus result limit |
| `--recall-at-k <list>` | corpus value | Override ordered cutoffs such as `1,3,5,10` |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | Embeddings endpoint for vector queries |
| `--engine <path>` | discovered DLL | Rust vector-engine library |
| `--timeout <duration>` | `2m` | Per-case knowledge search timeout |
| `--output-dir <path>` | `evaluation/results` | JSON and Markdown result directory |
| `--output-prefix <name>` | generated | Shared result filename prefix |

Reports include pass rate, required-section recall, recall@k, MRR, irrelevant selections, vector usage/errors, per-case latency, aggregate median/p95 latency, and deterministic per-case rankings. This path does not invoke source retrieval or the legacy source evaluator.

## `grimoire eval retrieval`

Run a repository-owned judged retrieval corpus against one or more query modes:

```bash
grimoire eval retrieval --cases <path> --root <repository> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--cases <path>` | none | Required judged corpus JSON |
| `--root <path>` | `.` | Repository being evaluated |
| `--state <path>` | `<root>/.grimoire` | Prepared source state |
| `--modes <list>` | `lexical` | Production deterministic source-retrieval mode |
| `--variant <name>` | `standalone` | Result label for paired comparisons |
| `--budget <n>` | case budget | Optional fixed budget override for every case |
| `--adaptive` | `false` | Replace case budgets with query-shape targets and evidence-coverage assembly |
| `--candidate-limit <n>` | `200` | Normal ranked candidate limit |
| `--probe-limit <n>` | `800` | Broader diagnostic ranking probe used only for failure attribution |
| `--selection-file-penalty <n>` | `10` | Evaluation-only curation penalty for each previously selected chunk from the same file |
| `--selection-subsystem-penalty <n>` | `18` | Evaluation-only curation penalty for each previously selected chunk from the same subsystem |
| `--selection-adjacent-primaries <n>` | `3` | Evaluation-only number of diversified primaries whose immediate prepared neighbors are promoted |
| `--compiler-facet-protection <bool>` | `true` | Protect query-facet implementation owners during final fitting |
| `--compiler-facet-file-depth <n>` | `2` | Distinct implementation files protected for eligible mechanism facets |
| `--compiler-companion-depth <n>` | `1` | Additional same-file chunks protected for each selected facet file |
| `--compiler-required-link-protection <bool>` | `true` | Retain complete provider-declared required source-link groups atomically |
| `--endpoint <url>` | `http://127.0.0.1:9876/v1` | Embeddings endpoint used only by optional Arcana semantic graph seeds |
| `--structural-providers <list>` | `none` | `none`, `lexicon`, or `lexicon,arcana` |
| `--structure-timeout <duration>` | `30s` | Per-case structural-provider timeout |
| `--lexicon-facts <path>` | automatic snapshot export | Explicit Lexicon JSONL export directory override |
| `--lexicon-state <path>` | `<root>/.lexicon` | Lexicon immutable state directory |
| `--lexicon-command <path>` | `lexicon` | Executable used for immutable snapshot export |
| `--arcana-state <path>` | `<root>/.arcana` | Arcana immutable graph-state directory |
| `--arcana-command <path>` | `arcana` | Executable used for graph synchronization and protocol queries |
| `--timeout <duration>` | `10s` | Per-case source-retrieval timeout |
| `--output-dir <path>` | `evaluation/results` | JSON and Markdown result directory |
| `--output-prefix <name>` | generated | Shared result filename prefix |

The corpus is separate from deterministic unit-test fixtures. A case may require source evidence, structural evidence, or both. Source expectations use `required`, `supporting`, and `forbidden`. Structural expectations use `required_structural`, `supporting_structural`, and `forbidden_structural`.

Structural expectations require `provider` and `kind`. Optional assertions include subject `symbol` and `path`, relationship `relation`, `direction`, and `certainty`, related `target_symbol` and `target_path`, an ordered `chain` subsequence, and unresolved-reference `expression`. Before retrieval, the runner verifies every referenced source path and any symbol paired with a path.

`--structural-providers none` runs the source-only baseline. `lexicon` executes immutable Lexicon export and symbol matching. `lexicon,arcana` additionally synchronizes and queries Arcana against the same snapshot. Arcana cannot be enabled without Lexicon because Lexicon-matched symbols are its bounded graph-query seeds.

For each case the runner records source and structural timings, provider warnings, selected source chunks, retained structural facts, immutable provider snapshots, final serialized package tokens, separate source and structural recall, separate irrelevant-evidence rates, and failure attribution. `--adaptive` also records the selected automatic budget, curated and assembled candidate counts, represented evidence coverage, and the assembly stop reason. Source and structural failures distinguish adaptive assembly loss from later budget-fitting loss. `--adaptive` cannot be combined with a fixed `--budget` override. The broad source-ranking probe does not contribute to reported context latency.

The three `--selection-*` flags substitute explicit values into the production curation implementation for judged experiments. They do not exist on `grimoire context`, and omitting them evaluates the current production defaults. This keeps calibration on the real algorithm rather than a parallel evaluator-only implementation.

Outputs are a machine-readable JSON report and a concise Markdown comparison grouped by category. Package comparison includes median and p95 tokens, median selected chunks, and median budget utilization. A case passes only when every required source and structural expectation survives into the final context package.

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
| `GRIMOIRE_EMBEDDING_MAX_TOKENS` | Override client-side per-input token limit; zero disables preflight enforcement |
| `GRIMOIRE_VECTOR_ENGINE` | Explicit Rust vector-engine DLL |

## Error behavior

Errors remain human-readable and do not yet have stable diagnostic codes or exit-code classes.

## Related documentation

- [Embedding model](embedding-model.md)
- [Vector store](vector-store.md)
- [Indexing](indexing.md)
- [Context package](context-package.md)
- [Prepared index](../architecture/prepared-index.md)
