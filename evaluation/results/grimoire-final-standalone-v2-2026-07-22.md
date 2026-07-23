# Retrieval evaluation: Grimoire

Generated: 2026-07-22 18:11:52-07:00  
Variant: `final-standalone-v2`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 8.3% | 13.3% | 0.0% | 78.7% | 2179.8 ms | 8974.6 ms |
| full | 8.3% | 8.9% | 5.3% | 77.6% | 2200.4 ms | 8767.4 ms |
| lexical | 0.0% | 6.7% | 0.0% | 75.8% | 1992.8 ms | 3682.9 ms |
| quality | 8.3% | 8.9% | 0.0% | 76.7% | 2158.4 ms | 13491.6 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 15.6% | 0.0% | 72.4% | 2205.7 ms | 2453.6 ms |
| call-chain-investigation | 0.0% | 12.5% | 0.0% | 82.0% | 2714.3 ms | 2860.6 ms |
| direct-location | 25.0% | 25.0% | 0.0% | 71.4% | 1011.6 ms | 1184.6 ms |
| long-mixed-query | 0.0% | 1.9% | 4.2% | 83.2% | 4022.2 ms | 21684.5 ms |
| mechanism-explanation | 0.0% | 5.6% | 0.0% | 72.4% | 1987.9 ms | 2178.7 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 25.0% | 0.0% | 66.7% | 2215.1 ms | 2243.7 ms |
| fast/call-chain-investigation | 0.0% | 16.7% | 0.0% | 87.1% | 2844.8 ms | 2892.0 ms |
| fast/direct-location | 33.3% | 33.3% | 0.0% | 70.6% | 1152.8 ms | 1208.8 ms |
| fast/long-mixed-query | 0.0% | 0.0% | 0.0% | 82.4% | 9512.9 ms | 14358.0 ms |
| fast/mechanism-explanation | 0.0% | 11.1% | 0.0% | 77.8% | 2117.9 ms | 2170.5 ms |
| full/architecture-ownership | 0.0% | 0.0% | 0.0% | 85.0% | 2440.3 ms | 2480.2 ms |
| full/call-chain-investigation | 0.0% | 16.7% | 0.0% | 80.6% | 2714.3 ms | 2724.7 ms |
| full/direct-location | 33.3% | 33.3% | 0.0% | 70.6% | 948.1 ms | 1138.5 ms |
| full/long-mixed-query | 0.0% | 0.0% | 16.7% | 82.5% | 9318.4 ms | 14277.2 ms |
| full/mechanism-explanation | 0.0% | 11.1% | 0.0% | 68.6% | 1875.2 ms | 1991.9 ms |
| lexical/architecture-ownership | 0.0% | 25.0% | 0.0% | 63.2% | 2008.0 ms | 2206.0 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 82.8% | 2621.6 ms | 2634.6 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 76.9% | 989.9 ms | 1028.9 ms |
| lexical/long-mixed-query | 0.0% | 7.7% | 0.0% | 83.3% | 3704.0 ms | 3893.9 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 69.0% | 1899.9 ms | 2067.1 ms |
| quality/architecture-ownership | 0.0% | 12.5% | 0.0% | 73.7% | 2158.4 ms | 2158.7 ms |
| quality/call-chain-investigation | 0.0% | 16.7% | 0.0% | 77.4% | 2666.7 ms | 2725.7 ms |
| quality/direct-location | 33.3% | 33.3% | 0.0% | 68.8% | 950.5 ms | 1037.5 ms |
| quality/long-mixed-query | 0.0% | 0.0% | 0.0% | 84.8% | 14568.7 ms | 24262.6 ms |
| quality/mechanism-explanation | 0.0% | 0.0% | 0.0% | 73.5% | 2142.1 ms | 2177.7 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | fast | false | 0.0% | 0.0% | 71.4% | 2971 | 1152.8 ms | budget-fitting loss |
| grimoire-dl-01 | full | false | 0.0% | 0.0% | 71.4% | 2943 | 812.9 ms | budget-fitting loss |
| grimoire-dl-01 | quality | false | 0.0% | 0.0% | 71.4% | 2943 | 797.4 ms | budget-fitting loss |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 2990 | 898.7 ms | budget-fitting loss |
| grimoire-dl-02 | fast | true | 100.0% | 0.0% | 66.7% | 2929 | 1036.6 ms |  |
| grimoire-dl-02 | full | true | 100.0% | 0.0% | 66.7% | 2905 | 948.1 ms |  |
| grimoire-dl-02 | quality | true | 100.0% | 0.0% | 66.7% | 2905 | 950.5 ms |  |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 60.0% | 2944 | 989.9 ms | budget-fitting loss |
| grimoire-dl-03 | fast | false | 0.0% | 0.0% | 75.0% | 2999 | 1215.0 ms | budget-fitting loss |
| grimoire-dl-03 | full | false | 0.0% | 0.0% | 75.0% | 2946 | 1159.7 ms | budget-fitting loss |
| grimoire-dl-03 | quality | false | 0.0% | 0.0% | 66.7% | 2880 | 1047.2 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 66.7% | 2917 | 1033.2 ms | budget-fitting loss |
| grimoire-me-01 | fast | false | 0.0% | 0.0% | 100.0% | 5995 | 2176.3 ms | budget-fitting loss |
| grimoire-me-01 | full | false | 50.0% | 0.0% | 72.7% | 5900 | 2004.9 ms | budget-fitting loss |
| grimoire-me-01 | quality | false | 0.0% | 0.0% | 100.0% | 5991 | 2142.1 ms | budget-fitting loss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 87.5% | 5984 | 1875.5 ms | budget-fitting loss |
| grimoire-me-02 | fast | false | 0.0% | 0.0% | 75.0% | 5968 | 1971.0 ms | budget-fitting loss |
| grimoire-me-02 | full | false | 0.0% | 0.0% | 75.0% | 6000 | 1859.4 ms | budget-fitting loss, exact recovery miss |
| grimoire-me-02 | quality | false | 0.0% | 0.0% | 75.0% | 5989 | 1930.1 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 70.0% | 5975 | 1899.9 ms | budget-fitting loss |
| grimoire-me-03 | fast | false | 33.3% | 0.0% | 61.5% | 5975 | 2117.9 ms | budget-fitting loss |
| grimoire-me-03 | full | false | 0.0% | 0.0% | 58.3% | 5868 | 1875.2 ms | budget-fitting loss |
| grimoire-me-03 | quality | false | 0.0% | 0.0% | 50.0% | 5976 | 2181.7 ms | budget-fitting loss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 54.5% | 5899 | 2085.7 ms | budget-fitting loss |
| grimoire-ao-01 | fast | false | 25.0% | 0.0% | 80.0% | 5991 | 2246.9 ms | budget-fitting loss |
| grimoire-ao-01 | full | false | 0.0% | 0.0% | 100.0% | 5956 | 2484.6 ms | budget-fitting loss, vector ranking miss |
| grimoire-ao-01 | quality | false | 0.0% | 0.0% | 100.0% | 5937 | 2158.7 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 5998 | 2228.0 ms | budget-fitting loss |
| grimoire-ao-02 | fast | false | 25.0% | 0.0% | 50.0% | 5903 | 2183.3 ms | budget-fitting loss |
| grimoire-ao-02 | full | false | 0.0% | 0.0% | 70.0% | 5903 | 2395.9 ms | budget-fitting loss |
| grimoire-ao-02 | quality | false | 25.0% | 0.0% | 50.0% | 5940 | 2158.0 ms | budget-fitting loss |
| grimoire-ao-02 | lexical | false | 50.0% | 0.0% | 22.2% | 5896 | 1788.0 ms | budget-fitting loss |
| grimoire-cc-01 | fast | false | 0.0% | 0.0% | 94.1% | 7898 | 2792.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | full | false | 0.0% | 0.0% | 94.1% | 7978 | 2725.9 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | quality | false | 0.0% | 0.0% | 94.1% | 7943 | 2601.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 7971 | 2636.0 ms | budget-fitting loss |
| grimoire-cc-02 | fast | false | 33.3% | 0.0% | 78.6% | 7991 | 2897.3 ms | budget-fitting loss |
| grimoire-cc-02 | full | false | 33.3% | 0.0% | 64.3% | 7953 | 2702.6 ms | budget-fitting loss |
| grimoire-cc-02 | quality | false | 33.3% | 0.0% | 57.1% | 7991 | 2732.2 ms | budget-fitting loss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 61.5% | 7995 | 2607.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | fast | false | 0.0% | 0.0% | 69.2% | 11982 | 14896.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | full | false | 0.0% | 0.0% | 76.5% | 11981 | 14828.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | quality | false | 0.0% | 0.0% | 76.9% | 11829 | 25339.7 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 73.3% | 11859 | 3915.0 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | fast | false | 0.0% | 0.0% | 90.5% | 11945 | 4129.5 ms | budget-fitting loss |
| grimoire-lm-02 | full | false | 0.0% | 50.0% | 87.0% | 11960 | 3808.6 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | quality | false | 0.0% | 0.0% | 90.0% | 11852 | 3797.7 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | lexical | false | 14.3% | 0.0% | 93.3% | 11982 | 3493.0 ms | budget-fitting loss, exact recovery miss |

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
- `grimoire-me-02` / `full`: budget-fitting loss, exact recovery miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: exact recovery miss
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
- `grimoire-ao-01` / `full`: budget-fitting loss, vector ranking miss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: vector ranking miss
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
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
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
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: vector ranking miss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/retrieve/exact.go` symbols `Exact`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`, `uniqueNonOverlapping`, `diversify`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: vector ranking miss
- `grimoire-lm-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: vector ranking miss
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
