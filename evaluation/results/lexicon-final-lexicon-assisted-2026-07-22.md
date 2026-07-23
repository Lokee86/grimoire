# Retrieval evaluation: Lexicon

Generated: 2026-07-22 18:20:07-07:00  
Variant: `final-lexicon-assisted`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 8.3% | 23.8% | 20.0% | 80.9% | 1965.2 ms | 3613.9 ms |
| full | 8.3% | 21.4% | 20.0% | 83.4% | 1914.8 ms | 3498.9 ms |
| lexical | 0.0% | 16.7% | 26.7% | 85.7% | 1887.9 ms | 3540.0 ms |
| quality | 8.3% | 19.0% | 20.0% | 84.6% | 1937.5 ms | 3428.7 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 12.5% | 37.5% | 87.0% | 1931.7 ms | 1990.3 ms |
| call-chain-investigation | 0.0% | 11.1% | 25.0% | 93.8% | 2549.8 ms | 2648.0 ms |
| direct-location | 25.0% | 20.0% | 0.0% | 88.7% | 1088.2 ms | 1130.9 ms |
| long-mixed-query | 0.0% | 37.5% | 37.5% | 75.4% | 3539.8 ms | 3647.7 ms |
| mechanism-explanation | 0.0% | 12.5% | 8.3% | 78.8% | 1883.6 ms | 2144.8 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 16.7% | 50.0% | 83.3% | 1921.9 ms | 1961.0 ms |
| fast/call-chain-investigation | 0.0% | 11.1% | 0.0% | 94.3% | 2620.5 ms | 2658.9 ms |
| fast/direct-location | 33.3% | 20.0% | 0.0% | 89.5% | 1126.2 ms | 1130.8 ms |
| fast/long-mixed-query | 0.0% | 50.0% | 50.0% | 66.7% | 3620.3 ms | 3678.7 ms |
| fast/mechanism-explanation | 0.0% | 10.0% | 0.0% | 79.4% | 1965.0 ms | 1967.6 ms |
| full/architecture-ownership | 0.0% | 16.7% | 50.0% | 83.3% | 1944.8 ms | 1946.5 ms |
| full/call-chain-investigation | 0.0% | 11.1% | 33.3% | 94.1% | 2459.2 ms | 2515.5 ms |
| full/direct-location | 33.3% | 20.0% | 0.0% | 89.5% | 1045.4 ms | 1051.7 ms |
| full/long-mixed-query | 0.0% | 41.7% | 25.0% | 76.1% | 3504.5 ms | 3555.3 ms |
| full/mechanism-explanation | 0.0% | 10.0% | 0.0% | 79.4% | 1880.6 ms | 1886.1 ms |
| lexical/architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 1884.8 ms | 1916.9 ms |
| lexical/call-chain-investigation | 0.0% | 22.2% | 66.7% | 86.7% | 2506.3 ms | 2571.7 ms |
| lexical/direct-location | 0.0% | 20.0% | 0.0% | 85.7% | 1117.5 ms | 1129.2 ms |
| lexical/long-mixed-query | 0.0% | 16.7% | 25.0% | 84.2% | 3543.5 ms | 3574.7 ms |
| lexical/mechanism-explanation | 0.0% | 20.0% | 33.3% | 77.4% | 1855.2 ms | 2120.4 ms |
| quality/architecture-ownership | 0.0% | 16.7% | 50.0% | 83.3% | 1923.6 ms | 1995.7 ms |
| quality/call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 2556.1 ms | 2613.5 ms |
| quality/direct-location | 33.3% | 20.0% | 0.0% | 89.5% | 1053.4 ms | 1095.5 ms |
| quality/long-mixed-query | 0.0% | 41.7% | 50.0% | 76.2% | 3437.4 ms | 3515.5 ms |
| quality/mechanism-explanation | 0.0% | 10.0% | 0.0% | 78.8% | 1871.2 ms | 2113.8 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| lexicon-dl-01 | fast | false | 0.0% | 0.0% | 100.0% | 2902 | 1126.2 ms | budget-fitting loss |
| lexicon-dl-01 | full | false | 0.0% | 0.0% | 100.0% | 2886 | 828.7 ms | budget-fitting loss |
| lexicon-dl-01 | quality | false | 0.0% | 0.0% | 100.0% | 2886 | 1100.2 ms | budget-fitting loss |
| lexicon-dl-01 | lexical | false | 50.0% | 0.0% | 60.0% | 2940 | 1098.5 ms | budget-fitting loss |
| lexicon-dl-02 | fast | false | 0.0% | 0.0% | 100.0% | 2987 | 1077.9 ms | budget-fitting loss, vector ranking miss |
| lexicon-dl-02 | full | false | 0.0% | 0.0% | 100.0% | 2963 | 1052.4 ms | budget-fitting loss, vector ranking miss |
| lexicon-dl-02 | quality | false | 0.0% | 0.0% | 100.0% | 2963 | 1053.4 ms | budget-fitting loss, vector ranking miss |
| lexicon-dl-02 | lexical | false | 0.0% | 0.0% | 100.0% | 2871 | 1130.5 ms | budget-fitting loss |
| lexicon-dl-03 | fast | true | 100.0% | 0.0% | 66.7% | 2947 | 1131.3 ms |  |
| lexicon-dl-03 | full | true | 100.0% | 0.0% | 66.7% | 2927 | 1045.4 ms |  |
| lexicon-dl-03 | quality | true | 100.0% | 0.0% | 66.7% | 2927 | 1049.0 ms |  |
| lexicon-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 2878 | 1117.5 ms | budget-fitting loss |
| lexicon-me-01 | fast | false | 33.3% | 0.0% | 75.0% | 5965 | 1958.2 ms | budget-fitting loss |
| lexicon-me-01 | full | false | 33.3% | 0.0% | 75.0% | 5946 | 1863.0 ms | budget-fitting loss |
| lexicon-me-01 | quality | false | 33.3% | 0.0% | 72.7% | 5939 | 2140.8 ms | budget-fitting loss |
| lexicon-me-01 | lexical | false | 33.3% | 0.0% | 45.5% | 5898 | 2149.8 ms | budget-fitting loss |
| lexicon-me-02 | fast | false | 0.0% | 0.0% | 100.0% | 5945 | 1967.9 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-02 | full | false | 0.0% | 0.0% | 100.0% | 5913 | 1886.7 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-02 | quality | false | 0.0% | 0.0% | 100.0% | 5913 | 1871.2 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-02 | lexical | false | 0.0% | 0.0% | 100.0% | 5927 | 1855.2 ms | budget-fitting loss, embedding miss |
| lexicon-me-03 | fast | false | 0.0% | 0.0% | 63.6% | 5956 | 1965.0 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-03 | full | false | 0.0% | 0.0% | 63.6% | 5947 | 1880.6 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-03 | quality | false | 0.0% | 0.0% | 63.6% | 5979 | 1856.6 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-03 | lexical | false | 25.0% | 100.0% | 90.0% | 5957 | 1833.2 ms | budget-fitting loss, embedding miss |
| lexicon-ao-01 | fast | false | 33.3% | 100.0% | 66.7% | 5979 | 1965.3 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-01 | full | false | 33.3% | 100.0% | 66.7% | 5935 | 1946.7 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-01 | quality | false | 33.3% | 100.0% | 66.7% | 5935 | 1843.5 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 5902 | 1920.5 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-02 | fast | false | 0.0% | 0.0% | 100.0% | 5900 | 1878.5 ms | budget-fitting loss |
| lexicon-ao-02 | full | false | 0.0% | 0.0% | 100.0% | 5991 | 1943.0 ms | budget-fitting loss |
| lexicon-ao-02 | quality | false | 0.0% | 0.0% | 100.0% | 5967 | 2003.7 ms | budget-fitting loss |
| lexicon-ao-02 | lexical | false | 0.0% | 0.0% | 100.0% | 5878 | 1849.1 ms | budget-fitting loss |
| lexicon-cc-01 | fast | false | 0.0% | 0.0% | 100.0% | 7965 | 2663.2 ms | budget-fitting loss |
| lexicon-cc-01 | full | false | 0.0% | 0.0% | 100.0% | 7896 | 2521.7 ms | budget-fitting loss |
| lexicon-cc-01 | quality | false | 0.0% | 0.0% | 100.0% | 7969 | 2619.9 ms | budget-fitting loss |
| lexicon-cc-01 | lexical | false | 33.3% | 50.0% | 86.7% | 7947 | 2578.9 ms | budget-fitting loss |
| lexicon-cc-02 | fast | false | 16.7% | 0.0% | 87.5% | 7921 | 2577.9 ms | budget-fitting loss, vector ranking miss |
| lexicon-cc-02 | full | false | 16.7% | 100.0% | 88.2% | 7961 | 2396.7 ms | budget-fitting loss, vector ranking miss |
| lexicon-cc-02 | quality | false | 0.0% | 0.0% | 100.0% | 7896 | 2492.4 ms | budget-fitting loss, vector ranking miss |
| lexicon-cc-02 | lexical | false | 16.7% | 100.0% | 86.7% | 7882 | 2433.6 ms | budget-fitting loss |
| lexicon-lm-01 | fast | false | 28.6% | 50.0% | 69.6% | 11883 | 3685.2 ms | budget-fitting loss, vector ranking miss |
| lexicon-lm-01 | full | false | 14.3% | 0.0% | 86.4% | 11921 | 3561.0 ms | budget-fitting loss, vector ranking miss |
| lexicon-lm-01 | quality | false | 14.3% | 50.0% | 85.7% | 11997 | 3524.2 ms | budget-fitting loss, vector ranking miss |
| lexicon-lm-01 | lexical | false | 0.0% | 0.0% | 100.0% | 11893 | 3508.8 ms | budget-fitting loss, vector ranking miss |
| lexicon-lm-02 | fast | false | 80.0% | 50.0% | 63.6% | 11942 | 3555.5 ms | budget-fitting loss |
| lexicon-lm-02 | full | false | 80.0% | 50.0% | 66.7% | 11957 | 3448.1 ms | budget-fitting loss |
| lexicon-lm-02 | quality | false | 80.0% | 50.0% | 66.7% | 11907 | 3350.6 ms | budget-fitting loss |
| lexicon-lm-02 | lexical | false | 40.0% | 50.0% | 70.0% | 11996 | 3578.2 ms | budget-fitting loss |

