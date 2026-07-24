# BM25 and intent-driven retrieval proof

Date: 2026-07-23  
Branch: `codex/intent-driven-retrieval`  
Corpus: `evaluation/retrieval/grimoire.json`  
Cases: 12  
Prepared index, Lexicon snapshot, and Arcana snapshot were held constant between runs.

## Implemented

- Merged `codex/bm25-lexical-retrieval` into current `main`.
- Runs BM25 alongside semantic retrieval instead of using lexical retrieval only after semantic failure.
- Derives bounded retrieval intents before provider execution.
- Runs separate BM25 and exact-recovery passes for mixed-query clauses.
- Shares one tokenized BM25 corpus across all intent queries in a request.
- Applies intent-specific source weighting for direct location, call-chain, architecture, and mechanism requests.
- Reserves front-of-ranking coverage for specific clauses in mixed queries.
- Attaches intent and primary/supporting/context roles to candidate descriptors.
- Uses the most graph-relevant clause for Lexicon and Arcana lookup.
- Leaves vector retrieval on the established full-query path, then fuses it with intent-ranked BM25.

## Paired lexical full-stack evaluation

Both runs enabled Lexicon and Arcana and used adaptive package construction.

| Metric | Current main | BM25 + intents | Change |
| --- | ---: | ---: | ---: |
| Corpus pass rate | 0.0% | 8.3% | +8.3 pp |
| Required source recall | 6.7% | 11.1% | +4.4 pp |
| Required R@10 | 0.0% | 11.1% | +11.1 pp |
| Required R@20 | 8.3% | 20.1% | +11.8 pp |
| MRR | 0.025 | 0.065 | +0.040 |
| Irrelevant source selections | 87.5% | 81.6% | -5.9 pp |
| Median latency | 1629.3 ms | 2028.6 ms | +399.3 ms / +24.5% |
| p95 latency | 2199.9 ms | 2362.0 ms | +162.1 ms / +7.4% |

## Source recall by category

| Category | Current main | BM25 + intents |
| --- | ---: | ---: |
| Direct location | 0.0% | 33.3% |
| Architecture ownership | 12.5% | 12.5% |
| Call-chain investigation | 16.7% | 16.7% |
| Long mixed query | 0.0% | 7.7% |
| Mechanism explanation | 0.0% | 0.0% |

Mechanism queries did not recover required evidence, but their irrelevant-selection rate improved from 100.0% to 76.7%. Call-chain recall remained unchanged while irrelevant selections worsened from 75.0% to 80.6%; this remains a calibration target.

## Hybrid verification limitation

A valid paired hybrid corpus result could not be produced in the current environment. Current `main` and the integration branch both terminate with Windows access violation `0xc0000005` when the prepared vector snapshot is opened through the current native DLL. Explicitly supplying `native/vector-engine/target/release/grimoire_vector_ffi.dll` produces the same failure. A run without an explicit engine merely reports the engine as unavailable and is not a retrieval score.

Vector-clause execution was therefore not merged. The implementation retains the established full-query semantic path and fuses its results with the verified intent-ranked BM25 provider. The native runtime failure is separate follow-up work.

## Verification

The final branch was checked with:

```text
go test ./...
go vet ./...
go test -race ./...
```
