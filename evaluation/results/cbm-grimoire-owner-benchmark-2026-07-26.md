# CBM owner benchmark: Grimoire

CBM: `codebase-memory-mcp 0.9.0`  
Repository HEAD: `4353543aa06e72015f896d4fd4c3a38096b9615f`  
Dirty working tree: `True`

The Arcana corpus labels are human judgments used as a common benchmark, not absolute ground truth.

| Mode | Pass | Required seed recall | MRR | Median latency |
| --- | ---: | ---: | ---: | ---: |
| bm25 | 0.0% | 0.0% | 0.000 | 51.0 ms |
| semantic | 0.0% | 0.0% | 0.000 | 45.2 ms |

## bm25

- `arcana-semantic-seeds`: ranks=[0]; results=filters_synthetic_and_anonymous_entry_points, serves_bounded_graph_analysis_queries, precompiled_publication_matches_standalone_validation_without_recompiling, overlay_snapshot_matches_materialized_graph, new, new
- `arcana-seed-composition`: ranks=[0, 0]; results=all_six_combinations_are_deterministic_and_valid, test_ids_and_output_are_deterministic, test_ids_and_output_are_deterministic, test_dependency_manifests_and_local_imports_are_deterministic, renders_identifier_and_path_terms_for_conceptual_retrieval, TestReportsAreDeterministicAndMeasureRepeatability
- `arcana-deterministic-expansion`: ranks=[0]; results=sync_builds_reuses_and_registers_a_lexicon_snapshot, compiler_preserves_and_validates_unresolved_references, resolves_typed_values_and_standard_callback_inputs, resolves_self_variants_generated_defaults_and_closure_captures, resolves_inherent_field_alias_constructor_ufcs_and_macro_calls, resolves_local_methods_through_standard_combinators_and_trait_impls
- `arcana-graph-documents`: ranks=[0, 0]; results=renders_node_and_immediate_graph_neighborhood, binds_graph_catalogue_unresolved_and_source_facts, with_unresolved, embedding_batch_concurrency_is_bounded, serves_bounded_graph_analysis_queries, parses_and_formats_generated_target
- `arcana-vector-search`: ranks=[0]; results=catalogue_file_is_immutable_and_validated, test_cli_output_is_stable_and_limit_is_applied, cloned_packed_graphs_share_immutable_backing_bytes, logical_edge_order_does_not_change_packed_bytes, renders_node_and_immediate_graph_neighborhood, sync_builds_reuses_and_registers_a_lexicon_snapshot

## semantic

- `arcana-semantic-seeds`: ranks=[0]; results=put_u32, put_u32, new, required, parse_span, put_u64
- `arcana-seed-composition`: ranks=[0, 0]; results=SearchManyWithConfig, nearestDeclarationAlias, TestCodecRoundTripPreservesPostingState, scoreChunk, TestSearchDeclarationAliasPromotesMatchingDeclaration, hasTerm
- `arcana-deterministic-expansion`: ranks=[0]; results=eligibleAliasToken, receiverStart, isPath, classifySignal, identifierTerms, readMacroCallee
- `arcana-graph-documents`: ranks=[0, 0]; results=facts, protocol_parts_transfer_owned_components, repository_identity_from_checksum, binds_graph_catalogue_unresolved_and_source_facts, write_immutable, decode
- `arcana-vector-search`: ranks=[0]; results=enumeratedRetrievalClauses, scoreRetrievalClause, splitStrongPunctuation, packageStructuralSelections, DefaultQueryOptions, testQueryOptions
