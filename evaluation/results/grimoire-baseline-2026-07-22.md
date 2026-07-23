# Retrieval evaluation: Grimoire

Generated: 2026-07-22 17:23:09-07:00  
Variant: `standalone-baseline`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 0.0% | 8.9% | 0.0% | 81.3% | 2617.3 ms | 10093.3 ms |
| full | 0.0% | 13.3% | 0.0% | 80.1% | 2625.7 ms | 11081.8 ms |
| lexical | 0.0% | 8.9% | 0.0% | 80.8% | 2347.4 ms | 4767.4 ms |
| quality | 0.0% | 11.1% | 0.0% | 80.3% | 2544.6 ms | 14675.0 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 15.6% | 0.0% | 76.5% | 2443.6 ms | 2749.2 ms |
| call-chain-investigation | 0.0% | 16.7% | 0.0% | 81.6% | 3282.9 ms | 3549.7 ms |
| direct-location | 0.0% | 0.0% | 0.0% | 79.7% | 1177.4 ms | 1400.6 ms |
| long-mixed-query | 0.0% | 1.9% | 0.0% | 85.6% | 5222.8 ms | 24174.0 ms |
| mechanism-explanation | 0.0% | 13.9% | 0.0% | 77.6% | 2479.9 ms | 2967.2 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 12.5% | 0.0% | 71.4% | 2617.3 ms | 2662.5 ms |
| fast/call-chain-investigation | 0.0% | 16.7% | 0.0% | 87.1% | 3429.2 ms | 3561.8 ms |
| fast/direct-location | 0.0% | 0.0% | 0.0% | 80.0% | 1210.6 ms | 1351.7 ms |
| fast/long-mixed-query | 0.0% | 0.0% | 0.0% | 86.1% | 10647.4 ms | 15634.1 ms |
| fast/mechanism-explanation | 0.0% | 11.1% | 0.0% | 77.8% | 2488.9 ms | 2735.2 ms |
| full/architecture-ownership | 0.0% | 12.5% | 0.0% | 79.2% | 2625.7 ms | 2776.5 ms |
| full/call-chain-investigation | 0.0% | 25.0% | 0.0% | 79.4% | 3315.6 ms | 3344.0 ms |
| full/direct-location | 0.0% | 0.0% | 0.0% | 80.0% | 1215.8 ms | 1418.8 ms |
| full/long-mixed-query | 0.0% | 0.0% | 0.0% | 84.2% | 11775.9 ms | 18022.4 ms |
| full/mechanism-explanation | 0.0% | 22.2% | 0.0% | 77.1% | 2365.5 ms | 2884.3 ms |
| lexical/architecture-ownership | 0.0% | 25.0% | 0.0% | 75.0% | 2164.4 ms | 2182.6 ms |
| lexical/call-chain-investigation | 0.0% | 8.3% | 0.0% | 82.1% | 3094.4 ms | 3135.5 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 85.7% | 1153.2 ms | 1187.0 ms |
| lexical/long-mixed-query | 0.0% | 7.7% | 0.0% | 86.7% | 4819.3 ms | 5287.2 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 75.0% | 2470.9 ms | 2638.2 ms |
| quality/architecture-ownership | 0.0% | 12.5% | 0.0% | 80.0% | 2373.9 ms | 2423.6 ms |
| quality/call-chain-investigation | 0.0% | 16.7% | 0.0% | 78.1% | 3329.6 ms | 3482.9 ms |
| quality/direct-location | 0.0% | 0.0% | 0.0% | 73.3% | 1164.1 ms | 1280.6 ms |
| quality/long-mixed-query | 0.0% | 0.0% | 0.0% | 85.7% | 15805.7 ms | 25981.9 ms |
| quality/mechanism-explanation | 0.0% | 22.2% | 0.0% | 80.0% | 2660.2 ms | 2964.3 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | fast | false | 0.0% | 0.0% | 80.0% | 2940 | 1110.5 ms | budget-fitting loss |
| grimoire-dl-01 | full | false | 0.0% | 0.0% | 80.0% | 2919 | 927.0 ms | budget-fitting loss |
| grimoire-dl-01 | quality | false | 0.0% | 0.0% | 80.0% | 2919 | 912.8 ms | budget-fitting loss |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 80.0% | 2993 | 1190.7 ms | budget-fitting loss |
| grimoire-dl-02 | fast | false | 0.0% | 0.0% | 83.3% | 2921 | 1210.6 ms | budget-fitting loss |
| grimoire-dl-02 | full | false | 0.0% | 0.0% | 83.3% | 2895 | 1441.3 ms | budget-fitting loss |
| grimoire-dl-02 | quality | false | 0.0% | 0.0% | 83.3% | 2895 | 1293.6 ms | budget-fitting loss |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 80.0% | 2949 | 1153.2 ms | budget-fitting loss |
| grimoire-dl-03 | fast | false | 0.0% | 0.0% | 75.0% | 2924 | 1367.4 ms | budget-fitting loss |
| grimoire-dl-03 | full | false | 0.0% | 0.0% | 75.0% | 2939 | 1215.8 ms | budget-fitting loss |
| grimoire-dl-03 | quality | false | 0.0% | 0.0% | 50.0% | 2900 | 1164.1 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 2995 | 1121.1 ms | budget-fitting loss |
| grimoire-me-01 | fast | false | 0.0% | 0.0% | 83.3% | 5985 | 2317.2 ms | budget-fitting loss |
| grimoire-me-01 | full | false | 50.0% | 0.0% | 80.0% | 5903 | 2228.9 ms | budget-fitting loss |
| grimoire-me-01 | quality | false | 50.0% | 0.0% | 81.8% | 5988 | 2263.2 ms | budget-fitting loss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 75.0% | 5968 | 2223.9 ms | budget-fitting loss |
| grimoire-me-02 | fast | false | 0.0% | 0.0% | 83.3% | 5896 | 2488.9 ms | budget-fitting loss |
| grimoire-me-02 | full | false | 0.0% | 0.0% | 81.8% | 5992 | 2365.5 ms | budget-fitting loss |
| grimoire-me-02 | quality | false | 0.0% | 0.0% | 83.3% | 5933 | 2998.1 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 80.0% | 5971 | 2656.7 ms | budget-fitting loss |
| grimoire-me-03 | fast | false | 33.3% | 0.0% | 66.7% | 5873 | 2762.5 ms | budget-fitting loss |
| grimoire-me-03 | full | false | 33.3% | 0.0% | 71.4% | 5995 | 2942.0 ms | budget-fitting loss |
| grimoire-me-03 | quality | false | 33.3% | 0.0% | 75.0% | 5920 | 2660.2 ms | budget-fitting loss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 70.0% | 5966 | 2470.9 ms | budget-fitting loss |
| grimoire-ao-01 | fast | false | 0.0% | 0.0% | 81.8% | 5984 | 2667.5 ms | budget-fitting loss |
| grimoire-ao-01 | full | false | 0.0% | 0.0% | 90.9% | 5963 | 2793.2 ms | budget-fitting loss |
| grimoire-ao-01 | quality | false | 0.0% | 0.0% | 90.0% | 5960 | 2429.1 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 6000 | 2184.6 ms | budget-fitting loss |
| grimoire-ao-02 | fast | false | 25.0% | 0.0% | 60.0% | 5943 | 2567.1 ms | budget-fitting loss |
| grimoire-ao-02 | full | false | 25.0% | 0.0% | 69.2% | 5996 | 2458.2 ms | budget-fitting loss |
| grimoire-ao-02 | quality | false | 25.0% | 0.0% | 70.0% | 5983 | 2318.6 ms | budget-fitting loss |
| grimoire-ao-02 | lexical | false | 50.0% | 0.0% | 44.4% | 5953 | 2144.3 ms | budget-fitting loss |
| grimoire-cc-01 | fast | false | 0.0% | 0.0% | 94.1% | 7999 | 3281.8 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | full | false | 16.7% | 0.0% | 88.9% | 7981 | 3347.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | quality | false | 0.0% | 0.0% | 88.2% | 7991 | 3159.3 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 94.1% | 7989 | 3048.8 ms | budget-fitting loss |
| grimoire-cc-02 | fast | false | 33.3% | 0.0% | 78.6% | 7935 | 3576.6 ms | budget-fitting loss |
| grimoire-cc-02 | full | false | 33.3% | 0.0% | 68.8% | 7945 | 3283.9 ms | budget-fitting loss |
| grimoire-cc-02 | quality | false | 33.3% | 0.0% | 66.7% | 7968 | 3499.9 ms | budget-fitting loss |
| grimoire-cc-02 | lexical | false | 16.7% | 0.0% | 63.6% | 7900 | 3140.0 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | fast | false | 0.0% | 0.0% | 80.0% | 11983 | 16188.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | full | false | 0.0% | 0.0% | 80.0% | 11983 | 18716.5 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | quality | false | 0.0% | 0.0% | 80.0% | 11983 | 27112.6 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 80.0% | 11990 | 5339.1 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | fast | false | 0.0% | 0.0% | 90.5% | 11885 | 5106.5 ms | budget-fitting loss |
| grimoire-lm-02 | full | false | 0.0% | 0.0% | 87.0% | 11990 | 4835.3 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | quality | false | 0.0% | 0.0% | 90.0% | 11839 | 4498.9 ms | budget-fitting loss |
| grimoire-lm-02 | lexical | false | 14.3% | 0.0% | 93.3% | 11935 | 4299.6 ms | budget-fitting loss, exact recovery miss |

