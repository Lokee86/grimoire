# Retrieval evaluation: Grimoire

Generated: 2026-07-22 17:40:32-07:00  
Variant: `neighbor-promotion`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 8.3% | 11.1% | 0.0% | 78.4% | 2565.8 ms | 9653.1 ms |
| full | 8.3% | 11.1% | 0.0% | 79.6% | 2535.1 ms | 10295.0 ms |
| lexical | 0.0% | 4.4% | 0.0% | 75.4% | 2291.8 ms | 4289.9 ms |
| quality | 8.3% | 8.9% | 0.0% | 80.5% | 2586.9 ms | 14332.3 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 15.6% | 0.0% | 70.9% | 2463.6 ms | 2601.4 ms |
| call-chain-investigation | 0.0% | 12.5% | 0.0% | 85.4% | 3010.8 ms | 3341.4 ms |
| direct-location | 25.0% | 25.0% | 0.0% | 74.1% | 1286.8 ms | 1569.5 ms |
| long-mixed-query | 0.0% | 0.0% | 0.0% | 85.2% | 4574.3 ms | 23275.7 ms |
| mechanism-explanation | 0.0% | 5.6% | 0.0% | 72.0% | 2515.8 ms | 2970.7 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 12.5% | 0.0% | 70.0% | 2521.9 ms | 2607.3 ms |
| fast/call-chain-investigation | 0.0% | 16.7% | 0.0% | 84.4% | 3152.7 ms | 3409.8 ms |
| fast/direct-location | 33.3% | 33.3% | 0.0% | 71.4% | 1516.2 ms | 1622.7 ms |
| fast/long-mixed-query | 0.0% | 0.0% | 0.0% | 84.8% | 10211.7 ms | 15238.8 ms |
| fast/mechanism-explanation | 0.0% | 11.1% | 0.0% | 74.3% | 2514.9 ms | 2762.5 ms |
| full/architecture-ownership | 0.0% | 12.5% | 0.0% | 85.0% | 2506.2 ms | 2511.6 ms |
| full/call-chain-investigation | 0.0% | 16.7% | 0.0% | 87.5% | 3010.8 ms | 3070.6 ms |
| full/direct-location | 33.3% | 33.3% | 0.0% | 73.3% | 1193.2 ms | 1262.7 ms |
| full/long-mixed-query | 0.0% | 0.0% | 0.0% | 82.5% | 10936.4 ms | 16708.8 ms |
| full/mechanism-explanation | 0.0% | 11.1% | 0.0% | 68.6% | 2558.0 ms | 2890.5 ms |
| lexical/architecture-ownership | 0.0% | 25.0% | 0.0% | 55.0% | 2093.8 ms | 2143.3 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 81.5% | 2739.8 ms | 2822.6 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 78.6% | 1197.2 ms | 1292.5 ms |
| lexical/long-mixed-query | 0.0% | 0.0% | 0.0% | 89.7% | 4298.0 ms | 4371.1 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 67.9% | 2350.3 ms | 2394.8 ms |
| quality/architecture-ownership | 0.0% | 12.5% | 0.0% | 73.7% | 2421.6 ms | 2557.6 ms |
| quality/call-chain-investigation | 0.0% | 16.7% | 0.0% | 87.5% | 3159.8 ms | 3161.2 ms |
| quality/direct-location | 33.3% | 33.3% | 0.0% | 73.3% | 1361.7 ms | 1462.2 ms |
| quality/long-mixed-query | 0.0% | 0.0% | 0.0% | 84.8% | 15435.4 ms | 25363.3 ms |
| quality/mechanism-explanation | 0.0% | 0.0% | 0.0% | 76.5% | 2600.9 ms | 2981.2 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | fast | false | 0.0% | 0.0% | 80.0% | 2871 | 1634.6 ms | budget-fitting loss |
| grimoire-dl-01 | full | false | 0.0% | 0.0% | 83.3% | 2978 | 1270.5 ms | budget-fitting loss |
| grimoire-dl-01 | quality | false | 0.0% | 0.0% | 83.3% | 2978 | 1171.8 ms | budget-fitting loss |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 2991 | 1197.2 ms | budget-fitting loss |
| grimoire-dl-02 | fast | true | 100.0% | 0.0% | 60.0% | 2937 | 1371.0 ms |  |
| grimoire-dl-02 | full | true | 100.0% | 0.0% | 60.0% | 2917 | 1193.2 ms |  |
| grimoire-dl-02 | quality | true | 100.0% | 0.0% | 60.0% | 2917 | 1361.7 ms |  |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 60.0% | 2944 | 1147.3 ms | budget-fitting loss |
| grimoire-dl-03 | fast | false | 0.0% | 0.0% | 75.0% | 2925 | 1516.2 ms | budget-fitting loss |
| grimoire-dl-03 | full | false | 0.0% | 0.0% | 75.0% | 2939 | 1166.5 ms | budget-fitting loss |
| grimoire-dl-03 | quality | false | 0.0% | 0.0% | 75.0% | 2983 | 1473.4 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 75.0% | 2954 | 1303.1 ms | budget-fitting loss |
| grimoire-me-01 | fast | false | 0.0% | 0.0% | 90.0% | 5999 | 2514.9 ms | budget-fitting loss |
| grimoire-me-01 | full | false | 50.0% | 0.0% | 80.0% | 5949 | 2326.1 ms | budget-fitting loss |
| grimoire-me-01 | quality | false | 0.0% | 0.0% | 100.0% | 5939 | 3023.5 ms | budget-fitting loss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 87.5% | 5913 | 2233.2 ms | budget-fitting loss |
| grimoire-me-02 | fast | false | 0.0% | 0.0% | 75.0% | 5910 | 2790.0 ms | budget-fitting loss |
| grimoire-me-02 | full | false | 0.0% | 0.0% | 75.0% | 5982 | 2927.5 ms | budget-fitting loss |
| grimoire-me-02 | quality | false | 0.0% | 0.0% | 75.0% | 5975 | 2600.9 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 70.0% | 5999 | 2350.3 ms | budget-fitting loss |
| grimoire-me-03 | fast | false | 33.3% | 0.0% | 61.5% | 5999 | 2359.7 ms | budget-fitting loss |
| grimoire-me-03 | full | false | 0.0% | 0.0% | 53.8% | 5996 | 2558.0 ms | budget-fitting loss |
| grimoire-me-03 | quality | false | 0.0% | 0.0% | 58.3% | 5980 | 2516.7 ms | budget-fitting loss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 50.0% | 5913 | 2399.8 ms | budget-fitting loss |
| grimoire-ao-01 | fast | false | 0.0% | 0.0% | 100.0% | 5941 | 2616.8 ms | budget-fitting loss |
| grimoire-ao-01 | full | false | 0.0% | 0.0% | 100.0% | 5930 | 2500.1 ms | budget-fitting loss |
| grimoire-ao-01 | quality | false | 0.0% | 0.0% | 100.0% | 6000 | 2270.4 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 90.0% | 5965 | 2038.9 ms | budget-fitting loss |
| grimoire-ao-02 | fast | false | 25.0% | 0.0% | 40.0% | 5993 | 2427.0 ms | budget-fitting loss |
| grimoire-ao-02 | full | false | 25.0% | 0.0% | 72.7% | 5959 | 2512.2 ms | budget-fitting loss |
| grimoire-ao-02 | quality | false | 25.0% | 0.0% | 50.0% | 5989 | 2572.8 ms | budget-fitting loss |
| grimoire-ao-02 | lexical | false | 50.0% | 0.0% | 20.0% | 5994 | 2148.8 ms | budget-fitting loss |
| grimoire-cc-01 | fast | false | 0.0% | 0.0% | 94.1% | 7911 | 3438.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | full | false | 0.0% | 0.0% | 94.1% | 7978 | 2944.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | quality | false | 0.0% | 0.0% | 94.1% | 7956 | 3161.3 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 7971 | 2831.8 ms | budget-fitting loss |
| grimoire-cc-02 | fast | false | 33.3% | 0.0% | 73.3% | 7891 | 2867.0 ms | budget-fitting loss |
| grimoire-cc-02 | full | false | 33.3% | 0.0% | 80.0% | 7895 | 3077.3 ms | budget-fitting loss |
| grimoire-cc-02 | quality | false | 33.3% | 0.0% | 80.0% | 7957 | 3158.2 ms | budget-fitting loss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 54.5% | 7895 | 2647.8 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | fast | false | 0.0% | 0.0% | 75.0% | 11908 | 15797.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | full | false | 0.0% | 0.0% | 75.0% | 11946 | 17350.1 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | quality | false | 0.0% | 0.0% | 75.0% | 11926 | 26466.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 78.6% | 11900 | 4216.7 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | fast | false | 0.0% | 0.0% | 90.5% | 11892 | 4626.0 ms | budget-fitting loss |
| grimoire-lm-02 | full | false | 0.0% | 0.0% | 87.5% | 11979 | 4522.7 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | quality | false | 0.0% | 0.0% | 90.5% | 11940 | 4404.4 ms | budget-fitting loss |
| grimoire-lm-02 | lexical | false | 0.0% | 0.0% | 100.0% | 11978 | 4379.3 ms | budget-fitting loss, exact recovery miss |

## Concrete failures

- `grimoire-dl-01` / `fast`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-01` / `full`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-01` / `quality`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
- `grimoire-dl-01` / `lexical`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
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
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
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
  - `internal/app/vector_ingest.go` symbols `ingestVectorBatch`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `writeVectorSnapshotManifest`, `readVectorSnapshotManifest`: budget-fitting loss
- `grimoire-me-03` / `quality`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_ingest.go` symbols `ingestVectorBatch`: budget-fitting loss
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
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
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
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
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
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: budget-fitting loss
