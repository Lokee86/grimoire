# Grimoire codemap

This codemap maps product behavior to the source that owns it. It is intended for maintainers, agents, and reviewers who need to locate the correct implementation boundary before changing Grimoire.

## Product entry points

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Executable entry | `cmd/grimoire/main.go` | Creates the application and delegates command handling. |
| Top-level command dispatch | `internal/app/run.go` | Parses the first command, prints help, and routes to the owning command implementation. |
| Public discovery CLI | `internal/app/query.go` | Parses `orient`, `search`, `trace`, `impact`, `inspect`, and compatibility `query` requests. |
| MCP server | `internal/app/mcp.go`, `internal/mcpserver/` | Exposes the `grimoire_discover` stdio tool and JSON-RPC framing. |
| State/status commands | `internal/app/status.go`, `internal/repostate/` | Inspects or aligns Grimoire, Lexicon, Arcana, document, and vector state. |
| Provider command namespaces | `internal/app/engine_commands.go`, `internal/app/engine_specs.go` | Resolves and forwards `grimoire lexicon ...` and `grimoire arcana ...` without absorbing provider ownership. |

## Discovery request flow

```text
CLI or MCP request
  -> internal/app
  -> internal/repostate.Ensure
  -> internal/agentruntime
       -> internal/agentquery
            -> exact/BM25 source retrieval
            -> Lexicon symbol facts
            -> Arcana relationships
       -> internal/knowledge
            -> document BM25
            -> optional document vectors
       -> internal/investigation
            -> session deduplication and handle ledger
  -> grimoire.discovery.v1 response
```

### Request and response contracts

| Area | Primary files | Notes |
| --- | --- | --- |
| Public schemas | `internal/agentquery/schema.go`, `internal/agentquery/model.go` | Defines the provider-neutral discovery request, lanes, handles, warnings, and assessment fields. |
| Search orchestration | `internal/agentquery/search.go`, `search_budget.go`, `search_seeds.go`, `search_relationships.go` | Executes balanced or narrow discovery and preserves lane budgets. |
| Orientation | `internal/agentquery/orient.go` | Produces compact entry points for unfamiliar repositories. |
| Handle resolution | `internal/agentquery/inspect.go`, `handle.go`, `resolve.go` | Resolves stable source, symbol, and relationship handles. |
| Graph expansion | `internal/agentquery/trace.go`, `impact.go` | Performs bounded path and dependent expansion from a selected anchor. |
| Evidence assessment | `internal/agentquery/assessment.go` | Reports observed and missing investigation dimensions without claiming proof. |
| Final response shaping | `internal/agentquery/excerpt.go`, `diversity.go`, `trace_shape.go` | Limits duplication and controls returned evidence shape. |

## Source indexing and retrieval

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Prepared source build | `internal/index/build.go`, `repository.go` | Walks relevant files and publishes prepared source state. |
| Chunking | `internal/index/chunk.go` | Creates stable source ranges used by search and handles. |
| Immutable identities | `internal/index/objects.go`, `codec.go`, `file_codec.go` | Stores content-addressed source objects and manifests. |
| Exclusions | `internal/index/exclusions.go`, `internal/ignore/` | Applies permanent and repository-configured traversal exclusions. |
| Exact search | `internal/retrieve/exact.go`, `exact_signals.go` | Finds literal identifiers, paths, routes, and configuration keys. |
| BM25 source search | `internal/retrieve/bm25.go`, `search.go` | Ranks implementation ranges lexically. |
| Scoped/file search | `internal/retrieve/scoped_search.go`, `file_search.go` | Narrows retrieval to selected files or scopes. |

## Documentation lane

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Document discovery | `internal/knowledge/discover.go` | Finds repository documentation independently from source evidence. |
| Section model | `internal/knowledge/sections.go`, `types.go` | Represents headings, ranges, citations, and document identity. |
| Document storage | `internal/knowledge/store.go`, `identity.go` | Publishes deterministic document state and freshness metadata. |
| BM25 document search | `internal/knowledge/search.go` | Ranks document sections without requiring embeddings. |
| Code links | `internal/knowledge/links.go` | Preserves explicit documentation-to-code references. |
| Optional vectors | `internal/knowledgevector/` | Builds and queries the supplemental document-vector lane. |
| Embedding runtime | `internal/embedding/` | Owns endpoint configuration, managed runtime lifecycle, request batching, and verification. |
| Vector storage boundary | `internal/vectorstore/` | Defines the Lodestone compatibility boundary used by Grimoire. |

## Lexicon integration

Grimoire does not parse languages. It consumes immutable Lexicon state.

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Lexicon state loading | `internal/lexiconfacts/load.go`, `state.go` | Resolves and validates the current Lexicon snapshot/export state. |
| Fact model | `internal/lexiconfacts/model.go`, `facts.go` | Maps Lexicon nodes, relationships, spans, and unresolved evidence into Grimoire structures. |
| Symbol candidates | `internal/lexiconfacts/candidates.go`, `query.go`, `rank.go` | Resolves and ranks declarations for discovery. |
| Relationship fallback | `internal/lexiconfacts/query_graph.go`, `relationship_attributes.go`, `relationship_provenance.go` | Supplies direct structural evidence when Arcana is unavailable. |
| Provider-neutral conversion | `internal/structure/` | Prevents the public response from depending on Lexicon-specific wire types. |

