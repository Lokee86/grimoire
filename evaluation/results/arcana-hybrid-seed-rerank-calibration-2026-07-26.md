# Arcana hybrid seed rerank calibration

Date: 2026-07-26  
Selected semantic candidate bound: `1536`  
Selected graph-neighborhood bound: `256`  
Final seed bound: `6`

## Change

The previous production path discarded Arcana vector scores and alternated a few semantic matches with Lexicon seeds. The replacement retains semantic score and rank, widens semantic recall before the final cutoff, and deterministically fuses:

- Arcana semantic similarity and rank;
- Lexicon rank and provider agreement;
- query-to-identifier and query-to-path evidence;
- declaration quality; and
- bounded incoming, outgoing, unresolved, and one-hop graph-neighborhood evidence.

The graph stage degrades to the deterministic non-graph fusion when Arcana graph lookup fails.

## Selected bound

The `1024` and `256` graph-neighborhood runs used the same prepared repository state and judged corpus within each repository. Median overhead is the paired vector-mode median minus the paired Lexicon-only median from the same run.

| Repository | Graph candidates | Pass | Required seed recall | MRR | Structural recall | Median vector overhead |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Grimoire | 1024 | 0.0% | 10.0% | 0.200 | 20.0% | 2639.6 ms |
| Grimoire | **256** | 0.0% | **10.0%** | **0.200** | **20.0%** | **1651.2 ms** |
| Space Rocks | 1024 | 22.2% | 22.2% | 0.074 | 22.2% | 4743.3 ms |
| Space Rocks | **256** | **22.2%** | **22.2%** | **0.074** | **22.2%** | **3229.8 ms** |

The 256-candidate bound preserves every measured quality metric while reducing median paired overhead by 988.4 ms on Grimoire and 1513.5 ms on Space Rocks. The wider bound is therefore rejected.

## Prior-v5 comparison

The earlier v5 seed alternation produced no required Grimoire seed or structural recall and produced 22.2% pass/seed/structural recall with 0.059 MRR on Space Rocks. The selected reranker reaches 10.0% required seed recall and 20.0% structural recall on the corrected current Grimoire corpus, while retaining Space Rocks' 22.2% recall and raising MRR to 0.074.

The Grimoire before/after values are directional rather than a strict paired comparison because the judged corpus was corrected to name the current mechanism owners (`RankedSemanticSeeds`, `RerankSeeds`, and private `search_index`) and the repository snapshot changed. The Space Rocks corpus and graph snapshot are unchanged, so its ranking comparison is directly comparable; absolute latency still varies with system load, which is why the paired overhead is reported.

## Remaining limitation

The reranker improves mechanism-owner recall but does not solve it. Grimoire still has no fully passing case: it retrieves `graph_documents` at rank 1 and its complete required structural neighborhood, but misses the paired supporting declaration and still prefers conceptually adjacent helpers for the other cases. The current implementation should be treated as a better bounded seed-selection seam, not as a finished semantic owner resolver.

## Reports

- `arcana-grimoire-hybrid-rerank-graph256-2026-07-26.{json,md}`
- `arcana-space-rocks-hybrid-rerank-graph256-2026-07-26.{json,md}`
- `arcana-grimoire-paired-v5-final-2026-07-26.{json,md}`
- `arcana-space-rocks-paired-v5-final-2026-07-26.{json,md}`
