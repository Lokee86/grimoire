# Evidence descriptors

`internal/evidence` defines the provider-neutral metadata shared by source retrieval, Lexicon facts, Arcana graph evidence, and response shaping.

## Responsibility

The package owns:

- stable identities for prepared source ranges and derived evidence;
- investigation intent labels such as direct location, mechanism, call chain, and architecture;
- evidence roles such as primary, supporting, structural, and context;
- links between related evidence identities;
- bounded graph-ranking signals;
- deterministic merging of metadata contributed by independent providers.

It does not retrieve source, query a graph, rank final responses, or own any provider wire format.

## Core contract

`Descriptor` is the shared coordination structure. Producers populate only the fields they own. Consumers merge descriptors without discarding metadata from another provider.

Merge behavior is intentionally conservative:

- set-like fields retain stable first-seen order;
- the stronger exact-match value is retained;
- the larger token estimate is retained;
- the best facet rank is retained per facet;
- graph distance keeps the shortest observed distance;
- graph proximity and centrality keep the strongest observed value;
- links are deduplicated without reordering existing entries.

`RangeIdentity` normalizes repository paths and binds an identity to one exact line range. `StableID` derives a compact deterministic identifier from a namespace and ordered parts.

## Consumers

| Package | Use |
| --- | --- |
| `internal/retrieve` | Adds retrieval intent, ranking, exact-match, range, and redundancy metadata to source candidates. |
| `internal/lexiconfacts` | Adds symbol and direct-relationship evidence metadata. |
| `internal/arcanagraph` | Adds graph distance, relation, role, centrality, and linked-evidence metadata. |
| `internal/structure` | Re-exports the provider-neutral evidence contract at the structural boundary. |
| `internal/agentquery` | Uses merged descriptors during lane shaping, diversity, and progressive response construction. |

## File map

- `descriptor.go` — intents, roles, links, graph signals, descriptor merge rules, stable identities, and cloning.
- `descriptor_test.go` — identity normalization and provider-metadata merge behavior.
- `graph_test.go` — graph-signal merge behavior.

## Change guidance

A new field belongs here only when multiple evidence providers or provider-neutral response logic need the same concept. Provider-specific payloads remain in the owning package. Any mergeable field must define deterministic conflict behavior and have focused tests.
