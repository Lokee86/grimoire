# Retrieval evaluation: Grimoire

Generated: 2026-07-23 06:49:17-07:00  
Variant: `ranking-metrics-baseline`  
Cases: 12  
Runs: 12  
Structural providers: ``

## Mode comparison

| Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 0.0% | 0.0% | 0.0% | 97.9% | 0.0% | 2253.2 ms | 4470.0 ms |

## Pre-curation source ranking

These metrics score the retrieved order before exact-result merging, curation, and package fitting.

| Mode | Queries | Required R@10 | Required R@20 | MRR | Relevant @10 | Relevant @20 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 12 | 8.3% | 8.3% | 0.017 | 1.7% | 2.1% |

## Category comparison

| Category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 2407.8 ms | 2514.2 ms |
| call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 3155.4 ms | 3176.5 ms |
| direct-location | 0.0% | 0.0% | 0.0% | 92.3% | 0.0% | 1079.5 ms | 1098.3 ms |
| long-mixed-query | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 4535.6 ms | 5125.4 ms |
| mechanism-explanation | 0.0% | 0.0% | 0.0% | 95.2% | 0.0% | 2063.6 ms | 2201.5 ms |

## Mode by category

| Mode/category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical/architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 2407.8 ms | 2514.2 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 3155.4 ms | 3176.5 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 92.3% | 0.0% | 1079.5 ms | 1098.3 ms |
| lexical/long-mixed-query | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 4535.6 ms | 5125.4 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 95.2% | 0.0% | 2063.6 ms | 2201.5 ms |

## Per-case results

| Case | Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 2923 | 862.6 ms | budget-fitting loss |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 80.0% | 0.0% | 2938 | 1100.4 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 2765 | 1079.5 ms | embedding miss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 5834 | 2063.6 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 87.5% | 0.0% | 5993 | 2038.7 ms | embedding miss, exact recovery miss, vector ranking miss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 5921 | 2216.8 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 5952 | 2526.0 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-ao-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 5938 | 2289.6 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 7921 | 3132.0 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 7889 | 3178.8 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 11602 | 5191.0 ms | embedding miss, exact recovery miss |
| grimoire-lm-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 11911 | 3880.1 ms | embedding miss, exact recovery miss |

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
- `grimoire-me-02` / `lexical`: embedding miss, exact recovery miss, vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: embedding miss
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
- `grimoire-cc-01` / `lexical`: budget-fitting loss, embedding miss, vector ranking miss
  - `cmd/grimoire/main.go` symbols `main`: embedding miss
  - `internal/app/run.go` symbols `Run`: embedding miss
  - `internal/app/context.go` symbols `runContext`: vector ranking miss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: vector ranking miss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: embedding miss
- `grimoire-cc-02` / `lexical`: budget-fitting loss, embedding miss, vector ranking miss
  - `internal/app/run.go` symbols `Run`: embedding miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: vector ranking miss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: embedding miss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: vector ranking miss
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
