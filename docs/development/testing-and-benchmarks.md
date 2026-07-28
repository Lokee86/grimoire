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

The canonical end-to-end runner is:

```bash
python evaluation/run_agent_benchmark.py --help
python evaluation/run_agent_benchmark.py --check
```

`--check` validates the task catalogue and selected packaged dependencies without creating worktrees, profiles, indexes, or agent runs. Output and checkout paths are resolved to absolute paths before subprocess launch so usage and MCP audit files cannot drift into detached worktrees. `python evaluation/revalidate_agent_benchmark.py --task <task-id>` rechecks saved answers and rewrites grounding reports without rerunning agents, recreating and removing pinned worktrees when needed. `import_agent_benchmark_run.py` imports an intentionally isolated result root into the canonical suite summary. Successful runs remove CBM caches and the partial summary automatically; `cleanup_agent_benchmark.py` removes those transients after interrupted runs. Do not run the benchmark itself as part of ordinary verification. The task catalogue is `evaluation/agent_benchmark_tasks.v2.json`. It contains natural problem reports plus hidden rubrics for architectural exploration, unclear ownership, cross-language change, impact analysis, and source-plus-rationale investigation. The prompt receives only the problem report and one generic evidence envelope; rubric dimensions and expected ownership areas are not disclosed to the agent.

`evaluation/benchmark_grounding.py` validates every backticked `path:line` citation and every structured evidence path/range against the pinned checkout. Missing files, path traversal, line overruns, malformed or empty evidence, refusals, and nonzero process exits make the run invalid. When Grimoire evidence includes an inspected source-range handle, it must resolve through the MCP audit log to the same canonical path and lines. Handle coverage is measured separately so direct source verification remains valid. Hidden-rubric path-family coverage is also reported without changing factual grounding validity; semantic completeness remains a separate answer-quality judgment. Each condition writes `<condition>.grounding.json`, and the suite summary records `valid: false` for failed grounding even when the agent process completed.

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
- automatic citation and structured-deliverable validation against the pinned checkout;
- strict validation of any supplied Grimoire source-range handles against `grimoire.mcp.audit.v1` records, plus separate handle-coverage reporting;
- no free preassembled context package or hidden prepared answer.

### Task suitability

A mechanically fair comparison can still be a poor discovery benchmark. Prompts that name the subsystem, enumerate every expected ownership area, and provide the likely identifiers are specifically favorable to a strong model using `rg`. They pre-solve much of the discovery problem and measure checklist execution more than uncertainty reduction.

Benchmark task selection should distinguish two context regimes:

- **Small/direct working set:** exact identifiers, compact call chains, or prompts that already provide the search plan. Additional discovery results can create a lost-in-the-middle problem by competing with a small amount of obvious evidence.
- **Large/ambiguous working set:** unclear ownership, cross-language boundaries, transitive impact, generated contracts, conflicting source and documentation, or incomplete problem reports. Structured discovery can prevent lost-in-the-middle by reducing and organizing the context the model must retain.

This task-size inversion should be treated as a hypothesis to test explicitly. A useful suite needs both regimes and should include tasks whose starting vocabulary does not reveal the complete investigation plan. Citation validity, irrelevant branches opened, context volume, and whether key evidence survives into the final answer should be measured separately.

The version 2 runner records answer bytes and, for Grimoire, audited discovery response bytes, operation counts, and newly emitted evidence counts. Cold preparation summaries retain provider actions plus explicit timing buckets for Lexicon, Arcana, source indexing, documentation indexing, inspection, lock wait, marker overhead, and final source verification. This allows later optimization to target measured stages rather than total wall time alone.

Grimoire should be exercised through `search`, `inspect`, `trace`, and `impact`, while allowing the agent to use direct source inspection whenever it is cheaper. Do not require a minimum number of Grimoire calls.

## Report interpretation

Checked-in reports are evidence for their exact corpus, repository revision, provider state, and date. They are not permanent product guarantees.

Do not compare source-lane scores directly with document, symbol, or relationship scores. Cross-provider scores are not globally calibrated. Compare end-to-end evidence coverage and agent outcomes instead.

Historical context-package reports remain useful for measuring the retired pipeline but must be labeled historical.

Current end-to-end results and task-shape interpretation are summarized in [Agent benchmark findings](agent-benchmark-findings.md). The checked-in raw reports include the final [Space Rocks network-interest benchmark](../../evaluation/results/network-interest-agent-benchmark-2026-07-27-v4/report.md) and the completed [HikariCP/Detekt/Now in Android unfamiliar-repository benchmark](../../evaluation/results/multi-repo-agent-benchmark-2026-07-27-v1/report.md). Raw reports remain authoritative for exact conditions.
