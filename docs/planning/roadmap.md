# Grimoire Roadmap

This roadmap describes intended implementation order, not release commitments. The lexical baseline is current; every item below remains future work unless linked to an implemented reference.

## Current baseline

Implemented today:

- incremental file records with unchanged-file reuse;
- standard Git-ignore traversal and protected tool-state exclusions;
- content-addressed go-git object storage with atomic snapshot publication;
- deterministic fallback chunks;
- fixed lexical ranking with inspectable reasons;
- whole-chunk budget fitting; and
- versioned JSON context packages.

## 1. Lexicon structural-chunk consumer

Define the smallest stable Grimoire boundary for consuming Lexicon output.

Required behavior:

- accept language-neutral file and structural-range facts;
- translate Lexicon ranges into Grimoire chunk records;
- preserve stable source identity and diagnostics where useful;
- fall back to the current chunker when Lexicon is unavailable, has no adapter, or cannot parse a file; and
- keep Lexicon-specific execution outside ranking and budgeting.

Dependency: Lexicon's normalized adapter output and invocation contract. This item is partially blocked until that contract exists.

## 2. Prepared lexical postings

Replace per-query scanning of every prepared chunk with an index-maintenance-time lexical structure.

Goals:

- an established lexical scorer such as BM25 rather than a custom corpus-ranking algorithm;
- incremental update of changed chunk postings;
- exact phrase, filename, path, symbol, and heading boosts as separate inspectable signals;
- deterministic tie-breaking; and
- benchmark comparison against the current 10,000-chunk linear baseline.

This work is not blocked by Lexicon. The postings model must support both fallback and future structural chunks.

## 3. Model-aware token costs

Use existing tokenizer libraries to count package cost for declared target models.

Goals:

- cache costs during index maintenance where practical;
- identify the tokenizer and version in package metadata;
- retain the current heuristic for unknown models or unavailable tokenizers; and
- test that selected package cost remains within the declared model budget.

This work is not blocked by Lexicon.

## 4. Selection quality

Improve package construction only after better chunks, lexical retrieval, and token costs can be measured.

Candidate improvements:

- overlap removal;
- file and subsystem diversity;
- query-intent or evidence-class reservations;
- adjacent-chunk expansion;
- stable package fingerprints; and
- explicit omission reasons beyond budget pressure.

Every improvement must preserve inspectability and be evaluated against fixtures or repository tasks.

## 5. Incremental maintenance runtime

Add standalone change-driven maintenance so prepared state stays current without manual indexing.

The standalone mode should own only Grimoire behavior. When hosted by Warlock, Grimoire should consume shared repository change events, lifecycle supervision, and health reporting rather than duplicate the complete runtime stack.

This work should follow a stable incremental index contract and should not be required for one-shot CLI use.

## 6. Optional semantic retrieval

Evaluate a small local embedding provider only after the lexical baseline has quality and latency measurements.

Constraints:

- local and offline-capable;
- changed chunks embedded incrementally;
- no generative model or autonomous retrieval loop;
- no remote embedding API or required external vector database;
- strict provider deadline; and
- immediate lexical fallback when semantic work is unavailable or late.

Semantic evidence must supplement rather than replace inspectable lexical and metadata evidence.

## 7. Optional Warlock evidence providers

Add bounded provider interfaces for:

- Arcana relationship candidates;
- Demon Docs authority, identity, linkage, validation, and staleness evidence; and
- repository-change or other shared Warlock facts when a stable owner exists.

Grimoire remains responsible for ranking, budgeting, and final package construction. Providers remain responsible for their own domain facts.

## 8. Stable external contracts

Before a stable release, define:

- CLI compatibility and exit behavior;
- machine-readable diagnostics;
- prepared-index migration policy;
- context-package compatibility policy;
- provider deadline and partial-result metadata; and
- benchmark gates for latency, memory, and retrieval quality.

## Graduation rule

When a roadmap item becomes implemented:

1. Update the owning package README.
2. Add or update current architecture documentation.
3. Add or update exact reference documentation.
4. Remove or narrow the corresponding current limitation.
5. Replace roadmap detail with links to the implemented references and any unresolved follow-on work.
