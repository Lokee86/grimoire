# Knowledge-evaluation package

`internal/knowledgeevaluation` owns the judged documentation-retrieval corpus contract, deterministic scoring, aggregate metrics, and report rendering.

## Responsibilities

- Load and validate frozen documentation cases identified by repository path and stable heading information.
- Score production `internal/knowledge` responses without implementing another retrieval algorithm.
- Report pass rate, required-section recall, recall@k, MRR, irrelevant selections, vector usage/errors, latency, and per-case rankings.
- Preserve corpus case order and the deterministic ordering returned by `internal/knowledge`.

The application package owns CLI flag handling, knowledge-state loading, timeout measurement, and execution through `knowledge.Search`. Optional documentation vectors remain supplemental to BM25 through `knowledge.VectorRanker`.

## Boundary

This package does not evaluate source chunks, invoke source retrieval, build knowledge indexes, or tune ranking policy. The legacy source evaluator remains under `internal/evaluation`.
