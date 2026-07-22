# CLI Reference

## Invocation

```text
grimoire <command> [flags]
```

The current commands are `index`, `context`, and `version`. Command output is written to standard output. Flag parsing and errors are written through the command's standard error path.

Running without a command returns an error. Unknown commands return an error.

## `grimoire index`

Prepare or incrementally update retrieval state.

```bash
grimoire index [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root to index |
| `--state <path>` | `<root>/.grimoire` | Prepared index repository path |
| `--ignore-file <path>` | root and nested `.gitignore` files | Replacement Git-ignore-syntax file |
| `--max-file-bytes <n>` | 2 MiB | Maximum eligible file size; non-positive values use the default |

Relative `--state` and `--ignore-file` paths are resolved from the absolute repository root. An absolute path is used directly.

The command loads prior prepared state when available, rebuilds changed records, reuses unchanged records, publishes the new snapshot, and writes JSON:

```json
{
  "state": "/absolute/path/to/repository/.grimoire",
  "files": 21,
  "stats": {
    "scanned": 21,
    "reused": 20,
    "updated": 1,
    "removed": 0
  }
}
```

Field meanings:

| Field | Meaning |
| --- | --- |
| `state` | Resolved prepared-state path |
| `files` | Total file records in the published snapshot |
| `stats.scanned` | Eligible source files encountered during this run |
| `stats.reused` | Eligible files whose prior records were reused |
| `stats.updated` | New or changed eligible files rebuilt during this run |
| `stats.removed` | Prior file records absent from the new snapshot |

Ignored, unsupported, oversized, binary, and non-regular files are not included in `scanned`.

## `grimoire context`

Compile a bounded context package from prepared state.

```bash
grimoire context [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root used to resolve the default or relative state path |
| `--state <path>` | `<root>/.grimoire` | Prepared index repository path |
| `--query <text>` | none | Required task or retrieval query |
| `--budget <n>` | `2000` | Maximum `o200k_base` tokens in the emitted JSON package |
| `--candidate-limit <n>` | `200` | Maximum ranked candidates before budget fitting; non-positive values disable this cap |

`--query` must be non-empty. `--budget` must be positive. The prepared index must already exist and contain a valid published state.

The command writes a versioned JSON context package. See [Context package](context-package.md).

The request path does not read repository source files. It loads prepared state, ranks stored chunks, and selects complete chunks while counting the exact indented JSON output. The budget includes package metadata and formatting, not only selected source content.

## `grimoire version`

Print the current development version:

```bash
grimoire version
```

Current value:

```text
0.1.0-dev
```

## Error behavior

Errors are returned for conditions including:

- missing or unknown commands;
- invalid flags;
- missing query;
- non-positive budget;
- missing configured ignore file;
- repository traversal or file-read failures;
- malformed or incompatible prepared state;
- budgets too small for the package metadata;
- invalid state paths or record paths; and
- conflicting concurrent index publication.

The current CLI does not yet define stable numeric exit-code classes or machine-readable error envelopes.

## Related documentation

- [Indexing](indexing.md)
- [Context package](context-package.md)
- [Prepared index](../architecture/prepared-index.md)
