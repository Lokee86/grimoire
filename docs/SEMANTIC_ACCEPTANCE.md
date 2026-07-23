# Semantic analysis acceptance gates

Lexicon semantic work is accepted by observable fact behavior rather than parser implementation claims.

## Dataflow

Every supported adapter family must have fixtures that prove:

- a resolved value use emits `reads`;
- a resolved assignment target emits `writes`;
- compound updates emit both relations;
- member or field access resolves only with sound local evidence;
- lexical shadowing selects the nearest declaration;
- unresolved, external, and built-in names do not receive fabricated targets.

JavaScript and Svelte are covered through the TypeScript adapter but require their own fixture coverage.

## Dispatch

Each language must test the dispatch mechanisms it actually provides. Formal interface or trait languages must distinguish contract declarations from concrete runtime targets. Dynamic languages must retain unsupported reflection, monkey patching, and runtime mutation as unresolved evidence.

A single proven target emits `calls`. Multiple sound concrete targets emit `possible-calls`. An interface or trait declaration must not be substituted for an implementation target merely because its method name matches.

## Dependencies

Each ecosystem must test normal dependencies, development or test dependencies, local path dependencies where supported, malformed or dynamic declarations, and deterministic ordering. Manifest readers never execute project code.

`imports` remains source-level import evidence. `depends-on` records package, module, plugin, resource, or manifest dependency evidence and does not replace `imports`.

## Runtime evidence

Runtime observations use `spec/runtime-evidence-v1.md`. Acceptance requires validation of canonical ordering, repository and stable-ID matching, and reconciliation into confirmed, possible, unmodeled, external, or stale-ID categories. A missing runtime observation is never negative evidence against a static edge.

## Integration

Before merging a semantic stream:

1. run its adapter-specific tests;
2. run all root and adapter suites affected by the changes;
3. validate representative emitted JSONL through `tools/validate_jsonl.py`;
4. summarize emitted relations through `tools/semantic_report.py`;
5. verify deterministic output by comparing two identical runs;
6. inspect unresolved counts and explicitly document unsupported forms.

A stream is incomplete when its promised relation count is zero for its acceptance fixtures, even when its tests otherwise pass.

The repeatable real-repository implementation of these gates lives under `evaluation/`. Run `python evaluation/bootstrap_corpus.py`, `python evaluation/run_tests.py`, and `python evaluation/run_validation.py --jobs 3`. A successful complete validation replaces the tracked `evaluation/validation/baseline.json`; details and current results are documented in `SEMANTIC_CORPUS_VALIDATION.md`.
