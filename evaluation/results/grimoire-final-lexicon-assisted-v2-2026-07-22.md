# Retrieval evaluation: Grimoire

Generated: 2026-07-22 18:14:57-07:00  
Variant: `final-lexicon-assisted-v2`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 8.3% | 13.3% | 0.0% | 79.3% | 1966.6 ms | 9441.4 ms |
| full | 8.3% | 8.9% | 5.3% | 76.9% | 1907.5 ms | 8947.8 ms |
| lexical | 0.0% | 6.7% | 0.0% | 75.8% | 1884.5 ms | 3684.5 ms |
| quality | 8.3% | 8.9% | 0.0% | 76.1% | 1894.7 ms | 13626.5 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 15.6% | 0.0% | 72.4% | 1905.0 ms | 1987.4 ms |
| call-chain-investigation | 0.0% | 12.5% | 0.0% | 83.6% | 2583.3 ms | 2727.4 ms |
| direct-location | 25.0% | 25.0% | 0.0% | 71.4% | 1031.6 ms | 1164.2 ms |
| long-mixed-query | 0.0% | 1.9% | 4.2% | 81.9% | 4309.1 ms | 21755.5 ms |
| mechanism-explanation | 0.0% | 5.6% | 0.0% | 71.4% | 1901.2 ms | 2129.7 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 25.0% | 0.0% | 66.7% | 1933.8 ms | 1956.3 ms |
| fast/call-chain-investigation | 0.0% | 16.7% | 0.0% | 90.3% | 2652.8 ms | 2654.9 ms |
| fast/direct-location | 33.3% | 33.3% | 0.0% | 70.6% | 1145.1 ms | 1183.2 ms |
| fast/long-mixed-query | 0.0% | 0.0% | 0.0% | 82.4% | 10012.9 ms | 15156.0 ms |
| fast/mechanism-explanation | 0.0% | 11.1% | 0.0% | 77.1% | 1974.4 ms | 2004.6 ms |
| full/architecture-ownership | 0.0% | 0.0% | 0.0% | 85.0% | 1907.5 ms | 1944.8 ms |
| full/call-chain-investigation | 0.0% | 16.7% | 0.0% | 83.9% | 2607.4 ms | 2689.5 ms |
| full/direct-location | 33.3% | 33.3% | 0.0% | 70.6% | 1029.8 ms | 1080.5 ms |
| full/long-mixed-query | 0.0% | 0.0% | 16.7% | 80.0% | 9507.4 ms | 14544.0 ms |
| full/mechanism-explanation | 0.0% | 11.1% | 0.0% | 65.7% | 1858.9 ms | 2236.7 ms |
| lexical/architecture-ownership | 0.0% | 25.0% | 0.0% | 63.2% | 1922.8 ms | 1994.8 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 82.8% | 2447.3 ms | 2461.2 ms |
| lexical/direct-location | 0.0% | 0.0% | 0.0% | 76.9% | 1033.4 ms | 1036.5 ms |
| lexical/long-mixed-query | 0.0% | 7.7% | 0.0% | 83.3% | 3694.4 ms | 3783.8 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 69.0% | 1854.8 ms | 1908.2 ms |
| quality/architecture-ownership | 0.0% | 12.5% | 0.0% | 73.7% | 1889.2 ms | 1900.0 ms |
| quality/call-chain-investigation | 0.0% | 16.7% | 0.0% | 77.4% | 2628.1 ms | 2731.4 ms |
| quality/direct-location | 33.3% | 33.3% | 0.0% | 68.8% | 1015.2 ms | 1052.9 ms |
| quality/long-mixed-query | 0.0% | 0.0% | 0.0% | 82.4% | 14660.6 ms | 23967.3 ms |
| quality/mechanism-explanation | 0.0% | 0.0% | 0.0% | 73.5% | 1888.3 ms | 1945.6 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | fast | false | 0.0% | 0.0% | 71.4% | 2976 | 941.5 ms | budget-fitting loss |
| grimoire-dl-01 | full | false | 0.0% | 0.0% | 71.4% | 2948 | 870.2 ms | budget-fitting loss |
| grimoire-dl-01 | quality | false | 0.0% | 0.0% | 71.4% | 2948 | 869.2 ms | budget-fitting loss |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 2872 | 964.6 ms | budget-fitting loss |
| grimoire-dl-02 | fast | true | 100.0% | 0.0% | 66.7% | 2908 | 1145.1 ms |  |
| grimoire-dl-02 | full | true | 100.0% | 0.0% | 66.7% | 2884 | 1086.2 ms |  |
| grimoire-dl-02 | quality | true | 100.0% | 0.0% | 66.7% | 2884 | 1015.2 ms |  |
| grimoire-dl-02 | lexical | false | 0.0% | 0.0% | 60.0% | 2970 | 1033.4 ms | budget-fitting loss |
| grimoire-dl-03 | fast | false | 0.0% | 0.0% | 75.0% | 2990 | 1187.4 ms | budget-fitting loss |
| grimoire-dl-03 | full | false | 0.0% | 0.0% | 75.0% | 2962 | 1029.8 ms | budget-fitting loss |
| grimoire-dl-03 | quality | false | 0.0% | 0.0% | 66.7% | 2886 | 1057.1 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 66.7% | 2922 | 1036.9 ms | budget-fitting loss |
| grimoire-me-01 | fast | false | 0.0% | 0.0% | 100.0% | 5862 | 1974.4 ms | budget-fitting loss |
| grimoire-me-01 | full | false | 50.0% | 0.0% | 72.7% | 5940 | 1858.0 ms | budget-fitting loss |
| grimoire-me-01 | quality | false | 0.0% | 0.0% | 100.0% | 5996 | 1951.9 ms | budget-fitting loss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 87.5% | 5984 | 1810.9 ms | budget-fitting loss |
| grimoire-me-02 | fast | false | 0.0% | 0.0% | 75.0% | 5996 | 2007.9 ms | budget-fitting loss |
| grimoire-me-02 | full | false | 0.0% | 0.0% | 66.7% | 5892 | 2278.6 ms | budget-fitting loss |
| grimoire-me-02 | quality | false | 0.0% | 0.0% | 75.0% | 5962 | 1888.3 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 70.0% | 6000 | 1914.1 ms | budget-fitting loss |
| grimoire-me-03 | fast | false | 33.3% | 0.0% | 61.5% | 5979 | 1942.7 ms | budget-fitting loss |
| grimoire-me-03 | full | false | 0.0% | 0.0% | 58.3% | 5874 | 1858.9 ms | budget-fitting loss |
| grimoire-me-03 | quality | false | 0.0% | 0.0% | 50.0% | 5979 | 1879.7 ms | budget-fitting loss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 54.5% | 5904 | 1854.8 ms | budget-fitting loss |
| grimoire-ao-01 | fast | false | 25.0% | 0.0% | 80.0% | 5925 | 1958.8 ms | budget-fitting loss |
| grimoire-ao-01 | full | false | 0.0% | 0.0% | 100.0% | 5975 | 1866.1 ms | budget-fitting loss |
| grimoire-ao-01 | quality | false | 0.0% | 0.0% | 100.0% | 5945 | 1877.2 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 5976 | 2002.8 ms | budget-fitting loss |
| grimoire-ao-02 | fast | false | 25.0% | 0.0% | 50.0% | 5959 | 1908.8 ms | budget-fitting loss |
| grimoire-ao-02 | full | false | 0.0% | 0.0% | 70.0% | 5871 | 1949.0 ms | budget-fitting loss |
| grimoire-ao-02 | quality | false | 25.0% | 0.0% | 50.0% | 5988 | 1901.2 ms | budget-fitting loss |
| grimoire-ao-02 | lexical | false | 50.0% | 0.0% | 22.2% | 5955 | 1842.8 ms | budget-fitting loss |
| grimoire-cc-01 | fast | false | 0.0% | 0.0% | 100.0% | 7906 | 2650.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | full | false | 0.0% | 0.0% | 100.0% | 7990 | 2516.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | quality | false | 0.0% | 0.0% | 100.0% | 7942 | 2513.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 7976 | 2462.8 ms | budget-fitting loss |
| grimoire-cc-02 | fast | false | 33.3% | 0.0% | 78.6% | 7905 | 2655.1 ms | budget-fitting loss |
| grimoire-cc-02 | full | false | 33.3% | 0.0% | 64.3% | 7984 | 2698.7 ms | budget-fitting loss |
| grimoire-cc-02 | quality | false | 33.3% | 0.0% | 50.0% | 7980 | 2742.8 ms | budget-fitting loss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 61.5% | 7839 | 2431.9 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | fast | false | 0.0% | 0.0% | 69.2% | 11999 | 15727.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | full | false | 0.0% | 0.0% | 76.5% | 11997 | 15103.6 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | quality | false | 0.0% | 0.0% | 71.4% | 11982 | 25001.3 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 73.3% | 11875 | 3793.8 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | fast | false | 0.0% | 0.0% | 90.5% | 11972 | 4298.3 ms | budget-fitting loss |
| grimoire-lm-02 | full | false | 0.0% | 50.0% | 82.6% | 11986 | 3911.3 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | quality | false | 0.0% | 0.0% | 90.0% | 11879 | 4319.8 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | lexical | false | 14.3% | 0.0% | 93.3% | 11881 | 3595.0 ms | budget-fitting loss, exact recovery miss |

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
