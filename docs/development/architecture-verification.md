# Architecture verification

Parent index: [Development](INDEX.md)

## Purpose

This document defines the Pitlord policy and focused test gates that protect Grimoire's component ownership, repository layout, generated-state boundaries, protocol lifecycle, and pinned cross-repository dependencies.

## Overview

Pitlord is the executable architecture-policy owner. It inspects required paths, forbidden source dependencies, and prohibited repository content without relying on a repository-specific checker. Focused tests remain the executable proof for detailed runtime behavior and exact external source identities.

## Pitlord gate

Run:

```bash
pitlord validate --policy tools/pitlord/policy.json
pitlord check --repo . --policy tools/pitlord/policy.json
```

The policy protects:

- the accepted ADR set and architecture-owner documents;
- independently buildable Grimoire, Lexicon, and Arcana roots;
- forbidden Grimoire-to-Lexicon and Lexicon-to-Grimoire implementation imports;
- explicit ownership of generated-state ignores and traversal exclusions;
- explicit ownership and focused tests for MCP negotiation, bounded admission, and cancellation;
- explicit ownership and focused tests for bounded Lexicon consumer execution;
- explicit ownership and focused tests for pinned Lodestone module and source verification.

Released Pitlord `v0.1.2` owns structural, path, and forbidden-content policy. Detailed positive runtime values remain protected by the focused Go and workflow tests rather than by a second repository-specific checker.

## Workflow integration

`python scripts/workflow.py test` runs Pitlord before documentation and component tests. Pull requests and pushes run the same policy in `.github/workflows/documentation-standard.yml`.

The release workflow independently checks out the pinned Lodestone source identity, and `scripts/workflow.py` verifies that checkout before tests or release builds.

## Review requirements

Changes to component ownership, module identities, public protocols, persisted formats, process boundaries, or source-of-truth rules require:

1. an update to the canonical architecture or contract owner;
2. an ADR when the decision is consequential, surprising, cross-cutting, or difficult to reverse;
3. focused tests and an updated Pitlord rule when the invariant is mechanically enforceable;
4. migration and compatibility notes when existing state or consumers are affected.

A passing Pitlord check does not waive those review requirements.

## Code map

| Verification responsibility | Implementation | Related tests or gates |
| --- | --- | --- |
| Canonical policy composition | `tools/pitlord/policy.json` | `pitlord validate` in the root workflow and CI |
| Repository architecture rules | `tools/pitlord/repository.json` | `pitlord check` in the root workflow and CI |
| Combined workflow ordering | `scripts/workflow.py` | `scripts/test_workflow.py` |
| Detailed MCP lifecycle behavior | `internal/mcpserver/` | `internal/mcpserver/server_test.go` |
| Detailed Lexicon timeout behavior | `lexicon/internal/consumer/` | `lexicon/internal/consumer/runner_test.go` |
| Lodestone source identity | `.github/workflows/release.yml`, `scripts/workflow.py` | workflow verification and packaging tests |

## Related docs

- [Operations and trust boundaries](../architecture/operations-and-trust.md)
- [Component architecture](../architecture/components.md)
- [Behavioral contract matrix](behavioral-contract-matrix.md)
- [Testing and benchmarks](testing-and-benchmarks.md)
- [Architecture decisions](../decisions/INDEX.md)

## Notes

New Pitlord rules should protect a named durable invariant. Do not turn architecture policy into a style checker or encode speculative package structures.
