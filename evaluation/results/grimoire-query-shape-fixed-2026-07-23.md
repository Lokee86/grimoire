# Retrieval evaluation: Grimoire

Generated: 2026-07-23 13:44:52-07:00  
Variant: `fixed`  
Cases: 12  
Runs: 12  
Structural providers: ``

## Mode comparison

| Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 0.0% | 0.0% | 0.0% | 98.9% | 0.0% | 1912.4 ms | 3659.1 ms |

## Package comparison

| Mode | Median tokens | p95 tokens | Median chunks | Median budget use |
| --- | ---: | ---: | ---: | ---: |
| lexical | 5937 | 11900 | 8.0 | 99.0% |

## Pre-curation source ranking

These metrics score the retrieved order before exact-result merging, curation, and package fitting.

| Mode | Queries | Required R@10 | Required R@20 | MRR | Relevant @10 | Relevant @20 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 12 | 0.0% | 8.3% | 0.015 | 0.8% | 2.1% |

## Category comparison

| Category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1912.4 ms | 1915.3 ms |
| call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 2430.5 ms | 2497.9 ms |
| direct-location | 0.0% | 0.0% | 0.0% | 91.7% | 0.0% | 873.4 ms | 885.9 ms |
| long-mixed-query | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 3736.6 ms | 4434.0 ms |
| mechanism-explanation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1815.9 ms | 1920.2 ms |

## Mode by category

| Mode/category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical/architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1912.4 ms | 1915.3 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 2430.5 ms | 2497.9 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 91.7% | 0.0% | 873.4 ms | 885.9 ms |
| lexical/long-mixed-query | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 3736.6 ms | 4434.0 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1815.9 ms | 1920.2 ms |

## Per-case results

| Case | Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Budget | Tokens | Curated | Assembled | Stop | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | --- |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 3000 | 2964 | 202 | 202 | - | 873.4 ms | budget-fitting loss |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 75.0% | 0.0% | 3000 | 2786 | 202 | 202 | - | 806.2 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 3000 | 2971 | 201 | 201 | - | 887.3 ms | embedding miss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5905 | 204 | 204 | - | 1931.8 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5980 | 201 | 201 | - | 1761.0 ms | exact recovery miss, vector ranking miss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5932 | 201 | 201 | - | 1815.9 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5942 | 205 | 205 | - | 1909.2 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-ao-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5919 | 204 | 204 | - | 1915.6 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 8000 | 7966 | 202 | 202 | - | 2505.4 ms | candidate merge loss, embedding miss, vector ranking miss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 8000 | 7995 | 205 | 205 | - | 2355.6 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11878 | 204 | 204 | - | 4511.5 ms | embedding miss, exact recovery miss |
| grimoire-lm-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11928 | 202 | 202 | - | 2961.7 ms | embedding miss, exact recovery miss |

## Query profile shadow output

These classifications are observational and do not change retrieval, curation, or package assembly.

| Case | Mode | Expected | Actual | Match | Specificity | Breadth | Ambiguity | Subsystems | Graph regions | Budget mode | Mismatches |
| --- | --- | --- | --- | ---: | --- | --- | --- | ---: | ---: | --- | --- |
| grimoire-dl-01 | lexical | bounded | bounded | true | medium | low | medium | 5 | 0 | fixed |  |
| grimoire-dl-02 | lexical | bounded | bounded | true | medium | low | medium | 5 | 0 | fixed |  |
| grimoire-dl-03 | lexical | bounded | bounded | true | medium | low | low | 2 | 0 | fixed |  |
| grimoire-me-01 | lexical | bounded | bounded | true | medium | medium | medium | 6 | 0 | fixed |  |
| grimoire-me-02 | lexical | bounded | bounded | true | medium | medium | medium | 2 | 0 | fixed |  |
| grimoire-me-03 | lexical | bounded | bounded | true | medium | medium | low | 3 | 0 | fixed |  |
| grimoire-ao-01 | lexical | exploratory | exploratory | true | medium | high | low | 3 | 0 | fixed |  |
| grimoire-ao-02 | lexical | exploratory | exploratory | true | high | high | low | 1 | 0 | fixed |  |
| grimoire-cc-01 | lexical | exploratory | exploratory | true | high | high | low | 8 | 0 | fixed |  |
| grimoire-cc-02 | lexical | exploratory | exploratory | true | medium | high | low | 3 | 0 | fixed |  |
| grimoire-lm-01 | lexical | exploratory | exploratory | true | high | high | low | 6 | 0 | fixed |  |
| grimoire-lm-02 | lexical | exploratory | exploratory | true | medium | high | low | 2 | 0 | fixed |  |

## Query profile calibration

| Mode | Judged profiles | Matches | Match rate |
| --- | ---: | ---: | ---: |
| lexical | 12 | 12 | 100.0% |

## Concrete failures

- `grimoire-dl-01` / `lexical`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-02` / `lexical`: budget-fitting loss
  - `internal/retrieve/exact_signals.go` symbols `exactSignals`, `classifySignal`, `isPath`, `isConfigKey`, `isIdentifier`, `addTerminalSignal`: budget-fitting loss
- `grimoire-dl-03` / `lexical`: embedding miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `defaultEvaluationPrefix`, `writeEvaluationSummary`: embedding miss
- `grimoire-me-01` / `lexical`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
- `grimoire-me-02` / `lexical`: exact recovery miss, vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: vector ranking miss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: vector ranking miss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: exact recovery miss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: exact recovery miss
- `grimoire-me-03` / `lexical`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_ingest.go` symbols `ingestVectorBatch`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `writeVectorSnapshotManifest`, `readVectorSnapshotManifest`: budget-fitting loss
- `grimoire-ao-01` / `lexical`: budget-fitting loss, embedding miss, vector ranking miss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: vector ranking miss
  - `internal/selection/selection.go` symbols `Curate`: embedding miss
  - `internal/compiler/compiler.go` symbols `Compile`: budget-fitting loss
- `grimoire-ao-02` / `lexical`: budget-fitting loss, embedding miss, vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: embedding miss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: vector ranking miss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: embedding miss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: budget-fitting loss
- `grimoire-cc-01` / `lexical`: candidate merge loss, embedding miss, vector ranking miss
  - `cmd/grimoire/main.go` symbols `main`: embedding miss
  - `internal/app/run.go` symbols `Run`: embedding miss
  - `internal/app/context.go` symbols `runContext`: vector ranking miss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: vector ranking miss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: candidate merge loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: embedding miss
- `grimoire-cc-02` / `lexical`: budget-fitting loss, embedding miss, vector ranking miss
  - `internal/app/run.go` symbols `Run`: embedding miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: vector ranking miss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: embedding miss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: embedding miss
- `grimoire-lm-01` / `lexical`: embedding miss, exact recovery miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: embedding miss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: exact recovery miss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: embedding miss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: embedding miss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: embedding miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: embedding miss
- `grimoire-lm-02` / `lexical`: embedding miss, exact recovery miss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: embedding miss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: embedding miss
  - `internal/retrieve/exact.go` symbols `Exact`: exact recovery miss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: embedding miss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: exact recovery miss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: embedding miss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: embedding miss
