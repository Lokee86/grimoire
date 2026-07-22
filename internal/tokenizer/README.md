# Tokenizer

`internal/tokenizer` owns Grimoire's single token-counting policy and hides the third-party tokenizer implementation from indexing and package compilation.

## Owns

- the fixed `o200k_base` tokenizer identity;
- one lazily initialized shared codec;
- exact token counting for prepared chunk text; and
- exact token counting for serialized context packages.

## Does not own

- retrieval terms, postings, or ranking;
- language parsing or chunk boundaries;
- package selection policy;
- model selection; or
- chat, tool, or agent wrapper overhead outside Grimoire's emitted JSON.

## Dependency

The implementation uses `github.com/tiktoken-go/tokenizer` with its embedded `o200k_base` vocabulary. It does not download vocabulary data at runtime.

## Contract

Grimoire supports one tokenizer. `Name` is part of both the prepared-index manifest and context-package schema. A prepared index using another tokenizer identity is incompatible and must be rebuilt.

## Related documentation

- [Indexing reference](../../docs/reference/indexing.md)
- [Context package](../../docs/reference/context-package.md)
- [Prepared index](../../docs/architecture/prepared-index.md)
