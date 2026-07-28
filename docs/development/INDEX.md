# Development

Development documentation defines how Grimoire is verified and how retrieval claims are measured.

- [Testing and benchmarks](testing-and-benchmarks.md) — test suites, evaluation commands, benchmark controls, and report artifacts.
- [Agent benchmark findings](agent-benchmark-findings.md) — current task-shape conclusions, Space Rocks comparisons, and the HikariCP/Detekt/Now in Android unfamiliar-repository suite.
- [Recent changes — July 2026](recent-changes-2026-07.md) — progressive-discovery consolidation, lexical-first retrieval, packaging, skill, and benchmark summary.
- [Release workflow](release-workflow.md) — root orchestration, local installation, skill packaging, version injection, and release artifacts.
- [Retrieval quality](retrieval-quality.md) — corpus schema, pipeline-loss attribution, and metric interpretation.
- [Ranking calibration corpus](ranking-calibration-corpus.md) — judged case design and expansion rules.

The checked-in documentation evaluation corpus is [`evaluation/knowledge/grimoire.json`](../../evaluation/knowledge/grimoire.json). Paired Arcana graph-seed corpora are checked in for [Grimoire](../../evaluation/arcana/grimoire.json) and the external [Space Rocks](../../evaluation/arcana/space-rocks.json) repository. Their execution and validation paths are documented in [Testing and benchmarks](testing-and-benchmarks.md).

Checked-in evaluation reports are evidence for the exact repository and state recorded by the report. They are not permanent product guarantees and must not be summarized without their mode, corpus, provider set, and date.
