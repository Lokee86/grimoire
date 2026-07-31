# Lexical index

`internal/lexical` provides the reusable deterministic lexical document representation used by Grimoire's source and documentation retrieval paths.

It is a low-level indexing package. BM25 scoring, exact source matching, result assembly, and provider orchestration remain in their owning packages.

## Responsibilities

- tokenize prose, paths, and identifiers, including camel-case and letter/digit boundaries;
- analyze text into deterministic term frequencies and specialized token fields;
- build postings for scoring and candidate selection;
- preserve declaration vocabulary statistics;
- reuse unchanged analyzed documents during rebuilds;
- encode and decode the versioned lexical index representation.

## Document model

`Analyze` converts one `Input` into a `Document` containing:

- total token length and sorted term frequencies;
- base-name tokens;
- full path tokens;
- first-line tokens;
- declaration-oriented tokens from the path and the first bounded non-comment source lines.

The specialized fields allow callers to select and score likely files or sections without reparsing raw text during each query.

## Candidate selection

The index maintains two related structures:

- term-frequency postings used by scoring callers;
- candidate postings containing all searchable field terms.

`CandidateDocuments` returns the stable union of documents matching any requested term. `CandidateDocumentsAll` intersects postings, beginning with the rarest term to bound work. Returned document indexes are deterministic.

## Rebuild and persistence

`Rebuild` reuses a previous analyzed `Document` when an input key is unchanged and analyzes only new keys. `New` validates sorted fields, positive frequencies, and document identities before constructing postings.

`Encode` and `Decode` use a versioned JSON representation of analyzed documents. Decode rebuilds and revalidates the in-memory postings rather than trusting serialized derived state.

## File map

- `tokenize.go` — normalized token streams, unique token sets, and identifier splitting.
- `document.go` — input/document types, term-frequency analysis, and bounded declaration headers.
- `index.go` — validated index construction, incremental document reuse, postings, and vocabulary.
- `candidates.go` — any-term and all-term candidate selection.
- `codec.go` — versioned persistence boundary.
- `tokenize_test.go` — identifier and text tokenization behavior.
- `codec_test.go` — persistence and validation behavior.

## Boundaries

- Retrieval packages decide ranking formulas and weights.
- Index and knowledge packages decide which source chunks or document sections become lexical inputs.
- This package does not interpret programming-language semantics; Lexicon owns that work.
- Format changes require a version change and compatibility decision rather than silent reinterpretation.
