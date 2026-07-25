# Lexicon Facts Provider

`internal/lexiconfacts` owns Grimoire's optional Lexicon enrichment path.

## Owns

- resolving `.lexicon/CURRENT`;
- creating and reusing cached immutable `lexicon export` directories;
- loading exported normalized nodes, occurrence-level edges, and typed edge attributes;
- ranking query-matched symbols;
- aggregating repeated relationships without discarding bounded source sites;
- preserving matched symbols, exact source spans, call-resolution evidence, macro expansion chains and substitutions, direct argument flow, identities, and immediate relationships as structural evidence;
- traversing one additional bounded hop only through normalized interstack contract nodes; and
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
- `rank.go` — symbol matching, one-hop candidate expansion, and bounded two-hop interstack bridging.
- `candidates.go` — prepared-source mapping.
- `evidence.go` — first-class symbol and aggregated relationship evidence.
- `relationship_provenance.go` — bounded occurrence sites and typed call, macro, and argument-flow provenance.
- `terms.go` — deterministic query and identifier normalization.