## Concrete failures

- `grimoire-dl-01` / `fast`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-01` / `full`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-01` / `quality`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-01` / `lexical`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-02` / `fast`: budget-fitting loss
  - `internal/retrieve/exact_signals.go` symbols `exactSignals`, `classifySignal`, `isPath`, `isConfigKey`, `isIdentifier`, `addTerminalSignal`: budget-fitting loss
- `grimoire-dl-02` / `full`: budget-fitting loss
  - `internal/retrieve/exact_signals.go` symbols `exactSignals`, `classifySignal`, `isPath`, `isConfigKey`, `isIdentifier`, `addTerminalSignal`: budget-fitting loss
- `grimoire-dl-02` / `quality`: budget-fitting loss
  - `internal/retrieve/exact_signals.go` symbols `exactSignals`, `classifySignal`, `isPath`, `isConfigKey`, `isIdentifier`, `addTerminalSignal`: budget-fitting loss
- `grimoire-dl-02` / `lexical`: budget-fitting loss
  - `internal/retrieve/exact_signals.go` symbols `exactSignals`, `classifySignal`, `isPath`, `isConfigKey`, `isIdentifier`, `addTerminalSignal`: budget-fitting loss
- `grimoire-dl-03` / `fast`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `defaultEvaluationPrefix`, `writeEvaluationSummary`: budget-fitting loss
- `grimoire-dl-03` / `full`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `defaultEvaluationPrefix`, `writeEvaluationSummary`: budget-fitting loss
- `grimoire-dl-03` / `quality`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `defaultEvaluationPrefix`, `writeEvaluationSummary`: budget-fitting loss
- `grimoire-dl-03` / `lexical`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `defaultEvaluationPrefix`, `writeEvaluationSummary`: budget-fitting loss
- `grimoire-me-01` / `fast`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
- `grimoire-me-01` / `full`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
- `grimoire-me-01` / `quality`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
- `grimoire-me-01` / `lexical`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
- `grimoire-me-02` / `fast`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-02` / `full`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-02` / `quality`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-02` / `lexical`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-03` / `fast`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `writeVectorSnapshotManifest`, `readVectorSnapshotManifest`: budget-fitting loss
- `grimoire-me-03` / `full`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `writeVectorSnapshotManifest`, `readVectorSnapshotManifest`: budget-fitting loss
- `grimoire-me-03` / `quality`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `writeVectorSnapshotManifest`, `readVectorSnapshotManifest`: budget-fitting loss
- `grimoire-me-03` / `lexical`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_ingest.go` symbols `ingestVectorBatch`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `writeVectorSnapshotManifest`, `readVectorSnapshotManifest`: budget-fitting loss
- `grimoire-ao-01` / `fast`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`: budget-fitting loss
- `grimoire-ao-01` / `full`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`: budget-fitting loss
- `grimoire-ao-01` / `quality`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`: budget-fitting loss
- `grimoire-ao-01` / `lexical`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`: budget-fitting loss
- `grimoire-ao-02` / `fast`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-ao-02` / `full`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: budget-fitting loss
- `grimoire-ao-02` / `quality`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-ao-02` / `lexical`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-cc-01` / `fast`: budget-fitting loss, vector ranking miss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: vector ranking miss
- `grimoire-cc-01` / `full`: budget-fitting loss, vector ranking miss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: vector ranking miss
- `grimoire-cc-01` / `quality`: budget-fitting loss, vector ranking miss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: vector ranking miss
- `grimoire-cc-01` / `lexical`: budget-fitting loss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: budget-fitting loss
- `grimoire-cc-02` / `fast`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-cc-02` / `full`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-cc-02` / `quality`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-cc-02` / `lexical`: budget-fitting loss, vector ranking miss
  - `internal/app/run.go` symbols `Run`: vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-lm-01` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: budget-fitting loss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: vector ranking miss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
- `grimoire-lm-01` / `full`: budget-fitting loss, vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: budget-fitting loss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: vector ranking miss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
- `grimoire-lm-01` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: budget-fitting loss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: vector ranking miss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
- `grimoire-lm-01` / `lexical`: budget-fitting loss, vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: budget-fitting loss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: vector ranking miss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
- `grimoire-lm-02` / `fast`: budget-fitting loss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/retrieve/exact.go` symbols `Exact`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: budget-fitting loss
- `grimoire-lm-02` / `full`: budget-fitting loss, vector ranking miss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/retrieve/exact.go` symbols `Exact`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: vector ranking miss
- `grimoire-lm-02` / `quality`: budget-fitting loss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/retrieve/exact.go` symbols `Exact`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: budget-fitting loss
- `grimoire-lm-02` / `lexical`: budget-fitting loss, exact recovery miss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/retrieve/exact.go` symbols `Exact`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: exact recovery miss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
