# Query shape and assembly

Grimoire can select a context target and assembly policy without an LLM. The policy is deterministic, inspectable, and emitted in the context package.

## Activation

`grimoire context` has two modes:

- `--budget 0` or omitted: activate automatic query-shape policy and evidence-coverage assembly.
- positive `--budget`: retain fixed fit-to-budget behavior while emitting the calculated policy in shadow form.

The evaluation runner follows the same distinction. `grimoire eval retrieval --adaptive` replaces case budgets with automatic targets; a fixed `--budget` override cannot be combined with `--adaptive`.

## Profile signals

The profile records:

- task intent;
- specificity;
- breadth;
- ambiguity;
- cross-system scope;
- required evidence types; and
- deterministic reasons for each decision.

Prompt semantics are evaluated separately from candidate ranking. After retrieval, ranking confidence, path dispersion, and structural dispersion can widen or narrow the provisional scope.

## Current tiers

| Scope | Minimum | Target | Maximum | Intended shape |
| --- | ---: | ---: | ---: | --- |
| Focused | 2,000 | 3,000 | 6,000 | Concrete symbol, path, or narrowly localized question |
| Bounded | 3,000 | 6,000 | 10,000 | Mechanism spanning a small number of related regions |
| Exploratory | 6,000 | 12,000 | 18,000 | Architecture, ownership, call-chain, or cross-system investigation |

The target is the current automatic budget. Minimum and maximum values are policy safety bounds and calibration metadata; callers cannot currently request a min/max range through the CLI.

## Evidence assembly

Automatic assembly receives curated candidates in ranked order and preserves enough alternatives for final package fitting. It applies different coverage requirements by scope:

- Focused assembly stays near an exact or highest-ranked anchor region.
- Bounded assembly requires evidence from at least two represented regions.
- Exploratory assembly requires evidence from at least three represented regions.

Each scope also has deterministic candidate and structural-evidence caps. Assembly stops when minimum coverage and reserve requirements are satisfied, when a cap is reached, or when the candidate set is exhausted.

The compiler still enforces the selected token boundary. Evidence coverage determines when assembly has supplied enough alternatives; token accounting determines what survives in the serialized package.

## Package metadata

Automatic packages record:

- scope;
- candidates considered and selected;
- candidate tokens retained for fitting;
- structural evidence considered and selected;
- represented path regions and evidence roles; and
- stop reason.

Explicit-budget packages omit the assembly decision because they retain the fixed candidate path.

## Determinism and limits

The classifier is heuristic rather than learned. Identical queries and retrieval state produce identical profiles and policies. The current three tiers and coverage rules are deliberately concrete, but they remain calibration targets rather than proof that a package is sufficient for every task.

See [Current limitations](../limits/current-limitations.md) and [Retrieval quality](../development/retrieval-quality.md).
