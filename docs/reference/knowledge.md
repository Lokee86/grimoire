# Knowledge retrieval

`grimoire knowledge` indexes repository rationale separately from production-source chunks. It discovers Markdown/MDX and text documentation, architecture and planning notes, ADR-like files, issue/export notes, and supported prose-heavy configuration. `.git`, worktrees, Lexicon/Arcana/Grimoire state, generated content, `.gitignore` matches, and explicit exclusions are not indexed.

```bash
grimoire knowledge index --root .
grimoire knowledge search --root . --query "why is snapshot freshness validated" --top-k 5
grimoire knowledge inspect --root . --handle 'knowledge://docs/architecture/index.md#...'
grimoire knowledge vector build --root .
grimoire knowledge vector info --root .
```

The root aliases `grimoire vector build` and `grimoire vector info` invoke the same documentation-vector commands. There is no production source-vector build or search command.

## Persistent state

The knowledge index defaults to `.grimoire/knowledge/index.json`. Documents and sections have stable path/heading-derived identities, exact byte and line spans, hashes, Git metadata when available, and deterministic BM25 terms.

Optional vectors live under `.grimoire/knowledge/vectors/<embedding-identity>/`. Their manifest is bound only to documentation paths, kinds, document hashes, section IDs, and section hashes. Repository moves, new commits, and unrelated source-code changes do not invalidate a current documentation snapshot. Documentation content or section-structure changes do.

## Search

Knowledge search always runs deterministic BM25 first. It supports `--path`, `--kind`, `--heading`, `--commit`, `--since`, and `--until` filters and returns cited section text, stable handles, ranking reasons, and exact code-link hints when documented symbols, repository paths, endpoints, messages, or configuration contract names also occur in repository files.

When a current documentation-vector snapshot exists, search embeds the query and adds a bounded supplemental vector score through the `knowledge.VectorRanker` seam. BM25 remains the primary ranking path. Use `--vectors=false` to force lexical-only search.

Missing, stale, incompatible, timed-out, or unavailable vectors do not fail knowledge retrieval. The response keeps BM25 results, sets `vector_used` to false, and exposes the failure through `vector_error`. Freshness is validated before query embedding.

## Build behavior

The vector builder deduplicates identical section text, reuses immutable content-addressed vector objects, writes successful batches immediately, and publishes a packed exact-search snapshot only after all required vectors are available. A failed build leaves completed objects reusable by the next run. An unchanged current snapshot returns immediately without object probes or rematerialization.

Documentation vectors are consumed by `knowledge search` and the MCP knowledge lane only. They never enter source-context ranking and never affect `grimoire context` readiness or warnings.
