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
        "semantic vector similarity"
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
| `retrieval_sources` | string array | Candidate providers used to construct this package; normally `vector`, or `lexical` after semantic fallback |
| `selections` | object array | Ranked chunks retained under the complete package budget |
| `omitted_for_budget` | integer | Ranked candidates rejected because adding the complete chunk would exceed the package budget |

## Selection fields

| Field | Type | Meaning |
| --- | --- | --- |
| `path` | string | Repository-relative slash-normalized source path |
| `start_line` | integer | One-based inclusive source start line |
| `end_line` | integer | One-based inclusive source end line |
| `score` | number | Provider-native ranking score; vector similarity for semantic retrieval or the fallback lexical score |
| `retrieval_source` | string | Provider that produced the candidate, currently `vector` or `lexical` |
| `retrieval_rank` | integer | One-based rank assigned by that provider before package budgeting |
| `reasons` | string array | Inspectable provider explanation |
| `token_count` | integer | Exact `o200k_base` count of the prepared chunk text |
| `content` | string | Exact prepared chunk text |

Scores from different providers are not declared comparable. Consumers should use `retrieval_source` and `retrieval_rank` when interpreting provider-native scores.

The selection-level count is useful evidence metadata, but it is not sufficient to calculate the package total. JSON syntax, escaped content, indentation, query text, paths, reasons, and all other package fields also consume tokens.

## Ordering

Vector selections preserve the exact native search order, including the vector engine's deterministic tie-breaking. Lexical fallback selections use:

1. descending lexical score;
2. ascending path; and
3. ascending source start line.

Package budgeting preserves candidate order but may omit a larger candidate and later retain a smaller one that still fits.

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

The normal path uses the configured embedding endpoint and exact vector snapshot. Before query embedding, Grimoire validates the snapshot model and chunk count. It validates dimensions after embedding and rejects returned IDs absent from prepared state.

When semantic retrieval is unavailable or incompatible, `context` writes a warning to stderr and emits a package built from the deterministic lexical fallback. The output package identifies the actual provider through `retrieval_sources` and each selection's provenance fields.

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
