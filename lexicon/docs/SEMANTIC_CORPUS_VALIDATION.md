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

## C-family calibration added July 24, 2026

Two pinned C calibration cases supplement the July 23 cross-adapter baseline:

| Case | Definite calls | Possible-call edges | Macro references | Passes to | Reads | Writes | Unresolved calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Git `9a0c470` | 65,235 | 2,720 | 24,923 | 44,615 | 262,277 | 56,674 | 23,435 |
| Codebase Memory C backend `97ce23f` | 31,125 | 23 | 3,109 | 42,472 | 116,753 | 14,745 | 22,476 |

Raw edge counts are not a call-site coverage percentage: one indirect call site can emit several possible targets while retaining an unresolved dynamic record, and one macro invocation can contribute several body calls at the same source span. The site-level report therefore classifies each observed expression and expansion expression separately. Git contains 88,511 such sites; 66,314 have repository-callable evidence, of which 64,456 have definite targets, 1,858 have possible targets only, and 3 retain both classes. That is 97.2% definite and 100% definite-or-possible for repository-target sites. No Git or CBM call remains classified as `missing-target`.

The C-family passes now include include-closure translation units for included `.c` fragments and headers, bounded macro aliases and function-like bodies, nested argument substitution, function-pointer typedef propagation, and bounded target flow from initializers, assignments, designated dispatch tables, callback arguments, and macro-expanded calls. Conditional macro and indirect function-pointer targets remain possible rather than definite. External APIs and genuinely dynamic calls remain unresolved by design.

The CBM case intentionally targets `internal/cbm`, the independently meaningful C backend, rather than mixing duplicate frontend/application definitions and generated vendored grammars into one judgment surface. Its unresolved groups remain dominated by C-library and Tree-sitter APIs.

The C cases use exact node and source-target edge judgments, relation-count gates, and expected-zero unresolved-call-reason gates. They protect included-C translation units, header-to-source static calls, macro-mediated body calls, macro aliases, function-pointer typedefs and dispatch tables, definition selection, includer-driven header attribution, and repository-local include resolution.

## C-family semantic depth added July 25, 2026

Adapter 0.5.0 adds a bounded semantic interpretation of function-like macros rather than replaying a compiler preprocessor. It records macro parameters and body calls, binds invocation arguments, substitutes safe identifier tokens, follows nested repository-visible macros to a fixed depth, and emits each provable call at the invocation span. Every expanded call retains its macro-definition span, expansion chain, body callee, substituted arguments, and ordinary resolution evidence. Unsupported token pasting, stringification, variadic substitution, cycles, and argument mismatches remain explicit unresolved records.

| Case | Macro-mediated occurrences / pairs | Body-call occurrences / pairs | Substitution-backed occurrences / pairs | Passes-to occurrences / pairs | Maximum macro depth |
| --- | ---: | ---: | ---: | ---: | ---: |
| Git | 6,549 / 4,541 | 5,190 / 3,471 | 5,157 / 3,451 | 44,615 / 36,992 | 2 |
| Codebase Memory C backend | 7,763 / 43 | 7,763 / 43 | 7,475 / 41 | 42,472 / 8,554 | 0 |
| LevelDB | 0 / 0 | 0 / 0 | 0 / 0 | 1,784 / 1,296 | 0 |
| fmt | 1,106 / 407 | 1,094 / 395 | 1,106 / 407 | 2,801 / 2,154 | 5 |
| Catch2 `src/` | 0 / 0 | 0 / 0 | 0 / 0 | 512 / 445 | 0 |
| nlohmann/json holdout | 0 / 0 | 0 / 0 | 0 / 0 | 480 / 267 | 0 |

The CBM backend is the direct comparison target. The previous Lexicon graph contained 5,076 unique definite function-call pairs; adapter 0.5.0 contains 5,115, a net gain of 39. Its 43 macro-body pairs cover the macro-mediated class previously observed only in CBM while preserving substantially deeper occurrence-level provenance. The same graph includes 8,554 unique direct caller-value-to-callee-parameter pairs, which CBM's compact C graph does not represent.

All six cases passed twice with byte-identical output. The nlohmann/json holdout retained its prior 1,857 definite calls and zero call-level `missing-target` records while gaining direct argument-flow evidence, supporting that the new flow model is language-general rather than CBM-specific.

## C++ corpus added July 24, 2026

Four pinned C++ cases add distinct semantic pressure without tuning against the holdout:

| Case | Split | Definite calls | Possible-call edges | Extends | Reads | Writes | Unresolved calls |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| LevelDB `99b3c03` | Calibration | 4,623 | 8,453 | 13 | 13,255 | 3,224 | 4,564 |
| fmt `407c905` | Calibration | 6,088 | 76,205 | 232 | 17,892 | 4,176 | 7,612 |
| Catch2 `src/` `191fa38` | Validation | 1,890 | 4,790 | 87 | 7,401 | 1,360 | 2,163 |
| nlohmann/json `include/nlohmann/` `55f9368` | Holdout | 1,737 | 4,079 | 34 | 5,638 | 841 | 2,082 |

LevelDB represents conventional production C++ with classes, virtual interfaces, callbacks, and separate headers and implementations. fmt provides concentrated template, constexpr, specialization, and overload pressure. Catch2 exercises macro-heavy framework code and reporter, matcher, generator, and session hierarchies. nlohmann/json remains a header-only holdout so general improvements can be checked against code that was not used to select changes.

A second calibration pass added conservative callable-arity pruning, explicit namespace/type qualification, proven enclosing-type ownership, and direct receiver types from parameters, locals, and directly owned fields. Unsupported or ambiguous evidence falls back to the prior candidate set. Site-level results changed as follows:

| Case | Definite sites | Possible sites | p90 fanout | Missing targets |
| --- | ---: | ---: | ---: | ---: |
| LevelDB | 4,623 → 4,855 | 1,752 → 1,520 | 10 → 10 | 0 → 0 |
| fmt | 6,088 → 6,889 | 4,697 → 3,897 | 92 → 13 | 55 → 0 |
| Catch2 `src/` | 1,890 → 2,184 | 1,150 → 856 | 7 → 6 | 12 → 0 |
| nlohmann/json holdout | 1,737 → 1,857 | 1,119 → 999 | 8 → 7 | 48 → 0 |

The holdout improved without repository-specific rules. Remaining unresolved calls are classified as external, dynamic, or ambiguous rather than missing repository targets. Each case runs twice and is byte-identical. Exact gates protect representative class and function nodes plus concrete call and inheritance edges, and all four C++ cases now require zero unresolved call `missing-target` records. Catch2 is scoped to `src/` to exclude the generated amalgamation in `extras/`; nlohmann/json is scoped to `include/nlohmann/` to isolate the distributed library surface.

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
