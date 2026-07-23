# Retrieval evaluation: Grimoire

Generated: 2026-07-22 18:05:26-07:00  
Variant: `final-lexicon-assisted`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 0.0% | 13.3% | 21.1% | 81.6% | 1930.9 ms | 8942.7 ms |
| full | 0.0% | 13.3% | 21.1% | 80.6% | 1825.1 ms | 8985.8 ms |
| lexical | 0.0% | 13.3% | 21.1% | 81.0% | 1800.4 ms | 4060.2 ms |
| quality | 0.0% | 13.3% | 21.1% | 81.6% | 1821.1 ms | 13873.9 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 37.5% | 0.0% | 64.0% | 1825.1 ms | 1935.4 ms |
| call-chain-investigation | 0.0% | 0.0% | 0.0% | 92.9% | 2508.0 ms | 2603.1 ms |
| direct-location | 0.0% | 0.0% | 33.3% | 90.9% | 980.4 ms | 1043.3 ms |
| long-mixed-query | 0.0% | 15.4% | 16.7% | 77.8% | 4258.6 ms | 22213.9 ms |
| mechanism-explanation | 0.0% | 11.1% | 40.0% | 80.8% | 1816.5 ms | 1944.3 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 37.5% | 0.0% | 64.0% | 1930.9 ms | 1944.4 ms |
| fast/call-chain-investigation | 0.0% | 0.0% | 0.0% | 93.8% | 2532.8 ms | 2545.9 ms |
| fast/direct-location | 0.0% | 0.0% | 33.3% | 90.9% | 1023.9 ms | 1058.9 ms |
| fast/long-mixed-query | 0.0% | 15.4% | 16.7% | 78.0% | 9477.4 ms | 14289.4 ms |
| fast/mechanism-explanation | 0.0% | 11.1% | 40.0% | 81.4% | 1905.7 ms | 1982.9 ms |
| full/architecture-ownership | 0.0% | 37.5% | 0.0% | 64.0% | 1825.1 ms | 1828.0 ms |
| full/call-chain-investigation | 0.0% | 0.0% | 0.0% | 90.6% | 2455.1 ms | 2493.4 ms |
| full/direct-location | 0.0% | 0.0% | 33.3% | 90.9% | 995.9 ms | 999.0 ms |
| full/long-mixed-query | 0.0% | 15.4% | 16.7% | 79.1% | 9541.6 ms | 14544.2 ms |
| full/mechanism-explanation | 0.0% | 11.1% | 40.0% | 79.1% | 1814.7 ms | 1838.5 ms |
| lexical/architecture-ownership | 0.0% | 37.5% | 0.0% | 64.0% | 1777.5 ms | 1784.6 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 93.5% | 2487.7 ms | 2598.7 ms |
| lexical/direct-location | 0.0% | 0.0% | 33.3% | 90.9% | 951.0 ms | 963.5 ms |
| lexical/long-mixed-query | 0.0% | 15.4% | 16.7% | 76.2% | 4089.9 ms | 4357.0 ms |
| lexical/mechanism-explanation | 0.0% | 11.1% | 40.0% | 81.4% | 1815.5 ms | 1872.5 ms |
| quality/architecture-ownership | 0.0% | 37.5% | 0.0% | 64.0% | 1827.6 ms | 1842.8 ms |
| quality/call-chain-investigation | 0.0% | 0.0% | 0.0% | 93.8% | 2486.9 ms | 2578.2 ms |
| quality/direct-location | 0.0% | 0.0% | 33.3% | 90.9% | 997.7 ms | 1024.5 ms |
| quality/long-mixed-query | 0.0% | 15.4% | 16.7% | 78.0% | 14980.3 ms | 24938.0 ms |
| quality/mechanism-explanation | 0.0% | 11.1% | 40.0% | 81.4% | 1817.4 ms | 1824.1 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| grimoire-dl-01 | fast | false | 0.0% | 0.0% | 100.0% | 2981 | 884.5 ms | budget-fitting loss |
| grimoire-dl-01 | full | false | 0.0% | 0.0% | 100.0% | 2981 | 778.8 ms | budget-fitting loss |
| grimoire-dl-01 | quality | false | 0.0% | 0.0% | 100.0% | 2981 | 777.5 ms | budget-fitting loss |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 2972 | 783.2 ms | budget-fitting loss |
| grimoire-dl-02 | fast | false | 0.0% | 100.0% | 71.4% | 2926 | 1023.9 ms | budget-fitting loss |
| grimoire-dl-02 | full | false | 0.0% | 100.0% | 71.4% | 2926 | 995.9 ms | budget-fitting loss |
| grimoire-dl-02 | quality | false | 0.0% | 100.0% | 71.4% | 2926 | 997.7 ms | budget-fitting loss |
| grimoire-dl-02 | lexical | false | 0.0% | 100.0% | 71.4% | 2917 | 951.0 ms | budget-fitting loss |
| grimoire-dl-03 | fast | false | 0.0% | 0.0% | 100.0% | 2957 | 1062.8 ms | budget-fitting loss |
| grimoire-dl-03 | full | false | 0.0% | 0.0% | 100.0% | 2968 | 999.4 ms | budget-fitting loss |
| grimoire-dl-03 | quality | false | 0.0% | 0.0% | 100.0% | 2966 | 1027.5 ms | budget-fitting loss |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 2947 | 964.9 ms | budget-fitting loss |
| grimoire-me-01 | fast | false | 0.0% | 0.0% | 92.9% | 5988 | 1991.4 ms | budget-fitting loss |
| grimoire-me-01 | full | false | 0.0% | 0.0% | 92.9% | 5991 | 1841.2 ms | budget-fitting loss |
| grimoire-me-01 | quality | false | 0.0% | 0.0% | 92.9% | 5988 | 1817.4 ms | budget-fitting loss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 92.9% | 5989 | 1815.5 ms | budget-fitting loss |
| grimoire-me-02 | fast | false | 25.0% | 100.0% | 50.0% | 5995 | 1905.7 ms | budget-fitting loss |
| grimoire-me-02 | full | false | 25.0% | 100.0% | 50.0% | 5995 | 1814.7 ms | budget-fitting loss |
| grimoire-me-02 | quality | false | 25.0% | 100.0% | 50.0% | 5995 | 1824.8 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 25.0% | 100.0% | 50.0% | 5999 | 1785.0 ms | budget-fitting loss |
| grimoire-me-03 | fast | false | 0.0% | 0.0% | 100.0% | 5934 | 1803.4 ms | budget-fitting loss |
| grimoire-me-03 | full | false | 0.0% | 0.0% | 93.3% | 5999 | 1725.2 ms | budget-fitting loss |
| grimoire-me-03 | quality | false | 0.0% | 0.0% | 100.0% | 5934 | 1807.1 ms | budget-fitting loss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 100.0% | 5925 | 1878.8 ms | budget-fitting loss |
| grimoire-ao-01 | fast | false | 50.0% | 0.0% | 71.4% | 5977 | 1945.9 ms | budget-fitting loss |
| grimoire-ao-01 | full | false | 50.0% | 0.0% | 71.4% | 5963 | 1828.3 ms | budget-fitting loss |
| grimoire-ao-01 | quality | false | 50.0% | 0.0% | 71.4% | 5998 | 1844.5 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 50.0% | 0.0% | 71.4% | 5968 | 1785.3 ms | budget-fitting loss |
| grimoire-ao-02 | fast | false | 25.0% | 0.0% | 54.5% | 5977 | 1915.9 ms | budget-fitting loss |
| grimoire-ao-02 | full | false | 25.0% | 0.0% | 54.5% | 5977 | 1821.9 ms | budget-fitting loss |
| grimoire-ao-02 | quality | false | 25.0% | 0.0% | 54.5% | 5977 | 1810.7 ms | budget-fitting loss |
| grimoire-ao-02 | lexical | false | 25.0% | 0.0% | 54.5% | 5968 | 1769.6 ms | budget-fitting loss |
| grimoire-cc-01 | fast | false | 0.0% | 0.0% | 94.1% | 7898 | 2547.4 ms | budget-fitting loss |
| grimoire-cc-01 | full | false | 0.0% | 0.0% | 88.2% | 7992 | 2497.7 ms | budget-fitting loss |
| grimoire-cc-01 | quality | false | 0.0% | 0.0% | 94.1% | 7925 | 2385.4 ms | budget-fitting loss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 94.1% | 7977 | 2364.2 ms | budget-fitting loss |
| grimoire-cc-02 | fast | false | 0.0% | 0.0% | 93.3% | 7965 | 2518.3 ms | budget-fitting loss |
| grimoire-cc-02 | full | false | 0.0% | 0.0% | 93.3% | 7881 | 2412.5 ms | budget-fitting loss |
| grimoire-cc-02 | quality | false | 0.0% | 0.0% | 93.3% | 7946 | 2588.3 ms | budget-fitting loss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 92.9% | 7922 | 2611.1 ms | budget-fitting loss |
| grimoire-lm-01 | fast | false | 0.0% | 0.0% | 71.4% | 11964 | 14824.1 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | full | false | 0.0% | 0.0% | 75.0% | 11883 | 15100.0 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | quality | false | 0.0% | 0.0% | 71.4% | 11982 | 26044.4 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 66.7% | 11877 | 4386.7 ms | budget-fitting loss |
| grimoire-lm-02 | fast | false | 28.6% | 50.0% | 81.5% | 11874 | 4130.6 ms | budget-fitting loss |
| grimoire-lm-02 | full | false | 28.6% | 50.0% | 81.5% | 11884 | 3983.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | quality | false | 28.6% | 50.0% | 81.5% | 11874 | 3916.2 ms | budget-fitting loss, vector ranking miss |
| grimoire-lm-02 | lexical | false | 28.6% | 50.0% | 81.5% | 11881 | 3793.1 ms | budget-fitting loss |

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
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
- `grimoire-me-01` / `quality`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
- `grimoire-me-01` / `lexical`: budget-fitting loss
  - `internal/embedding/query.go` symbols `ParseQueryMode`, `PlanQuery`, `queryWindows`, `Validate`: budget-fitting loss
  - `internal/embedding/query_batch.go` symbols `EmbedQueryPlan`, `queryBatches`, `embedQueryBatch`: budget-fitting loss
