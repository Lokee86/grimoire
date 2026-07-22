# Context Package

## Purpose

`grimoire context` emits a versioned, agent-independent JSON package containing selected repository chunks and an inspectable record of how they were retrieved.

The current package version is `3`.

## Example

```json
{
  "version": 3,
  "query": "where is player damage resolved",
  "budget": 2000,
  "tokenizer": "o200k_base",
  "token_count": 241,
  "index_version": 2,
  "retrieval_sources": [
    "vector"
  ],
  "selections": [
    {
      "path": "internal/game/damage/resolver.go",
      "start_line": 1,
      "end_line": 42,
      "score": 0.8125,
      "retrieval_source": "vector",
      "retrieval_rank": 1,
      "reasons": [
        "semantic vector similarity from split window 1/1"
      ],
      "token_count": 126,
      "content": "..."
    }
  ],
  "omitted_for_budget": 0
}
```

The values above illustrate the schema. The recorded `token_count` in real output is calculated from the complete emitted JSON document, including indentation and its trailing newline.

## Package fields

| Field | Type | Meaning |
| --- | --- | --- |
| `version` | integer | Context-package schema version; currently `3` |
| `query` | string | Original query supplied by the caller |
| `budget` | integer | Maximum `o200k_base` tokens permitted in the emitted package |
| `tokenizer` | string | Tokenizer used for chunk and package accounting; currently `o200k_base` |
| `token_count` | integer | Exact token count of the complete emitted JSON package |
| `index_version` | integer | Prepared-index format version |
| `retrieval_sources` | string array | Candidate sources involved in construction, ordered by first curated use; may include `exact`, `vector`, `lexical`, and `adjacent` |
| `selections` | object array | Ranked chunks retained under the complete package budget |
| `omitted_for_budget` | integer | Ranked candidates rejected because adding the complete chunk would exceed the package budget |

## Selection fields

| Field | Type | Meaning |
| --- | --- | --- |
| `path` | string | Repository-relative slash-normalized source path |
| `start_line` | integer | One-based inclusive source start line |
| `end_line` | integer | One-based inclusive source end line |
| `score` | number | Provider-native ranking score; vector candidates use their best similarity across query inputs, while adjacent expansion uses zero |
| `retrieval_source` | string | Source that produced the candidate: `exact`, `vector`, `lexical`, or `adjacent` |
| `retrieval_rank` | integer | One-based provider rank after deterministic same-provider query-vector merging and before curation; adjacent expansion uses zero |
| `reasons` | string array | Inspectable provider explanation |
| `token_count` | integer | Exact `o200k_base` count of the prepared chunk text |
| `content` | string | Exact prepared chunk text |

Scores from different providers are not declared comparable. Consumers should use `retrieval_source` and `retrieval_rank` when interpreting provider-native scores.

The selection-level count is useful evidence metadata, but it is not sufficient to calculate the package total. JSON syntax, escaped content, indentation, query text, paths, reasons, and all other package fields also consume tokens.

## Ordering

Each provider assigns its own deterministic order and rank. Exact recovery candidates are merged before vector or lexical candidates because they represent concrete literal evidence. Duplicate chunk IDs are retained once, with another-provider evidence recorded in the reasons.

Curation then:

1. removes later duplicate and overlapping ranges;
2. applies bounded soft file/subsystem diversity while preserving all unique primaries;
3. emits the first four diversified primaries;
4. emits their deduplicated immediate prepared neighbours; and
5. emits the remaining primaries.

Provider ranks are provenance only and are never compared across providers. Package budgeting preserves curated order but may omit a larger candidate and later retain a smaller one that still fits.

## Budget behavior

Grimoire never truncates a selected chunk. For each ranked candidate:

1. Append the complete selection to a tentative package.
2. Serialize the package using the same indented JSON format used for output.
3. Stabilize the self-referential package-level `token_count` field.
4. Count the exact serialized bytes with `o200k_base`.
5. Retain the selection only when the resulting package does not exceed `budget`.

A rejected candidate increments `omitted_for_budget`; later smaller candidates may still fit. The compiler performs a final count and removes the lowest-ranked retained selections if metadata changes would otherwise push the final package over budget.

A successful package always satisfies:

```text
token_count <= budget
```

If the budget cannot fit even the package metadata with no selections, the command returns an error and emits no package.

## Retrieval behavior

The normal path uses the configured embedding endpoint and exact vector snapshot. Before query embedding, Grimoire requires the persistent vector manifest's prepared identity to exactly match the current content-addressed prepared-index root. It then validates model identity, dimensions, vector count, and returned chunk IDs.

Fast mode retains the complete query, divides it into non-overlapping 16-token windows, groups windows into requests containing at most 64 query tokens, and runs at most two embedding requests concurrently. Full mode submits the complete query once. Quality mode submits the full query plus the bounded split-window requests. A nonzero `--query-max-tokens` value is an explicit optional limit; zero leaves the query untruncated. Duplicate vector hits retain the best similarity, record every matching query input in `reasons`, and receive one deterministic merged vector rank.

Concrete query signals—such as quoted phrases, paths, filenames, identifiers, configuration keys, error codes, and versions—also activate targeted exact recovery. Exact and semantic candidates are merged before deterministic curation and exact-budget fitting.

When semantic retrieval is unavailable or incompatible, `context` writes a warning to stderr and substitutes the deterministic lexical fallback. Targeted exact recovery and candidate curation still run. The output identifies actual candidate sources through `retrieval_sources` and each selection's provenance fields.

## Budget boundary

The budget covers exactly the JSON bytes emitted by Grimoire, including the trailing newline. It does not cover system prompts, chat-message framing, tool schemas, transport envelopes, or any other content added by a consumer after emission.

The count is exact for `o200k_base`. Models using another tokenizer may count the same package differently.

## Compatibility

Version 3 changes selection scores from integer-only lexical values to provider-native numbers and adds `retrieval_source` and `retrieval_rank` to every selection.

Version 2 introduced exact `token_count` fields and the fixed tokenizer name. Consumers should use `version` to select a package decoder and tolerate additional fields in future compatible revisions.

The schema is pre-release and may change before a stable Grimoire release.

## Related documentation

- [CLI](cli.md)
- [System overview](../architecture/system-overview.md)
- [Prepared index](../architecture/prepared-index.md)
- [Current limitations](../limits/current-limitations.md)
