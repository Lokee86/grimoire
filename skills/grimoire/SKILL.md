---
name: grimoire
description: Use Grimoire's unified repository-discovery interface for architecture, ownership, call paths, impact, and cross-language code investigation. Use it alongside normal shell search, Git inspection, and direct file reads; prefer the cheapest reliable path.
---

# Grimoire repository discovery

Grimoire is a discovery aid, not a replacement for normal repository inspection. Use it when structural, cross-file, cross-language, documentation, or ownership relationships would otherwise require repeated search and reading. Use direct shell search or file reads when the task already has precise lexical anchors and they are cheaper.

## Route before retrieval

Choose the cheapest reliable path before calling Grimoire:

| Known state | First action |
| --- | --- |
| Exact file and symbol are known | Read the file directly. |
| Exact symbol or literal is known, but not the file | Use normal repository search, then read the matching file. |
| Likely ownership is localized but uncertain | Use one `breadth: "narrow"` search. |
| Ownership is distributed, cross-language, cross-service, or documentation-dependent | Use `breadth: "balanced"`. |
| Repository is unfamiliar and the task has no useful terms | Use `orient`, then a concrete search. |

Do not call Grimoire merely to confirm evidence already established by direct inspection. Escalate from direct tools to narrow search, and from narrow to balanced search, only when a named material question remains unresolved.

## Default workflow

1. Apply the routing table above.
2. Reuse one short `session` name for the whole Grimoire investigation. This lets Grimoire return compact deltas and prior handles instead of replaying evidence.
3. Set `include_documents: false` or `code_only: true` unless documentation is materially relevant.
4. Prefer `state_mode: current-only` when the repository state is known to be prepared. Use `refresh-if-needed` when state may be missing or stale. Do not use `force-refresh` unless the user explicitly asks for a rebuild or evidence proves the prepared state is invalid.
5. Treat search as discovery only. Narrow search defaults to handle-only results; inspect selected handles to read exact evidence.
6. Use `trace` or `impact` only for a named unresolved relationship. Do not expand merely because expansion is available.
7. Verify material conclusions with exact source inspection. Source defines current behavior; documentation supplies intent, rationale, constraints, plans, or history.
8. Stop querying when the evidence is sufficient. Fall back to shell search and direct reads whenever they are faster or clearer.

## Narrow-task workflow

Use `breadth: "narrow"` for localized ownership, predicate, API-boundary, or small impact questions where a few symbols are likely to contain the answer. Narrow search applies one combined evidence budget across exact, symbol, and source lanes instead of returning the full limit from every lane.

For a narrow task:

1. Prefer direct shell or file inspection first when exact paths or symbols are already known.
2. Otherwise make one `search` request with `breadth: "narrow"`, normally leaving the default `limit: 4`, disabling documents unless they are required, and reusing a session.
3. Inspect only the returned handles needed to establish the owning type or function, controlling predicate or transition, caller-visible boundary, and relevant tests.
4. Stop after the first search and targeted inspection when those four dimensions are established. Do not issue another broad search merely to collect corroborating evidence.
5. A third Grimoire call requires naming the specific unresolved behavior, boundary, or failure mode that the call must answer. Use direct reads instead when the relevant files are already known.

A narrow result is not evidence that the repository was exhaustively searched. Expand or switch to `balanced` only when a material dimension remains unresolved.

## Evidence assessment

Responses include an `assessment` object with conservative workflow guidance:

- `status`: whether discovery is empty, partial, ready for targeted inspection, or ready to synthesize.
- `observed_dimensions`: evidence already present for owner, control flow, public boundary, and tests.
- `missing_dimensions`: dimensions not yet represented in the returned evidence.
- `next_action`: the smallest justified continuation.

Treat `assessment` as a stopping aid, not proof of correctness. When it reports `ready-to-synthesize`, synthesize rather than searching for redundant corroboration. When it reports `ready-for-targeted-inspection`, inspect selected handles; do not issue another broad search. When dimensions are missing, obtain only those required by the actual task.

## Mode selection

| Need | Mode |
| --- | --- |
| Find relevant implementation, symbols, documents, or direct relationships | `search` |
| Establish initial repository anchors without a concrete query | `orient` |
| Read exact source or document evidence from a returned handle | `inspect` |
| Follow a bounded structural path from a known handle | `trace` |
| Find bounded incoming, outgoing, or bidirectional dependents | `impact` |

## Efficient request patterns

Narrow ownership search:

```json
{
  "schema": "grimoire.discovery.v1",
  "mode": "search",
  "root": ".",
  "query": "automatic background compaction scheduling manual compaction",
  "breadth": "narrow",
  "detail": "handles",
  "code_only": true,
  "include_documents": false,
  "state_mode": "current-only",
  "session": "compaction-owner"
}
```

Balanced code-first search:

```json
{
  "schema": "grimoire.discovery.v1",
  "mode": "search",
  "root": ".",
  "query": "camera visibility network interest realtime snapshot",
  "breadth": "balanced",
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
- Use `breadth: "narrow"` when the task should fit within one owner, one controlling mechanism, one public boundary, and nearby tests.
- Prefer one strong search followed by exact handle inspection over many loosely phrased searches.
- Treat search as discovery and `inspect` as evidence expansion; do not request full inline search evidence unless explicitly necessary.
- For narrow work, do not make more than two Grimoire calls unless a named material question remains unresolved.
- Report concrete files and symbols, not Grimoire metadata, unless preparation or coverage limitations affect confidence.