For parser, adapter, object-store, snapshot, or scan changes, start in [`lexicon/docs/CODEMAP.md`](../../lexicon/docs/CODEMAP.md).

## Arcana integration

Grimoire queries Arcana through an explicit process/protocol boundary.

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Client and sessions | `internal/arcanagraph/client.go`, `session.go` | Starts or reuses Arcana protocol sessions. |
| Wire protocol | `internal/arcanagraph/protocol.go` | Encodes requests and decodes `arcana.query.v1` responses. |
| Query orchestration | `internal/arcanagraph/query.go` | Resolves symbols and requests relationships, paths, and impact. |
| Candidate conversion | `internal/arcanagraph/candidates.go`, `evidence.go`, `model.go` | Converts graph responses into Grimoire structural evidence. |
| Structural reranking | `internal/arcanagraph/rerank*.go`, `semantic.go` | Ranks graph-derived candidates while keeping exact graph traversal authoritative. |
| State validation | `internal/arcanagraph/state.go` | Checks whether Arcana state matches the consumed Lexicon snapshot. |

For graph ingestion, storage, snapshot, protocol, or traversal changes, start in [`arcana/docs/CODEMAP.md`](../../arcana/docs/CODEMAP.md).

## Repository state ownership

| State | Owner | Main implementation |
| --- | --- | --- |
| `.grimoire/` source state | Grimoire | `internal/index/`, `internal/repostate/` |
| `.grimoire/knowledge/` | Grimoire | `internal/knowledge/`, `internal/knowledgevector/` |
| `.lexicon/` | Lexicon | `lexicon/internal/objectstore/`, `lexicon/internal/scan/` |
| `.arcana/` | Arcana | `arcana/src/repository/`, `arcana/src/snapshot/`, `arcana/src/storage/` |

`internal/repostate` coordinates freshness; it does not replace the owning component's builder or mutate another component's private state directly.

## Investigation sessions

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Ledger storage | `internal/investigation/ledger.go`, `json.go` | Persists returned evidence and session metadata. |
| Snapshot binding | `internal/investigation/handles.go`, `resolve.go` | Rejects stale handles and resolves current evidence. |
| Deduplication | `internal/investigation/retrieval.go`, `response.go` | Replaces repeated evidence with compact references. |
| Lifecycle | `internal/investigation/status.go`, `internal/app/investigation.go` | Creates, inspects, and closes sessions. |

## Evaluation and verification

| Area | Primary files | Responsibility |
| --- | --- | --- |
| Retrieval-quality corpus | `internal/app/testdata/retrieval-quality/` | Small deterministic source/document retrieval cases. |
| Document evaluation | `internal/knowledgeevaluation/`, `evaluation/knowledge/` | Scores document discovery and ranking. |
| Arcana evaluation | `internal/arcanaevaluation/`, `evaluation/arcana/` | Scores structural graph evidence. |
| Agent outcome benchmark | `evaluation/agent_discovery/`, `evaluation/agent_benchmark_tasks.v2.json` | Measures progressive discovery in end-to-end agent tasks. |
| Workflow orchestration | `scripts/workflow.py` | Runs bounded build, test, smoke, packaging, and installation workflows. |

## Common changes

### Add or change a discovery field

Start with `internal/agentquery/schema.go`, then update the producing mode, MCP conversion in `internal/app/mcp.go`, CLI output if applicable, contract documentation, and schema tests.

### Change source ranking

Start in `internal/retrieve/`. Keep exact matches and BM25 source matches distinct, then update retrieval-quality cases and response-shaping tests.

### Change documentation ranking

Start in `internal/knowledge/` or `internal/knowledgevector/`. Do not move document evidence into source or structural lanes.

### Change language semantics

Start in the owning adapter under `lexicon/adapters/`, not Grimoire. Update the normalized Lexicon contract only when the change affects cross-language facts.

### Change graph traversal or storage

Start in `arcana/src/protocol/`, `arcana/src/repository/`, `arcana/src/snapshot/`, or `arcana/src/storage/`. Grimoire should only change when the public integration contract changes.

### Change refresh behavior

Start in `internal/repostate/` for coordination and in the owning provider for its actual build. Preserve the explicit `current-only`, `refresh-if-needed`, and `force-refresh` modes.

## Boundaries to preserve

- Source, documentation, symbols, and relationships remain separate evidence classes.
- Lexicon owns language analysis; Arcana owns graph semantics and packed graph state.
- Grimoire owns discovery orchestration and the provider-neutral response.
- Optional vectors supplement deterministic discovery; they do not become required for source or graph correctness.
- Stable handles are snapshot-qualified and must not silently resolve against unrelated state.
- Planned architecture belongs in planning documents, not this codemap.
