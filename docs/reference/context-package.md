# Context Package

## Purpose

`grimoire context` emits a versioned, agent-independent JSON package containing selected source chunks and an inspectable record of why they were included.

The current package version is `2`.

## Exact empty-package example

This is an actual version-2 package emitted with no matching candidates:

```json
{
  "version": 2,
  "query": "zzzzunlikelyterm",
  "budget": 200,
  "tokenizer": "o200k_base",
  "token_count": 83,
  "index_version": 2,
  "retrieval_sources": [
    "lexical"
  ],
  "selections": [],
  "omitted_for_budget": 0
}
```

The recorded `token_count` includes the complete indented JSON document and its trailing newline.

## Package fields

| Field | Type | Meaning |
| --- | --- | --- |
| `version` | integer | Context-package schema version; currently `2` |
| `query` | string | Original query supplied by the caller |
| `budget` | integer | Maximum `o200k_base` tokens permitted in the emitted package |
| `tokenizer` | string | Tokenizer used for chunk and package accounting; currently `o200k_base` |
| `token_count` | integer | Exact token count of the complete emitted JSON package |
| `index_version` | integer | Prepared-index format version; currently `2` |
| `retrieval_sources` | string array | Candidate providers that contributed; currently only `lexical` |
| `selections` | object array | Ranked chunks retained under the complete package budget |
| `omitted_for_budget` | integer | Ranked candidates rejected because adding the complete chunk would exceed the package budget |

## Selection fields

| Field | Type | Meaning |
| --- | --- | --- |
| `path` | string | Repository-relative slash-normalized source path |
| `start_line` | integer | One-based inclusive source start line |
| `end_line` | integer | One-based inclusive source end line |
| `score` | integer | Fixed lexical relevance score |
| `reasons` | string array | Inspectable score contributions |
| `token_count` | integer | Exact `o200k_base` count of the prepared chunk text |
| `content` | string | Exact prepared chunk text |

The selection-level count is useful evidence metadata, but it is not sufficient to calculate the package total. JSON syntax, escaped content, indentation, query text, paths, reasons, and all other package fields also consume tokens.

## Ordering

Selections preserve candidate ranking order. Current candidate ordering is:

1. descending score;
2. ascending path; and
3. ascending source start line.

Only candidates with a positive score are considered.

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

## Budget boundary

The budget covers exactly the JSON bytes emitted by Grimoire, including the trailing newline. It does not cover system prompts, chat-message framing, tool schemas, transport envelopes, or any other content added by a consumer after emission.

The count is exact for `o200k_base`. Models using another tokenizer may count the same package differently.

## Scoring reasons

Current reason strings may include:

```text
exact query phrase in content
filename matches <term>
path matches <term>
leading line matches <term>
content matches <term>
```

Reason strings are inspectable output, but they are not yet declared a stable compatibility API.

## Compatibility

Version 2 replaces the version-1 `estimated_tokens` fields with exact `token_count` fields and adds `tokenizer`. Consumers should use `version` to select a package decoder and tolerate additional fields in future compatible revisions.

The schema is pre-release and may change before a stable Grimoire release.

## Related documentation

- [CLI](cli.md)
- [System overview](../architecture/system-overview.md)
- [Prepared index](../architecture/prepared-index.md)
- [Current limitations](../limits/current-limitations.md)
