# Completeness and Status Claims

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document defines the evidence required before reporting documentation as complete, current, compliant, or up to standard.

## Overview

Documentation status is an auditable claim. File count, a long README, or the existence of a docs folder is not evidence of completeness.

## Allowed claims

### Documented

Use only when the changed responsibility has a canonical current owner that accurately explains its public contract, ownership, flow, state, failure behavior, recovery, and tests as applicable.

### Structurally compliant

Use only when the shared checker and repository structural tooling pass. This means required files, indexes, links, sections, and configured coverage mappings satisfy deterministic checks.

### Current

Use only when current code and public behavior were inspected against the owning documentation. Structural compliance alone does not establish semantic currency.

### Complete

Use only when all are true:

```text
structural checks pass
implementation coverage has no known unmapped production boundary
current public and operational contracts were inspected
planning and research are not acting as current owners
known limitations are disclosed
no material documentation debt is known
```

If any known gap remains, state the scope that is complete and list the gap.

## Required implementation report

```text
Documentation impact:
- Inspected: <canonical owners and affected surfaces>
- Updated: <files or none with reason>
- Not affected: <surfaces checked and why>
- Compliance check: <commands and actual result>
- Known documentation gaps: <none or explicit list>
```

## Prohibited claims

Do not say:

```text
documentation is complete
fully documented
up to standard
all docs are current
no documentation work remains
```

unless the evidence above exists.

Do not infer semantic completeness from a passing link checker, generated index, coverage table, or successful CI run.

## Related docs

- [Documentation standard](documentation-standard.md)
- [Change-impact rules](change-impact.md)
- [Adoption and enforcement](adoption.md)

## Notes

When uncertainty remains, report the verified scope and uncertainty directly rather than choosing a stronger status label.
