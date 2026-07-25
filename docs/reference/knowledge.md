# Knowledge retrieval

`grimoire knowledge` indexes repository rationale separately from production-source chunks. It discovers Markdown/MDX and text documentation, architecture and planning notes, ADR-like files, issue/export notes, and supported prose-heavy configuration. `.git`, worktrees, Lexicon/Arcana/Grimoire state, generated content, `.gitignore` matches, and explicit exclusions are not indexed.

```bash
grimoire knowledge index --root .
grimoire knowledge search --root . --query "why is snapshot freshness validated" --top-k 5
grimoire knowledge inspect --root . --handle 'knowledge://docs/architecture/index.md#...'
```

The persistent state defaults to `.grimoire/knowledge/index.json`. Documents and sections have stable path/heading-derived identities, exact byte and line spans, hashes, Git HEAD metadata when available, and deterministic BM25 terms. Search returns the cited text, a stable handle, ranking reasons, filters (`--path`, `--kind`, `--heading`, `--commit`, `--since`, and `--until`), and exact code-link hints only when a documented symbol, repository path, endpoint, message, or configuration contract name is also present in repository files.

Knowledge search is lexical-first and works without the embedding model, endpoint, or vector state. Package callers may provide an optional vector ranker; it is passed knowledge sections only and is supplemental to BM25 rather than a replacement.
