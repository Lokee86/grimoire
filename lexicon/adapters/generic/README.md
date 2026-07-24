# Generic adapter

The generic adapter provides conservative fallback coverage for source languages that do not yet have a dedicated Lexicon adapter.

Lexicon selects it only for a curated set of source-code extensions. Dedicated adapters always take precedence. Each extension receives a deterministic language identity such as `generic-java`, `generic-lua`, or `generic-zig`, while all variants share this implementation. C and C++ are owned by the dedicated `c-family` adapter.

## Emitted facts

For every accepted UTF-8 source file, the adapter emits:

- a file node and content identity;
- a module node owned by that file;
- high-confidence type and function declarations using shared and language-family recognizers;
- static import/include/use evidence;
- comment and string masking before declaration recognition, preventing source-like text from becoming facts;
- unresolved `imports` records for import targets that the generic adapter cannot safely resolve.

It deliberately does not emit resolved calls, dispatch targets, inheritance, dataflow, or other language-specific semantics. Those belong in a dedicated adapter.

## Exclusions

The registry uses a positive source-extension allowlist, so documentation, configuration, data, and binary files are not claimed. The adapter also excludes dependency, build, cache, worktree, Git, and Warlock state directories. Non-UTF-8, NUL-containing, and generated files are skipped.

## Usage

```text
lexicon-generic --repo <repository> --language generic-java --output facts.jsonl
```

Incremental invocations may repeat `--changed-file` and `--removed-file`. Output follows the shared [facts-v1 contract](../../spec/facts-v1.md) and is byte-deterministic for identical input.
