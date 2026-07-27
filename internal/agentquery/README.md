# Agent-query package

`internal/agentquery` owns Grimoire's provider-neutral, snapshot-qualified source and structural discovery contract.

## Modes

- `orient` returns compact source and symbol anchors.
- `search` returns independently limited exact, source, symbol, and relationship lanes.
- `trace` expands one stable structural handle through bounded paths.
- `impact` returns bounded dependents and affected source evidence.
- `inspect` resolves exact source and structural handles without repeating broad discovery.

Documentation is intentionally not owned here. `internal/agentruntime` adds the separate document lane from `internal/knowledge`.

## Responsibilities

- Define the `grimoire.discovery.v1` request and response schemas.
- Preserve independent lane limits so exact, lexical, symbol, and relationship evidence cannot suppress one another.
- Produce stable handles qualified by prepared-source, Lexicon, or Arcana snapshot identity.
- Resolve source ranges against prepared chunks and structural nodes against provider snapshots.
- Preserve provider provenance, certainty, relationship direction, evidence sites, and lane-specific truncation.
- Use Arcana for direct graph relationships and Lexicon as the structural fallback.
- Keep discovery progressive so callers expand returned evidence rather than receiving a preassembled package.

## Boundary

This package does not prepare repository state, persist investigation history, index documentation, or own Lexicon/Arcana semantics. `internal/agentruntime` prepares state and combines the independent lanes; `internal/investigation` records returned evidence; provider packages own analysis and graph storage.
