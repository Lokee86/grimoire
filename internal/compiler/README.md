# Compiler package

`internal/compiler` converts ranked source selections and structural evidence into the versioned Grimoire context package.

## Entry points

- `Compile` — source-only fixed-budget package.
- `CompileWithEvidence` — fixed-budget package with provider state and structural evidence.
- `CompileAdaptiveWithEvidence` — automatic-budget package with an explicit assembly decision.

## Current schema

The current package version is 5. The package records:

- query and selected budget;
- prepared and embedding identities;
- retrieval and structural sources;
- immutable provider state;
- query profile and retrieval policy;
- adaptive assembly metadata when applicable;
- selected source chunks and structural evidence;
- exact token count; and
- source and structural omission counts.

## Invariants

- Token accounting uses `o200k_base` over the serialized package representation.
- The compiler never exceeds the supplied positive budget.
- Source and structural evidence retain deterministic input order subject to fitting.
- Unsupported package versions must be rejected by consumers.
- Explicit-budget packages do not fabricate adaptive assembly metadata.

## Boundary

The compiler owns fitting, accounting, omission reporting, and serialization. It does not classify queries, rank candidates, decide evidence coverage, call providers, or persist retrieval state.
