# Grimoire Pitlord policy

This directory owns Grimoire's executable architecture policy.

## Policies

- `repository.json` enforces required decisions, component boundaries, forbidden implementation imports, generated-state ownership, lifecycle-test ownership, and cross-repository source-verification ownership.
- `policy.json` is the canonical composed policy entry point.

Policy belongs here rather than in a repository-specific checker. Pitlord owns rule validation, repository evidence, and diagnostics. Focused Go, Rust, and workflow tests remain the executable proof for detailed lifecycle behavior and exact external source identities.

## Check

The normal root workflow and CI run:

```text
pitlord validate --policy tools/pitlord/policy.json
pitlord check --repo . --policy tools/pitlord/policy.json
```

Set `PITLORD` when the executable is not available through `PATH` or the sibling Pitlord checkout. Grimoire pins CI to released Pitlord `v0.1.2`.

## Ownership

The policy declares invariants. The owning implementation and architecture documents remain the source of truth for why those invariants exist. Add a Pitlord rule only for a named durable boundary with deterministic evidence.
