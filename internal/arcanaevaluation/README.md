# Arcana evaluation package

`internal/arcanaevaluation` owns the judged paired corpus, seed-ranking and structural-evidence scoring, aggregation, deltas, and JSON/Markdown reports for Arcana semantic graph retrieval.

It compares two explicit evaluation modes against the same immutable prepared source, Lexicon export, and Arcana graph snapshot:

- `lexicon-seeds` bypasses the Arcana vector index;
- `lexicon-plus-vector` interleaves semantic graph seeds with the same Lexicon seeds before the same deterministic graph expansion.

Payload bytes are the serialized ranked seed list plus the serialized final Arcana structural evidence observed by the evaluator. Provider calls count evaluator-visible provider invocations: Lexicon seed search, optional Arcana semantic query, and Arcana graph expansion. They do not claim to count internal embedding HTTP requests or Arcana JSONL batch lines.

This package measures behavior only. It does not build vector state, alter production defaults, or participate in runtime discovery response assembly.
