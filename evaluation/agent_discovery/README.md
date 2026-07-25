# Agent discovery benchmark

This evaluation-only harness compares recorded repository discovery work. It does not run an agent or alter Grimoire query behavior. Corpus files use [`schema.v1.json`](schema.v1.json); the initial Space Rocks corpus is [`space-rocks.v1.json`](space-rocks.v1.json), pinned to revision `ff882f636706e4917f86e156ce1ed7f40b467e83`.

A case records the expected ownership boundary, required source and structural evidence, explicit unsupported conclusions, and completion criteria. Scores record input/output and repeated-input tokens, discovery/tool calls, opened source ranges, evidence timing, irrelevant branches, unsupported claims, correctness, and repeatability across equivalent runs.

Run a generic progressive-query recording:

```powershell
go run ./evaluation/agent_discovery/cmd/agent-discovery `
  --cases evaluation/agent_discovery/space-rocks.v1.json `
  --adapter progressive-jsonl --input .\recordings\progressive.jsonl `
  --output-dir evaluation\results --name progressive-space-rocks
```

Score a `grimoire context` JSON package for one case:

```powershell
grimoire context --root C:\!bin\workspace\space-rocks --query "Find config key readers" > context.json
go run ./evaluation/agent_discovery/cmd/agent-discovery `
  --cases evaluation/agent_discovery/space-rocks.v1.json `
  --adapter grimoire-context --input context.json `
  --case space-rocks-config-key-readers
```

`raw` accepts JSONL/JSON raw tool records. It maps `tool=open_file` or `read_*`, `arguments.path/start_line/end_line/symbol`, and `usage.input_tokens/output_tokens`. `progressive-jsonl` accepts one event per line with `adapter`, `run_id`, `case_id`, `time_ms`, `kind`, token usage, path/range/symbol, optional branch/relevance, and claims. A line may instead contain a complete `{events:[...]}` transcript.

CBM execution is external. A CBM exporter can ingest its own transcript shape by importing this package and calling `agentdiscovery.RegisterAdapter("cbm", adapter)`; no CBM dependency is embedded here. The runner writes a stable JSON comparison report and matching Markdown table. Multiple records with the same adapter and case are compared for score repeatability.
