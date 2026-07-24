# Lexicon binary fact object contract v1

Lexicon stores immutable per-file and shared language fact objects in a deterministic binary format. Snapshot manifests, `CURRENT`, `PENDING`, configuration, and JSONL exports remain text formats.

## Identity and compatibility

The object ID is SHA-256 over:

```text
lexicon:fact-object:v1\0<object bytes>
```

The hash domain is unchanged from legacy JSON objects. Readers identify binary objects by the magic bytes and must continue to accept legacy canonical JSON objects. Writers emit binary objects only.

```text
4c 58 4f 42 4a 00 01 00
 L  X  O  B  J    v1
```

A different binary encoding produces a different content-addressed object ID. Existing objects are never rewritten.

## Primitive encodings

- Unsigned integers use unsigned varints; writers use the shortest encoding.
- Byte strings use `<length varint><bytes>`.
- Text is UTF-8 stored once in the object string table.
- String references are unsigned-varint indexes into that table.
- String-table index `0` is always the empty string.
- Optional strings use the empty-string reference.
- All counts, lengths, and indexes are bounds checked by readers.

## Object layout

```text
magic[8]
object_version varint
fact_schema_version varint
string_count varint
string[string_count]
language string_ref
owner string_ref
source_content_id string_ref
adapter_version string_ref
analysis_config_id string_ref
node_section bytes
edge_section bytes
unresolved_section bytes
```

Each record section is independently length-prefixed so a consumer can skip sections it does not require. Each section begins with a record count followed by fixed-order records.

Strings enter the table in deterministic first-use order: object metadata, nodes, edges, then unresolved records. Repeated strings are represented by the same index.

## Node section

```text
node_count varint
repeat node_count:
    attributes bytes
    content_id string_ref
    id string_ref
    kind string_ref
    name string_ref
    owner string_ref
    path string_ref
    qualified_name string_ref
    span
```

## Edge section

```text
edge_count varint
repeat edge_count:
    attributes bytes
    owner string_ref
    relation string_ref
    source string_ref
    span
    target string_ref
```

## Unresolved section

```text
unresolved_count varint
repeat unresolved_count:
    attributes bytes
    candidate_name string_ref
    candidate_namespace string_ref
    expression string_ref
    owner string_ref
    reason string_ref
    relation string_ref
    source string_ref
    span
```

## Span encoding

```text
present byte
if present == 1:
    path string_ref
    start_line varint
    start_column varint
    end_line varint
    end_column varint
```

Only `0` and `1` are valid presence values.

## Attributes

Attributes remain compact deterministic JSON bytes inside the binary record. They are adapter-specific provenance data rather than graph-routing fields. Consumers that do not use attributes can skip them without parsing JSON.

## Reader requirements

Readers must reject:

- invalid magic or unsupported object versions;
- malformed, overflowing, or truncated varints;
- counts and lengths above implementation safety limits;
- invalid UTF-8 strings;
- string references outside the table;
- invalid span flags;
- truncated sections;
- unconsumed trailing bytes;
- bytes whose content hash does not match the requested object ID.

Lexicon's Go reader also validates embedded attribute JSON when materializing the legacy `FactObject` representation. Arcana reads structural fields directly from binary sections and does not reconstruct JSONL.

## Determinism

For the same metadata and ordered fact records, the encoder must produce byte-identical output across runs. The Go encoder and Rust reader share a committed golden object fixture to prevent format drift.
