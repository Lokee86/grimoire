# Documentation Procedure

Parent index: [Grimoire Documentation](INDEX.md)

## Purpose

This document defines the required workflow for Grimoire, Lexicon, and Arcana documentation changes.

## Overview

Documentation changes accompany implementation changes and are applied to the narrowest canonical product or component owner.

## Procedure

1. Identify whether the changed responsibility belongs to Grimoire, Lexicon, Arcana, Lodestone integration, or a cross-component contract.
2. Update exact public reference and implemented architecture owners in the same change.
3. Update the owning component documentation rather than relying on a root summary.
4. Update `docs/development/documentation-coverage.md` when production ownership, commands, packages, flows, or contracts change.
5. Update `docs/development/behavioral-contract-matrix.md` when an invariant or focused protecting test changes.
6. Update benchmarks or research only when the method, corpus, result, or interpretation changes; do not generalize measured outcomes into universal claims.
7. Move implemented work out of planning and record unresolved gaps in limits.
8. Update every affected root or component index and relative link.
9. Run the root, Lexicon, and Arcana documentation checks plus affected implementation tests.
10. Report documentation impact and known gaps explicitly.

## Verification

```bash
python .standards/docs_policy/check.py --repo .
python .standards/docs_policy/check.py --repo . --config docs-standard.lexicon.json
python .standards/docs_policy/check.py --repo . --config docs-standard.arcana.json
python scripts/workflow.py test
```

Use `--changed-from` on the root checker for pull-request change-impact enforcement.

## Related docs

- [Documentation policy](documentation-policy.md)
- [Documentation coverage](development/documentation-coverage.md)
- [Behavioral contract matrix](development/behavioral-contract-matrix.md)
- [Testing and benchmarks](development/testing-and-benchmarks.md)

## Notes

Passing one tree does not establish that the other component trees are current or complete.
