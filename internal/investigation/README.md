# Investigation ledger

`internal/investigation` owns Grimoire's persistent agent-facing discovery ledger. It is a storage and delta boundary, not a query engine: the discovery runtime supplies a repository/provider snapshot and typed evidence, then receives new evidence bodies plus stable references to evidence already returned.

## Storage contract

A ledger is published at `<grimoire-state>/investigations/<session-id>/`. `manifest.json` is the compact versioned index. Immutable evidence records live once under `records/`; response event files under `responses/` contain only snapshot-qualified record keys.

Session readers and writers use one filesystem lock. Every mutation reloads the manifest after acquiring the lock. Manifest publication and response writes are staged and atomically replaced. Windows uses `MoveFileEx` with replace-existing and write-through semantics plus a bounded retry for transient sharing violations; other platforms use atomic rename.

Sessions bind all responses to one repository snapshot and provider snapshot map. Opening with a different snapshot, changing the session identity, corrupting a manifest or record, or closing a session prevents stale handles from being reused.

Handles are opaque typed values (`NodeHandle`, `SourceRangeHandle`, `GraphPathHandle`, and `DocumentHandle`) whose payload includes ledger version, evidence kind, snapshot digest, content key, and checksum.

## Runtime use

`internal/agentruntime` passes flattened discovery responses through `RecordResponse` when a session is named. The ledger deduplicates source nodes, source ranges, direct relationships represented as graph paths, longer graph paths, and document sections.

Repeated evidence is returned through prior handles instead of replayed content.

## API

- `Create` and `Open`
- `RecordResponse` and `DeltaFor`
- `EvidenceAlreadyReturned`
- `Status`
- `Close`

The package does not infer repository or provider snapshots and does not own query ranking, graph semantics, source extraction, document indexing, or MCP transport.

Sessions can also be managed explicitly:

```text
grimoire investigation create --session <id> --snapshot <repository-id> [--provider name=identity]
grimoire investigation status --session <id>
grimoire investigation close --session <id>
```
