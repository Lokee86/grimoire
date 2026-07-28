# Tokenizer

`internal/tokenizer` owns Grimoire's single token-counting and token-splitting policy and hides the third-party tokenizer implementation from indexing and embedding request safety.

## Owns

- the fixed `o200k_base` tokenizer identity;
- one lazily initialized shared codec;
- exact token counting for prepared source chunks;
- encode/decode support for hard chunk splitting; and
- final embedding-input token limits.

## Does not own

- retrieval terms, postings, or ranking;
- language parsing or semantic chunk boundaries;
- discovery result limits;
- model selection; or
- chat, tool, or agent wrapper overhead outside Grimoire's indexed text and embedding requests.

## Dependency

The implementation uses `github.com/tiktoken-go/tokenizer` with its embedded `o200k_base` vocabulary. It does not download vocabulary data at runtime.

## Contract

Grimoire supports one tokenizer. `Name` is part of the prepared-index manifest. A prepared index using another tokenizer identity is incompatible and must be rebuilt.

## Related documentation

- [Indexing reference](../../docs/reference/indexing.md)
- [Prepared-index architecture](../../docs/architecture/prepared-index.md)
- [Embedding model](../../docs/reference/embedding-model.md)