- `grimoire-me-02` / `fast`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-02` / `full`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-02` / `quality`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-02` / `lexical`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`, `Marshal`: budget-fitting loss
- `grimoire-me-03` / `fast`: budget-fitting loss
  - `internal/app/vector_build.go` symbols `runVectorBuild`, `embedMissing`, `writeVectorRecords`: budget-fitting loss
  - `internal/app/vector_ingest.go` symbols `ingestVectorBatch`: budget-fitting loss
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
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
- `grimoire-ao-01` / `full`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
- `grimoire-ao-01` / `quality`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
- `grimoire-ao-01` / `lexical`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/selection/selection.go` symbols `Curate`: budget-fitting loss
- `grimoire-ao-02` / `fast`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-ao-02` / `full`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-ao-02` / `quality`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-ao-02` / `lexical`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `AggregateRuns`: budget-fitting loss
- `grimoire-cc-01` / `fast`: budget-fitting loss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: budget-fitting loss
- `grimoire-cc-01` / `full`: budget-fitting loss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: budget-fitting loss
- `grimoire-cc-01` / `quality`: budget-fitting loss
  - `cmd/grimoire/main.go` symbols `main`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/context.go` symbols `runContext`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidates`, `queryVectorCandidates`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `curateContextCandidates`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `Marshal`: budget-fitting loss
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
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-cc-02` / `full`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-cc-02` / `quality`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-cc-02` / `lexical`: budget-fitting loss
  - `internal/app/run.go` symbols `Run`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `parseEvaluationModes`, `packageSelections`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`: budget-fitting loss
