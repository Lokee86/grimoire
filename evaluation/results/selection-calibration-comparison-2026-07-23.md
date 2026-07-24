# Retrieval calibration and failure correction

Date: 2026-07-24  
Integration branch: `codex/retrieval-calibration-integration`  
Base: `1093d22` (`feat: add Arcana semantic graph vectors`)  
Corpus: `evaluation/retrieval/grimoire.json`  
Mode: lexical  
Structural providers: disabled

## Scope

This work rebased the earlier candidate-selection calibration onto the final BM25, query-decomposition, and Arcana-vector tree. It rejected stale constants, recalibrated curation against a fresh 716-file prepared snapshot, and corrected the largest remaining judged retrieval failures.

## Prepared state

The paired runs used the same refreshed prepared state:

| Field | Value |
| --- | ---: |
| Eligible files | 716 |
| Reused files | 185 |
| Updated files | 531 |
| Generated files skipped | 7 |

## Selected curation configuration

| Parameter | Previous production | Selected |
| --- | ---: | ---: |
| Same-file repeat penalty | 4 | 10 |
| Same-subsystem repeat penalty | 2 | 18 |
| Primaries receiving neighbor promotion | 4 | 3 |

The earlier `4/10/2` calibration regressed after query decomposition changed the candidate stream. `10/10/3` was an intermediate winner before Arcana semantic-vector files expanded the repository. A final local grid tested neighboring file penalties, subsystem penalties from 8 through 24, and two through four neighbor anchors. Subsystem penalties 18 and 20 tied on recall, pass rate, ranking, and irrelevant-selection rate; 24 regressed. The lower plateau value, 18, was selected.

## Confirmed defects corrected

The judged traces exposed four general defects:

1. Ownership responsibilities and call-chain phases separated by commas were treated as one retrieval query.
2. Human phase descriptions such as “aggregate construction” and “package serialization” did not match normal implementation vocabulary.
3. Generated evaluation reports and corpus fixtures could outrank implementation source.
4. A generic candidate repeated across every intent pass could occupy the fused front; duplicate suppression then failed to choose a replacement anchor for later phases.

The implementation now:

- decomposes ownership lists and call-chain phase lists;
- retains distinct mechanism actions instead of merging same-topic actions;
- adds bounded deterministic implementation vocabulary for common repository phases;
- prioritizes source declarations for direct-location, mechanism, architecture, and call-chain intents;
- strongly demotes generated evaluation artifacts unless the query explicitly names their path or filename;
- reserves the first unseen candidate from every specific intent pass before global fused leaders;
- preserves up to six specific clauses plus the low-weight complete-query context entry;
- exposes the production curation parameters through evaluation-only flags.

## Final adaptive comparison

Both variants used the same corpus, prepared snapshot, provider set, and automatic budgets. Runs were sequential.

| Metric | Current `main` | Calibrated retrieval | Change |
| --- | ---: | ---: | ---: |
| Case pass rate | 8.3% | 16.7% | +8.4 pp |
| Required source recall | 8.9% | 26.7% | +17.8 pp |
| Required R@10 | 11.1% | 43.4% | +32.3 pp |
| Required R@20 | 13.9% | 51.5% | +37.6 pp |
| MRR | 0.054 | 0.332 | +0.278 |
| Irrelevant source selections | 89.4% | 72.2% | -17.2 pp |
| Median latency | 1206.4 ms | 1164.1 ms | -42.3 ms |
| p95 latency | 1665.0 ms | 1662.8 ms | -2.2 ms |

## Final fixed-budget result

| Metric | Selected configuration |
| --- | ---: |
| Case pass rate | 8.3% |
| Required source recall | 15.6% |
| Required R@10 | 43.4% |
| Required R@20 | 51.5% |
| MRR | 0.332 |
| Irrelevant source selections | 72.1% |
| Median latency | 1465.5 ms |
| p95 latency | 2467.7 ms |

The ranking gains therefore apply to both budget modes. Adaptive assembly converts more of that improvement into final-package recall.

## Final adaptive category recall

| Category | Pass rate | Required recall |
| --- | ---: | ---: |
| Direct location | 66.7% | 66.7% |
| Architecture ownership | 0.0% | 25.0% |
| Call-chain investigation | 0.0% | 25.0% |
| Long mixed query | 0.0% | 30.8% |
| Mechanism explanation | 0.0% | 11.1% |

## Remaining limitations

- Ten of twelve cases still do not fully pass because every judged item must survive to the package for a case to pass.
- Most remaining losses occur during adaptive assembly or exact token fitting after relevant evidence has entered the candidate set.
- Some required chunks remain genuine ranking misses, especially later call-chain and large structured-query evidence.
- Exact recovery cannot infer unnamed symbols from conceptual prose.
- Fixed-budget package fitting converts less of the ranking gain into final recall than adaptive assembly.
- Hybrid source evaluation remains blocked by the existing native Windows vector-engine access violation; this report does not claim semantic-mode proof.
- The 12-case corpus is a calibration signal, not evidence that the constants are universally optimal.

## Verification contract

Deterministic tests cover current defaults, explicit evaluator configuration, ownership and call-chain decomposition, mechanism-action preservation, implementation-source priority, generated-artifact demotion, and unseen per-phase anchor reservation.
