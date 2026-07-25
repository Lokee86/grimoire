# Diff Context

`diffcontext` converts bounded Git diff output into current-tree source spans, changed prepared-index candidates, and structural change evidence.

Ownership:

- invoke Git without a shell for named scopes or one revision/range;
- include untracked, non-ignored files for `working-tree`;
- parse zero-context unified diffs, renames, additions, deletions, and file-level changes;
- map changed spans to prepared chunks with explicit `git-diff` provenance;
- construct a bounded retrieval-only query containing changed paths and declaration anchors.

It does not perform graph traversal or package assembly. Grimoire passes the retrieval-only query to Lexicon and Arcana, then merges their impact evidence with changed source candidates in `internal/app`.
