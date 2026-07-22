# Go repository adapter

The adapter scans a Go repository and writes canonical ArcanaGraph repository facts.
It uses only the Go standard library.

From this directory:

```text
go run . -repo /path/to/repository -output /path/to/repository.facts.tsv
go test ./...
```

The repository must contain a `go.mod`. The scanner skips `.git`, `.worktrees`,
`.workingtrees`, and `vendor` directories. The output is a deterministic UTF-8
TSV file with `version\t1`, `N` node records, and `E` edge records. Existing output
files are replaced by the command.

Node identities and file content IDs use FNV-1a 64-bit with the same offset basis
and prime as ArcanaGraph's Rust `StableHasher`. Identity strings are documented by
the implementation and include a kind prefix to keep categories distinct.