## Concrete failures

- `lexicon-dl-01` / `fast`: budget-fitting loss
  - `internal/objectstore/ingest_parse.go` symbols `ValidateOutput`, `parseOutput`, `validateHeader`, `validateFullHeader`: budget-fitting loss
  - `internal/objectstore/ingest.go` symbols `IngestLanguage`: budget-fitting loss
- `lexicon-dl-01` / `full`: budget-fitting loss
  - `internal/objectstore/ingest_parse.go` symbols `ValidateOutput`, `parseOutput`, `validateHeader`, `validateFullHeader`: budget-fitting loss
  - `internal/objectstore/ingest.go` symbols `IngestLanguage`: budget-fitting loss
- `lexicon-dl-01` / `quality`: budget-fitting loss
  - `internal/objectstore/ingest_parse.go` symbols `ValidateOutput`, `parseOutput`, `validateHeader`, `validateFullHeader`: budget-fitting loss
  - `internal/objectstore/ingest.go` symbols `IngestLanguage`: budget-fitting loss
- `lexicon-dl-01` / `lexical`: budget-fitting loss
  - `internal/objectstore/ingest_parse.go` symbols `ValidateOutput`, `parseOutput`, `validateHeader`, `validateFullHeader`: budget-fitting loss
- `lexicon-dl-02` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`, `Load`: budget-fitting loss
  - `internal/objectstore/manifest.go` symbols `BuildManifest`: vector ranking miss
- `lexicon-dl-02` / `full`: budget-fitting loss, vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`, `Load`: budget-fitting loss
  - `internal/objectstore/manifest.go` symbols `BuildManifest`: vector ranking miss
