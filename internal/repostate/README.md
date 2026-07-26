# Repository state preparation

`repostate.Ensure` is the high-level preparation boundary for query callers. Status schema version 2 reports Git/source identity, Lexicon, Arcana, prepared source, documentation knowledge, Arcana-vector, and documentation-vector state without requiring an LLM to infer freshness.

Modes are explicit:

- `current-only` inspects without mutation;
- `refresh-if-needed` serializes and performs only missing or stale refreshes; and
- `force-refresh` reruns Lexicon, Arcana synchronization, Grimoire source preparation, and documentation knowledge indexing.

Refreshes invoke the existing `lexicon`, `arcana`, `grimoire index`, and `grimoire knowledge index` commands. No analyzer, indexer, or vector builder is duplicated. A process mutex plus `.grimoire/repostate.lock` prevents concurrent refresh loops, and state is re-inspected after the lock is acquired. Both vector lanes are only inspected and never built. Git repositories use index blob identities plus changed working-tree content for fast fingerprints; non-Git repositories retain the full walk-and-hash fallback.
