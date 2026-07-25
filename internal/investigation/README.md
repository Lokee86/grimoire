# Investigation ledger

`internal/investigation` owns the persistent, agent-facing discovery ledger. It is a storage and delta boundary, not a query engine: a future query or MCP layer supplies a repository/provider snapshot and a typed response, then receives new evidence bodies plus stable references to evidence already returned.

## Storage contract

A ledger is published at `<grimoire-state>/investigations/<session-id>/`. `manifest.json` is the compact versioned index. Immutable evidence records live once under `records/`; response event files under `responses/` contain only snapshot-qualified record keys. Manifest publication and response writes are staged and renamed atomically. Session writers take a filesystem lock and reload the manifest before each mutation.

Sessions bind all responses to one repository snapshot and provider snapshot map. Opening with a different snapshot, changing the session identity, corrupting a manifest/record, or closing a session prevents stale handles from being reused. Handles are opaque typed values (`NodeHandle`, `SourceRangeHandle`, `GraphPathHandle`, and `DocumentHandle`) whose internal payload includes the ledger version, evidence kind, snapshot digest, content key, and checksum.

Use `Create`/`Open`, `RecordResponse`, `DeltaFor`, `EvidenceAlreadyReturned`, `Status`, and `Close`. `RecordResponse` is the normal integration path: it persists deduplicated immutable records and returns a `Delta`. A future query layer should pass provider outputs through this package before emitting agent-facing evidence; it should not write the ledger files or infer handle formats itself.

The ledger deliberately does not infer repository or provider snapshots and does not own query ranking, graph semantics, source extraction, or MCP transport. The current CLI only creates, reports, and closes sessions:

```text
grimoire investigation create --session <id> --snapshot <repository-id> [--provider name=identity]
grimoire investigation status --session <id>
grimoire investigation close --session <id>
```
