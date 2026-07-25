# Extraction

`extraction` owns the boundary between ranked/curated retrieval candidates and the source spans passed to adaptive assembly and package compilation.

The core accepts an explicit ordered list of `Discoverer` implementations. The first discoverer that returns usable spans owns that candidate; otherwise the original prepared-index chunk is preserved. No discoverer is activated implicitly.

`LanguageDiscoverer` adapts `internal/spandiscovery` declaration and section boundaries into query-ranked extraction spans. The generic `LineWindowDiscoverer` remains available as a fallback building block, but is not used by the context pipeline.

Extraction preserves retrieval score, rank, source, reasons, facet membership, and provider metadata. Refined ranges receive deterministic identities and corrected token estimates. When one candidate produces several separated spans, the first span carries required `extracted_companion` links to its siblings so adaptive assembly and final fitting treat them as one evidence unit.

The application integration is deliberately conservative. `grimoire context --span-extraction` and `grimoire eval retrieval --span-extraction` enable language-aware extraction for adaptive packages only, and the application requires at least two useful spans before replacing the original candidate. The flag is disabled by default because the initial Grimoire paired evaluation was quality-neutral and added latency.
