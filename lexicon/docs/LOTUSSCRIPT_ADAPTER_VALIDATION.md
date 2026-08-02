# LotusScript adapter validation

Parent index: [Lexicon Documentation](README.md)

## Purpose

This document records calibration and validation evidence for the Lexicon LotusScript adapter.

## Overview

The evidence uses retained calibration, validation, and holdout repositories to test declaration extraction, script-library scope, call resolution, inheritance, and conservative dataflow.

## Research status

Retained dated adapter-calibration evidence. It informs current semantic limits but is not a universal language guarantee.

## Question

Does the LotusScript adapter preserve repository structure without inventing cross-library calls or dataflow, while recovering enough definite relationships for downstream graph use?

## Method

Run pinned corpus splits twice, require byte-identical output, enforce selected node and edge gates, audit every cross-file call/read/write against the transitive `Use` graph, and retain unresolved evidence when static proof is unavailable.

## Corpus or inputs

Validation date: August 1, 2026.

| Split | Repository | Revision | Primary evidence |
| --- | --- | --- | --- |
| Calibration | `MrArtemAA/DUnit` | `e013a21cc511fcadc93afa1e9095ac9da0b50984` | Script libraries, inheritance, `Use` resolution, visibility, legacy base64 ODP agents |
| Validation | `dpastov/jsonparser-ls` | `2406bc4431225ef83276700bda853ae247bd9cfd` | Typed receivers, constructors, class fields, array/list indexing, expression-heavy parser code |
| Holdout | `HCL-TECH-SOFTWARE/volt-mx-ls-toolkit` | `48d0c5b14cef215dabeefa3c182caa11036c0ad8` | Structured DXL agents, transitive imports, cross-file calls/dataflow, Domino runtime boundaries |

The repositories are bootstrapped by `evaluation/bootstrap_corpus.py`. Semantic gates are declared in `evaluation/corpus.json` and executed by `evaluation/run_validation.py`.

## Audit findings

The version 0.3.0 audit found and corrected several graph-distorting behaviors:

1. Unqualified declarations and duplicate class names were resolved across the entire repository rather than through the script-library import scope.
2. `Use` edges existed, but call and inheritance resolution did not use them.
3. module visibility, `Option Public`, and private declarations were not enforced.
4. class maps were keyed by class name rather than stable class identity, allowing same-named classes to collide.
5. `Static` and `ReDim` declarations could be misclassified, unmodified class/`Type` members were omitted, and colon-separated statements and `With` member shorthand were incomplete.
6. no `reads` or `writes` relationships were emitted.

Version 0.3.0 now uses transitive `Use` scope for cross-file resolution, models declaration visibility, preserves same-name class boundaries, handles the audited syntax forms, and emits conservative symbol-based reads and writes.

## Results

All three cases passed twice with byte-identical output and no missing required nodes, edges, or relations. The gates now require `reads` and `writes` in addition to the prior call/import/inheritance expectations and include selected dataflow exemplar edges.

| Case | Split | Nodes | Calls | Imports | Extends | Reads | Writes | Unresolved calls |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `lotusscript-dunit` | Calibration | 292 | 159 | 4 | 8 | 375 | 73 | 161 |
| `lotusscript-jsonparser` | Validation | 87 | 20 | 0 | 0 | 157 | 77 | 37 |
| `lotusscript-volt-mx-toolkit` | Holdout | 433 | 217 | 12 | 2 | 792 | 307 | 253 |

A separate relationship audit found:

| Case | Cross-file calls | Calls outside transitive import scope | Cross-file reads/writes | Dataflow outside transitive import scope |
| --- | ---: | ---: | ---: | ---: |
| DUnit | 64 | 0 | 11 | 0 |
| jsonparser-ls | 0 | 0 | 0 | 0 |
| Volt MX toolkit | 127 | 0 | 44 | 0 |

The holdout gates verify that a structured DXL agent is extracted, its local script-library dependencies resolve, repository-local inheritance remains connected, and imported constants or variables produce dataflow only through visible libraries.

## Accepted unresolved evidence

Unresolved calls remain intentional when the repository does not provide enough static evidence. The largest categories are receiver dispatch through Domino runtime objects such as `NotesSession` and `NotesDocument`, external product interfaces, and built-in functions. Built-ins remain explicit `builtin-target` records rather than invented declarations.

The adapter does not currently provide:

- content detection for LotusScript stored under generic `.txt`, `.bas`, or `.vb` extensions;
- non-ASCII LMBCS decoding for legacy base64 DXL source payloads;
- assignment-only type inference when a declaration supplies no usable type;
- implicit undeclared-variable modeling;
- alias analysis, container element identity, or interprocedural value flow;
- Domino form, view, field, or design-element graph extraction;
- complete computed or runtime dispatch reconstruction;
- incremental narrowing.

## Reproduction

```text
python evaluation/bootstrap_corpus.py
python evaluation/run_validation.py --adapter lotusscript --jobs 3
```

Focused adapter verification:

```text
cd adapters/lotusscript
go test ./...
```

## Limitations

The available LotusScript corpus is narrow, source conventions vary substantially, and dynamic Notes/Domino behavior cannot be reconstructed completely from static source.

## Interpretation

The accepted result is scoped repository-local evidence with explicit unresolved cases, not complete runtime reconstruction. The important graph guarantee is that definite cross-file edges require visibility through the script-library import graph.

## Retained artifacts

Corpus definitions, semantic gates, repeat-run summaries, unresolved classifications, and reproduction commands are retained in the referenced evaluation outputs.

## Related docs

- [LotusScript adapter README](../adapters/lotusscript/README.md)
- [Semantic acceptance gates](SEMANTIC_ACCEPTANCE.md)
- [Semantic corpus validation](SEMANTIC_CORPUS_VALIDATION.md)
- [Development and verification](DEVELOPMENT.md)

## Notes

Results apply only to the recorded corpus, revisions, adapter implementation, and acceptance policy.
