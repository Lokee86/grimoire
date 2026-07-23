# Context Package

## Purpose

`grimoire context` emits a versioned, agent-independent JSON package containing selected repository source plus bounded structural evidence from Lexicon and Arcana when their repository state is available.

The current package version is `5`.

## Example

```json
{
  "version": 5,
  "query": "trace player damage resolution",
  "budget": 2000,
  "tokenizer": "o200k_base",
  "token_count": 1184,
  "index_version": 2,
  "retrieval_sources": [
    "vector"
  ],
  "structural_sources": [
    "lexicon",
    "arcana"
  ],
  "structural_state": [
    {
      "provider": "lexicon",
      "snapshot": "sha256:..."
    },
    {
      "provider": "arcana",
      "snapshot": "sha256:..."
    }
  ],
  "assembly": {
    "scope": "focused",
    "candidates_considered": 5,
    "candidates_selected": 3,
    "candidate_tokens": 2460,
    "structural_considered": 2,
    "structural_selected": 2,
    "regions_represented": [
      "internal/game"
    ],
    "roles_represented": [
      "implementation",
      "verification"
    ],
    "stop_reason": "focused evidence coverage satisfied"
  },
  "structural_evidence": [
    {
      "provider": "lexicon",
      "kind": "symbol",
      "rank": 1,
      "score": 41,
      "reasons": [
        "query names Lexicon symbol ResolveDamage"
      ],
      "node": {
        "identity": "sha256:...",
        "kind": "function",
        "name": "ResolveDamage",
        "qualified_name": "damage.ResolveDamage",
        "path": "internal/game/damage/resolver.go",
        "span": {
          "path": "internal/game/damage/resolver.go",
          "start_line": 18,
          "end_line": 61
        }
      },
      "relationships": [
        {
          "direction": "outgoing",
          "relation": "calls",
          "certainty": "definite",
          "node": {
            "identity": "sha256:...",
            "kind": "function",
            "name": "ApplyShield",
            "path": "internal/game/damage/shield.go"
          }
        }
      ]
    },
    {
      "provider": "arcana",
      "kind": "operational_role",
      "rank": 1,
      "node": {
        "identity": "sha256:...",
        "node_id": 148,
        "kind": "function",
        "name": "ResolveDamage",
        "path": "internal/game/damage/resolver.go"
      },
      "summary": "ResolveDamage has 3 definite caller(s), 0 possible caller(s), 4 definite callee(s), 0 possible target(s), and 2 incoming reference(s)."
    }
  ],
  "selections": [
    {
      "path": "internal/game/damage/resolver.go",
      "start_line": 1,
      "end_line": 72,
      "score": 0.8125,
      "retrieval_source": "vector",
      "retrieval_rank": 1,
      "reasons": [
        "semantic vector similarity from split window 1/1"
      ],
      "token_count": 214,
      "content": "..."
    }
  ],
  "omitted_structural_for_budget": 0,
  "omitted_for_budget": 0
}
```

The values illustrate the schema. A real `token_count` is calculated from the complete indented JSON document, including its trailing newline.

## Package fields

| Field | Type | Meaning |
| --- | --- | --- |
| `version` | integer | Context-package schema version; currently `5` |
| `query` | string | Original query supplied by the caller |
| `budget` | integer | Maximum `o200k_base` tokens permitted in the emitted package |
| `tokenizer` | string | Tokenizer used for chunk and package accounting; currently `o200k_base` |
| `token_count` | integer | Exact token count of the complete emitted JSON package |
| `index_version` | integer | Prepared-index format version |
| `retrieval_sources` | string array | Source-chunk providers involved in construction, ordered by first curated use |
| `structural_sources` | string array | Structural providers retained in the package, ordered by first retained evidence |
| `structural_state` | object array | Immutable provider snapshot identities for retained structural evidence |
| `structural_evidence` | object array | Bounded Lexicon and Arcana facts retained under the package budget |
| `assembly` | object | Automatic query scope, coverage, selected/considered counts, and stop reason; omitted for explicit budgets |
| `selections` | object array | Ranked source chunks retained under the package budget |
| `omitted_structural_for_budget` | integer | Structural facts rejected because the complete fact did not fit |
| `omitted_for_budget` | integer | Source selections rejected because the complete chunk did not fit |

`structural_sources`, `structural_state`, and `structural_evidence` are omitted when no structural evidence is available or retained. A state entry is omitted when an explicit legacy export has no recoverable immutable snapshot identity.

## Structural evidence

Every structural item includes:

| Field | Meaning |
| --- | --- |
| `provider` | `lexicon` or `arcana` |
| `kind` | Concrete evidence shape |
| `rank` | One-based provider-local order before cross-provider interleaving |
| `score` | Optional provider-native relevance score |
| `reasons` | Inspectable explanation for inclusion |
| `node` | Subject symbol, when the evidence has one |
| `truncated` | Whether the provider bounded a larger result |

