# Unified discovery contract

Grimoire exposes one progressive repository-discovery interface over prepared source, repository documentation, Lexicon symbols, and Arcana relationships. Consumers do not select a provider. Grimoire routes each operation internally and returns provider provenance with every result.

The schema is `grimoire.discovery.v1`.

## Modes

```bash
grimoire orient --root .
grimoire search --root . --query "SubmitLogin"
grimoire trace --root . --anchor "<handle>" --depth 3
grimoire impact --root . --anchor "<handle>" --direction incoming
grimoire inspect --root . --handle "<handle>" --adjacent-context 3
```

`grimoire query <mode>` remains a compatibility spelling for the same interface.

- `orient` returns compact source and symbol anchors plus suggested expansions.
- `search` returns ranked discovery evidence. Balanced search preserves independent exact, source, document, and symbol budgets; narrow search applies one combined code-evidence budget and defers expansion.
- `trace` expands one stable structural handle through bounded paths.
- `impact` performs bounded incoming, outgoing, or bidirectional traversal.
- `inspect` reads exact source or document evidence by stable handle. Supplied handles are never fuzzily rediscovered.

## Search lanes

A search response may contain:

| Field | Meaning |
| --- | --- |
| `exact_matches` | Literal source matches for concrete identifiers, paths, routes, and configuration keys |
| `source_matches` | BM25-ranked implementation ranges |
| `document_matches` | Separately indexed documentation sections, including text, line ranges, freshness metadata, reasons, and code links |
| `symbol_matches` | Lexicon-grounded declarations and definitions |
| `relationship_matches` | Direct Arcana graph relationships, with Lexicon relationship fallback |

With `breadth: "balanced"`, `limit` applies independently to each lane. A full exact lane does not suppress source, symbol, relationship, or document results. With `breadth: "narrow"`, one combined limit is round-robin allocated across exact, symbol, and source evidence while overlapping cross-lane representations are suppressed. Narrow search defaults to four combined results.

Narrow search also defaults to `detail: "handles"`: results retain stable inspectable handles, paths, names, kinds, ranks, and reasons while inline excerpts and duplicated node spans are deferred to `inspect`. `truncated_lanes` identifies ranked evidence that remains available beyond the returned budget.

Documentation never appears in `exact_matches`, `source_matches`, or `symbol_matches`. Use `--code-only` or `include_documents: false` to omit the document lane entirely.

## Source and documentation semantics

Source and documentation are separate evidence classes:

- Source describes current executable behavior.
- Documentation describes intent, rationale, constraints, plans, or historical decisions.

A document result may be relevant while stale. Its path, line range, commit metadata, reasons, and stable `knowledge://` handle remain visible so the consumer can assess it independently rather than allowing it to displace implementation evidence.

## Relationship results

Each `relationship_matches` entry contains:

- `subject` and `object` nodes with stable handles;
- `direction` and typed `relation`;
- certainty where the provider distinguishes definite and possible edges;
- provider provenance;
- source spans and occurrence evidence when available.

The lane is intended for direct, useful relationships around discovered symbols. Longer paths belong in `trace`; transitive dependents belong in `impact`.

## Request fields

A complete request may be supplied as JSON:

```bash
grimoire query --request '{"schema":"grimoire.discovery.v1","mode":"search","root":".","query":"SubmitLogin","limit":8}'
```

Common fields:

- `mode`, `root`, `state`, and `state_mode`;
- `query`, `anchor`, `target`, and `handles`;
- `limit`, `depth`, `direction`, and `relations`;
- `adjacent_context`, `detail`, and search `breadth`;
- `code_only`, `include_documents`, and `use_document_vectors`;
- optional repository-provider state or executable overrides.

Public CLI and MCP search default to 12 results per lane in balanced mode and four combined results in narrow mode. Trace defaults to eight. `detail` accepts `handles`, `summary`, or `full`; narrow search defaults to `handles`. `limit` is bounded to 200, `depth` to 16, and adjacent inspection context to 200 lines.

## Handles

Every source range and structural node includes a stable snapshot-qualified handle. Document sections expose `knowledge://` handles.

- Source handles identify an exact prepared-index range.
- Lexicon handles identify a durable symbol in one immutable Lexicon snapshot.
- Arcana handles identify a durable node plus its snapshot-local graph ID.
- Document handles identify one exact indexed section.

Follow-up operations should use handles rather than repeating a broad search. Handles are rejected when their snapshot no longer matches active repository state.

## Other response fields

All responses include `schema`, `mode`, and `snapshot`. Depending on mode they may also contain:

- `paths` and `unresolved` for trace;
- `dependents` for impact;
- `inspections` for source inspection;
- `document_matches` for document inspection;
- `assessment`, `coverage`, `deferred_expansions`, `suggestions`, `warnings`, `preparation`, and investigation-session `delta` metadata.

`assessment` is conservative workflow guidance. It reports observed and missing `owner`, `control_flow`, `public_boundary`, and `tests` dimensions plus a `status` and `next_action`. It does not claim exhaustive correctness. Narrow session-backed search records discovery nodes and retrieval hits but does not materialize source ranges until `inspect`.

Provider degradation is reported in `warnings`; remaining lanes still return when possible.
