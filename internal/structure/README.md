# Structural Evidence

`internal/structure` defines Grimoire's provider-neutral structural context schema.

The schema represents immutable provider state, symbols, source spans, directed relationships, graph-depth results, ordered paths, unresolved references, provider provenance, and truncation. Lexicon and Arcana normalize their concrete results into these types before the compiler performs exact package budgeting.

This package contains data contracts only. It does not execute providers, rank tasks, inspect repositories, or fit token budgets.
