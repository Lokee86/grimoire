# Evaluation package

`internal/evaluation` owns Grimoire's judged corpus model, validation, scoring, aggregation, and report rendering.

## Responsibilities

- Load and validate source, structural, forbidden, and query-profile expectations.
- Record source stages from index presence through final package inclusion.
- Record structural stages from provider production through final inclusion.
- Distinguish retrieval, merge, curation, adaptive assembly, and budget-fitting losses.
- Calculate ranking recall, MRR, final recall, irrelevant-selection rates, latency, package size, and budget utilization.
- Aggregate by mode and category.
- Emit machine-readable JSON and reviewable Markdown.

The application package executes the production pipeline for each case and supplies stage snapshots. Evaluation must not contain an alternate retrieval implementation.

## Boundary

This package measures behavior; it does not tune rankings or policies automatically. Corpus expectations are authored evidence and must be corrected when they no longer describe the repository.
