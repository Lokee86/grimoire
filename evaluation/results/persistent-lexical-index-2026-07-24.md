# Persistent lexical index proof

Date: 2026-07-24

## Scope

This stream replaces query-time prepared-corpus tokenization with a persistent identifier-aware lexical sidecar. The prepared snapshot now stores per-chunk term frequencies, document lengths, basename/path/leading-line/declaration tokens, and stable document identities. Loading reconstructs content and candidate postings in memory.

BM25 uses those postings for candidate generation and scoring. Exact retrieval intersects lexical tokens within each exact signal, unions separate signals, and then preserves the existing case-sensitive and boundary-aware verification against source text and paths.

Prepared-index format advances from version 3 to version 4. Existing prepared state rebuilds once on the next `grimoire index` run.

## Benchmark

Command:

```text
go test ./internal/retrieve -run ^$ -bench "Benchmark(Search|Exact)TenThousandChunks" -benchmem -count=1
```

Environment:

```text
windows/amd64
11th Gen Intel Core i9-11900H
10,000 prepared chunks
```

| Benchmark | Main | Persistent index | Change |
| --- | ---: | ---: | ---: |
| BM25 time | 313,759,175 ns/op | 30,170,667 ns/op | 10.4x faster |
| BM25 bytes | 45,872,232 B/op | 14,687,045 B/op | 68.0% lower |
| BM25 allocations | 640,076 allocs/op | 100,151 allocs/op | 84.4% lower |
| Exact time | 2,457,293 ns/op | 286,087 ns/op | 8.6x faster |
| Exact bytes | 729,868 B/op | 724,809 B/op | 0.7% lower |
| Exact allocations | 14 allocs/op | 39 allocs/op | 25 additional allocations |

The benchmark excludes initial sidecar construction from the query loop, matching production behavior after a prepared snapshot has been built and loaded.

## Verification

```text
go test ./internal/lexical ./internal/index ./internal/retrieve
go test ./...
```

Both passed after the final exact-posting correction.
