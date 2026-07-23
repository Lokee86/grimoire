# Lexicon Facts Provider

`internal/lexiconfacts` owns Grimoire's optional Lexicon enrichment path.

## Owns

- resolving `.lexicon/CURRENT`;
- creating and reusing cached immutable `lexicon export` directories;
- loading exported normalized nodes and edges;
- ranking query-matched symbols;
- preserving matched symbols, source spans, identities, and immediate relationships as structural evidence; and
- mapping matched structural ranges back to prepared source candidates.

## Does not own

- language parsing or adapter execution;
- Lexicon's object store;
- Arcana traversal;
- source-candidate curation; or
- final package budgeting.

Repositories without Lexicon state continue through Grimoire's standalone source-retrieval path.

## Main files

- `state.go` — immutable snapshot discovery and cached export publication.
- `load.go` — exported JSONL loading.
- `rank.go` — symbol matching and one-hop candidate expansion.
- `candidates.go` — prepared-source mapping.
- `evidence.go` — first-class symbol and relationship evidence.
- `terms.go` — deterministic query and identifier normalization.
