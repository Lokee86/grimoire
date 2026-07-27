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
| Release concurrency bounds | `scripts/test_workflow.py` |

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

A fair Grimoire-versus-CBM comparison must use:

- the same repository revision and task wording;
- equivalent warm or cold state;
- the same agent model and completion criteria;
- all tool calls, source opens, tokens, and elapsed time;
- no free preassembled Grimoire context package.

Grimoire should be exercised through `search`, `inspect`, `trace`, and `impact`, beginning with the same information available to the CBM agent.

## Report interpretation

Checked-in reports are evidence for their exact corpus, repository revision, provider state, and date. They are not permanent product guarantees.

Do not compare source-lane scores directly with document, symbol, or relationship scores. Cross-provider scores are not globally calibrated. Compare end-to-end evidence coverage and agent outcomes instead.

Historical context-package reports remain useful for measuring the retired pipeline but must be labeled historical.