- `lexicon-dl-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`, `Load`: budget-fitting loss
  - `internal/objectstore/manifest.go` symbols `BuildManifest`: vector ranking miss
- `lexicon-dl-02` / `lexical`: budget-fitting loss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`, `Load`: budget-fitting loss
  - `internal/objectstore/manifest.go` symbols `BuildManifest`: budget-fitting loss
- `lexicon-dl-03` / `lexical`: budget-fitting loss
  - `internal/objectstore/dependencies.go` symbols `ImpactedFiles`, `DependencyScope`, `closure`, `addRelation`: budget-fitting loss
- `lexicon-me-01` / `fast`: budget-fitting loss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-01` / `full`: budget-fitting loss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-01` / `quality`: budget-fitting loss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-01` / `lexical`: budget-fitting loss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-02` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: vector ranking miss
  - `internal/library/header.go` symbols `SetSharedComplete`: budget-fitting loss
- `lexicon-me-02` / `full`: budget-fitting loss, vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: vector ranking miss
  - `internal/library/header.go` symbols `SetSharedComplete`: budget-fitting loss
- `lexicon-me-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: vector ranking miss
  - `internal/library/header.go` symbols `SetSharedComplete`: budget-fitting loss
- `lexicon-me-02` / `lexical`: budget-fitting loss, embedding miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: embedding miss
  - `internal/library/header.go` symbols `SetSharedComplete`: budget-fitting loss
