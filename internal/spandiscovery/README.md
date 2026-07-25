# Source Span Discovery

`internal/spandiscovery` identifies meaningful source ranges without changing prepared-index chunk persistence.

## Owns

- deterministic language selection from repository paths;
- Markdown heading sections, excluding fenced-code headings;
- TOML table sections;
- indentation-bounded Python, Ruby, and GDScript types, functions, and methods;
- brace-bounded Go, Rust, JavaScript, TypeScript, Java, C#, C, and C++ declarations;
- ignoring braces in comments and common string/character literals;
- nested type/method classification; and
- narrowest-containing and overlap range helpers.

## Does not own

- prepared-index chunk creation or serialization;
- query matching and ranking;
- reading repository files;
- extracting text from a selected range; or
- context-package assembly and token-budget decisions.

Unsupported paths return no spans. Consumers must retain their existing fallback behavior when discovery yields nothing.
