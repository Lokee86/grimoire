# AGENTS.md

Grimoire is the unified repository-discovery product. Lexicon and Arcana remain independently usable components with explicit process, state, documentation, and release boundaries.

## Read first

- `README.md`
- `docs/INDEX.md`
- `docs/architecture/system-overview.md`
- `docs/architecture/components.md`
- `docs/architecture/analysis-stack.md`
- `docs/documentation-policy.md`
- `docs/documentation-procedure.md`
- `docs/development/documentation-coverage.md`
- `docs/development/behavioral-contract-matrix.md`

For component work:

- `lexicon/docs/README.md`
- `arcana/docs/README.md`

## Ownership rules

- Grimoire owns unified discovery, prepared source/document state, stable handles, sessions, and cross-component orchestration.
- Lexicon owns language adapters, normalized facts, immutable analysis objects, and snapshots.
- Arcana owns graph ingestion, packed graph state, traversal, impact, paths, overlays, compaction, and graph protocol.
- Lodestone owns native vector storage and exact search.
- Keep component behavior in its owning component rather than adding translation or policy layers at the Grimoire boundary.
- Preserve source and documentation as separate evidence classes.
- Preserve deterministic result ordering, stable handles, snapshot alignment, and bounded progressive investigation.
- Exclude `.worktrees/`, `.workingtrees/`, generated targets, caches, and tool state from repository-wide scans.

## Documentation discipline

Documentation is part of the implementation.

Update the owning product or component documentation in the same change. Update `docs/development/documentation-coverage.md` when a command, package, component, stateful flow, or machine-readable contract changes. Update `docs/development/behavioral-contract-matrix.md` when a durable invariant or protecting test changes. Keep current behavior separate from planning, research, recent-change records, and limitations.

Do not report documentation as complete or current unless all configured product and component checks pass and known semantic gaps are disclosed.

## Completion report

```text
Documentation impact:
- Inspected:
- Updated:
- Not affected:
- Compliance check:
- Known documentation gaps:
```
