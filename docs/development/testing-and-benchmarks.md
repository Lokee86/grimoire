# Testing and benchmarks

Verification is split by owning component and by discovery outcome.

## CPU-bounded root workflow

The root workflow defaults to one worker:

```bash
python scripts/workflow.py test
python scripts/workflow.py build --version dev
python scripts/workflow.py release --version 0.1.0
```

This constrains Go package concurrency, Go test parallelism, Cargo build jobs, and Rust test threads. Increase concurrency only explicitly:

```bash
python scripts/workflow.py test --jobs 2
```

Do not use a high `--jobs` value as a routine default. Component test suites may each contain many packages and test binaries.

The workflow smoke suite does not compile the full product:

```bash
python scripts/test_workflow.py
```

After a build, verify the actual combined-ZIP consumer path:

```bash
python scripts/test_installed_mcp.py --source build --version installed-smoke
```

This extracts the release bundle, runs its embedded installer, launches the installed MCP server from a clean temporary repository, refreshes managed provider state, and verifies a session handle through exact inspection and graph trace.

## Direct bounded Go verification

For focused Grimoire work:

```bash
GOMAXPROCS=1 go test -p 1 -parallel 1 ./internal/agentquery ./internal/agentruntime ./internal/app
GOMAXPROCS=1 go test -p 1 -parallel 1 ./...
```

## Lexicon and Arcana

Lexicon retains its owning Go tests under `lexicon/`. Arcana retains its owning Cargo tests under `arcana/`.

```bash
cd lexicon
go test -p 1 -parallel 1 ./...

cd ../arcana
cargo test --jobs 1 --all-targets --locked -- --test-threads 1
```

## Discovery contract tests

The active Grimoire contract is covered by:

| Concern | Tests |
| --- | --- |
| Independent exact, source, symbol, and relationship limits | `internal/agentquery/query_test.go` |
| Bounded source excerpts | `internal/agentquery/query_test.go` |
| Separate document lane and document-handle inspection | `internal/agentruntime/runtime_test.go` |
| Session deduplication for nodes, documents, relationships, and paths | `internal/agentruntime/*_test.go` and `internal/investigation/*_test.go` |
| Direct CLI commands and retired context command | `internal/app/run_test.go`, `internal/app/exact_context_test.go` |
| MCP schema and state preparation | `internal/app/*_test.go`, `internal/repostate/*_test.go` |
| Release concurrency bounds and bundled adapter installation | `scripts/test_workflow.py` |
| Installed release MCP, managed provider state, opaque inspect/trace handles | `scripts/test_installed_mcp.py` |

## Documentation retrieval evaluation

The independent document lane uses the checked-in knowledge evaluator:

```bash
grimoire eval knowledge --cases evaluation/knowledge/grimoire.json --root .
```

Record corpus revision, document-index identity, vector mode, model identity, top-k, and date.

## Arcana evaluation

Graph discovery and optional semantic entry points use:

```bash
grimoire eval arcana --cases evaluation/arcana/grimoire.json --root .
grimoire eval arcana --cases evaluation/arcana/space-rocks.json --root C:/!bin/workspace/space-rocks
```

Record Lexicon snapshot, Arcana snapshot, vector mode, model identity, and date.

## Agent discovery evaluation

`evaluation/agent_discovery` scores complete progressive investigation traces. It measures:

- required source and structural evidence found;
- ownership-boundary identification;
- unsupported conclusions;
- discovery calls;
- input and output tokens;
- latency to first required evidence and completion;
- irrelevant branches opened.

The evaluator accepts progressive JSONL and generic raw tool traces. External CBM adapters can be registered without coupling Grimoire to CBM.

A fair assisted-agent comparison must use:

- the same repository revision, clean checkout state, and task wording;
- equivalent warm or cold state, reported explicitly;
- the same agent model, normal shell/Git/file tools, and completion criteria;
- exactly one optional discovery system per assisted condition;
- the product's installed skill rather than ad hoc prompt instructions;
- all setup, refresh, discovery, direct-read, token, model-call, and elapsed costs;
- citation and structured-deliverable validation;
- no free preassembled context package or hidden prepared answer.

### Task suitability

A mechanically fair comparison can still be a poor discovery benchmark. Prompts that name the subsystem, enumerate every expected ownership area, and provide the likely identifiers are specifically favorable to a strong model using `rg`. They pre-solve much of the discovery problem and measure checklist execution more than uncertainty reduction.

Benchmark task selection should distinguish two context regimes:

- **Small/direct working set:** exact identifiers, compact call chains, or prompts that already provide the search plan. Additional discovery results can create a lost-in-the-middle problem by competing with a small amount of obvious evidence.
- **Large/ambiguous working set:** unclear ownership, cross-language boundaries, transitive impact, generated contracts, conflicting source and documentation, or incomplete problem reports. Structured discovery can prevent lost-in-the-middle by reducing and organizing the context the model must retain.

This task-size inversion should be treated as a hypothesis to test explicitly. A useful suite needs both regimes and should include tasks whose starting vocabulary does not reveal the complete investigation plan. Citation validity, irrelevant branches opened, context volume, and whether key evidence survives into the final answer should be measured separately.

Grimoire should be exercised through `search`, `inspect`, `trace`, and `impact`, while allowing the agent to use direct source inspection whenever it is cheaper. Do not require a minimum number of Grimoire calls.

## Report interpretation

Checked-in reports are evidence for their exact corpus, repository revision, provider state, and date. They are not permanent product guarantees.

Do not compare source-lane scores directly with document, symbol, or relationship scores. Cross-provider scores are not globally calibrated. Compare end-to-end evidence coverage and agent outcomes instead.

Historical context-package reports remain useful for measuring the retired pipeline but must be labeled historical.

Current end-to-end results and task-shape interpretation are summarized in [Agent benchmark findings](agent-benchmark-findings.md). The checked-in raw reports include the final [Space Rocks network-interest benchmark](../../evaluation/results/network-interest-agent-benchmark-2026-07-27-v4/report.md) and the completed [HikariCP/Detekt/Now in Android unfamiliar-repository benchmark](../../evaluation/results/multi-repo-agent-benchmark-2026-07-27-v1/report.md). Raw reports remain authoritative for exact conditions.
