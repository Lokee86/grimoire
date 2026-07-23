# Retrieval evaluation: Grimoire

Generated: 2026-07-22 18:02:39-07:00  
Variant: `final-standalone`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 8.3% | 13.3% | 0.0% | 78.7% | 2100.2 ms | 8417.3 ms |
| full | 8.3% | 8.9% | 5.3% | 77.6% | 1917.0 ms | 10806.6 ms |
| lexical | 0.0% | 6.7% | 0.0% | 74.2% | 1907.7 ms | 3685.6 ms |
| quality | 8.3% | 8.9% | 0.0% | 76.1% | 1916.2 ms | 13352.2 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 15.6% | 0.0% | 72.4% | 1890.7 ms | 1947.6 ms |
| call-chain-investigation | 0.0% | 12.5% | 0.0% | 82.0% | 2528.5 ms | 2792.3 ms |
| direct-location | 25.0% | 25.0% | 0.0% | 71.4% | 1144.5 ms | 1501.0 ms |
| long-mixed-query | 0.0% | 1.9% | 4.2% | 81.9% | 3907.5 ms | 23162.6 ms |
| mechanism-explanation | 0.0% | 5.6% | 0.0% | 71.6% | 1935.3 ms | 2198.2 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 25.0% | 0.0% | 66.7% | 1922.4 ms | 1951.1 ms |
| fast/call-chain-investigation | 0.0% | 16.7% | 0.0% | 87.1% | 2777.2 ms | 2822.5 ms |
| fast/direct-location | 33.3% | 33.3% | 0.0% | 70.6% | 1301.9 ms | 1630.6 ms |
| fast/long-mixed-query | 0.0% | 0.0% | 0.0% | 82.4% | 8911.3 ms | 13357.2 ms |
| fast/mechanism-explanation | 0.0% | 11.1% | 0.0% | 77.8% | 2173.4 ms | 2223.0 ms |
| full/architecture-ownership | 0.0% | 0.0% | 0.0% | 85.0% | 1879.9 ms | 1882.9 ms |
| full/call-chain-investigation | 0.0% | 16.7% | 0.0% | 80.6% | 2528.5 ms | 2541.3 ms |
| full/direct-location | 33.3% | 33.3% | 0.0% | 70.6% | 1178.0 ms | 1304.9 ms |
| full/long-mixed-query | 0.0% | 0.0% | 16.7% | 82.5% | 11580.3 ms | 18543.5 ms |
| full/mechanism-explanation | 0.0% | 11.1% | 0.0% | 68.6% | 1945.4 ms | 1955.7 ms |
| lexical/architecture-ownership | 0.0% | 25.0% | 0.0% | 63.2% | 1884.5 ms | 1890.3 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 82.8% | 2475.0 ms | 2503.8 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 76.9% | 1024.1 ms | 1113.4 ms |
| lexical/long-mixed-query | 0.0% | 7.7% | 0.0% | 80.0% | 3699.9 ms | 3829.2 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 65.5% | 1924.4 ms | 1925.2 ms |
| quality/architecture-ownership | 0.0% | 12.5% | 0.0% | 73.7% | 1916.2 ms | 1933.3 ms |
| quality/call-chain-investigation | 0.0% | 16.7% | 0.0% | 77.4% | 2532.9 ms | 2600.7 ms |
| quality/direct-location | 33.3% | 33.3% | 0.0% | 68.8% | 1165.7 ms | 1345.2 ms |
| quality/long-mixed-query | 0.0% | 0.0% | 0.0% | 82.4% | 14432.3 ms | 24153.1 ms |
| quality/mechanism-explanation | 0.0% | 0.0% | 0.0% | 73.5% | 1895.1 ms | 2042.9 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | fast | false | 0.0% | 0.0% | 71.4% | 2971 | 1667.1 ms | budget-fitting loss |
| grimoire-dl-01 | full | false | 0.0% | 0.0% | 71.4% | 2943 | 1319.0 ms | budget-fitting loss |
| grimoire-dl-01 | quality | false | 0.0% | 0.0% | 71.4% | 2943 | 1365.1 ms | budget-fitting loss |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 2990 | 1123.4 ms | budget-fitting loss |
| grimoire-dl-02 | fast | true | 100.0% | 0.0% | 66.7% | 2929 | 1072.2 ms |  |
| grimoire-dl-02 | full | true | 100.0% | 0.0% | 66.7% | 2905 | 1069.3 ms |  |
| grimoire-dl-02 | quality | true | 100.0% | 0.0% | 66.7% | 2905 | 999.5 ms |  |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 60.0% | 2944 | 1019.5 ms | budget-fitting loss |
| grimoire-dl-03 | fast | false | 0.0% | 0.0% | 75.0% | 2999 | 1301.9 ms | budget-fitting loss |
| grimoire-dl-03 | full | false | 0.0% | 0.0% | 75.0% | 2946 | 1178.0 ms | budget-fitting loss |
| grimoire-dl-03 | quality | false | 0.0% | 0.0% | 66.7% | 2880 | 1165.7 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 66.7% | 2917 | 1024.1 ms | budget-fitting loss |
| grimoire-me-01 | fast | false | 0.0% | 0.0% | 100.0% | 5995 | 2173.4 ms | budget-fitting loss |
| grimoire-me-01 | full | false | 50.0% | 0.0% | 72.7% | 5900 | 1945.4 ms | budget-fitting loss |
| grimoire-me-01 | quality | false | 0.0% | 0.0% | 100.0% | 5991 | 2059.3 ms | budget-fitting loss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 87.5% | 5984 | 1925.3 ms | budget-fitting loss |
| grimoire-me-02 | fast | false | 0.0% | 0.0% | 75.0% | 5910 | 2228.5 ms | budget-fitting loss |
| grimoire-me-02 | full | false | 0.0% | 0.0% | 75.0% | 5947 | 1956.9 ms | budget-fitting loss, exact recovery miss |
| grimoire-me-02 | quality | false | 0.0% | 0.0% | 75.0% | 5975 | 1895.1 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 60.0% | 5985 | 1924.4 ms | budget-fitting loss |
| grimoire-me-03 | fast | false | 33.3% | 0.0% | 61.5% | 5975 | 2027.0 ms | budget-fitting loss |
| grimoire-me-03 | full | false | 0.0% | 0.0% | 58.3% | 5868 | 1888.6 ms | budget-fitting loss |
| grimoire-me-03 | quality | false | 0.0% | 0.0% | 50.0% | 5976 | 1891.2 ms | budget-fitting loss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 54.5% | 5899 | 1881.3 ms | budget-fitting loss |
| grimoire-ao-01 | fast | false | 25.0% | 0.0% | 80.0% | 5991 | 1954.3 ms | budget-fitting loss |
| grimoire-ao-01 | full | false | 0.0% | 0.0% | 100.0% | 5956 | 1883.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-ao-01 | quality | false | 0.0% | 0.0% | 100.0% | 5937 | 1897.2 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 5904 | 1890.9 ms | budget-fitting loss |
| grimoire-ao-02 | fast | false | 25.0% | 0.0% | 50.0% | 5903 | 1890.5 ms | budget-fitting loss |
| grimoire-ao-02 | full | false | 0.0% | 0.0% | 70.0% | 5903 | 1876.5 ms | budget-fitting loss |
| grimoire-ao-02 | quality | false | 25.0% | 0.0% | 50.0% | 5940 | 1935.2 ms | budget-fitting loss |
| grimoire-ao-02 | lexical | false | 50.0% | 0.0% | 22.2% | 5896 | 1878.0 ms | budget-fitting loss |
| grimoire-cc-01 | fast | false | 0.0% | 0.0% | 94.1% | 7898 | 2726.8 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | full | false | 0.0% | 0.0% | 94.1% | 7978 | 2514.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | quality | false | 0.0% | 0.0% | 94.1% | 7943 | 2457.6 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 7971 | 2507.0 ms | budget-fitting loss |
| grimoire-cc-02 | fast | false | 33.3% | 0.0% | 78.6% | 7991 | 2827.5 ms | budget-fitting loss |
| grimoire-cc-02 | full | false | 33.3% | 0.0% | 64.3% | 7953 | 2542.8 ms | budget-fitting loss |
| grimoire-cc-02 | quality | false | 33.3% | 0.0% | 57.1% | 7991 | 2608.2 ms | budget-fitting loss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 61.5% | 7993 | 2443.0 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | fast | false | 0.0% | 0.0% | 69.2% | 11992 | 13851.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | full | false | 0.0% | 0.0% | 76.5% | 11981 | 19317.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | quality | false | 0.0% | 0.0% | 76.9% | 11829 | 25233.1 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 73.3% | 11859 | 3843.6 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | fast | false | 0.0% | 0.0% | 90.5% | 11879 | 3971.4 ms | budget-fitting loss |
| grimoire-lm-02 | full | false | 0.0% | 50.0% | 87.0% | 11890 | 3843.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | quality | false | 0.0% | 0.0% | 85.7% | 11971 | 3631.5 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | lexical | false | 14.3% | 0.0% | 86.7% | 11977 | 3556.3 ms | budget-fitting loss, exact recovery miss |

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
