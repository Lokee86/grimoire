# Grimoire MCP interface

`grimoire mcp` serves the unified discovery contract over stdio.

```bash
grimoire mcp --root .
```

The server exposes one tool: `grimoire_discover`.

## Normal workflow

1. Call `search` with the repository question.
2. Review the independent exact, source, document, symbol, and relationship lanes.
3. Use returned handles with `inspect`, `trace`, or `impact`.
4. Reuse one `session` name for the investigation so previously returned evidence is represented by compact prior handles rather than replayed.

`orient` is useful when the repository is unfamiliar, but agents should normally begin a concrete task with `search`.

## Input

The tool accepts the `grimoire.discovery.v1` fields documented in [Unified discovery contract](agent-query.md), plus automatic-state and session fields:

- `state_mode`: `current-only`, `refresh-if-needed`, or `force-refresh`;
- `session`: optional investigation-session name;
- `include_documents`: whether the separate documentation lane is returned;
- `use_document_vectors`: whether current documentation vectors augment document BM25 ranking.

Provider override fields are accepted for controlled environments, but agents should not choose between Grimoire, Lexicon, and Arcana. Grimoire owns that routing.

## Output

The MCP response is the same flattened discovery response used by the CLI:

- `exact_matches`
- `source_matches`
- `document_matches`
- `symbol_matches`
- `relationship_matches`
- mode-specific `paths`, `dependents`, or `inspections`
- `snapshot`, `preparation`, `warnings`, and `truncated_lanes`

When `session` is supplied, newly discovered evidence is returned through `delta`; repeated evidence is represented by prior handles.

## Evidence semantics

Source and documentation must be interpreted separately:

- Source is the authority for current implementation behavior.
- Documentation supplies intent, rationale, plans, or historical constraints.

Documentation is never inserted into source or symbol lanes and never consumes their result limits.

## State preparation

The MCP runtime aligns Grimoire, Lexicon, Arcana, and documentation state before querying when `state_mode` permits refresh. Missing or failed structural providers are reported as warnings. Source and document discovery continue when possible.

Vector state is never required for source, symbol, or relationship discovery. Documentation vectors are optional and are validated for freshness before use.
