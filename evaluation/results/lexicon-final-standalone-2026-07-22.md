# Retrieval evaluation: Lexicon

Generated: 2026-07-22 18:17:51-07:00  
Variant: `final-standalone`  
Cases: 12  
Runs: 48

## Mode comparison

| Mode | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast | 8.3% | 23.8% | 20.0% | 82.3% | 2007.2 ms | 3742.4 ms |
| full | 8.3% | 21.4% | 20.0% | 83.5% | 1933.4 ms | 3589.2 ms |
| lexical | 0.0% | 14.3% | 26.7% | 86.5% | 1902.7 ms | 3780.6 ms |
| quality | 8.3% | 19.0% | 20.0% | 85.4% | 1968.9 ms | 3632.7 ms |

## Category comparison

| Category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| architecture-ownership | 0.0% | 12.5% | 37.5% | 88.2% | 1935.9 ms | 2014.2 ms |
| call-chain-investigation | 0.0% | 11.1% | 25.0% | 93.8% | 2548.7 ms | 3015.8 ms |
| direct-location | 25.0% | 20.0% | 0.0% | 89.0% | 1110.3 ms | 1586.9 ms |
| long-mixed-query | 0.0% | 35.4% | 37.5% | 77.3% | 3739.5 ms | 3806.4 ms |
| mechanism-explanation | 0.0% | 12.5% | 8.3% | 78.8% | 1931.1 ms | 2031.1 ms |

## Mode by category

| Mode/category | Pass rate | Required recall | Supporting recall | Irrelevant | Median | p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fast/architecture-ownership | 0.0% | 16.7% | 50.0% | 84.0% | 2010.7 ms | 2021.1 ms |
| fast/call-chain-investigation | 0.0% | 11.1% | 0.0% | 94.3% | 2988.5 ms | 3070.2 ms |
| fast/direct-location | 33.3% | 20.0% | 0.0% | 89.5% | 1176.8 ms | 1840.8 ms |
| fast/long-mixed-query | 0.0% | 50.0% | 50.0% | 71.1% | 3746.8 ms | 3785.8 ms |
| fast/mechanism-explanation | 0.0% | 10.0% | 0.0% | 79.4% | 1932.6 ms | 2007.0 ms |
| full/architecture-ownership | 0.0% | 16.7% | 50.0% | 83.3% | 1929.6 ms | 1936.5 ms |
| full/call-chain-investigation | 0.0% | 11.1% | 33.3% | 94.1% | 2632.3 ms | 2752.1 ms |
| full/direct-location | 33.3% | 20.0% | 0.0% | 90.0% | 1059.3 ms | 1276.7 ms |
| full/long-mixed-query | 0.0% | 41.7% | 25.0% | 76.1% | 3600.4 ms | 3701.4 ms |
| full/mechanism-explanation | 0.0% | 10.0% | 0.0% | 79.4% | 1929.5 ms | 1954.0 ms |
| lexical/architecture-ownership | 0.0% | 0.0% | 0.0% | 100.0% | 1872.9 ms | 1899.0 ms |
| lexical/call-chain-investigation | 0.0% | 22.2% | 66.7% | 86.7% | 2492.2 ms | 2525.0 ms |
| lexical/direct-location | 0.0% | 20.0% | 0.0% | 85.7% | 1090.9 ms | 1141.0 ms |
| lexical/long-mixed-query | 0.0% | 8.3% | 25.0% | 86.8% | 3782.1 ms | 3796.3 ms |
| lexical/mechanism-explanation | 0.0% | 20.0% | 33.3% | 77.4% | 1903.4 ms | 2031.5 ms |
| quality/architecture-ownership | 0.0% | 16.7% | 50.0% | 87.5% | 1939.3 ms | 1943.6 ms |
| quality/call-chain-investigation | 0.0% | 0.0% | 0.0% | 100.0% | 2548.7 ms | 2563.7 ms |
| quality/direct-location | 33.3% | 20.0% | 0.0% | 90.0% | 1129.7 ms | 1299.9 ms |
| quality/long-mixed-query | 0.0% | 41.7% | 50.0% | 76.7% | 3649.0 ms | 3794.9 ms |
| quality/mechanism-explanation | 0.0% | 10.0% | 0.0% | 78.8% | 1993.8 ms | 2016.7 ms |

## Per-case results

