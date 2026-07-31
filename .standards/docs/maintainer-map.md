# Engineering Standards Maintainer Map

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document routes maintainers to the canonical documentation and implementation boundary for common changes to the shared engineering standards.

## Overview

Use this map when the owning standard or enforcement surface is unclear. It is a navigation aid, not a replacement for the normative standard pages, checker implementation, templates, or repository-local adoption policy.

## Change-area routing

| Change area | Canonical documentation | Primary implementation boundary | Verification |
| --- | --- | --- | --- |
| Documentation taxonomy, required shapes, and ownership rules | [Documentation standard](documentation-standard.md) | `docs/documentation-standard.md` | Shared checker self-check |
| Repository-type requirements and capability additions | [Repository profiles](profiles.md) | `docs/profiles.md`, `templates/docs-standard.json` | Shared checker self-check |
| Documentation workflow and agent behavior | [Documentation procedure](documentation-procedure.md) | `docs/documentation-procedure.md`, `skills/documentation/SKILL.md`, `templates/AGENTS-documentation.md` | Shared checker self-check |
| Change-to-document impact rules | [Change-impact rules](change-impact.md) | `docs/change-impact.md`, repository `docs-standard.json` mappings | Changed-from checker tests |
| Completeness claims and legacy baselines | [Completeness and status claims](completeness.md) | `tools/docs_policy/baseline.py`, `tools/docs_policy/check.py` | `tools/docs_policy/tests/` |
| Required paths, sections, indexes, links, and coverage | [Adoption and enforcement](adoption.md) | `tools/docs_policy/`, `policies/pitlord/documentation-core.json` | `tools/docs_policy/tests/` and self-check |
| Vendored standard snapshots | [Adoption and enforcement](adoption.md) | `tools/sync_checker.py` | Sync into an isolated adopted repository and run its check |
| Repository adoption inventory | [Adoption and enforcement](adoption.md) | `adoption/repositories.json` | Validate adopted repository checks |

## Boundaries

- Normative rules belong in `docs/`; enforcement mechanics belong in `tools/docs_policy/` and `policies/`.
- Templates illustrate compliant repository-local surfaces but do not override the normative documents.
- Vendored `.standards/` directories are generated snapshots and are not edited directly.
- Repository-local documentation policies may specialize the standard only through explicit documented exceptions.

## Related docs

- [Documentation standard](documentation-standard.md)
- [Documentation procedure](documentation-procedure.md)
- [Adoption and enforcement](adoption.md)

## Notes

Use this map to find the owner first. Use the owning page or implementation tests for the detailed contract.
