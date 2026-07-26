# Agent-query package

`internal/agentquery` owns Grimoire's progressive, snapshot-qualified repository query contract.

## Modes

- `orient` returns compact repository anchors and role summaries.
- `search` performs code-first deterministic retrieval over prepared source and available structural evidence.
- `trace` expands one exact node or source handle through bounded relationships and call paths.
- `impact` returns bounded dependents and affected source evidence.
- `inspect` resolves exact source, node, and relationship handles without repeating broad discovery.

## Responsibilities

- Define the versioned request, response, node, edge, range, warning, and handle schemas.
- Produce stable handles qualified by prepared, Lexicon, or Arcana snapshot identity.
- Resolve source ranges against prepared chunks and structural nodes against provider snapshots.
- Preserve provider provenance, certainty, relationship direction, evidence sites, and bounded truncation.
- Keep query modes progressive so callers can expand prior evidence instead of rebuilding a monolithic context package.

## Boundary

This package does not prepare repository state, persist investigation history, index documentation, or own Lexicon/Arcana graph semantics. `internal/agentruntime` prepares state and combines code and knowledge lanes; `internal/investigation` records returned evidence; provider packages own structural retrieval.
