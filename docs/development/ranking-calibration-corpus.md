# Ranking Calibration Corpus

Grimoire's ranking calibration corpus must include clean controls as well as the existing heterogeneous and pathological repositories. A ranker tuned only against tangled or under-documented code can learn compensating heuristics that damage ordinary repositories; a ranker tuned only against clean code can collapse when ownership and terminology are ambiguous.

## Clean control candidates

Repositories are admitted only after an exact commit is pinned and the first judged cases are reviewed. The repository names below are candidates, not floating dependencies.

| Repository | Language | Calibration role | Initial scope |
| --- | --- | --- | --- |
| `charmbracelet/gum` | Go | Small-to-medium clean control with visible command ownership and clear user-facing behavior | Whole repository |
| `httpie/cli` | Python | Mature application with separate implementation, tests, and documentation | Whole repository |
| `actualbudget/actual` | TypeScript | Real business logic, persistence, and calculation flows without requiring the entire monorepo | `packages/loot-core` |
| `sharkdp/fd` | Rust | Compact mature CLI with clear source, test, and documentation boundaries | Whole repository |
| `rubocop/rubocop` | Ruby | Documentation-heavy control with many similarly named implementations and tests | Whole repository, added after the initial wave |
| `gdquest-demos/godot-2d-space-game` | GDScript | Clean, lightly documented gameplay control for scenes, autoloads, signals, and ownership | Whole repository, added after the initial wave |

Secondary candidates are `jendrikseipp/vulture` for a smaller Python control, `puma/puma` for a harder Ruby concurrency boundary, and `BurntSushi/ripgrep` for a larger multi-crate Rust tier.

## Initial wave

Begin with Gum, HTTPie CLI, Actual's `loot-core`, and fd. Each repository contributes five judged cases:

1. direct implementation location;
2. mechanism explanation across multiple symbols;
3. architecture or responsibility ownership;
4. call-chain or data-flow investigation; and
5. one realistic mixed task involving implementation plus tests, configuration, or persistence.

This produces twenty clean-control queries before adding the docs-heavy Ruby and lightly documented GDScript controls.

## Judgement policy

Judgements are written before ranking changes and stored in the existing `evaluation/retrieval` corpus format.

- **Required evidence** is necessary to answer the task correctly.
- **Supporting evidence** materially improves the answer but is not required for a pass.
- **Forbidden evidence** is used only for known, plausible distractors; it is not a catalogue of every irrelevant file.
- Paths and symbols must be confirmed by inspecting the pinned repository revision.
- A case is rejected when ownership is too ambiguous to establish defensible ground truth.

Each repository revision, scope, and judgement authoring date must be recorded beside its corpus. Do not update a pinned revision and silently preserve old judgements.

## Calibration sequence

1. Pin and import one repository revision.
2. Write five judged cases without changing retrieval behavior.
3. Record the standalone baseline in all supported query modes.
4. Inspect pre-curation rank metrics and final-package survival separately.
5. Reduce confirmed ranking failures into deterministic fixtures.
6. Change one ranking feature or weight family at a time.
7. Rerun every frozen corpus and reject changes that improve one tier by materially regressing another.

The first algorithm-tuning boundary is measurement, not a new weight guess. Grimoire reports required-evidence recall at rank 10 and 20, first-required reciprocal rank, and judged-path relevance at rank 10 and 20 before merge, curation, and package fitting. These metrics expose whether the retrieval order itself improved; final-package recall continues to measure downstream usefulness.

## Initial measurements

Lexical-only runs on July 23, 2026 established the first pre-curation baselines:

| Corpus | Cases | Required R@10 | Required R@20 | MRR | Relevant @10 | Relevant @20 | Final required recall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Grimoire | 12 | 8.3% | 8.3% | 0.017 | 1.7% | 2.1% | 0.0% |
| Gum `716d8b5d0221558f944b5a078dbbcca8572534fb` | 5 | 26.7% | 31.7% | 0.476 | 34.0% | 24.0% | 28.6% |

The Gum corpus is the first pinned clean control. It contains five manually judged cases covering command dispatch, filter behavior, timeout and exit ownership, the spin process call chain, and file-picker behavior. All five cases failed complete-package recall despite materially better initial ranking than the Grimoire self-corpus; the final packages recovered only 28.6% of required evidence and selected 74.1% unjudged paths. This separates two problems that must be tuned independently: weak initial ordering and loss during package fitting.

These measurements describe the current lexical path, not a tuning target by themselves. The remaining clean-control cases must be frozen before changing ranking weights.

## Score-attribution pass

The pinned Gum checkout now lives at `C:\!bin\workspace\grimoire-corpora\gum`; its prepared index is repository-local under `.grimoire`. Grimoire evaluation reports now retain numeric lexical, exact, and semantic scoring signals for each candidate and show its retrieved, exact, merged, curated, and included positions. Adjacency is reported as a curation insertion rather than misrepresented as an additive score.

The first C-drive attribution run found that the current lexical scorer is dominated by weak terms:

| Gum case | Top-20 score from common stopwords |
| --- | ---: |
| `gum-dl-01` | 74.6% |
| `gum-me-01` | 49.6% |
| `gum-ao-01` | 49.0% |
| `gum-cc-01` | 35.9% |
| `gum-lm-01` | 34.0% |

Substring matching compounds the problem: the query term `and` receives a filename boost against `command.go`. Common words such as `the`, `to`, `and`, and `are` frequently contribute the maximum capped content score, outranking ownership-bearing identifiers and paths. This is now a measured ranking defect rather than a subjective package review.

No ranking behavior changed during the lexical attribution pass. Stopword suppression and boundary-aware matching remain required lexical corrections, but the provider comparison changed the tuning order.

## Provider-attribution result

The frozen Gum and Grimoire corpora were rerun in lexical-only, vector-only, and neutral rank-interleaved hybrid modes. Vector retrieval was measurably different from lexical retrieval and usually stronger before final package fitting.

| Corpus | Mode | Required R@10 | Required R@20 | MRR | Final required recall |
| --- | --- | ---: | ---: | ---: | ---: |
| Gum | lexical | 26.7% | 31.7% | 0.476 | 28.6% |
| Gum | vector | 45.0% | 60.0% | 0.477 | 50.0% |
| Gum | hybrid | 50.0% | 55.0% | 0.557 | 42.9% |
| Grimoire | lexical | 10.4% | 25.0% | 0.058 | 2.2% |
| Grimoire | vector | 16.5% | 29.0% | 0.080 | 2.2% |
| Grimoire | hybrid | 12.5% | 20.6% | 0.059 | 2.2% |

On Grimoire, lexical, vector, and hybrid all finished at the same 2.2% required-evidence recall even though vector ranking improved. Of 45 required evidence items, lexical retrieved 34 and curated 33, vector retrieved 35 and curated 34, and hybrid retrieved and curated 33. Every mode included only one required item in the final package. Hybrid additionally carried 46 required vector candidate chunks through curation but included only four of those chunks.

The earlier apparent lack of vector benefit was therefore a final-package measurement artifact, not evidence that vector retrieval returned the same candidates or added no value. Package fitting and final selection are now the next controlled tuning target. Lexical normalization follows as a separate provider correction. The complete analysis is recorded in `evaluation/results/provider-attribution-analysis-2026-07-23.md`.
