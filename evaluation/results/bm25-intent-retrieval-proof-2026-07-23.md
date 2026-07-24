# BM25 and intent-driven retrieval proof

Date: 2026-07-23  
Branch: `codex/intent-driven-retrieval`  
Corpus: `evaluation/retrieval/grimoire.json`  
Cases: 12

## Implemented

- Merged `codex/bm25-lexical-retrieval` into current Grimoire.
- Runs BM25 alongside semantic retrieval instead of using lexical retrieval only after semantic failure.
- Derives bounded retrieval intents before provider execution.
- Runs separate BM25 and exact-recovery passes for mixed-query clauses.
- Shares one tokenized BM25 corpus across all intent queries in a request.
- Applies intent-specific source weighting for direct location, call-chain, architecture, and mechanism requests.
- Reserves front-of-ranking coverage for specific clauses in mixed queries.
- Attaches intent and primary/supporting/context roles to candidate descriptors.
- Uses the most graph-relevant clause for Lexicon and Arcana lookup.
- Leaves vector retrieval on the established full-query path, then fuses it with intent-ranked BM25.

## Final consolidated-tree comparison

This comparison was run after the in-repo Arcana and Lexicon consolidation was merged. Both builds used the same fresh prepared snapshot built from 690 eligible files. Structural providers were disabled so the result isolates source retrieval, intent ranking, exact recovery, curation, and package fitting.

| Metric | Current main | BM25 + intents | Change |
| --- | ---: | ---: | ---: |
| Corpus pass rate | 8.3% | 8.3% | unchanged |
| Required source recall | 2.2% | 8.9% | +6.7 pp |
| Required R@10 | 0.0% | 11.1% | +11.1 pp |
| Required R@20 | 8.3% | 13.9% | +5.6 pp |
| MRR | 0.014 | 0.049 | +0.035 |
| Irrelevant source selections | 93.5% | 89.5% | -4.0 pp |
| Median latency | 923.6 ms | 1216.1 ms | +292.5 ms / +31.7% |
| p95 latency | 1815.6 ms | 1616.5 ms | -199.1 ms / -11.0% |

## Final recall by category

| Category | Current main | BM25 + intents |
| --- | ---: | ---: |
| Direct location | 33.3% | 33.3% |
| Mechanism explanation | 0.0% | 33.3% |
| Architecture ownership | 0.0% | 0.0% |
| Call-chain investigation | 0.0% | 0.0% |
| Long mixed query | 0.0% | 0.0% |

The improvement is real but incomplete. Mechanism recall improved materially and top-20 aggregate ranking improved. Architecture, call-chain, and long mixed cases still retrieved no judged required evidence. Long mixed irrelevant selections worsened from 88.0% to 96.6%, so clause planning and mixed-query ranking remain explicit calibration targets.

## Earlier full-stack structural comparison

Before the component consolidation changed the repository corpus, a paired run with Lexicon and Arcana enabled showed:

| Metric | Current main | BM25 + intents |
| --- | ---: | ---: |
| Required source recall | 6.7% | 11.1% |
| Required R@10 | 0.0% | 11.1% |
| Required R@20 | 8.3% | 20.1% |
| MRR | 0.025 | 0.065 |
| Irrelevant source selections | 87.5% | 81.6% |
| Median latency | 1629.3 ms | 2028.6 ms |
| p95 latency | 2199.9 ms | 2362.0 ms |

That result is retained as secondary evidence only; the consolidated-tree comparison above is authoritative for the final branch.

## Hybrid verification limitation

A valid paired hybrid corpus result could not be produced in the current environment. Current `main` and the integration branch both terminate with Windows access violation `0xc0000005` when the prepared vector snapshot is opened through the current native DLL. Explicitly supplying `native/vector-engine/target/release/grimoire_vector_ffi.dll` produces the same failure. A run without an explicit engine merely reports the engine as unavailable and is not a retrieval score.

Vector-clause execution was therefore not merged. The implementation retains the established full-query semantic path and fuses its results with the verified intent-ranked BM25 provider. The native runtime failure is separate follow-up work.

## Verification

The final branch passed:

```text
go test ./...
go vet ./...
go test -race ./...
```