| Case | Mode | Pass | Required | Supporting | Irrelevant | Tokens | Latency | Failure |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| lexicon-dl-01 | fast | false | 0.0% | 0.0% | 100.0% | 2894 | 1914.6 ms | budget-fitting loss |
| lexicon-dl-01 | full | false | 0.0% | 0.0% | 100.0% | 3000 | 1300.9 ms | budget-fitting loss |
| lexicon-dl-01 | quality | false | 0.0% | 0.0% | 100.0% | 3000 | 1318.8 ms | budget-fitting loss |
| lexicon-dl-01 | lexical | false | 50.0% | 0.0% | 60.0% | 2924 | 1090.9 ms | budget-fitting loss |
| lexicon-dl-02 | fast | false | 0.0% | 0.0% | 100.0% | 2984 | 1050.1 ms | budget-fitting loss, vector ranking miss |
| lexicon-dl-02 | full | false | 0.0% | 0.0% | 100.0% | 2960 | 1059.3 ms | budget-fitting loss, vector ranking miss |
| lexicon-dl-02 | quality | false | 0.0% | 0.0% | 100.0% | 2960 | 1129.7 ms | budget-fitting loss, vector ranking miss |
| lexicon-dl-02 | lexical | false | 0.0% | 0.0% | 100.0% | 2943 | 1146.6 ms | budget-fitting loss |
| lexicon-dl-03 | fast | true | 100.0% | 0.0% | 66.7% | 2931 | 1176.8 ms |  |
| lexicon-dl-03 | full | true | 100.0% | 0.0% | 66.7% | 2911 | 1047.5 ms |  |
| lexicon-dl-03 | quality | true | 100.0% | 0.0% | 66.7% | 2911 | 1030.2 ms |  |
| lexicon-dl-03 | lexical | false | 0.0% | 0.0% | 100.0% | 2873 | 976.3 ms | budget-fitting loss |
| lexicon-me-01 | fast | false | 33.3% | 0.0% | 75.0% | 5939 | 1932.6 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-01 | full | false | 33.3% | 0.0% | 75.0% | 5987 | 1956.7 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-01 | quality | false | 33.3% | 0.0% | 72.7% | 5912 | 1993.8 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-01 | lexical | false | 33.3% | 0.0% | 45.5% | 5860 | 1792.1 ms | budget-fitting loss |
| lexicon-me-02 | fast | false | 0.0% | 0.0% | 100.0% | 5926 | 1916.1 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-02 | full | false | 0.0% | 0.0% | 100.0% | 5890 | 1929.5 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-02 | quality | false | 0.0% | 0.0% | 100.0% | 5890 | 1861.7 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-02 | lexical | false | 0.0% | 0.0% | 100.0% | 5995 | 2045.7 ms | budget-fitting loss, embedding miss |
| lexicon-me-03 | fast | false | 0.0% | 0.0% | 63.6% | 5966 | 2015.3 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-03 | full | false | 0.0% | 0.0% | 63.6% | 5907 | 1911.4 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-03 | quality | false | 0.0% | 0.0% | 63.6% | 5978 | 2019.3 ms | budget-fitting loss, vector ranking miss |
| lexicon-me-03 | lexical | false | 25.0% | 100.0% | 90.0% | 5978 | 1903.4 ms | budget-fitting loss, embedding miss |
| lexicon-ao-01 | fast | false | 33.3% | 100.0% | 66.7% | 5962 | 1999.1 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-01 | full | false | 33.3% | 100.0% | 75.0% | 5977 | 1937.2 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-01 | quality | false | 33.3% | 100.0% | 75.0% | 5977 | 1944.1 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-01 | lexical | false | 0.0% | 0.0% | 100.0% | 5897 | 1901.9 ms | budget-fitting loss |
| lexicon-ao-02 | fast | false | 0.0% | 0.0% | 100.0% | 5999 | 2022.3 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-02 | full | false | 0.0% | 0.0% | 91.7% | 5959 | 1921.9 ms | vector ranking miss |
| lexicon-ao-02 | quality | false | 0.0% | 0.0% | 100.0% | 5951 | 1934.5 ms | budget-fitting loss, vector ranking miss |
| lexicon-ao-02 | lexical | false | 0.0% | 0.0% | 100.0% | 5873 | 1843.9 ms | budget-fitting loss, candidate merge loss |
| lexicon-cc-01 | fast | false | 0.0% | 0.0% | 100.0% | 7965 | 2897.7 ms | budget-fitting loss |
| lexicon-cc-01 | full | false | 0.0% | 0.0% | 100.0% | 7904 | 2499.2 ms | budget-fitting loss |
| lexicon-cc-01 | quality | false | 0.0% | 0.0% | 100.0% | 7964 | 2532.0 ms | budget-fitting loss |
| lexicon-cc-01 | lexical | false | 33.3% | 50.0% | 86.7% | 7997 | 2528.7 ms | budget-fitting loss |
| lexicon-cc-02 | fast | false | 16.7% | 0.0% | 87.5% | 7895 | 3079.3 ms | budget-fitting loss, vector ranking miss |
| lexicon-cc-02 | full | false | 16.7% | 100.0% | 88.2% | 7974 | 2765.4 ms | budget-fitting loss, vector ranking miss |
| lexicon-cc-02 | quality | false | 0.0% | 0.0% | 100.0% | 7891 | 2565.4 ms | budget-fitting loss, vector ranking miss |
| lexicon-cc-02 | lexical | false | 16.7% | 100.0% | 86.7% | 7855 | 2455.7 ms | budget-fitting loss |
| lexicon-lm-01 | fast | false | 28.6% | 50.0% | 73.9% | 11872 | 3790.2 ms | budget-fitting loss, exact recovery miss, vector ranking miss |
| lexicon-lm-01 | full | false | 14.3% | 0.0% | 86.4% | 11905 | 3712.6 ms | budget-fitting loss, exact recovery miss, vector ranking miss |
| lexicon-lm-01 | quality | false | 14.3% | 50.0% | 85.7% | 11983 | 3811.1 ms | budget-fitting loss, exact recovery miss, vector ranking miss |
| lexicon-lm-01 | lexical | false | 0.0% | 0.0% | 100.0% | 11877 | 3766.5 ms | budget-fitting loss, exact recovery miss, vector ranking miss |
| lexicon-lm-02 | fast | false | 80.0% | 50.0% | 68.2% | 11962 | 3703.4 ms | budget-fitting loss |
| lexicon-lm-02 | full | false | 80.0% | 50.0% | 66.7% | 11895 | 3488.2 ms | budget-fitting loss |
| lexicon-lm-02 | quality | false | 80.0% | 50.0% | 68.2% | 11969 | 3486.8 ms | budget-fitting loss |
| lexicon-lm-02 | lexical | false | 20.0% | 50.0% | 75.0% | 11964 | 3797.8 ms | budget-fitting loss |

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
- `lexicon-me-01` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: vector ranking miss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-01` / `full`: budget-fitting loss, vector ranking miss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: vector ranking miss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-01` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: vector ranking miss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-01` / `lexical`: budget-fitting loss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`, `mergeLanguages`: budget-fitting loss
  - `internal/scan/analysis.go` symbols `analyzePlans`, `analyzePlan`, `analysisRequest`: budget-fitting loss
