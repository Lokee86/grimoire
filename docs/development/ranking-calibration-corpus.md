# Historical ranking calibration corpus

Parent index: [Development Documentation](INDEX.md)

## Purpose

This document records the status and permitted use of the frozen ranking corpora from Grimoire's retired context-package pipeline.

## Overview

The artifacts remain historical calibration evidence for earlier ranking and assembly work. They are not executable current tests and must not be presented as evidence for the active progressive-discovery interface.

## Research status

Retained historical evidence for a removed implementation. The corpus is frozen and no longer a release gate.

## Question

What ranking and package-assembly behavior was measured before Grimoire adopted progressive discovery?

## Method

The retired evaluator scored source retrieval, query-shape classification, candidate selection, evidence assembly, and token-budget fitting against frozen cases.

## Corpus or inputs

The retained inputs live under `evaluation/retrieval/` and may reference implementation files that no longer exist.

`evaluation/retrieval/` contains frozen corpora for the retired context-package retrieval pipeline. They calibrated source ranking, query-shape classification, candidate curation, evidence assembly, and final package fitting before Grimoire moved to progressive discovery.

The runner and owning source-evaluation implementation have been removed. The corpora may reference deleted source files and must not be treated as executable current tests.

## Results

These artifacts remain useful for interpreting checked-in reports about:

- exact and BM25 source ranking;
- cross-repository retrieval controls;
- query decomposition and intent ranking;
- candidate diversification;
- coverage-aware assembly;
- token-budget fitting;
- provider attribution.

They document why older ranking and package decisions were accepted or rejected. They do not define the current Grimoire interface.

## Interpretation

The historical evidence explains decisions in the removed context-package implementation. It does not validate the current progressive-discovery architecture, whose lanes and agent outcomes require different evaluation.

## Current replacement

The active system returns independent exact, source, document, symbol, and relationship lanes. Current evaluation should therefore measure:

- per-lane recall and rank;
- exact handle inspection;
- relationship and path coverage;
- evidence found across a progressive investigation;
- unsupported conclusions;
- tool calls, source opens, token use, and latency;
- irrelevant branches explored.

Use the repository-owned agent-discovery evaluator described in [Testing and benchmarks](testing-and-benchmarks.md). Documentation retrieval and Arcana graph retrieval retain their separate judged evaluators.

## Historical report discipline

When citing an older report, label it as context-pipeline history and preserve its repository revision, corpus, provider configuration, state freshness, and date. Do not compare its package recall directly with current discovery-lane or agent-outcome metrics.

## Limitations

The runner and owning implementation were removed, some referenced paths are stale, and package-level metrics are not comparable to current independent discovery lanes.

## Retained artifacts

Frozen corpora and checked-in historical reports remain under `evaluation/retrieval/` and the corresponding dated results directories.

## Related docs

- [Testing and benchmarks](testing-and-benchmarks.md)
- [Discovery quality](retrieval-quality.md)
- [Agent benchmark findings](agent-benchmark-findings.md)

## Notes

Historical calibration data may inform interpretation, but current product claims require current discovery-lane or agent-outcome evidence.
