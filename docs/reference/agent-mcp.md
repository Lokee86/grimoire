# Agent MCP runtime

`grimoire mcp` serves one progressive repository tool over MCP stdio:

```bash
grimoire mcp --root /path/to/repository
```

The exposed tool is `grimoire_query`. It accepts the `grimoire.query.v1` orient, search, trace, impact, and inspect fields plus:

- `state_mode`: `current-only`, `refresh-if-needed`, or `force-refresh`;
- `session`: a stable investigation name used to deduplicate evidence across calls;
- `include_knowledge`: whether documentation and design-rationale retrieval is included;
- `use_knowledge_vectors`: whether the optional documentation-vector ranker is used; the default is BM25-only;
- `code_only`: an explicit override for excluding documentation from source results.

The default state mode is `refresh-if-needed`. Grimoire checks the source fingerprint and repository-local `.lexicon`, `.arcana`, and `.grimoire` state before each call. Provider commands are resolved in this order: explicit request path, repository `.grimoire/providers.json`, executables beside Grimoire, a canonical Grimoire checkout, and finally `PATH`. Checkout discovery accepts `GRIMOIRE_HOME` and can derive the checkout from the target repository's Lexicon `adapter_root`, so child processes do not depend on inherited shell `PATH`. Structural-provider failures become warnings and source retrieval remains available; a required Grimoire source-index refresh still fails the request if it cannot complete.

Repository provider configuration is optional:

```json
{
  "version": 1,
  "lexicon_command": "C:/tools/grimoire/bin/lexicon.exe",
  "arcana_command": "C:/tools/grimoire/bin/arcana.exe"
}
```

Relative command paths are resolved from the analyzed repository root.

## Progressive response behavior

Without `session`, the response contains typed query evidence and bounded knowledge excerpts. Search and orientation results include at most 1,200 bytes of section text, four ranking reasons, and eight code-link hints. Returned query handles can be passed to later `trace`, `impact`, or `inspect` calls. Inspecting a `knowledge://` handle returns the complete exact section.

With `session`, Grimoire records evidence under `.grimoire/investigations/<session>/`. The first call returns new nodes, source ranges, graph paths, documents, questions, and accepted or rejected branches. Repeated evidence is replaced with prior ledger handles rather than replayed content. Query handles and suggestions remain available so the caller can continue expanding the investigation.

Sessions are bound to the prepared source identity and active provider snapshot identities. A stale handle or a session opened against different state is rejected rather than silently rediscovered.

## Recommended agent loop

1. Call `orient` once for an unfamiliar repository.
2. Call `search` with a concrete behavior, symbol, contract, endpoint, or message.
3. Reuse a returned handle with `trace` or `impact`.
4. Use `inspect` only for exact source or knowledge handles needed for the implementation.
5. Reuse one `session` throughout the task to avoid evidence replay.

Code and knowledge are deliberately separate lanes. When knowledge retrieval is enabled, `orient` and `search` automatically keep documentation out of code results and return relevant documentation through the knowledge results instead.
