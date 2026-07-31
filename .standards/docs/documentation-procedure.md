# Documentation Procedure

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document defines the required process for documentation work.

## Overview

Use this procedure whenever implementation, public behavior, ownership, state, operations, testing, planning, research, or known limitations change.

## Procedure

1. **Identify the changed responsibility.** Determine the public surface, runtime owner, state owner, flow, invariant, or operational behavior affected.
2. **Classify each fact.** Separate guide, reference, architecture, operations, domain/design, data/protocol, development, research, planning, limits, and agent material.
3. **Find the canonical owner.** Reuse an existing document before creating another. Place information where it is owned, not merely where it is consumed.
4. **Decide whether a new boundary is justified.** Create a file only for a durable independently discoverable concern. Create a folder only for a durable boundary expected to contain several documents.
5. **Update indexes with the change.** Add, move, or remove every affected direct-file and direct-folder entry.
6. **Write the required shape.** Explain behavior and ownership in prose; use tables, code maps, commands, schemas, and examples as supporting evidence.
7. **Update implementation coverage.** Add or revise coverage-map entries for packages, commands, stateful flows, persistence, concurrency, machine-readable contracts, and recovery seams.
8. **Update behavioral contracts.** When a critical invariant or protecting test changes, update the behavioral-contract matrix.
9. **Graduate shipped work.** Move implemented facts out of planning. Retain research evidence but update the current or planning owner with the resulting decision.
10. **Record unresolved reality.** Put active defects and transitional gaps in limits; do not hide them in notes or leave them implied.
11. **Remove stale material.** Delete replaced legacy docs, stale duplicate claims, graduated stubs, obsolete links, and empty non-stub folders.
12. **Run compliance checks.** Run the repository's Demon Docs checks, shared documentation checker, Pitlord policy, and normal test gate as configured.
13. **Report documentation impact.** State what was inspected, updated, unaffected, checked, and still missing.

## Reuse before creation

Create a new document only when all are true:

- the concern has a clear durable owner;
- it has enough substance to stand alone;
- adding it elsewhere would blur ownership;
- readers will reasonably seek it independently.

Use `notes.md` for temporary or unclassified information. Use `stubs/` only when the eventual owner is clear but the document is not canonical yet.

## Planning and research graduation

When planned work ships:

```text
write or update current guide/reference/architecture/operations owners
update the plan status
link to implemented references
remove future-tense statements that are no longer true
leave only unresolved or later work in planning
```

When research changes a decision:

```text
retain the evidence and limitations in research
update the current or planning owner with the decision
avoid generalizing beyond the measured population
```

## Final verification

Before completion, verify:

```text
The type and owner are correct.
Every direct file and folder is indexed.
Relative links resolve.
Current behavior is not owned only by planning or research.
Architecture documents explain ownership, flow, state, invariants, and failure behavior.
Public surfaces have exact reference owners.
Stateful and recovery flows have focused current owners.
Coverage and behavioral-contract maps are current.
The root README remains an entry point.
Known gaps are disclosed.
The configured compliance checks pass.
```

## Related docs

- [Documentation standard](documentation-standard.md)
- [Change-impact rules](change-impact.md)
- [Adoption and enforcement](adoption.md)

## Notes

A documentation-only move can still break links, examples, fixtures, generated indexes, and codemap extraction. It requires verification.
