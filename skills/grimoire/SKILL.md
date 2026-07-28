---
name: grimoire
description: Use Grimoire's unified repository-discovery interface for architecture, ownership, call paths, impact, and cross-language code investigation. Use it alongside normal shell search, Git inspection, and direct file reads; prefer the cheapest reliable path.
---

# Grimoire repository discovery

Grimoire is a discovery aid, not a replacement for normal repository inspection. Use it when structural, cross-file, cross-language, documentation, or ownership relationships would otherwise require repeated search and reading. Use direct shell search or file reads when the task already has precise lexical anchors and they are cheaper.

## Default workflow

1. Start a concrete investigation with `grimoire_discover` in `search` mode. Use `orient` only when the repository is unfamiliar and the task has no useful search terms.
2. Reuse one short `session` name for the whole investigation. This lets Grimoire return compact deltas and prior handles instead of replaying evidence.
3. Keep the first request narrow: normally `limit: 4` to `8`. Set `include_documents: false` or `code_only: true` unless documentation is materially relevant.
4. Prefer `state_mode: current-only` when the repository state is known to be prepared. Use `refresh-if-needed` when state may be missing or stale. Do not use `force-refresh` unless the user explicitly asks for a rebuild or evidence proves the prepared state is invalid.
5. Follow returned stable handles with `inspect`, `trace`, or `impact`. Do not repeat broad searches when a handle already identifies the relevant source or symbol.
6. Verify material conclusions with exact source inspection. Source defines current behavior; documentation supplies intent, rationale, constraints, plans, or history.
7. Stop querying when the evidence is sufficient. Fall back to shell search and direct reads whenever they are faster or clearer.

## Mode selection

| Need | Mode |
| --- | --- |
| Find relevant implementation, symbols, documents, or direct relationships | `search` |
| Establish initial repository anchors without a concrete query | `orient` |
| Read exact source or document evidence from a returned handle | `inspect` |
| Follow a bounded structural path from a known handle | `trace` |
| Find bounded incoming, outgoing, or bidirectional dependents | `impact` |

## Efficient request patterns

Concrete code-first search:

```json
{
  "schema": "grimoire.discovery.v1",
  "mode": "search",
  "root": ".",
  "query": "camera visibility network interest realtime snapshot",
  "limit": 6,
  "code_only": true,
  "state_mode": "current-only",
  "session": "net-interest"
}
```

Inspect exact evidence:

```json
{
  "schema": "grimoire.discovery.v1",
  "mode": "inspect",
  "root": ".",
  "handles": ["<returned-handle>"],
  "adjacent_context": 3,
  "state_mode": "current-only",
  "session": "net-interest"
}
```

Trace only after discovering an anchor:

```json
{
  "schema": "grimoire.discovery.v1",
  "mode": "trace",
  "root": ".",
  "anchor": "<returned-handle>",
  "depth": 3,
  "limit": 6,
  "state_mode": "current-only",
  "session": "net-interest"
}
```

## Evidence handling

Search lanes are independent evidence classes:

- `exact_matches`: literal source matches for identifiers, paths, routes, and configuration keys.
- `source_matches`: ranked implementation ranges.
- `document_matches`: separately ranked documentation sections. These may be useful but stale.
- `symbol_matches`: Lexicon-grounded declarations and definitions.
- `relationship_matches`: direct Arcana or Lexicon relationships.

Do not treat documentation as proof of current implementation. Do not let repeated matches from multiple lanes inflate confidence; deduplicate them conceptually by file, symbol, and behavior.

## Completeness and negative claims

Before claiming that something has no callers, no dependents, no implementation, or no relevant files:

- inspect `warnings`, `preparation`, and `truncated_lanes`;
- expand all materially relevant truncated lanes or verify the bounded scope with normal repository search;
- account for unavailable or degraded providers;
- state unresolved limitations rather than presenting partial discovery as exhaustive.

## Practical rules

- Grimoire and normal shell/file tools may be mixed freely.
- Never choose between Grimoire's internal Lexicon and Arcana providers; Grimoire owns provider routing.
- Use `search` before `trace` when the exact structural handle is unknown.
- Use `impact` for transitive dependents, not repeated broad searches.
- Keep documentation disabled for implementation-only tasks.
- Avoid large result limits unless a smaller request demonstrably misses relevant evidence.
- Prefer one strong search followed by handle-based expansion over many loosely phrased searches.
- Report concrete files and symbols, not Grimoire metadata, unless preparation or coverage limitations affect confidence.
