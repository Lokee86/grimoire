# Testing and Benchmarks

## Required checks

Run from the Grimoire repository root:

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

Formatting should produce no diff after the final run.

## Test ownership

| Area | Primary tests |
| --- | --- |
| CLI integration and flag wiring | `internal/app/run_test.go` |
| Incremental traversal, ignore behavior, and exclusions | `internal/index/build_test.go` |
| Binary shard and file codecs | `internal/index/codec_test.go` |
| Prepared repository persistence, validation, reuse, and conflicts | `internal/index/store_test.go` |
| Lexical scoring and deterministic tie-breaking | `internal/retrieve/search_test.go` |
| Whole-chunk budget selection and package fields | `internal/compiler/compiler_test.go` |

Tests use temporary directories and locally constructed snapshots. They do not require external services.

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
- [Current limitations](../limits/current-limitations.md)
