# Repository state preparation

`repostate.Ensure` is the high-level preparation boundary for future query callers. It reports Git/source identity and the current `.lexicon`, `.arcana`, `.grimoire`, and vector state without requiring an LLM to infer freshness.

Modes are explicit:

- `current-only` inspects without mutation;
- `refresh-if-needed` serializes and performs only missing or stale refreshes; and
- `force-refresh` reruns Lexicon, Arcana synchronization, and Grimoire preparation.

Refreshes invoke the existing `lexicon`, `arcana`, and `grimoire index` commands. No analyzer or vector builder is duplicated. A process mutex plus `.grimoire/repostate.lock` prevents concurrent refresh loops, and state is re-inspected after the lock is acquired. Vectors are only inspected and never built.
