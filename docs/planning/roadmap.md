# Roadmap

This page contains work that is not yet complete. Implemented behavior is documented elsewhere.

## Completed foundation

- Lexicon and Arcana source trees and histories consolidated into the Grimoire repository while retaining independent applications and state boundaries.
- Content-addressed prepared source index with exact token counts.
- Managed local Qwen3 embedding setup and service commands.
- Content-addressed vector objects and packed exact-search snapshots.
- Incremental vector reuse and concurrent embedding requests with serialized deterministic ingestion.
- Semantic, lexical, and concrete exact retrieval.
- Deterministic ranking, judged curation calibration, and prepared-neighbour expansion.
- Lexicon facts and Arcana graph evidence as optional query-time enrichment.
- Source and structural judged evaluation with pipeline-loss attribution.
- Deterministic query-shape profiling and automatic budgets.
- Evidence-coverage assembly and fixed-versus-adaptive evaluation.
- Version 5 exact-budget context packages.

## Near-term priorities

1. Add root-level build, test, and release orchestration that verifies all three components without collapsing their build systems.
2. Define coordinated versioning and component-specific release artifacts from the monorepo.
3. Re-run ranking and adaptive-package calibration after all merged retrieval changes and tune targets against representative recall.
4. Expand frozen judged corpora across additional repositories, languages, sizes, and task categories.
5. Improve task-oriented evidence roles and stopping conditions without hiding decisions in opaque scoring.
6. Add stronger diagnostics for runtime selection, provider failures, state compatibility, and native-engine errors.
7. Add explicit prepared/vector/structural state status and maintenance commands suitable for Warlock supervision.

## Monorepo and distribution work

- Decide whether the former Arcana and Lexicon repositories should become automated subtree mirrors for compatibility.
- Preserve independently installable `arcana`, `lexicon`, and `grimoire` artifacts.
- Define canonical module/package import paths before a stable release.
- Add one installer that can install any subset of the components.
- Add repository-wide contribution and release documentation.

## Retrieval and package quality

- Add clean controls beyond the current self and Gum corpora.
- Preserve provider-attribution, ranking, curation, assembly, and fitting metrics as separate gates.
- Add caller-selectable automatic minimum/maximum policy bounds.
- Add stronger evidence-class allocation only when judged failures justify it.
- Add package fingerprints and more explicit omission reasons.
- Measure downstream agent discovery calls, latency, and usage in addition to evidence recall.

## Prepared-state maintenance

- Use Git-aware changed-file detection as a fast path while preserving non-Git fallback.
- Add optional repository watching or Warlock-fed change events without making one-shot commands dependent on a daemon.
- Add lazy or bounded prepared-state reads for very large repositories.
- Make file eligibility and generated-content policy configurable without weakening permanent state exclusions.
- Evaluate optional Lexicon-aligned source chunk preparation while retaining language-agnostic fallback.

## Vector-engine work

- Add safe reachability-based immutable-object cleanup.
- Add non-Windows Go dynamic-library loaders and release packaging.
- Benchmark float32 against float16 and int8 encodings.
- Optimize exact-scan kernels only when measurements show material benefit.
- Consider approximate indexing only when exact search is no longer acceptable and exact fallback remains available.
- Evaluate a more efficient ingestion boundary after measuring serialized JSONL persistence cost.

## Structural integration work

- Improve Lexicon seed matching through judged task-shaped cases.
- Expand Arcana operations only when specific graph-evidence failures justify them.
- Add conflict and provenance diagnostics across source, Lexicon, and Arcana evidence.
- Evaluate Demon Docs, Git-change, and other Warlock evidence providers behind concrete interfaces.
- Define a stable external provider contract only after the current integrations settle.

## Operational and compatibility work

- Add stable machine-readable diagnostics and documented exit classes.
- Define prepared-index, vector-index, Lexicon, Arcana, embedding-runtime, and context-package migration policy.
- Add managed runtime artifacts for additional platforms.
- Add Warlock lifecycle integration for model service, component discovery, and state maintenance while keeping every component independently usable.
- Establish release gates for latency, memory, retrieval quality, determinism, adapter correctness, graph correctness, and ABI stress.

## Longer-term investigation

- Learned or model-assisted policy components only where deterministic rules are insufficient and decisions remain inspectable.
- Repository-scale prioritization and packetized context delivery for very large codebases.
- Global package optimization only when deterministic whole-item fitting shows measured, reproducible failures.

Each roadmap item requires an owning seam, verification plan, and documentation update before it becomes current behavior.
