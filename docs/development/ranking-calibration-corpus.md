# Historical ranking calibration corpus

`evaluation/retrieval/` contains frozen corpora for the retired context-package retrieval pipeline. They calibrated source ranking, query-shape classification, candidate curation, evidence assembly, and final package fitting before Grimoire moved to progressive discovery.

The runner and owning source-evaluation implementation have been removed. The corpora may reference deleted source files and must not be treated as executable current tests.

## Historical value

These artifacts remain useful for interpreting checked-in reports about:

- exact and BM25 source ranking;
- cross-repository retrieval controls;
- query decomposition and intent ranking;
- candidate diversification;
- coverage-aware assembly;
- token-budget fitting;
- provider attribution.

They document why older ranking and package decisions were accepted or rejected. They do not define the current Grimoire interface.

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