Current evidence kinds are:

| Kind | Provider | Payload |
| --- | --- | --- |
| `symbol` | Lexicon | Durable symbol identity, kind, qualified name, source span, and immediate incoming/outgoing relationships |
| `operational_role` | Arcana | Graph summary plus bounded callers and callees |
| `impact` | Arcana | Bounded transitive dependents with graph depth |
| `call_chain` | Arcana | Ordered shortest call-chain nodes and relations between matched symbols |
| `unresolved` | Arcana | Unresolved expressions, candidate metadata, reasons, relations, and source spans owned by a matched symbol |

Lexicon identities are durable across consumers. Arcana `node_id` values are snapshot-local; `structural_state` records the exact immutable snapshot that makes those IDs meaningful.

Relationships record `direction`, `relation`, and `certainty`. Definite and possible call edges remain distinct rather than being collapsed.

## Selection fields

| Field | Type | Meaning |
| --- | --- | --- |
| `path` | string | Repository-relative slash-normalized source path |
| `start_line` | integer | One-based inclusive source start line |
| `end_line` | integer | One-based inclusive source end line |
| `score` | number | Provider-native ranking score |
| `retrieval_source` | string | Source that produced the candidate, such as `exact`, `vector`, `lexical`, `lexicon`, or `adjacent` |
| `retrieval_rank` | integer | One-based provider rank before curation; adjacent expansion uses zero |
| `reasons` | string array | Inspectable provider explanation |
| `token_count` | integer | Exact `o200k_base` count of the prepared chunk text |
| `content` | string | Exact prepared chunk text |

Scores from different providers are not declared comparable. Consumers should use provider identity, rank, and reasons rather than comparing raw scores across providers.

## Structural-state resolution

Structural enrichment is enabled by default but remains optional.

When `<root>/.lexicon/CURRENT` exists, Grimoire resolves its immutable snapshot ID and creates or reuses a cached standalone export under the Grimoire state directory. The export is produced through `lexicon export`; Grimoire does not inspect Lexicon's mutable private library.

When Lexicon produced matched symbols, Grimoire resolves `<root>/.arcana/CURRENT`. If Arcana is missing or does not represent the same Lexicon snapshot, Grimoire invokes one-shot `arcana sync`. It then queries the matching immutable snapshot through `arcana protocol --snapshot`.

A structural-provider failure writes a warning to stderr and does not disable standalone source retrieval. `--structure=false` explicitly skips both providers. Explicit state, executable, and exported-facts paths are available through the context command flags.

## Ordering and budget behavior

Lexicon and Arcana evidence are interleaved while preserving provider-local order, so one provider cannot consume the entire structural section before the other is considered.

The compiler attempts evidence in this order:

1. highest-ranked structural fact;
2. highest-ranked source selection;
3. remaining interleaved structural facts; and
4. remaining curated source selections.

This guarantees that first-class structural data does not remain merely a ranking hint, while reserving an early opportunity for the implementation source itself. Complete facts and complete chunks are never truncated. An item that does not fit is omitted, and later smaller items may still be retained.

For every tentative addition, Grimoire serializes the complete package, stabilizes the self-referential `token_count`, and measures the exact bytes using `o200k_base`. A successful package always satisfies:

```text
token_count <= budget
```

If the budget cannot fit package metadata with no evidence or selections, the command returns an error and emits no package.

## Source retrieval behavior

The normal source path uses the configured embedding endpoint and exact vector snapshot. Before query embedding, Grimoire requires the vector manifest's prepared identity to match the current content-addressed prepared-index root. It then validates model identity, dimensions, vector count, and returned chunk IDs.

Fast mode divides the complete query into non-overlapping windows and sends bounded requests. Full mode submits the complete query once. Quality mode submits both forms. Concrete repository literals also activate targeted exact recovery. Exact, semantic or fallback, and Lexicon-derived source candidates are merged before deterministic curation.

When semantic retrieval is unavailable or incompatible, `context` writes a warning to stderr and substitutes deterministic lexical retrieval. Structural enrichment can still run independently.

## Compatibility

Version 5 adds deterministic automatic budgeting and assembly metadata. Automatic requests may intentionally stop below their selected maximum after scope-specific evidence coverage is satisfied. Explicit-budget requests omit `assembly` and retain fit-to-budget behavior.

Version 4 adds first-class structural evidence, structural-provider provenance, and separate structural budget omissions. It also changes runtime behavior by automatically using matching Lexicon and Arcana repository state when available.

Version 3 introduced provider-native selection scores and source provenance. Version 2 introduced exact token-count fields and the fixed tokenizer name. Consumers should select a decoder using `version` and tolerate additional fields in future compatible revisions.

The schema remains pre-release.

## Related documentation

- [CLI](cli.md)
- [System overview](../architecture/system-overview.md)
- [Prepared index](../architecture/prepared-index.md)
- [Current limitations](../limits/current-limitations.md)
