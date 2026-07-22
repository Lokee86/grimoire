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
| Embedding client, query planning, and runtime contract | `internal/embedding/client_test.go`, `internal/embedding/query_test.go`, `internal/embedding/runtime_test.go` |
| Rust object, snapshot, search, and handle lifecycle | `native/vector-engine/crates/*` unit tests |
| Go-to-Rust ABI integration | `internal/vectorstore/integration_windows_test.go` |
| Index/embed/reuse/semantic-search application path | `internal/app/vector_test.go` |
| Incremental traversal, ignore behavior, and exclusions | `internal/index/build_test.go` |
| Binary shard and file codecs | `internal/index/codec_test.go` |
| Manifest tokenizer identity and incompatible-version detection | `internal/index/objects_test.go` |
| Prepared repository persistence, validation, reuse, and conflicts | `internal/index/store_test.go` |
| `o200k_base` initialization and known token count | `internal/tokenizer/tokenizer_test.go` |
| Lexical scoring and deterministic tie-breaking | `internal/retrieve/search_test.go` |
| Targeted exact recovery and signal extraction | `internal/retrieve/exact_test.go` |
| Candidate deduplication, overlap handling, diversity, and neighbours | `internal/selection/selection_test.go` |
| End-to-end deterministic retrieval-quality fixtures | `internal/app/retrieval_quality_test.go` |
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

## Retrieval and selection benchmarks

```bash
go test ./internal/retrieve ./internal/selection \
  -bench 'Benchmark(Search|Exact|Curate)' \
  -benchmem
```

The retrieval benchmarks construct prepared in-memory snapshots containing 10,000 chunks. They measure the lexical failure path and conditional exact recovery separately. The selection benchmark measures deterministic curation of 200 ranked candidates with prepared neighbours.

The benchmarks include query parsing, relevant chunk scans, reason construction, candidate sorting or reordering, deduplication, overlap checks, and neighbour lookup as appropriate. They exclude filesystem traversal, source-file reads, hashing, chunk construction, go-git loading, model inference, native vector scanning, and complete context-package serialization.

These are warm algorithm baselines, not complete command latency. To compare live query-embedding plans without prompt-cache bias, run the local model and execute:

```bash
GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT=http://127.0.0.1:8080/v1 \
  go test ./internal/embedding -run '^$' \
  -bench BenchmarkLiveQueryEmbeddingModes -benchtime=3x -count=3
```

Repository-scale semantic measurements, query-mode results, and the checked-in quality corpus are documented in [Retrieval quality and latency baselines](retrieval-quality.md).

## Interpreting performance

The lexical fallback and initial targeted exact recovery are linear in prepared chunk text. They are useful regression baselines, but they should not be treated as proof that full scans scale to every repository. Exact recovery is conditional and skipped for ordinary natural-language queries.

Any future compact exact index or postings structure should add side-by-side benchmarks rather than replacing the existing baselines immediately. That preserves evidence of the algorithmic change.

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
