# Context Package

## Purpose

`grimoire context` emits a versioned, agent-independent JSON package containing selected source chunks and an inspectable record of why they were included.

The current package version is `1`.

## Example

```json
{
  "version": 1,
  "query": "resolve damage",
  "budget": 100,
  "estimated_tokens": 18,
  "index_version": 1,
  "retrieval_sources": [
    "lexical"
  ],
  "selections": [
    {
      "path": "internal/damage/resolver.go",
      "start_line": 1,
      "end_line": 4,
      "score": 24,
      "reasons": [
        "filename matches resolver",
        "content matches damage"
      ],
      "estimated_tokens": 18,
      "content": "package damage\n\nfunc ResolveDamage() int { return 10 }"
    }
  ],
  "omitted_for_budget": 0
}
```

The example values are illustrative. Scores depend on the exact query, path, and content.

## Package fields

| Field | Type | Meaning |
| --- | --- | --- |
| `version` | integer | Context-package schema version; currently `1` |
| `query` | string | Original query supplied by the caller |
| `budget` | integer | Caller-supplied estimated content-token budget |
| `estimated_tokens` | integer | Sum of selected chunk estimates |
| `index_version` | integer | Prepared-index format version; currently `1` |
| `retrieval_sources` | string array | Candidate providers that contributed; currently only `lexical` |
| `selections` | object array | Ranked chunks that fit the budget |
| `omitted_for_budget` | integer | Ranked candidates skipped because each complete chunk did not fit the remaining budget |

## Selection fields

| Field | Type | Meaning |
| --- | --- | --- |
| `path` | string | Repository-relative slash-normalized source path |
| `start_line` | integer | One-based inclusive source start line |
| `end_line` | integer | One-based inclusive source end line |
| `score` | integer | Fixed lexical relevance score |
| `reasons` | string array | Inspectable score contributions |
| `estimated_tokens` | integer | Stored chunk-cost heuristic |
| `content` | string | Exact prepared chunk text |

## Ordering

Selections preserve candidate ranking order. Current candidate ordering is:

1. descending score;
2. ascending path; and
3. ascending source start line.

Only candidates with a positive score are considered.

## Budget behavior

Grimoire never truncates a selected chunk. For each ranked candidate:

- select it when its complete estimated cost fits the remaining budget;
- otherwise increment `omitted_for_budget` and continue; and
- allow a later smaller chunk to use the remaining budget.

`estimated_tokens` never exceeds `budget` in a successfully compiled package.

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

Consumers should use `version` to select a package decoder and should tolerate additional fields in future compatible revisions. The schema is pre-release and may change before a stable Grimoire release.

## Related documentation

- [CLI](cli.md)
- [System overview](../architecture/system-overview.md)
- [Current limitations](../limits/current-limitations.md)
