# Agent discovery benchmark

This evaluation-only harness compares recorded progressive repository investigations. It does not run an agent or alter Grimoire behavior.

Corpus files use [`schema.v1.json`](schema.v1.json). The initial Space Rocks corpus is [`space-rocks.v1.json`](space-rocks.v1.json), pinned to its recorded revision.

Each case defines:

- the expected ownership boundary;
- required source and structural evidence;
- forbidden unsupported conclusions;
- completion criteria;
- known relevant branches.

Scores include correctness, required-evidence recall, input/output and repeated-input tokens, discovery calls, source opens, evidence timing, irrelevant branches, unsupported claims, and repeatability.

## Progressive recordings

Record complete Grimoire `search`, `inspect`, `trace`, and `impact` interactions as JSONL, then score them:

```powershell
go run ./evaluation/agent_discovery/cmd/agent-discovery `
  --cases evaluation/agent_discovery/space-rocks.v1.json `
  --adapter progressive-jsonl --input .\recordings\grimoire.jsonl `
  --output-dir evaluation\results --name grimoire-space-rocks
```

`progressive-jsonl` accepts one event per line with `adapter`, `run_id`, `case_id`, `time_ms`, `kind`, token usage, path/range/symbol, optional branch/relevance, and claims. A line may instead contain a complete `{events:[...]}` transcript.

`raw` accepts generic JSON or JSONL tool records. It maps common open/read tool names, path and line arguments, symbols, token usage, and claims.

## CBM comparison

CBM execution remains external. A CBM exporter can register its transcript adapter with:

```go
agentdiscovery.RegisterAdapter("cbm", adapter)
```

No CBM dependency is embedded in Grimoire.

A fair paired run uses the same repository revision, task, agent model, completion criteria, and warm/cold state. Grimoire receives no free preassembled context package; its discovery calls and source inspections are counted normally.

## Historical adapter

The `grimoire-context` adapter remains only to score frozen historical context-package artifacts. It is not a current execution mode and must be labeled historical in reports.

The runner writes stable JSON and Markdown reports. Multiple records with the same adapter and case are compared for repeatability.
