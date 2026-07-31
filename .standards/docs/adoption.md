# Adoption and Enforcement

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document defines how repositories adopt and enforce the shared documentation standard.

## Overview

The shared repository owns the normative standard, profiles, checker, reusable Pitlord policies, templates, and adoption registry. Each product repository owns its current documentation, local specialization, exceptions, and CI configuration.

## Adoption steps

1. Select a primary profile and capabilities from [Repository profiles](profiles.md).
2. Add `docs-standard.json` using the template.
3. Ensure `README.md`, `AGENTS.md`, documentation index, policy, procedure, planning, limits, and coverage owners exist or add explicit justified exemptions.
4. Configure the repository's root and folder index convention.
5. Map production paths to canonical current documentation.
6. Add focused change-impact rules.
7. Run the shared checker and record existing debt explicitly.
8. Configure Demon Docs for index, link, frontmatter, move, and managed-surface maintenance when appropriate.
9. Include the reusable Pitlord documentation policy or equivalent repository rules.
10. Add CI checks and remove baselines as debt is repaired.

## Tool ownership

### Shared checker

Owns deterministic cross-repository checks for required paths, indexes, direct index coverage, relative links, required sections, configured coverage owners, and change-impact mappings.

The canonical implementation lives in `tools/docs_policy/`. `tools/sync_checker.py` publishes generated snapshots under each adopted repository's `.standards/` directory so CI does not depend on a sibling checkout or an unavailable network resource. The generated manifest records exact SHA-256 identities. Generated snapshots are not edited directly.

### Demon Docs

Owns deterministic maintenance and verification of documentation indexes, links, frontmatter, moves, reverse indexes, managed codemaps, and repository-local document policy.

### Pitlord

Owns repository-level governance rules, including required paths, forbidden paths/content, required policy language, and architecture-policy integration. Reusable policies live under `policies/pitlord/`.

### Warlock

Should surface compliance state, failed checks, missing owners, and repair actions. It is a presentation and orchestration surface, not the canonical policy owner.

## CI contract

The minimum CI gate inside an adopted repository is:

```bash
python .standards/docs_policy/check.py --repo .
```

Refresh the generated snapshot from the canonical repository with:

```bash
python tools/sync_checker.py /path/to/repository
```

Pull requests should also run:

```bash
python .standards/docs_policy/check.py --repo . --changed-from origin/main
```

Repositories using Demon Docs and Pitlord add their local commands after the shared check.

## Baselines

Create or refresh an explicit baseline with:

```bash
python /path/to/engineering-standards/tools/docs_policy/check.py \
  --repo . \
  --write-baseline docs-standard.baseline.json
```

Then set `baseline` in `docs-standard.json`. A baseline may suppress known legacy failures during rollout, but:

- every suppressed item remains visible as debt;
- new failures are blocked;
- a baseline is not evidence of compliance;
- the repository policy states the removal plan;
- semantic gaps are never hidden solely by a machine baseline;
- normal check output reports the number of baselined findings;
- fixed findings become stale baseline entries and should be removed from the baseline.

## Adoption registry

`adoption/repositories.json` records the selected profile, enforcement state, and known migration status for active repositories. It is coordination data, not a substitute for repository-local configuration.

## Related docs

- [Documentation standard](documentation-standard.md)
- [Repository profiles](profiles.md)
- [Completeness and status claims](completeness.md)

## Notes

Repository-local rules may be stricter. Space Rocks' type-specific service/protocol/data/design taxonomy and Demon Docs' coverage and behavioral-contract rules remain valid specializations.
