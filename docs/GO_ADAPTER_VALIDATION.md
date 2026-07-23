# Go adapter real-repository validation

This is a dated validation record, not a current performance benchmark. It captures the Go semantic adapter validation performed on July 22, 2026 before the later multi-module repository and adaptive semantic-parallelism changes were merged into `main`. The semantic categories and limitations remain relevant; counts, file totals, packed sizes, and elapsed behavior may differ on the current implementation.

Validated on July 22, 2026 against two existing Go modules without modifying
either source repository.

## Results

| Repository | Nodes | Edges | Indexed files | Packages | Call expressions | Definite call expressions | Possible target facts | Conversion expressions | Unresolved calls | Closures | Captures | Packed graph | Catalogue | Unresolved file |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Demon Docs | 4,687 | 17,176 | 353 | 33 | 17,993 | 17,276 | 287 | 675 | 0 | 357 | 1,055 | 281,248 B | 539,274 B | 10 B |
| Space Rocks game server | 5,596 | 19,853 | 528 | 53 | 18,430 | 17,594 | 731 | 765 | 0 | 389 | 526 | 327,936 B | 762,943 B | 10 B |

`Definite call expressions` counts call sites with one callable contract. This
includes internal, standard-library, third-party, built-in, interface-contract,
and uniquely resolved function-value calls. `Possible target facts` counts
conservative runtime targets emitted for interface dispatch, callbacks, method
values, and other function flows; multiple targets may belong to one call site.
Type conversions are represented separately through `converts-to` edges.

Neither repository produced an unresolved call fact. This does not mean every
runtime dispatch is claimed to be uniquely known. Definite calls use `calls`;
multi-target dispatch uses `possible-calls`; conversions use `converts-to`.
Arcana therefore retains uncertainty instead of turning every possible target
into a definite edge.

Both repositories completed Go package/type loading with zero reported semantic
errors.

## Resolution behavior

The adapter combines repository-wide AST extraction with
`golang.org/x/tools/go/packages`, Go type information, SSA, and variable-type
analysis. It models:

- same-package and cross-package functions and methods;
- recursive calls as valid graph self-edges;
- standard-library and external API symbols without indexing dependency source;
- built-ins as callable symbol nodes;
- type conversions as `converts-to` relationships;
- interface types, interface methods, embedded interfaces, and implementation
  relationships;
- concrete interface targets as conservative `possible-calls` relationships;
- function variables, callback parameters, method values, and returned function
  values;
- closures as independent function nodes, including nested closure bodies;
- closure captures as variable nodes reached through `references` edges;
- mutually exclusive build-tag declarations under one canonical package-level
  symbol identity;
- AST-only callable contracts for files excluded from the active host build.

External closures and compiler-generated wrappers with no source declaration
receive stable synthetic nodes. Reflection-heavy or opaque calls can therefore
retain a callable contract even when no concrete repository implementation can
be proven.

## Packed graph results

Demon Docs produced these unique packed relationships:

- 10,909 `calls`;
- 279 `possible-calls`;
- 445 `converts-to`;
- 688 capture `references`;
- 45 `implements`;
- zero unresolved references.

Multiple call sites between the same two nodes and relation become one packed
edge, so packed relationship counts are intentionally lower than call-site
counts. The query protocol does not calculate a percentage from those
incompatible units.

## Query checks

A real Demon Docs closure at `internal/reverseindex/watch.go:26:9` returned:

- two conservative `possible-calls` targets, `Lock` and `Unlock`;
- seven captured variables through `references` edges;
- zero unresolved references in the snapshot.

The protocol statistics reported all new relation categories separately.
Both repositories imported successfully into verified repository snapshots.

An incremental test removed one real `possible-calls` edge from that closure.
`update-facts` reported one changed file, one removed edge, and an overlay rather
than a rebuild. The updated query returned one remaining target, and snapshot
diff reported exactly one relationship-changed source node.

Both fact files were generated twice and compared byte-for-byte. Repeated output
was identical.

## Remaining semantic limits

The remaining limits are about precision and provenance rather than missing
call records:

- VTA is conservative and can over-approximate runtime targets;
- build-tag variants currently share one logical symbol rather than carrying a
  per-build-configuration execution view;
- reflection, plugins, cgo, assembly, and runtime-generated functions may retain
  contracts or synthetic targets instead of one proven concrete implementation;
- generated-code classification is not yet represented explicitly;
- dependency implementation graphs require indexing those dependencies as
  separate repositories;
- packed graph edges are deduplicated relationships, so exact call-site coverage
  requires a future call-site fact layer rather than reconstruction from edges.
