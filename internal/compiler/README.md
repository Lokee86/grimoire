# Compiler package

`internal/compiler` converts ranked source selections and structural evidence into the versioned Grimoire context package.

## Entry points

- `Compile` — source-only fixed-budget package.
- `CompileWithEvidence` — fixed-budget package with provider state and structural evidence.
- `CompileAdaptiveWithEvidence` — automatic-budget package with an explicit assembly decision.
- `CompileAdaptiveWithEvidenceConfig` — paired evaluator entry point for final-fitting policy comparisons.

## Current schema

The current package version is 8. The package records:

- query and selected budget;
- prepared and embedding identities;
- retrieval and structural sources;
- immutable provider state;
- query profile and retrieval policy;
- adaptive assembly metadata when applicable;
- selected source chunks and structural evidence;
- facet identities and protected facet claims;
- required linked-span group identities and protection summaries;
- facet-protection, protected-file-depth, and companion-depth summaries;
- exact token count; and
- source, structural, facet, and required-group omission counts.

## Adaptive fitting

Coverage-aware adaptive packages protect one primary implementation owner for each available query facet before spending the remaining budget on repeated evidence. Tests, documentation, and configuration remain eligible later, but cannot displace a primary owner solely because their lexical score is higher. When a candidate belongs to a provider-declared required source-link group, the complete group is attempted atomically at that candidate's existing rank rather than being elevated ahead of stronger evidence.

Single-facet or exploratory mechanism investigations may protect a second distinct implementation file for the same facet. Call-chain investigations stay at one protected file because held-out calibration showed that extra call-chain breadth displaced correct route and connection evidence. Bounded multi-facet requests keep one protected file per facet so secondary files cannot consume every owner slot. Each protected mechanism, call-chain, or direct-location file may also receive one same-file companion chunk when it contributes a lexical term not already represented by that file.

For semantic-only facet plans, file ranking remains authoritative, but declaration-bearing chunks are preferred within the selected file over headers or continuations. When semantic evidence provides no lexical score details, one declaration-bearing same-file companion may be retained to complete the owner. Lexical plans retain their calibrated candidate order. Architecture and mixed-intent owners do not receive this completion. All protection is deterministic and bounded; it does not change retrieval or ranking.

## Invariants

- Token accounting uses `o200k_base` over the serialized package representation.
- The compiler never exceeds the supplied positive budget.
- Legacy and fixed-budget packages retain deterministic input order.
- Coverage-aware adaptive packages retain deterministic required-group, facet, and companion order subject to fitting.
- A required linked group is retained or omitted as a complete unit; final trimming never leaves a partial protected group.
- Unsupported package versions must be rejected by consumers.
- Explicit-budget packages do not fabricate adaptive assembly metadata.

## Boundary

The compiler owns final fitting, accounting, omission reporting, and serialization. It does not classify queries, rank candidates, decide the candidate pool, call providers, or persist retrieval state.
