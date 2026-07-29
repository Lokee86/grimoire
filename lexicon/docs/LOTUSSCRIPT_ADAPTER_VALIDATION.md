# LotusScript adapter validation

Validation date: July 28, 2026.

This record describes the corpus calibration that moved the LotusScript adapter from version 0.1.0 to 0.2.0. Measurements are evidence from the pinned revisions below, not permanent performance guarantees.

## Corpus splits

| Split | Repository | Revision | Primary evidence |
| --- | --- | --- | --- |
| Calibration | `MrArtemAA/DUnit` | `e013a21cc511fcadc93afa1e9095ac9da0b50984` | Script libraries, inheritance, `Use` resolution, legacy base64 ODP agents |
| Validation | `dpastov/jsonparser-ls` | `2406bc4431225ef83276700bda853ae247bd9cfd` | Typed local receivers, constructors, list indexing, expression-heavy parser code |
| Holdout | `HCL-TECH-SOFTWARE/volt-mx-ls-toolkit` | `48d0c5b14cef215dabeefa3c182caa11036c0ad8` | Structured DXL agents, script-library imports, cross-file calls, Domino runtime boundaries |

The repositories are bootstrapped by `evaluation/bootstrap_corpus.py`. Semantic gates are declared in `evaluation/corpus.json` and executed by `evaluation/run_validation.py`.

## Calibration findings

The initial adapter exposed four material problems:

1. `%REM` blocks were treated as executable source, producing declarations and calls from documentation text.
2. Boolean operators, `ForAll`, array indexing, and list indexing could be misclassified as calls.
3. Calls through declared local, parameter, and field types remained dynamic even when the target class was present in the repository.
4. Domino ODP agents were skipped because `.lsa` source may be stored either in structured `<lotusscript>` DXL elements or in base64 `$AssistAction` raw-item payloads.

Version 0.2.0 addresses those cases. It also suppresses the duplicate plain-function candidate produced by `New TypeName()` and expands built-in classification for common LotusScript functions.

## Before-and-after measurements

A resolved-call rate is calculated as `calls / (calls + unresolved calls)`. The holdout comparison includes seven additional ODP agents after DXL extraction, so its raw unresolved count is not directly comparable without the coverage and rate columns.

| Repository | Files before | Files after | Call edges before | Call edges after | Unresolved before | Unresolved after | Resolved-call rate before | Resolved-call rate after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| DUnit | 3 | 5 | 137 | 155 | 211 | 170 | 39.4% | 47.7% |
| jsonparser-ls | 1 | 1 | 18 | 20 | 50 | 37 | 26.5% | 35.1% |
| Volt MX toolkit | 6 | 13 | 95 | 214 | 246 | 259 | 27.9% | 45.2% |

## Final semantic results

The three selected corpus cases passed twice with byte-identical output and no missing required nodes, edges, or relations.

| Case | Split | Nodes | Call edges | Import edges | Extends edges | Unresolved calls |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| `lotusscript-dunit` | Calibration | 291 | 155 | 4 | 8 | 170 |
| `lotusscript-jsonparser` | Validation | 86 | 20 | 0 | 0 | 37 |
| `lotusscript-volt-mx-toolkit` | Holdout | 421 | 214 | 12 | 2 | 259 |

The holdout gates verify that a structured DXL agent is extracted, its `Use "NotesHttpJsonRequestHelper"` dependency resolves to the local script-library module, its `Initialize` routine calls a local helper method, and repository-local inheritance remains connected.

## Accepted unresolved evidence

Unresolved calls remain intentional when the repository does not provide enough static evidence. The largest category in the holdout is receiver dispatch through Domino runtime objects such as `NotesSession`, `NotesDocument`, and JSON runtime interfaces. Built-ins remain explicit `builtin-target` records rather than invented declarations.

The adapter does not currently provide:

- content detection for LotusScript stored under generic `.txt`, `.bas`, or `.vb` extensions;
- non-ASCII LMBCS decoding for legacy base64 DXL source payloads;
- assignment-only type inference when a declaration supplies no usable type;
- Domino form, view, field, or design-element graph extraction;
- incremental narrowing.

## Reproduction

```text
python evaluation/bootstrap_corpus.py
python evaluation/run_validation.py --adapter lotusscript --jobs 3
```

Focused adapter verification remains:

```text
cd adapters/lotusscript
go test ./...
```
