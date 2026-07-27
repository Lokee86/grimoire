# CBM owner benchmark: Space Rocks

CBM: `codebase-memory-mcp 0.9.0`  
Repository HEAD: `8ecee9441d4522890b62f21a9e7e3ad4bf659ee5`  
Dirty working tree: `True`

The Arcana corpus labels are human judgments used as a common benchmark, not absolute ground truth.

| Mode | Pass | Required seed recall | MRR | Median latency |
| --- | ---: | ---: | ---: | ---: |
| bm25 | 0.0% | 0.0% | 0.000 | 139.5 ms |
| semantic | 0.0% | 0.0% | 0.000 | 62.5 ms |

## bm25

- `space-rocks-room-player-activation`: ranks=[0]; results=test_ffa_hides_team_column_and_adaptive_roster_scrolls, test_sampling_occurs_once_per_second_and_does_not_store_raw_frames, _on_connection_connected, test_attempt_and_connected_events_share_one_trace, _on_connected, test_close_before_connected_emits_connection_failed
- `space-rocks-realtime-lane-commit`: ranks=[0]; results=test_emits_devtools_enabled_after_successful_configuration, TestWriteResyncRequiredAndApplyRejectsStaleMatchAfterRoomAdvances, test_poll_emits_ready_only_after_all_channels_open, test_unparseable_baselines_survive_cleanup_until_capacity_overflow_clears_pending, TestAdvanceMetadataForSuccessfulWriteAdvancesEventLaneSequence, test_lobby_leave_sends_leave_and_returns_after_leave
- `space-rocks-client-inbound-binding`: ranks=[0]; results=test_debug_readouts_are_routed_to_dedicated_signals, test_websocket_and_webrtc_gameplay_packets_share_pipeline_application_path, test_presentation_adapter_forwards_decoded_session_state_to_local_lifecycle_flow, is_connected_to_server, test_compact_event_batch_expands_before_application_and_dedupes, test_packets_are_buffered_without_active_match_and_replayed_on_activation
- `space-rocks-client-gameplay-reset`: ranks=[0]; results=test_begin_accepting_gameplay_packets_activates_presentation_bridge, test_gameplay_composition_initializes_and_propagates_replay_availability, test_replay_waits_for_graceful_close_before_emitting_replay_requested, test_reset_clears_activation_and_pending_state, begin_accepting_gameplay_packets, _configure_realtime_replay_availability
- `space-rocks-webrtc-channel-recovery`: ranks=[0]; results=test_recovery_timeout_closes_replacement_and_emits_failure_once, test_webrtc_transport_reconnect_ownership_closes_previous_transport_and_starts_new_one, test_channel_close_replaces_transport_and_starts_new_offer_path, test_emits_one_request_failure_for_an_unhandled_exception, test_emits_one_request_failure_for_an_unhandled_exception, test_preserves_a_valid_trace_returns_it_and_emits_one_start_and_completion_with_duration
- `space-rocks-client-process-shutdown`: ranks=[0]; results=test_generates_a_trace_when_the_incoming_value_is_missing_or_invalid, test_generates_a_trace_when_the_incoming_value_is_missing_or_invalid, test_does_not_add_request_completion_after_a_specific_workflow_failure, test_does_not_add_request_completion_after_a_specific_workflow_failure, test_observability_ruby_is_a_language, test_reset_allows_reusing_a_projectile_id_in_a_new_match
- `space-rocks-account-result-persistence`: ranks=[0, 0, 0]; results=/internal/accounts, test_sampling_occurs_once_per_second_and_does_not_store_raw_frames, test_emits_devtools_enabled_exactly_once_across_repeated_configuration, newInMemoryHTTPListener, newInMemoryHTTPServer, test_auth_state_changed_outside_sign_in_screen_does_not_route
- `space-rocks-observability-contract-fanout`: ranks=[0]; results=test_realtime_wire_accepts_json_and_docs_outputs, test_observability_sync_plans_applies_and_detects_drift, test_valid_counter_commands_reuse_one_context_trace_for_event_and_packet, test_source_lookup_and_selection_use_metadata, test_emits_one_request_failure_for_an_unhandled_exception, test_emits_one_request_failure_for_an_unhandled_exception
- `space-rocks-canonical-event-admission`: ranks=[0, 0, 0]; results=test_mark_applied_rejects_the_same_sequence, test_counter_selectors_receive_the_same_target_set, test_all_player_target_selectors_receive_the_same_rows, before_each, test_generates_a_trace_when_the_incoming_value_is_missing_or_invalid, test_generates_a_trace_when_the_incoming_value_is_missing_or_invalid

## semantic

- `space-rocks-room-player-activation`: ranks=[0]; results=_set_file_writer_for_tests, test_left_click_with_pending_context_calls_pending_action_not_target_select, test_unexpected_close_after_connected_emits_disconnected, test_connection_waits_for_webrtc_ready_before_multiplayer_ready_then_sends_once, test_close_result_signal_distinguishes_expected_and_unexpected_without_changing_legacy_close_signal, assert_signal_emitted
- `space-rocks-realtime-lane-commit`: ranks=[0]; results=run, ClearGame, ClearGameInstance, normalized_score, newRoomTeams, render_invalid_input
- `space-rocks-client-inbound-binding`: ranks=[0]; results=_handle_channel_packet, send_raw_packet, Decode, DecodeType, compactWireEncodeBoundField, compactWireEncodeRecord
- `space-rocks-client-gameplay-reset`: ranks=[0]; results=before_each, after_each, reset, reset, get_local_lifecycle_flow, shell_debug
- `space-rocks-webrtc-channel-recovery`: ranks=[0]; results=configure, get_scripts, get_logger, get_gut, set_logger, reset
- `space-rocks-client-process-shutdown`: ranks=[0]; results=_on_show_server_hitboxes_toggled, _on_show_server_hitboxes_changed, _on_close_requested, set_enabled, set_server_hitboxes_enabled, configure
- `space-rocks-account-result-persistence`: ranks=[0, 0, 0]; results=stop, configure, _archive_files, _start_startup_maintenance, clear_bullets, SelectSendPlan
- `space-rocks-observability-contract-fanout`: ranks=[0]; results=_collect_pickup_types_from_scene, test_load_drop_table_parses_basic_asteroids_file, render_go_output, _load_dict_into, output_for_id, _validate_schema
- `space-rocks-canonical-event-admission`: ranks=[0, 0, 0]; results=_configure_observability_providers, close, connect_to_server, connect_to_server, _process, close
