# Extraction

`extraction` owns the boundary between ranked/curated retrieval candidates and the source spans passed to adaptive assembly and package compilation.

The core accepts an explicit ordered list of `Discoverer` implementations. The first discoverer that returns usable spans owns that candidate; otherwise the original prepared-index chunk is preserved. No discoverer is activated implicitly. Pipeline integration must choose the discovery policy and measure it against the judged corpus.

A conservative query-term `LineWindowDiscoverer` is included as a fallback building block. A language-aware discoverer can be placed ahead of it without changing retrieval, curation, assembly, or compiler ownership. Bounded multi-span output is supported through `Config.MaxSpans`.

Extraction preserves retrieval score, rank, source, reasons, facet membership, and provider metadata. Refined ranges receive deterministic identities and corrected token estimates. Refinement is rejected unless it produces a meaningful token saving, which keeps small or weakly matched chunks unchanged.
