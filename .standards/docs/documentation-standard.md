# Documentation Standard

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document defines the mandatory documentation standard for Laughing Skull repositories.

## Overview

Documentation is part of the implementation. A change is incomplete when the repository no longer explains its current public behavior, ownership boundaries, state, failure modes, operations, or safe extension points.

The standard separates facts by what they are and who owns them. Repository profiles choose the categories that fit the product; no repository may collapse plans, research, current behavior, and operational truth into one undifferentiated document set.

## Core rules

1. Every durable fact has one canonical current owner.
2. Current behavior, planning, research, limitations, and agent guidance remain visibly separate.
3. Root `README.md` is a product entry point, not the complete manual.
4. Architecture documentation explains ownership and behavior in prose; a file list or code map is supporting evidence, not a substitute.
5. Every non-minimal repository has a maintainer map that routes common change areas to canonical documentation and primary implementation boundaries.
6. Public commands, APIs, schemas, configuration, defaults, diagnostics, and compatibility behavior have exact reference owners.
7. Stateful flows, mutation boundaries, persistence models, concurrency boundaries, recovery seams, and critical invariants have explicit current owners.
8. Implemented facts must not live only in planning or research.
9. Every documentation folder is indexed according to the repository's configured index convention.
10. Internal links must resolve.
11. Documentation, maintainer maps, and affected coverage maps change in the same implementation change when ownership or navigation changes.
12. Exceptions are explicit, narrow, and justified in `docs-standard.json`.
13. Documentation status claims must follow [Completeness and status claims](completeness.md).

## Canonical document classes

Repositories select the classes appropriate to their profile.

### Guide

Task-oriented instructions for accomplishing a user or operator goal. Guides state prerequisites, ordered steps, expected result, and likely recovery paths. They link to reference material instead of duplicating exhaustive contracts.

### Reference

Exact public contracts: commands, flags, APIs, packets, schemas, configuration, defaults, precedence, diagnostics, exit behavior, compatibility, and machine-readable formats.

### Maintainer map

A repository-navigation artifact that routes common change intents to their canonical documentation, primary implementation boundary, and relevant verification surface. A maintainer map is not a repository inventory, package catalogue, generated dependency graph, or substitute for focused architecture and code maps.

The conventional repository-level location is `docs/architecture/maintainer-map.md` or `docs/maintainer-map.md`. Independently maintained components in a monorepo or umbrella product may use component-local `MAINTAINER_MAP.md` files linked from the repository-level map.

### Architecture

Implemented ownership, responsibilities, non-ownership boundaries, lifecycle or flow, state ownership, invariants, seams, failure behavior, code maps, and tests.

### Operations

Runtime behavior, deployment, observability, logs, backups, recovery, shutdown, migrations, unattended execution, and troubleshooting.

### Domain or systems design

Cross-system product flows, conceptual mechanics, authority rules, and durable design invariants. Domain documents do not replace implementation owners.

### Data or protocol

Source-of-truth data, generated outputs, persistence contracts, pipelines, communication surfaces, transport behavior, compatibility, and validation.

### Development

Repository layout, build and test workflows, fixtures, release gates, generated artifacts, extension procedures, and contributor constraints.

### Research

Question, method, corpus or inputs, measured results, limitations, interpretation, and retained artifacts. Research evidence is not automatically a product guarantee.

### Planning

Future, unresolved, proposed, sequenced, partially implemented, back-burnered, or superseded work. Planning states current status and links to implemented owners when work ships.

### Limits

Current defects, blockers, incomplete transitional behavior, and accepted practical ceilings. Intentional invariants belong in architecture or systems design; future work belongs in planning.

### Agent

Stable editing, testing, tooling, and repository-orientation rules. Long-lived product facts remain in canonical current documents.

### Notes and legacy

Notes are non-authoritative temporary material. Legacy documents are temporary migration sources and are deleted when useful facts have moved.

## Universal document shape

Normal canonical documents contain:

```text
Title
Parent index
Purpose
Overview
Type-specific sections
Related docs
Notes
```

Indexes, changelogs, licenses, generated references, and explicitly configured exceptions may use a different shape.

## Type-specific minimums

### Guide

```text
Prerequisites
Procedure or workflow
Expected result
Failure and recovery
```

### Reference

```text
Exact contract
Defaults or precedence when applicable
Diagnostics or failure behavior
Examples
```

### Maintainer map

```text
Scope and intended use
Change-area routing table
Canonical documentation owner for each route
Primary implementation boundary for each route
Relevant verification or test owner when useful
Cross-component boundaries and links to component maps
```

### Architecture

```text
Code root
Responsibilities
Does not own
Flow or lifecycle
State or data ownership
Invariants and safety boundaries
Code map
Tests
```

### Operations

```text
Operating model
Commands or controls
Runtime state and logs
Failure and recovery
Verification
```

### Research

```text
Research status
Question
Method
Corpus or inputs
Results
Limitations
Interpretation
Retained artifacts
```

### Planning

```text
Current status
Expected ownership or ownership boundary
Planned behavior
Implementation sequence
Acceptance criteria
Open decisions
Implemented references when applicable
```

### Development

```text
Workflow or repository boundary
Commands
Failure modes
Code map when implementation-facing
```

## Coverage requirements

Every production package, executable, public command family, independent stateful flow, mutation boundary, persistent model, concurrency boundary, machine-readable contract, and recovery seam maps to canonical current documentation.

Small utility packages may share the owner of the subsystem they concretely serve. A broad package with several independent flows requires focused owners when one umbrella page cannot explain their transitions and failure boundaries clearly.

A coverage-table entry is not sufficient by itself. The linked owner must actually explain responsibility, flow, state, non-ownership, invariants, failure or recovery, extension seams, and relevant tests or public contracts.

Critical behavioral invariants should map to focused tests and release gates through a behavioral-contract matrix for complex or stateful products.

Every non-minimal repository has one repository-level maintainer map. It must route by maintainer intent or change area rather than merely repeat the directory tree. Multi-component repositories add component maps when one top-level table cannot route maintainers to a sufficiently precise implementation boundary.

Maintainer maps point to canonical owners; they do not become duplicate owners. Focused code maps remain inside the architecture, reference, operations, or development documents that own the behavior.

## Index and link requirements

Each repository declares a root index and folder index convention in `docs-standard.json`. Every normal documentation folder contains its configured index. `stubs/` may be exempt.

Indexes list every direct Markdown file and direct documentation subfolder. Indexes remain navigational and do not duplicate full document bodies.

All relative Markdown links must resolve. Moves and deletions update inbound links and affected indexes in the same change.

## Related docs

- [Documentation procedure](documentation-procedure.md)
- [Repository profiles](profiles.md)
- [Change-impact rules](change-impact.md)
- [Completeness and status claims](completeness.md)

## Notes

Repositories may add specialized document classes, such as Space Rocks service, protocol, data, systems-design, and devtools documentation, when those classes have distinct ownership rules.
