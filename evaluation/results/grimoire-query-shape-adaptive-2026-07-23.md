# Retrieval evaluation: Grimoire

Generated: 2026-07-23 13:44:57-07:00  
Variant: `adaptive`  
Cases: 12  
Runs: 12  
Structural providers: ``

## Mode comparison

| Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 8.3% | 2.2% | 0.0% | 96.9% | 0.0% | 1212.4 ms | 1982.8 ms |

## Package comparison

| Mode | Median tokens | p95 tokens | Median chunks | Median budget use |
| --- | ---: | ---: | ---: | ---: |
| lexical | 8715 | 11990 | 9.5 | 99.2% |

## Pre-curation source ranking

These metrics score the retrieved order before exact-result merging, curation, and package fitting.

| Mode | Queries | Required R@10 | Required R@20 | MRR | Relevant @10 | Relevant @20 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical | 12 | 0.0% | 8.3% | 0.015 | 0.8% | 2.1% |

## Category comparison

| Category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1698.3 ms | 1793.0 ms |
| call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1202.5 ms | 1421.5 ms |
| direct-location | 33.3% | 33.3% | 0.0% | 87.0% | 0.0% | 1139.1 ms | 1154.0 ms |
| long-mixed-query | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1695.5 ms | 2151.3 ms |
| mechanism-explanation | 0.0% | 0.0% | 0.0% | 95.0% | 0.0% | 1235.8 ms | 1251.5 ms |

## Mode by category

| Mode/category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexical/architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1698.3 ms | 1793.0 ms |
| lexical/call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1202.5 ms | 1421.5 ms |
| lexical/direct-location | 33.3% | 33.3% | 0.0% | 87.0% | 0.0% | 1139.1 ms | 1154.0 ms |
| lexical/long-mixed-query | 0.0% | 0.0% | 0.0% | 100.0% | 0.0% | 1695.5 ms | 2151.3 ms |
| lexical/mechanism-explanation | 0.0% | 0.0% | 0.0% | 95.0% | 0.0% | 1235.8 ms | 1251.5 ms |

## Per-case results

| Case | Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Budget | Tokens | Curated | Assembled | Stop | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | --- |
| grimoire-dl-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5947 | 202 | 151 | bounded evidence coverage satisfied | 1139.1 ms | budget-fitting loss |
| grimoire-dl-02 | lexical | true | 100.0% | 0.0% | 66.7% | 0.0% | 6000 | 5941 | 202 | 156 | bounded evidence coverage satisfied | 1118.2 ms |  |
| grimoire-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5998 | 201 | 134 | bounded evidence coverage satisfied | 1155.6 ms | embedding miss |
| grimoire-me-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 5930 | 204 | 122 | bounded evidence coverage satisfied | 1115.6 ms | budget-fitting loss |
| grimoire-me-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 6000 | 6000 | 201 | 130 | bounded evidence coverage satisfied | 1235.8 ms | exact recovery miss, vector ranking miss |
| grimoire-me-03 | lexical | false | 0.0% | 0.0% | 87.5% | 0.0% | 6000 | 5953 | 201 | 127 | bounded evidence coverage satisfied | 1253.3 ms | budget-fitting loss |
| grimoire-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11718 | 205 | 98 | exploratory evidence coverage satisfied | 1593.0 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-ao-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11941 | 204 | 102 | exploratory evidence coverage satisfied | 1803.5 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-cc-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11988 | 202 | 91 | exploratory evidence coverage satisfied | 1445.8 ms | candidate merge loss, embedding miss, vector ranking miss |
| grimoire-cc-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11876 | 205 | 64 | exploratory evidence coverage satisfied | 959.3 ms | budget-fitting loss, embedding miss, vector ranking miss |
| grimoire-lm-01 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11993 | 204 | 64 | exploratory evidence coverage satisfied | 2201.9 ms | embedding miss, exact recovery miss |
| grimoire-lm-02 | lexical | false | 0.0% | 0.0% | 100.0% | 0.0% | 12000 | 11430 | 202 | 68 | exploratory evidence coverage satisfied | 1189.0 ms | embedding miss, exact recovery miss |

## Query profile shadow output

These classifications are observational and do not change retrieval, curation, or package assembly.

| Case | Mode | Expected | Actual | Match | Specificity | Breadth | Ambiguity | Subsystems | Graph regions | Budget mode | Mismatches |
| --- | --- | --- | --- | ---: | --- | --- | --- | ---: | ---: | --- | --- |
| grimoire-dl-01 | lexical | bounded | bounded | true | medium | low | medium | 5 | 0 | automatic |  |
| grimoire-dl-02 | lexical | bounded | bounded | true | medium | low | medium | 5 | 0 | automatic |  |
| grimoire-dl-03 | lexical | bounded | bounded | true | medium | low | low | 2 | 0 | automatic |  |
| grimoire-me-01 | lexical | bounded | bounded | true | medium | medium | medium | 6 | 0 | automatic |  |
| grimoire-me-02 | lexical | bounded | bounded | true | medium | medium | medium | 2 | 0 | automatic |  |
| grimoire-me-03 | lexical | bounded | bounded | true | medium | medium | low | 3 | 0 | automatic |  |
| grimoire-ao-01 | lexical | exploratory | exploratory | true | medium | high | low | 3 | 0 | automatic |  |
| grimoire-ao-02 | lexical | exploratory | exploratory | true | high | high | low | 1 | 0 | automatic |  |
| grimoire-cc-01 | lexical | exploratory | exploratory | true | high | high | low | 8 | 0 | automatic |  |
| grimoire-cc-02 | lexical | exploratory | exploratory | true | medium | high | low | 3 | 0 | automatic |  |
| grimoire-lm-01 | lexical | exploratory | exploratory | true | high | high | low | 6 | 0 | automatic |  |
| grimoire-lm-02 | lexical | exploratory | exploratory | true | medium | high | low | 2 | 0 | automatic |  |

## Query profile calibration

| Mode | Judged profiles | Matches | Match rate |
| --- | ---: | ---: | ---: |
| lexical | 12 | 12 | 100.0% |

## Concrete failures

- `grimoire-dl-01` / `lexical`: budget-fitting loss
  - `internal/app/vector_manifest.go` symbols `validateVectorSnapshotManifest`, `validateVectorEngineInfo`, `validateVectorSnapshotManifestFields`: budget-fitting loss
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
