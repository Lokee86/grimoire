# Testing and Benchmarks

## Required checks

Run from the Grimoire repository root:

```bash
cd native/vector-engine
cargo fmt --all --check
cargo test --workspace
cargo clippy --workspace --all-targets -- -D warnings
cargo build -p grimoire-vector-ffi --release
cd ../..
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

Formatting should produce no diff after the final run.

## Test ownership

| Area | Primary tests |
| --- | --- |
| CLI integration and flag wiring | `internal/app/run_test.go`, `internal/app/model_test.go` |
| Embedding client and runtime contract | `internal/embedding/client_test.go`, `internal/embedding/runtime_test.go` |
| Rust object, snapshot, search, and handle lifecycle | `native/vector-engine/crates/*` unit tests |
| Go-to-Rust ABI integration | `internal/vectorstore/integration_windows_test.go` |
| Index/embed/reuse/semantic-search application path | `internal/app/vector_test.go` |
| Incremental traversal, ignore behavior, and exclusions | `internal/index/build_test.go` |
| Binary shard and file codecs | `internal/index/codec_test.go` |
| Manifest tokenizer identity and incompatible-version detection | `internal/index/objects_test.go` |
| Prepared repository persistence, validation, reuse, and conflicts | `internal/index/store_test.go` |
| `o200k_base` initialization and known token count | `internal/tokenizer/tokenizer_test.go` |
| Lexical scoring and deterministic tie-breaking | `internal/retrieve/search_test.go` |
| Whole-chunk fitting and exact serialized-package accounting | `internal/compiler/compiler_test.go` |

Tests use temporary directories, local HTTP test servers, synthetic vectors, and locally constructed snapshots. They do not require an installed model or external service. Native integration tests require a built Rust DLL and skip when it is unavailable.

Embedding coverage verifies query instruction formatting, OpenAI-compatible response handling, response-index ordering, 1024-to-512 reduction, normalization, and malformed-vector rejection.

A real installed provider can be smoke-tested separately:

```bash
grimoire model serve
# in another shell
grimoire model probe
```

## Native vector verification

The Rust suite validates immutable-object reuse/conflict rejection, deterministic snapshot construction, malformed layout rejection, exact ranking, and handle-close behavior. The Go integration test crosses the real DLL ABI and verifies caller-owned buffers, metadata, search ordering, close, and search-after-close failure.

Before release, also run race-enabled Go tests and sanitizer-enabled Rust builds on supported toolchains. The C boundary is deliberately narrow enough for repeated open/search/close stress and malformed-input fuzzing.

## Retrieval benchmark

```bash
go test ./internal/retrieve \
  -bench BenchmarkSearchTenThousandChunks \
  -benchmem
```

The benchmark constructs a prepared in-memory snapshot containing 10,000 chunks, then measures `retrieve.Search`.

It includes:

- query parsing;
- scanning prepared chunks;
- lexical scoring and reason construction;
- candidate collection;
- sorting; and
- candidate limiting.

It excludes:

- filesystem traversal;
- source-file reads;
- hashing;
- chunk construction;
- go-git repository loading; and
- context-package budget fitting.

The benchmark therefore measures the current warm retrieval algorithm, not complete command latency.

## Interpreting performance

Current retrieval is linear in the number and size of prepared chunks, followed by candidate sorting. The benchmark is useful as a regression baseline, but it should not be treated as proof that the current full prepared-chunk scan will scale to every target repository.

Future postings-index work should add side-by-side benchmarks rather than replacing this baseline immediately. That preserves evidence of the algorithmic change.

## Documentation verification

When behavior changes:

1. Update the owning package README.
2. Update current architecture or reference documentation.
3. Update current limitations when the change removes or introduces one.
4. Move unresolved follow-on work into the roadmap rather than describing it as implemented.
5. Keep command examples aligned with `internal/app/run.go`.

## Related documentation

- [System overview](../architecture/system-overview.md)
- [Prepared index](../architecture/prepared-index.md)
- [Embedding model](../reference/embedding-model.md)
- [Vector store](../reference/vector-store.md)
- [Current limitations](../limits/current-limitations.md)
