# Vector Store Bridge

`internal/vectorstore` owns the Go side of the Rust vector-engine ABI.

## Owns

- locating and loading the native vector library;
- ABI-version validation;
- content-addressed object existence checks;
- JSONL ingestion and snapshot materialization calls;
- snapshot handle lifecycle;
- caller-owned search buffers and retry-on-capacity behavior;
- conversion of native hits into Go values; and
- `runtime.KeepAlive` protection for every borrowed Go buffer.

## Does not own

- embedding generation;
- chunk or source identity;
- packed storage formats;
- similarity computation;
- hybrid ranking; or
- context-package selection.

The current production bridge uses the Windows DLL ABI. Non-Windows builds return `ErrUnavailable` until equivalent platform loaders are added.
