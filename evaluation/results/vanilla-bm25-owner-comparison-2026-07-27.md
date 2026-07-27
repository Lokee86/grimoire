# Vanilla BM25 owner comparison

Date: 2026-07-27

## Purpose

Measure a deliberately ordinary lexical search baseline against the same human-labelled mechanism-owner corpora used for the Arcana hybrid evaluation. The labels are comparative judgments, not absolute ground truth.

The baseline uses conventional BM25 (`k1=1.2`, `b=0.75`) with camel-case and snake-case token splitting. It has no embeddings, graph traversal, query decomposition, declaration weighting, provider fusion, or owner heuristics.

Production-only runs exclude tests, generated/build directories, evaluation code, dependencies, and worktrees.

## Results

| Retrieval | Corpus | Unit | Required owner recall | Case pass | MRR | Median query |
| --- | --- | --- | ---: | ---: | ---: | ---: |
| Vanilla BM25 | Grimoire | 40-line chunks | 14.3% | 20.0% | 0.050 | 75.7 ms |
| Vanilla BM25 | Grimoire | whole files | 0.0% | 0.0% | 0.000 | 24.5 ms |
| Current Arcana hybrid | Grimoire | symbols | 10.0% | 0.0% | 0.200 | 6,597.7 ms |
| Vanilla BM25 | Space Rocks | 40-line chunks | 7.7% | 11.1% | 0.111 | 134.6 ms |
| Vanilla BM25 | Space Rocks | whole files | 46.2% | 44.4% | 0.389 | 42.4 ms |
| Current Arcana hybrid | Space Rocks | symbols | 22.2% | 22.2% | 0.074 | 13,174.4 ms |

The whole-file score counts a required owner only when the top-six ranked file both matches the judged path and contains the exact judged symbol name. It therefore measures owner-file discovery, not exact symbol ranking within that file.

## Interpretation

A single globally ranked representation is not consistently best:

- Grimoire's larger and denser files need chunk-level retrieval; whole-file BM25 dilutes the relevant terms.
- Space Rocks' smaller ownership-focused files strongly favour whole-file retrieval; BM25 finds six of thirteen required owners across four of nine cases.
- The hybrid graph/vector path improves some symbol-level cases, but it is not reliably better than ordinary lexical search and is orders of magnitude slower.
- Test and generated-output filtering materially improves lexical relevance and should be treated as corpus hygiene rather than semantic inference.

## Recommendation

Stop pursuing a globally deterministic mechanism-owner ranker through additional graph and scoring heuristics.

Use multi-granularity retrieval instead:

1. Run fast lexical ranking over production files and semantic chunks in parallel.
2. Merge a bounded set of top files and chunks without forcing them into one universal score scale.
3. Resolve declarations and call relationships only inside those retrieved files.
4. Use Arcana/Lexicon graph evidence for local expansion, explanation, and verification rather than global owner inference.

The next bounded experiment should test this file-first then symbol-within-file pipeline against the same corpora. It should be retained only if it beats the best vanilla baseline and the current hybrid without repository-specific rules.
