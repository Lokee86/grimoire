# Development

Development documentation defines how Grimoire is verified and how retrieval claims are measured.

- [Testing and benchmarks](testing-and-benchmarks.md) — test suites, evaluation commands, and report artifacts.
- [Retrieval quality](retrieval-quality.md) — corpus schema, pipeline-loss attribution, and metric interpretation.
- [Ranking calibration corpus](ranking-calibration-corpus.md) — judged case design and expansion rules.

The checked-in documentation evaluation corpus is [`evaluation/knowledge/grimoire.json`](../../evaluation/knowledge/grimoire.json); its execution path is documented in [Testing and benchmarks](testing-and-benchmarks.md).

Checked-in evaluation reports are evidence for the exact repository and state recorded by the report. They are not permanent product guarantees and must not be summarized without their mode, corpus, provider set, and date.
