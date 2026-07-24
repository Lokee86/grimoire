# Lexicon runtime evidence contract v1

Lexicon runtime evidence is a separate UTF-8 JSON Lines stream that records observed relationships against a specific static fact snapshot. Runtime evidence supplements static facts; it never rewrites or upgrades them by itself.

## Header

The first record is:

```json
{"record":"lexicon-runtime","repository":"example/module","run_id":"integration-2026-07-23","schema_version":1,"static_snapshot":"sha256:..."}
```

Required fields:

- `record`: always `lexicon-runtime`;
- `schema_version`: integer `1`;
- `repository`: the same repository identity used by the static fact stream;
- `run_id`: a caller-defined stable identity for this captured run.

Optional fields:

- `static_snapshot`: the Lexicon snapshot or manifest identity used to instrument the program;
- `build_id`: a build or artifact identity;
- `started_at` and `ended_at`: RFC 3339 timestamps;
- `attributes`: deterministic scalar values or sorted scalar arrays describing the environment.

A runtime stream describes observations from one run. Combining runs is a consumer operation and must preserve each `run_id`.

## Observation record

```json
{"count":4,"record":"observation","relation":"calls","source":"sha256:...","target":"sha256:..."}
```

Required fields:

- `record`: `observation`;
- `relation`: `calls`, `reads`, or `writes`;
- `source`: a static Lexicon node ID;
- `count`: a positive integer.

Exactly one target form is required:

- `target`: a static Lexicon node ID; or
- `external_target`: a deterministic runtime identity when no static node can be matched.

Optional fields:

- `first_seen_ns` and `last_seen_ns`: non-negative monotonic offsets from the capture start;
- `thread`: a deterministic thread, task, actor, or fiber identity;
- `owner`: repository-relative source ownership when instrumentation can retain it;
- `attributes`: deterministic scalar values or sorted scalar arrays, such as receiver type, instrumentation provider, sampling mode, or external-target reason.

Observations are aggregated by `(relation, source, target form, thread, attributes)`. Producers emit one record per aggregate rather than one record per event. Records are sorted by `(relation, source, target, external_target, thread, canonical attributes)` so ties remain deterministic.

## Reconciliation semantics

For observed `calls`:

- a matching static `calls` edge is **confirmed-definite**;
- a matching static `possible-calls` edge is **confirmed-possible** for that run, but remains `possible-calls` in the static graph;
- a known source and target without either static edge is **unmodeled-static-target**;
- an `external_target` is **external-runtime-target**;
- an unknown static source or target is **unknown-static-id** and indicates stale instrumentation, a mismatched snapshot, or corrupt evidence.

For observed `reads` and `writes`, the same categories apply using the corresponding static relation. There is no `possible-reads` or `possible-writes` relation in v1.

Absence of an observation is never proof that a static edge is impossible. A run may have incomplete coverage, sampling, disabled instrumentation, or an unexecuted path. Consumers must not delete, downgrade, or narrow static edges solely because they were not observed.

## Instrumentation boundary

Language adapters remain static analyzers and do not embed runtime tracing. Instrumentation providers may be language-specific, but they must map observations back to Lexicon stable IDs or emit an explicit `external_target`. Providers must not fabricate static IDs for generated or external code.

This separation allows Arcana and other consumers to retain a stable static graph while attaching run-specific evidence, coverage, and discrepancy reports.