- `lexicon-me-02` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: vector ranking miss
  - `internal/library/header.go` symbols `SetSharedComplete`: vector ranking miss
- `lexicon-me-02` / `full`: budget-fitting loss, vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: vector ranking miss
  - `internal/library/header.go` symbols `SetSharedComplete`: vector ranking miss
- `lexicon-me-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: budget-fitting loss
  - `internal/library/merge_stream.go` symbols `readStream`, `nodeOwners`, `recordOwner`, `directOwner`, `normalize`, `sortRecords`: vector ranking miss
  - `internal/library/header.go` symbols `SetSharedComplete`: vector ranking miss
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
- `lexicon-ao-01` / `lexical`: budget-fitting loss
  - `internal/state/git.go` symbols `Ensure`, `Open`, `SourceChanges`, `CommitState`, `RestoreLibrary`: budget-fitting loss
  - `internal/scan/scanner.go` symbols `Scan`, `scan`, `notifyConsumers`: budget-fitting loss
  - `internal/objectstore/store.go` symbols `Publish`, `Current`: budget-fitting loss
- `lexicon-ao-02` / `fast`: budget-fitting loss, vector ranking miss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: budget-fitting loss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: vector ranking miss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: vector ranking miss
- `lexicon-ao-02` / `full`: vector ranking miss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: vector ranking miss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: vector ranking miss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: vector ranking miss
- `lexicon-ao-02` / `quality`: budget-fitting loss, vector ranking miss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: budget-fitting loss
  - `internal/scan/plan.go` symbols `languageOwnsSource`, `plansFor`: vector ranking miss
  - `internal/scope/repository.go` symbols `Build`, `expandSemanticUnits`, `languageConfig`: vector ranking miss
- `lexicon-ao-02` / `lexical`: budget-fitting loss, candidate merge loss
  - `internal/languages/registry.go` symbols `ForPath`, `OwnsSource`, `SourceExtension`: candidate merge loss
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
- `lexicon-lm-01` / `fast`: budget-fitting loss, exact recovery miss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: vector ranking miss
  - `internal/scope/repository.go` symbols `Build`: vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: exact recovery miss
- `lexicon-lm-01` / `full`: budget-fitting loss, exact recovery miss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`: vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: exact recovery miss
  - `internal/scan/scanner.go` symbols `notifyConsumers`: budget-fitting loss
- `lexicon-lm-01` / `quality`: budget-fitting loss, exact recovery miss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: vector ranking miss
  - `internal/scope/repository.go` symbols `Build`: vector ranking miss
  - `internal/library/merge.go` symbols `Merge`: exact recovery miss
  - `internal/scan/scanner.go` symbols `notifyConsumers`: budget-fitting loss
- `lexicon-lm-01` / `lexical`: budget-fitting loss, exact recovery miss, vector ranking miss
  - `internal/state/git.go` symbols `SourceChanges`: vector ranking miss
  - `internal/scan/languages.go` symbols `languagesForChanges`, `libraryDriftLanguagesFor`: budget-fitting loss
  - `internal/scan/snapshot.go` symbols `adapterDriftLanguages`, `publishSnapshot`: budget-fitting loss
  - `internal/scan/plan.go` symbols `plansFor`: budget-fitting loss
  - `internal/scope/repository.go` symbols `Build`: budget-fitting loss
  - `internal/library/merge.go` symbols `Merge`: exact recovery miss
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
  - `internal/consumer/state.go` symbols `saveSnapshot`: budget-fitting loss
