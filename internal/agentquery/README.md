# Agent-query package

`internal/agentquery` owns Grimoire's provider-neutral, snapshot-qualified source and structural discovery contract.

## Modes

- `orient` returns compact source and symbol anchors.
- `search` defaults to balanced, independently limited exact, source, and symbol lanes; `breadth: narrow` applies one combined evidence budget across those lanes. Both report coverage and defer graph expansion.
- `trace` expands one stable structural handle through bounded paths.
- `impact` returns bounded dependents and affected source evidence.
- `inspect` resolves exact source and structural handles without repeating broad discovery.

Documentation is intentionally not owned here. `internal/agentruntime` adds the separate document lane from `internal/knowledge`.

## Responsibilities

- Define the `grimoire.discovery.v1` request and response schemas.
- Preserve independent lane limits for balanced discovery so exact, lexical, symbol, and documentation evidence cannot suppress one another.
- Provide an explicit narrow-search path that round-robins exact, symbol, and source evidence under one combined limit and suppresses overlapping cross-lane representations.
- Default narrow discovery to handle-only results so exact source expansion occurs through `inspect` rather than being repeated in search.
- Return conservative evidence assessments that identify observed and missing owner, control-flow, public-boundary, and test dimensions without claiming exhaustive correctness.
- Produce stable handles qualified by prepared-source, Lexicon, or Arcana snapshot identity.
- Resolve source ranges against prepared chunks and structural nodes against provider snapshots.
- Preserve provider provenance, certainty, relationship direction, evidence sites, and lane-specific truncation.
- Defer graph traversal from broad search; callers use a selected stable handle with `trace` or `impact`.
- Use Arcana for explicit graph traversal and Lexicon as the structural fallback.
- Report returned, deferred, and duplicate-suppression counts so compact discovery does not silently hide coverage.
- Keep discovery progressive so callers expand returned evidence rather than receiving a preassembled package.
- Preserve balanced search behavior for broad investigations while allowing narrow session deltas to defer source ranges until inspection.

## Boundary

This package does not prepare repository state, persist investigation history, index documentation, or own Lexicon/Arcana semantics. `internal/agentruntime` prepares state and combines the independent lanes; `internal/investigation` records returned evidence; provider packages own analysis and graph storage.
