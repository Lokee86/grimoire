# Lexicon command entry points

This directory owns executable entry points for the Lexicon application.

## Direct folders

| Folder | Responsibility |
| --- | --- |
| `lexicon/` | Build the `lexicon` CLI and delegate command execution to `internal/cli` |

## Does not own

Command entry points do not own application behavior, language analysis, storage, or contracts. They should remain thin process boundaries.

## Related documentation

- [Root README](../README.md)
- [Application commands](../docs/APPLICATION.md)
- [Architecture](../docs/ARCHITECTURE.md)
- [Development and build](../docs/DEVELOPMENT.md)

## Placement rules

Add a new executable only when it represents a separately deployable process with a clear lifecycle. Do not split ordinary subcommands into separate binaries.
