# Ignore Policy

`internal/ignore` owns Git-ignore pattern loading and matching for repository traversal.

## Owns

- root `.gitignore` loading;
- nested `.gitignore` loading as directories are entered;
- Git-ignore pattern scope and negation through go-git;
- replacement ignore-file loading; and
- exclusion of the configured replacement control file itself.

## Does not own

- permanent exclusions such as `.git`, `.grimoire`, `.ddocs`, `.lexicon`, `.arcana`, `.warlock`, or worktree containers;
- custom state-path exclusion;
- supported file types;
- size or binary filtering; or
- traversal statistics.

Those concerns remain in `internal/index`.

## Main file

- `policy.go` - policy construction, directory pattern loading, and matching.

## Related documentation

- [Indexing reference](../../docs/reference/indexing.md)
- [System overview](../../docs/architecture/system-overview.md)
