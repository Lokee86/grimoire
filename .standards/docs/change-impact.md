# Change-Impact Rules

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document defines when implementation changes require documentation changes in the same work.

## Overview

Documentation impact is determined by changed responsibility, not by file extension or commit label. `docs-standard.json` maps repository paths to required documentation owners and lets CI reject implementation changes that omit all mapped documentation.

## Mandatory triggers

| Change | Required documentation impact |
| --- | --- |
| CLI, API, packet, schema, config, default, diagnostic, or exit behavior | Update exact reference and affected guide |
| Ownership, dependency, process, lifecycle, or data flow | Update architecture and codemap |
| Persistent state, storage format, transaction, mutation, or migration | Update architecture, reference, operations, recovery, and coverage |
| Concurrency, scheduling, watch, daemon, or subprocess behavior | Update lifecycle/operations, failure recovery, and behavioral contracts |
| Build, test, fixture, CI, packaging, or release process | Update development or release documentation |
| Deployment, environment, health, logging, backup, or troubleshooting | Update operations |
| Material UI workflow or layout | Update user workflow and current screenshots |
| New defect, blocker, or transitional gap | Update limits |
| Future proposal or unresolved decision | Update planning only; do not present it as current |
| Research result or benchmark | Update research and any product/planning decision it changes |
| Package, component, or command added, removed, renamed, or reassigned | Update coverage map, ownership index, and codemap |
| Critical invariant or protecting test changes | Update behavioral-contract matrix |

## No-impact claims

A report may state that documentation was not affected only when it identifies the inspected owner and explains why the observable contract, ownership, flow, state, operations, or extension seam did not change.

Invalid reasons include:

```text
small change
internal change
only refactoring
no time
README still looks fine
```

A behavior-preserving refactor may still require codemap or ownership updates when files, packages, or seams moved.

## Configuration

Each change rule in `docs-standard.json` contains code path globs and one or more documentation path globs. When `--changed-from` is supplied, the checker requires at least one mapped documentation path to change whenever a code path matches.

Example:

```json
{
  "paths": ["internal/storage/**"],
  "docs": [
    "docs/architecture/**",
    "docs/development/documentation-coverage.md"
  ]
}
```

Rules should map to concrete owners, not a blanket `docs/**` escape hatch.

## Related docs

- [Documentation procedure](documentation-procedure.md)
- [Completeness and status claims](completeness.md)

## Notes

Change-impact checks prove that a mapped documentation surface changed. Review still determines whether the change is accurate and sufficient.
