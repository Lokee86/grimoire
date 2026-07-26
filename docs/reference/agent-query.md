# Agent query API

`grimoire query` is a progressive, JSON-first repository query interface. It
uses prepared source plus repository-local Lexicon and Arcana state; vectors are
not required. Evidence order is deterministic: literal source, Lexicon facts,
Arcana graph, then prepared-source BM25. `--code-only` excludes documentation
from the source and structural lanes so a separate knowledge consumer can own
repository rationale without duplicating it as code evidence.

## Modes

```bash
grimoire query orient --root .
grimoire query search --root . --query "SubmitLogin"
grimoire query trace --root . --anchor "<handle>" --depth 3
grimoire query impact --root . --anchor "<handle>" --direction incoming
grimoire query inspect --root . --handle "<handle>" --adjacent-context 3
```

- `orient` returns compact symbol, file, contract, test, and documentation
  anchors plus suggested `search`, `trace`, or `inspect` expansions.
- `search` combines exact, Lexicon, Arcana, and lexical results without source
  package assembly.
- `trace` returns behavior-ranked paths and Arcana unresolved alternatives.
  Summary detail is the default: containment is collapsed, equivalent paths are
  deduplicated, repeated entry-method variants are bounded, and each path
  returns a readable summary, relation list, exact evidence handles, and stable
  continuation handles. Use `--detail full` only when the complete node and
  step objects are required. Interstack `calls-endpoint`, `handled-by`,
  `publishes`, `consumes`, and `reads-config` relations are ordinary traversable
  edges.
- `impact` performs bounded incoming, outgoing, or bidirectional traversal with
  depth and repeatable `--relation` filters.
- `inspect` reads the exact prepared source selected by a handle, returning its
  declaration and containing span. `--adjacent-context` is bounded to 200
  lines. A supplied handle is never fuzzily rediscovered.

`trace` defaults to eight paths; other modes default to twelve. `--limit` is
bounded to 200 and `--depth` to 16. Query state discovery matches the existing
context workflow: `.grimoire`, `.lexicon`, and `.arcana` are found under
`--root` unless explicit state paths are supplied.

## Request and response contract

The request and response schema is `grimoire.query.v1`. A complete request can
be supplied as one JSON object:

```bash
grimoire query --request '{"schema":"grimoire.query.v1","mode":"search","root":".","query":"SubmitLogin","limit":8}'
```

Common request fields are `mode`, `root`, `state`, `query`, `anchor`, `target`,
`handles`, `limit`, `depth`, `direction`, `relations`, `adjacent_context`,
`code_only`, and `detail`. `detail` accepts `summary` or `full`. Provider
overrides are `lexicon_facts`, `lexicon_state`,
`lexicon_command`, `arcana_state`, and `arcana_command`.

Every returned node and source range includes a `handle`. Its `value` is the
string accepted by later calls, while its expanded fields expose provider,
snapshot, durable node identity or Arcana node ID, and normalized source range.
Source handles use the immutable prepared-index tree identity. Lexicon and
Arcana handles include their immutable snapshot ID when automatic state
discovery provides one. Handles are rejected if their snapshot does not match
the active state.

Responses use mode-specific arrays:

| Mode | Primary fields |
| --- | --- |
| `orient`, `search` | `results`, optionally `suggestions` |
| `trace` | `paths`, `unresolved`; compact paths contain `summary`, `relations`, `evidence`, and `continuation_handles` |
| `impact` | `dependents` |
| `inspect` | `inspections` |

All responses include `schema`, `mode`, and `snapshot`; degraded structural
providers are reported in `warnings`. The unified MCP runtime wraps this
contract directly and adds automatic state preparation, repository-knowledge
results, and investigation-session deltas.
