# Lexicon fact contract v1

Lexicon adapters emit UTF-8 JSON Lines. Every line is one JSON object. Object keys must be serialized in lexicographic order and records after the header must be sorted by the canonical sort keys below.

## Header

The first record is:

```json
{"adapter_version":"0.1.0","language":"python","record":"lexicon","repository":"example/module","schema_version":1}
```

Required fields:

- `record`: always `lexicon`;
- `schema_version`: integer `1`;
- `adapter_version`: adapter release version;
- `language`: canonical lower-case language name;
- `repository`: repository or module identity discovered by the adapter.

## Stable identities

Stable IDs use lower-case SHA-256:

```text
sha256:<64 hexadecimal characters>
```

A node identity is the digest of this UTF-8 string:

```text
lexicon:v1\0<language>\0<kind>\0<canonical identity>
```

A content identity is the digest of the unmodified file bytes.

Adapters must document each node kind's canonical identity. Identities must not include absolute checkout paths.

## Source spans

Spans use one-based inclusive start positions and one-based exclusive end positions:

```json
{"end_column":9,"end_line":12,"path":"src/example.py","start_column":1,"start_line":12}
```

A missing or synthetic span is omitted rather than encoded with sentinel values.

## Node record

```json
{"attributes":{},"content_id":"sha256:...","id":"sha256:...","kind":"file","name":"example.py","path":"src/example.py","qualified_name":"src/example.py","record":"node","span":{...}}
```

Required fields:

- `record`: `node`;
- `id`;
- `kind`;
- `name`;
- `path`;
- `qualified_name`.

Optional fields:

- `content_id` for files;
- `span`;
- `attributes`, containing deterministic scalar values or sorted scalar arrays.

Initial common node kinds:

- `repository`;
- `directory`;
- `file`;
- `module`;
- `namespace`;
- `import`;
- `type`;
- `interface`;
- `trait`;
- `function`;
- `method`;
- `constructor`;
- `field`;
- `variable`;
- `constant`;
- `parameter`;
- `test`.

Adapters may add language-specific kinds but should prefer common kinds when semantics align.

## Edge record

```json
{"record":"edge","relation":"defines","source":"sha256:...","span":{...},"target":"sha256:..."}
```

Required fields:

- `record`: `edge`;
- `source`;
- `target`;
- `relation`.

Optional fields:

- `span`;
- `attributes`.

Initial common relations:

- `contains`;
- `defines`;
- `imports`;
- `calls`;
- `possible-calls`;
- `references`;
- `extends`;
- `implements`;
- `uses-trait`;
- `overrides`;
- `reads`;
- `writes`;
- `annotates`.

`calls` means one definite statically identified callable contract. Multiple sound targets must use `possible-calls`.

## Unresolved record

```json
{"expression":"factory()","reason":"dynamic-target","record":"unresolved","relation":"calls","source":"sha256:...","span":{...}}
```

Required fields:

- `record`: `unresolved`;
- `source`;
- `relation`;
- `expression`;
- `reason`.

Optional fields:

- `candidate_name`;
- `candidate_namespace`;
- `span`;
- `attributes`.

Common reasons:

- `missing-target`;
- `ambiguous-target`;
- `unsupported-form`;
- `dynamic-target`;
- `external-target`;
- `builtin-target`;
- `generated-target`.

## Sorting

After the header:

1. nodes by `(id, kind, path, qualified_name)`;
2. edges by `(source, target, relation, span path/start/end)`;
3. unresolved records by `(source, relation, expression, reason, span path/start/end)`.

All node records precede edge records; all edge records precede unresolved records.

## Consumer boundary

This format is a transport and conformance contract, not Arcana's storage representation. Consumers may translate records into compact binary IDs, indexes, caches, policy models, or retrieval metadata.