- `lexicon-me-03` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/objectstore/gc.go` symbols `PlanGC`, `GarbageCollect`: budget-fitting loss
  - `internal/objectstore/gc_storage.go` symbols `listSnapshots`, `readConsumerPins`, `listObjects`: vector ranking miss
  - `internal/objectstore/gc_validate.go` symbols `validateGCPlan`, `rejectOverlap`: budget-fitting loss
  - `internal/objectstore/gc_execute.go` symbols `ExecuteGC`, `canonicalGCPlan`: budget-fitting loss
- `lexicon-me-03` / `full`: budget-fitting loss, vector ranking miss
  - `internal/objectstore/gc.go` symbols `PlanGC`, `GarbageCollect`: budget-fitting loss
  - `internal/objectstore/gc_storage.go` symbols `listSnapshots`, `readConsumerPins`, `listObjects`: vector ranking miss
  - `internal/objectstore/gc_validate.go` symbols `validateGCPlan`, `rejectOverlap`: budget-fitting loss
  - `internal/objectstore/gc_execute.go` symbols `ExecuteGC`, `canonicalGCPlan`: budget-fitting loss
- `lexicon-me-03` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/objectstore/gc.go` symbols `PlanGC`, `GarbageCollect`: budget-fitting loss
  - `internal/objectstore/gc_storage.go` symbols `listSnapshots`, `readConsumerPins`, `listObjects`: vector ranking miss
  - `internal/objectstore/gc_validate.go` symbols `validateGCPlan`, `rejectOverlap`: budget-fitting loss
  - `internal/objectstore/gc_execute.go` symbols `ExecuteGC`, `canonicalGCPlan`: budget-fitting loss
- `lexicon-me-03` / `lexical`: budget-fitting loss, embedding miss
  - `internal/objectstore/gc.go` symbols `PlanGC`, `GarbageCollect`: budget-fitting loss
  - `internal/objectstore/gc_storage.go` symbols `listSnapshots`, `readConsumerPins`, `listObjects`: embedding miss
  - `internal/objectstore/gc_execute.go` symbols `ExecuteGC`, `canonicalGCPlan`: budget-fitting loss
- `lexicon-ao-01` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `Ensure`, `Open`, `SourceChanges`, `CommitState`, `RestoreLibrary`: vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`: budget-fitting loss
- `lexicon-ao-01` / `full`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `Ensure`, `Open`, `SourceChanges`, `CommitState`, `RestoreLibrary`: vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`: budget-fitting loss
- `lexicon-ao-01` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `Ensure`, `Open`, `SourceChanges`, `CommitState`, `RestoreLibrary`: vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`: budget-fitting loss
- `lexicon-ao-01` / `lexical`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `Ensure`, `Open`, `SourceChanges`, `CommitState`, `RestoreLibrary`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `Scan`, `scan`, `notifyConsumers`: budget-fitting loss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`: vector ranking miss
- `lexicon-ao-02` / `fast`: budget-fitting loss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: budget-fitting loss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: budget-fitting loss
- `lexicon-ao-02` / `full`: budget-fitting loss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: budget-fitting loss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: budget-fitting loss
- `lexicon-ao-02` / `quality`: budget-fitting loss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: budget-fitting loss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: budget-fitting loss
- `lexicon-ao-02` / `lexical`: budget-fitting loss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: budget-fitting loss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: budget-fitting loss
- `lexicon-cc-01` / `fast`: budget-fitting loss
  - `adapters/gdscript/call_parser.go` symbols `findCalls`, `findCallsInTokens`, `terminalCall`: budget-fitting loss
  - `adapters/gdscript/call_resolution.go` symbols `resolveCall`, `resolveMethods`, `methodTargets`: budget-fitting loss
  - `adapters/gdscript/call_emission.go` symbols `processCalls`, `emitResolvedCall`, `emitPossibleTargets`: budget-fitting loss