- `grimoire-lm-01` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: budget-fitting loss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: budget-fitting loss
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
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: vector ranking miss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
- `grimoire-lm-01` / `lexical`: budget-fitting loss
  - `internal/app/eval_retrieval.go` symbols `runEval`, `validateEvaluationCase`, `parseEvaluationModes`, `packageSelections`, `writeEvaluationSummary`: budget-fitting loss
  - `internal/evaluation/model.go` symbols `FormatVersion`, `Evidence`, `Case`, `Corpus`, `CaseRun`, `Report`: budget-fitting loss
  - `internal/evaluation/corpus.go` symbols `LoadCorpus`, `validateEvidence`: budget-fitting loss
  - `internal/evaluation/score.go` symbols `ScoreCase`, `classifyEvidenceFailure`, `AggregateRuns`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Write`, `Markdown`: budget-fitting loss
  - `internal/app/context_evaluation.go` symbols `evaluateContext`, `chunksToEvaluation`, `candidatesToEvaluation`, `selectionsToEvaluation`: budget-fitting loss
- `grimoire-lm-02` / `fast`: budget-fitting loss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: budget-fitting loss
- `grimoire-lm-02` / `full`: budget-fitting loss, vector ranking miss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: vector ranking miss
- `grimoire-lm-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: vector ranking miss
- `grimoire-lm-02` / `lexical`: budget-fitting loss
  - `internal/embedding/query.go` symbols `PlanQuery`, `queryWindows`: budget-fitting loss
  - `internal/app/context_semantic.go` symbols `semanticCandidatesForEvaluation`, `searchQueryVectors`, `mergeSemanticHits`: budget-fitting loss
  - `internal/app/context_candidates.go` symbols `mergeContextCandidates`, `contextCandidateSources`: budget-fitting loss
  - `internal/compiler/compiler.go` symbols `Compile`, `stabilizeTokenCount`: budget-fitting loss
  - `internal/evaluation/report.go` symbols `BuildAggregates`, `Markdown`: budget-fitting loss
