# Grimoire CLI

## Invocation

```bash
grimoire <command> [flags]
```

Run `grimoire help` for the installed command summary.

## Discovery commands

### `grimoire search`

Search all repository evidence lanes through one interface:

```bash
grimoire search --root . --query "Where is session creation handled?"
```

The response uses schema `grimoire.discovery.v1` and may include:

- `exact_matches`
- `source_matches`
- `document_matches`
- `symbol_matches`
- `relationship_matches`

`--limit` applies to each lane independently. The default is 12 and the maximum is 200.

Useful flags:

```text
--root <path>                 Repository root
--state <path>                Grimoire state directory
--state-mode <mode>           current-only, refresh-if-needed, or force-refresh
--query <text>                Literal, symbol, behavior, or documentation query
--limit <n>                   Maximum results per lane
--code-only                   Omit the documentation lane
--include-documents=<bool>    Include separately ranked documentation
--document-vectors            Use current documentation vectors when available
--session <name>              Reuse one investigation ledger
--timeout <duration>          Complete operation timeout
```

Provider-state and executable overrides exist for controlled environments, but normal callers should allow Grimoire to discover and route Lexicon and Arcana internally.

### `grimoire orient`

Return compact source and symbol anchors for an unfamiliar repository:

```bash
grimoire orient --root .
```

Orient uses the same response schema and returns suggested follow-up operations. A concrete repository question should normally begin with `search` instead.

### `grimoire inspect`

Read exact evidence by stable handle:

```bash
grimoire inspect --root . --handle '<handle>'
grimoire inspect --root . --handle '<handle>' --adjacent-context 3
```

Source, Lexicon, and Arcana handles are snapshot-qualified. Documentation uses `knowledge://` section handles. Inspection does not fuzzily rerun the original query.

`--adjacent-context` applies to source inspection and is bounded to 200 lines.

### `grimoire trace`

Expand bounded graph paths from one returned symbol or relationship node:

```bash
grimoire trace --root . --anchor '<handle>' --depth 4
```

Trace defaults to eight paths. `--detail summary` is the default; `--detail full` returns complete node and step objects. Optional `--target`, `--direction`, and repeated `--relation` filters narrow traversal.

### `grimoire impact`

Find bounded incoming, outgoing, or bidirectional dependents:

```bash
grimoire impact --root . --anchor '<handle>' --direction incoming --depth 4
```

### Compatibility spelling

The former subcommand shape remains accepted:

```bash
grimoire query search --root . --query "SubmitLogin"
```

It executes the same discovery path. New integrations should use the direct commands.

### JSON requests

A complete request may be supplied as JSON:

```bash
grimoire query --request '{"schema":"grimoire.discovery.v1","mode":"search","root":".","query":"SubmitLogin","limit":8}'
```

See [Unified discovery contract](agent-query.md).

## `grimoire mcp`

Serve the same discovery interface over stdio:

```bash
grimoire mcp --root .
```

The exposed tool is `grimoire_discover`. The MCP server adds automatic repository preparation and optional investigation-session deduplication but does not change the evidence-lane contract.

See [Grimoire MCP interface](agent-mcp.md).

## Repository preparation

### `grimoire status`

Inspect current source, documentation, Lexicon, Arcana, and optional vector state:

```bash
grimoire status --root .
grimoire status --root . --refresh
grimoire status --root . --force-refresh
```

Normal discovery commands use `refresh-if-needed` unless another state mode is selected.

### `grimoire index`

Prepare source state directly:

```bash
grimoire index --root .
```

Useful flags:

```text
--state <path>
--ignore-file <path>
--exclude <path>              Repeatable
--max-file-bytes <n>
--include-generated
--lexicon-facts <path>
--lexicon-state <path>
--lexicon-command <path>
```

The source index owns exact and BM25 source discovery. Documentation is filtered from source results and indexed independently.

### `grimoire knowledge index`

Prepare the document lane:

```bash
grimoire knowledge index --root .
```

The standalone `knowledge search` and `knowledge inspect` commands remain available for diagnostics, but normal agents should use `grimoire search` and `grimoire inspect`.

## Documentation vectors

### `grimoire vector build`

Build or refresh optional document vectors:

```bash
grimoire vector build --root .
```

Vectors affect only `document_matches`. Exact, source, symbol, relationship, trace, and impact operations do not require them.

### `grimoire vector info`

Inspect document-vector freshness and storage:

```bash
grimoire vector info --root .
```

A stale or missing vector snapshot falls back to document BM25 before embedding the query.

## Embedding runtime

### `grimoire model setup`

Install or validate the managed embedding runtime:

```bash
grimoire model setup
```

### `grimoire model start`

Start the managed embedding service:

```bash
grimoire model start
```

### `grimoire model info`

Inspect runtime and model state:

```bash
grimoire model info
```

### `grimoire model serve`

Run the managed service in the foreground:

```bash
grimoire model serve
```

### `grimoire model probe`

Probe the configured embedding endpoint:

```bash
grimoire model probe
```

## Investigation sessions

Create, inspect, or close a persistent ledger explicitly:

```bash
grimoire investigation create --root . --session task-1 --snapshot <source-id>
grimoire investigation inspect --root . --session task-1
grimoire investigation close --root . --session task-1
```

Normal CLI and MCP discovery can create and reuse the ledger automatically when `--session` or `session` is supplied.

## Evaluation

### `grimoire eval knowledge`

Evaluate the independent documentation retriever against a frozen corpus:

```bash
grimoire eval knowledge --cases evaluation/knowledge/cases.json --root .
```

### `grimoire eval arcana`

Evaluate Arcana graph retrieval and optional semantic graph entry points:

```bash
grimoire eval arcana --cases evaluation/arcana/cases.json --root .
```

### Agent discovery evaluation

The repository-owned agent-discovery evaluator scores progressive discovery traces and end-to-end investigation outcomes. See [Testing and benchmarks](../development/testing-and-benchmarks.md) and `evaluation/agent_discovery/README.md`.

The former context-package command and package-focused retrieval evaluator are retired from the normal product workflow.

## `grimoire version`

```bash
grimoire version
```

Release builds override the development version through linker flags.

## Environment variables

Provider and embedding configuration is documented with the owning component. Common discovery does not require environment variables when the consolidated binaries are installed together.

`GRIMOIRE_HOME` may identify the consolidated checkout for provider discovery. Repository-local `.grimoire/providers.json` can pin Lexicon and Arcana commands.

## Error behavior

- Invalid requests fail before querying.
- Stale handles fail rather than being fuzzily rediscovered.
- Missing structural providers produce warnings while source and document discovery continue when possible.
- Missing or stale documentation vectors fall back to BM25.
- A repository that cannot produce current source state fails discovery instead of returning stale implementation evidence.

## Related documentation

- [Unified discovery contract](agent-query.md)
- [Grimoire MCP interface](agent-mcp.md)
- [System overview](../architecture/system-overview.md)
- [Current limitations](../limits/current-limitations.md)
