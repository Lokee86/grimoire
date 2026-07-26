# Roadmap

This page contains work that is not yet complete. Implemented behavior is documented elsewhere.

## Completed foundation

- Lexicon and Arcana source trees and histories consolidated into the Grimoire repository while retaining independent applications and state boundaries.
- Content-addressed prepared source index with exact token counts.
- Managed local Qwen3 embedding setup and service commands.
- Content-addressed documentation-vector objects and packed exact-search snapshots, plus Arcana-owned semantic graph indexes.
- Incremental documentation-vector reuse, resumable concurrent embedding requests, serialized deterministic ingestion, and current-snapshot fast reuse.
- Deterministic source BM25, concrete exact recovery, Lexicon facts, and Arcana graph retrieval without repository-wide source embeddings.
- Deterministic ranking, judged curation calibration, and prepared-neighbour expansion.
- Lexicon facts and Arcana graph evidence as optional query-time enrichment.
- Source and structural judged evaluation with pipeline-loss attribution.
- Deterministic query-shape profiling and automatic budgets.
- Evidence-coverage assembly and fixed-versus-adaptive evaluation.
- Version 8 exact-budget context packages with adaptive assembly, facet protection, and required-link-group fitting.
- Lexicon-aligned semantic source preparation with complete fallback coverage and per-file preparation identities.
- Judged documentation retrieval with an eight-case frozen self-corpus, BM25/vector attribution, recall@k, MRR, irrelevant-selection, latency, and per-case reports.
- Unified status schema for source, knowledge, Arcana vectors, and documentation vectors, with incremental refresh of all deterministic state.
- Repository-wide source-vector implementation and evaluator removed; source retrieval now has one production path.
- Root build/test orchestration, coordinated version injection, subset installation, independent component archives, combined bundles, and checksums.
- Git-aware repository fingerprints that reuse index blob identities and hash only changed working-tree content.

## Near-term priorities

1. Calibrate the documentation BM25/vector blend on external repositories and rationale-discovery tasks beyond the current self-corpus.
2. Re-run source ranking and adaptive-package calibration after all merged retrieval changes and tune targets against representative recall.
3. Expand frozen judged corpora across additional repositories, languages, sizes, and task categories.
4. Improve task-oriented evidence roles and stopping conditions without hiding decisions in opaque scoring.
5. Add stable diagnostic codes for runtime selection, provider failures, state compatibility, and native-engine errors.

## Monorepo and distribution work

- Decide whether the former Arcana and Lexicon repositories should become automated subtree mirrors for compatibility.
- Define canonical module/package import paths before a stable release.
- Add repository-wide contribution guidance beyond the implemented build and release documentation.

## Retrieval and package quality

- Add clean controls beyond the current self and Gum corpora.
- Preserve provider-attribution, ranking, curation, assembly, and fitting metrics as separate gates.
- Add caller-selectable automatic minimum/maximum policy bounds.
- Add stronger evidence-class allocation only when judged failures justify it.
- Add package fingerprints and more explicit omission reasons.
- Measure downstream agent discovery calls, latency, and usage in addition to evidence recall.

## Prepared-state maintenance

- Add optional repository watching or Warlock-fed change events without making one-shot commands dependent on a daemon.
- Add lazy or bounded prepared-state reads for very large repositories.
- Make file eligibility and generated-content policy configurable without weakening permanent state exclusions.
- Calibrate semantic declaration chunking against judged retrieval and downstream-agent token use.

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
- Define prepared-index, documentation-vector, Arcana-vector, Lexicon, Arcana, embedding-runtime, and context-package migration policy.
- Add managed runtime artifacts for additional platforms.
- Add Warlock lifecycle integration for model service, component discovery, and state maintenance while keeping every component independently usable.
- Establish release gates for latency, memory, retrieval quality, determinism, adapter correctness, graph correctness, and ABI stress.

## Longer-term investigation

- Learned or model-assisted policy components only where deterministic rules are insufficient and decisions remain inspectable.
- Repository-scale prioritization and packetized context delivery for very large codebases.
- Global package optimization only when deterministic whole-item fitting shows measured, reproducible failures.

Each roadmap item requires an owning seam, verification plan, and documentation update before it becomes current behavior.
