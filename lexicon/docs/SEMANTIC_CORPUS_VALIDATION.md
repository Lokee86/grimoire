# Cross-adapter semantic corpus validation

Validated on July 23, 2026 from `C:\!bin\workspace` using the tracked corpus manifest and harness under `evaluation/`.

## Result

All 12 non-Go corpus cases passed:

- every adapter completed both scans;
- every JSONL output passed contract validation;
- all repeated outputs were byte-identical;
- every required relation was present;
- every expected-zero relation remained absent;
- no case-level execution failure occurred.

The Go adapter is covered separately by `GO_ADAPTER_VALIDATION.md`, including two real repositories and repeat-run determinism.

## Corpus results

| Case | Split | Calls | Possible calls | Reads | Writes | Dependencies | Unresolved calls | Output |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| GDScript / Alien Attack | Calibration | 3 | 0 | 17 | 25 | 2 | 51 | 100 KB |
| GDScript / Space Rocks client | Validation | 14,434 | 2,167 | 10,343 | 9,726 | 735 | 11,588 | 40.2 MB |
| GDScript / Speedy Saucer | Holdout | 0 | 0 | 2 | 2 | 2 | 7 | 15 KB |
| Python / doc-ledger | Calibration | 633 | 39 | 1,420 | 955 | 44 | 1,356 | 3.0 MB |
| Python / Space Rocks tools | Validation | 2,176 | 37 | 2,788 | 1,988 | 191 | 2,750 | 7.3 MB |
| Ruby / Lexicon adapter | Calibration | 626 | 216 | 705 | 477 | 1 | 1,903 | 2.3 MB |
| Ruby / Space Rocks API | Validation | 876 | 453 | 1,083 | 912 | 31 | 3,509 | 5.2 MB |
| Rust / Grimoire vector engine | Calibration | 190 | 5 | 383 | 160 | 15 | 643 | 1.0 MB |
| Rust / Arcana | Validation | 1,636 | 147 | 2,499 | 1,222 | 6 | 3,885 | 7.4 MB |
| TypeScript / workspace-mcp | Calibration | 1,250 | 169 | 2,033 | 834 | 132 | 2,875 | 5.1 MB |
| TypeScript/Svelte / Lexicanter | Validation | 16,513 | 3,872 | 22,924 | 10,845 | 191 | 13,885 | 41.5 MB |
| TypeScript / Space Rocks Astro | Holdout | 186 | 19 | 494 | 148 | 34 | 757 | 1.5 MB |

The authoritative machine-readable values and SHA-256 output identities are stored in `evaluation/validation/baseline.json`.

## Defects exposed by the corpus

The first corpus attempt found two GDScript defects that fixture-only testing had not exposed:

1. malformed or incomplete call syntax could reach call parsing with no matching close parenthesis and panic;
2. dataflow resolution selected the first same-named local or member encountered in Go map iteration, making large-repository output nondeterministic and occasionally binding to a later local or an ambiguous field from another candidate owner.

The parser now rejects unterminated calls. Local dataflow resolution selects the nearest prior declaration and falls back to the function parameter. Member dataflow emits a definite edge only when the inferred owners produce exactly one repository member target. Focused regressions cover both behaviors.

Before the dataflow fix, repeated Space Rocks client scans differed by 77 edges. After the fix, both 40 MB outputs had SHA-256 `bc91f069f6811270d3728bc1be41315305a3d0005ec02a290ccc6bb648559550`.

## Interpretation

This run establishes that the added call, possible-call, read, write, dependency, inheritance, override, and related semantic streams are implemented, survive representative repositories, and are deterministic for the current corpus.

It does not establish perfect precision or recall. High unresolved-call counts are expected for built-ins, external libraries, dynamic dispatch, generated code, and forms the adapters intentionally decline to guess. Future calibration should sample those categories, label false positives and false negatives, and update fixtures or resolution rules only when the evidence is language-general.

## Reproduction

```text
python evaluation/bootstrap_corpus.py
python evaluation/run_tests.py
python evaluation/run_validation.py --jobs 3
```

Generated per-case summaries and audit samples are written beneath `evaluation/validation/generated/`. The complete run updates the tracked baseline only when every gate passes.