- `lexicon-cc-01` / `full`: budget-fitting loss
  - `adapters/gdscript/call_parser.go` symbols `findCalls`, `findCallsInTokens`, `terminalCall`: budget-fitting loss
  - `adapters/gdscript/call_resolution.go` symbols `resolveCall`, `resolveMethods`, `methodTargets`: budget-fitting loss
  - `adapters/gdscript/call_emission.go` symbols `processCalls`, `emitResolvedCall`, `emitPossibleTargets`: budget-fitting loss
- `lexicon-cc-01` / `quality`: budget-fitting loss
  - `adapters/gdscript/call_parser.go` symbols `findCalls`, `findCallsInTokens`, `terminalCall`: budget-fitting loss
  - `adapters/gdscript/call_resolution.go` symbols `resolveCall`, `resolveMethods`, `methodTargets`: budget-fitting loss
  - `adapters/gdscript/call_emission.go` symbols `processCalls`, `emitResolvedCall`, `emitPossibleTargets`: budget-fitting loss
- `lexicon-cc-01` / `lexical`: budget-fitting loss
  - `adapters/gdscript/call_parser.go` symbols `findCalls`, `findCallsInTokens`, `terminalCall`: budget-fitting loss
  - `adapters/gdscript/call_resolution.go` symbols `resolveCall`, `resolveMethods`, `methodTargets`: budget-fitting loss
- `lexicon-cc-02` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/cli/cli.go` symbols `Run`, `runScan`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `Scan`, `scan`, `notifyConsumers`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
  - `internal/adapters/runner.go` symbols `Run`, `command`: vector ranking miss
  - `internal/objectstore/store.go` symbols `Publish`: budget-fitting loss
- `lexicon-cc-02` / `full`: budget-fitting loss, vector ranking miss
  - `internal/cli/cli.go` symbols `Run`, `runScan`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
  - `internal/adapters/runner.go` symbols `Run`, `command`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `publishSnapshot`: budget-fitting loss
  - `internal/objectstore/store.go` symbols `Publish`: vector ranking miss
- `lexicon-cc-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/cli/cli.go` symbols `Run`, `runScan`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `Scan`, `scan`, `notifyConsumers`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
  - `internal/adapters/runner.go` symbols `Run`, `command`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `publishSnapshot`: budget-fitting loss
  - `internal/objectstore/store.go` symbols `Publish`: budget-fitting loss
- `lexicon-cc-02` / `lexical`: budget-fitting loss
  - `internal/cli/cli.go` symbols `Run`, `runScan`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
  - `internal/adapters/runner.go` symbols `Run`, `command`: budget-fitting loss
  - `internal/scan/snapshot.go` symbols `publishSnapshot`: budget-fitting loss
  - `internal/objectstore/store.go` symbols `Publish`: budget-fitting loss
- `lexicon-lm-01` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`: vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
- `lexicon-lm-01` / `full`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`: vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `notifyConsumers`: budget-fitting loss
- `lexicon-lm-01` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`: vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `notifyConsumers`: budget-fitting loss
- `lexicon-lm-01` / `lexical`: budget-fitting loss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`: budget-fitting loss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`: budget-fitting loss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `notifyConsumers`: budget-fitting loss
- `lexicon-lm-02` / `fast`: budget-fitting loss
  - `internal/consumer/registry.go` symbols `RunOne`, `validateDefinition`: budget-fitting loss
- `lexicon-lm-02` / `full`: budget-fitting loss
  - `internal/consumer/registry.go` symbols `RunOne`, `validateDefinition`: budget-fitting loss
- `lexicon-lm-02` / `quality`: budget-fitting loss
  - `internal/consumer/registry.go` symbols `RunOne`, `validateDefinition`: budget-fitting loss
- `lexicon-lm-02` / `lexical`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `notifyConsumers`: budget-fitting loss
  - `internal/consumer/registry.go` symbols `RunOne`, `validateDefinition`: budget-fitting loss
  - `internal/consumer/runner.go` symbols `Run`, `runOne`, `Validate`, `invoke`: budget-fitting loss
