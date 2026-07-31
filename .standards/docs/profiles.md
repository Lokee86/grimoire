# Repository Profiles

Parent index: [Documentation Standards](INDEX.md)

## Purpose

This document defines minimum documentation surfaces by repository type.

## Overview

Profiles prevent both under-documentation and cargo-cult folder creation. A repository selects one primary profile and may add explicit capabilities such as `service`, `ui`, `protocol`, `data-pipeline`, `research`, or `stateful`.

## Universal minimum

Every active repository has:

```text
README.md
AGENTS.md
docs root index
documentation policy
documentation procedure
planning owner
limits owner
implementation coverage map
docs-standard.json
```

Small archived experiments may use the `minimal` profile and explicitly omit coverage, planning, or limits.

## Minimal profile

For experiments, fixtures, and narrow utilities:

```text
README.md
AGENTS.md or explicit exemption
docs-standard.json
```

The README must still state purpose, status, usage, ownership, limitations, and whether the project is maintained.

## Library or engine profile

Adds:

```text
architecture
public API or format reference
state/storage format when applicable
development and testing
compatibility or migrations
```

## CLI product profile

Adds:

```text
getting-started guide
CLI reference
configuration reference
diagnostics and exit behavior
architecture
operations or troubleshooting
development and release workflow
```

## Service profile

Adds:

```text
service architecture
API or protocol reference
configuration and environment reference
deployment
observability and logging
health and readiness
backup/recovery or disposable-state declaration
security boundaries
failure and troubleshooting
```

## Application or desktop UI profile

Adds:

```text
user guide or product walkthrough
current screenshots for material workflows
UI architecture and state ownership
configuration
operations or local runtime behavior
accessibility or input behavior when relevant
release workflow
```

## Game profile

Adds repository-specific classes for:

```text
domain flows
services or runtimes
protocols
data and generated sources
systems design and authority
operations and deployment
devtools
player-facing guides where applicable
```

Space Rocks is the reference game profile.

## Umbrella product profile

For a product coordinating independently usable tools:

```text
system overview
component ownership and independent-use rules
shared runtime and process lifecycle
installation and component management
cross-component contracts
per-component references
operations and recovery
licensing and third-party obligations
```

Warlock uses this profile.

## Capability additions

### Stateful

Requires explicit persistence model, lifecycle, transactions or mutation boundaries, concurrency, failure recovery, migrations, and behavioral-contract mapping.

### Protocol

Requires participating systems, authority, message sequence, source-of-truth schemas, compatibility, validation, and implementation owners.

### Data pipeline

Requires source files, configuration, generated outputs, consumers, commands, validation, failure modes, and source maps.

### Research

Requires methodology, corpus, retained artifacts, limitations, and a clear boundary between measured evidence and product claims.

### UI

Requires current workflow documentation and screenshots when a screenshot materially communicates the current product surface.

## Related docs

- [Documentation standard](documentation-standard.md)
- [Adoption and enforcement](adoption.md)

## Notes

A profile defines minimum surfaces, not mandatory folder names. Repository-local policy chooses a taxonomy appropriate to the product while preserving the standard's semantic separation.
