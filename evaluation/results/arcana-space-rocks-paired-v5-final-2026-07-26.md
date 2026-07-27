# Arcana semantic graph retrieval evaluation: Space Rocks

Generated: 2026-07-26 20:58:09-07:00  
Variant: `paired-v5-final`  
Arcana snapshot: `sha256:1e7c96b463135cbcd7cf92860fe18e3a3445129407c94956a2b87cb6f7cf1177`  
Embedding identity: `qwen3-embedding-0.6b-q8_0-512d`  
Corpus cases: `9`

The paired modes use the same prepared source, Lexicon export, and Arcana graph snapshot. `lexicon-seeds` bypasses semantic lookup; `lexicon-plus-vector` adds the existing Arcana vector index before the same deterministic graph expansion.

## Aggregate comparison

| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | p95 latency | Median payload | p95 payload | Provider calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.000 | 0.0% | 9125.1 ms | 12089.5 ms | 13823 B | 61206 B | 2.00 |
| lexicon-plus-vector | 22.2% | 22.2% | 0.059 | 22.2% | 10954.3 ms | 15152.2 ms | 27176 B | 45206 B | 3.00 |

## Seed recall at k

| Mode | R@1 | R@3 | R@6 |
| --- | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.0% |
| lexicon-plus-vector | 0.0% | 11.1% | 22.2% |

## Vector-minus-baseline deltas

| Metric | Delta |
| --- | ---: |
| Pass rate | +22.2 pp |
| Required seed recall | +22.2 pp |
| Seed recall@1 | +0.0 pp |
| Seed recall@3 | +11.1 pp |
| Seed recall@6 | +22.2 pp |
| MRR | +0.059 |
| Required structural recall | +22.2 pp |
| Median latency | +1829.2 ms |
| Median payload | +13353 B |
| Mean provider calls | +1.00 |

## Cases

| Case | Mode | Pass | Seed recall | MRR | Structural recall | Latency | Payload | Calls | Error |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `space-rocks-room-player-activation` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 8531.0 ms | 5875 B | 2 |  |
| `space-rocks-room-player-activation` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 9942.4 ms | 28755 B | 3 |  |
| `space-rocks-realtime-lane-commit` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 8756.4 ms | 10511 B | 2 |  |
| `space-rocks-realtime-lane-commit` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 9608.5 ms | 26974 B | 3 |  |
| `space-rocks-client-inbound-binding` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 7576.7 ms | 18507 B | 2 |  |
| `space-rocks-client-inbound-binding` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 10843.3 ms | 24276 B | 3 |  |
| `space-rocks-client-gameplay-reset` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 9157.9 ms | 5600 B | 2 |  |
| `space-rocks-client-gameplay-reset` | lexicon-plus-vector | true | 100.0% | 0.200 | 100.0% | 10832.2 ms | 42709 B | 3 |  |
| `space-rocks-webrtc-channel-recovery` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 9430.9 ms | 7836 B | 2 |  |
| `space-rocks-webrtc-channel-recovery` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 12571.4 ms | 27176 B | 3 |  |
| `space-rocks-client-process-shutdown` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 9125.1 ms | 61206 B | 2 |  |
| `space-rocks-client-process-shutdown` | lexicon-plus-vector | true | 100.0% | 0.333 | 100.0% | 10954.3 ms | 44551 B | 3 |  |
| `space-rocks-account-result-persistence` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 8995.3 ms | 56886 B | 2 |  |
| `space-rocks-account-result-persistence` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 11560.6 ms | 45206 B | 3 |  |
| `space-rocks-observability-contract-fanout` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 9834.9 ms | 13823 B | 2 |  |
| `space-rocks-observability-contract-fanout` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 13131.7 ms | 23900 B | 3 |  |
| `space-rocks-canonical-event-admission` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 12089.5 ms | 42603 B | 2 |  |
| `space-rocks-canonical-event-admission` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 15152.2 ms | 27054 B | 3 |  |

## Per-case seed rankings

### `space-rocks-room-player-activation` / `lexicon-seeds`

Where does a connected lobby roster become live authoritative avatars while preserving per-connection identity and rolling back partial binds?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `connected` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 2 | lexicon | `back` | `client/addons/gut/printers.gd` | false | false |
| 3 | lexicon | `Identity` | `services/game-server/internal/authclient/types.go` | false | false |
| 4 | lexicon | `lobby` | `client/scripts/lobby` | false | false |

### `space-rocks-room-player-activation` / `lexicon-plus-vector`

Where does a connected lobby roster become live authoritative avatars while preserving per-connection identity and rolling back partial binds?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `LobbySessionState` | `client/scripts/lobby/lobby_session_state.gd` | false | false |
| 2 | lexicon | `connected` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 3 | vector | `ResetToLobbyForSession` | `services/game-server/internal/rooms/room_lifecycle.go` | false | false |
| 4 | lexicon | `back` | `client/addons/gut/printers.gd` | false | false |
| 5 | vector | `apply_lobby_state` | `client/scripts/ui/lobby/multiplayer_lobby.gd` | false | false |
| 6 | lexicon | `Identity` | `services/game-server/internal/authclient/types.go` | false | false |

### `space-rocks-realtime-lane-commit` / `lexicon-seeds`

Which server boundary refuses stale round context, reserves transport capacity for an entire update batch, and advances client baselines only after successful sends?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `update` | `client/scripts/shell/gameplay_hud_flow.gd` | false | false |
| 2 | lexicon | `Client` | `services/game-server/internal/authclient/client.go` | false | false |
| 3 | lexicon | `Context` | `services/game-server/internal/networking/tooling/router.go` | false | false |
| 4 | lexicon | `context` | `client/scripts/devtools/context` | false | false |

### `space-rocks-realtime-lane-commit` / `lexicon-plus-vector`

Which server boundary refuses stale round context, reserves transport capacity for an entire update batch, and advances client baselines only after successful sends?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `test_baseline_bucket_overflow_requests_resync_and_clears_pending_state` | `client/tests/unit/protocol/realtime/test_lifecycle_lane_gate.gd` | false | false |
| 2 | lexicon | `update` | `client/scripts/shell/gameplay_hud_flow.gd` | false | false |
| 3 | vector | `test_expired_bucket_discards_and_requests_each_lane_once` | `client/tests/unit/networking/realtime/test_realtime_packet_pipeline_match_boundary.gd` | false | false |
| 4 | lexicon | `Client` | `services/game-server/internal/authclient/client.go` | false | false |
| 5 | vector | `test_protocol_match_begin_end_preserves_transport_without_close` | `client/tests/unit/test_client_connection_service.gd` | false | false |
| 6 | lexicon | `Context` | `services/game-server/internal/networking/tooling/router.go` | false | false |

### `space-rocks-client-inbound-binding` / `lexicon-seeds`

Where are decoded control signals and unreliable gameplay families connected to their application and presentation consumers without making the network facade own packet application?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `connected` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 2 | lexicon | `presentation` | `client/scripts/gameplay/presentation` | false | false |
| 3 | lexicon | `decode` | `client/scripts/networking/packets/packet_codec.gd` | false | false |

### `space-rocks-client-inbound-binding` / `lexicon-plus-vector`

Where are decoded control signals and unreliable gameplay families connected to their application and presentation consumers without making the network facade own packet application?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `Decode` | `services/game-server/internal/protocol/packetcodec/codec.go` | false | false |
| 2 | lexicon | `connected` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 3 | vector | `packet_codec` | `client/scripts/networking/packets/packet_codec.gd` | false | false |
| 4 | lexicon | `presentation` | `client/scripts/gameplay/presentation` | false | false |
| 5 | vector | `gameplay_packet_applied` | `client/scripts/networking/realtime/realtime_packet_pipeline.gd` | false | false |
| 6 | vector | `HandleWebRTCSignalingPacket` | `services/game-server/internal/networking/inbound/webrtc.go` | false | false |

### `space-rocks-client-gameplay-reset` / `lexicon-seeds`

Which client session owner stops accepting game traffic and clears both the presentation bridge and gameplay composition before replay or route changes?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `test_begin_accepting_gameplay_packets_activates_presentation_bridge` | `client/tests/unit/test_gameplay_session_controller.gd` | false | false |
| 2 | lexicon | `presentation` | `client/scripts/gameplay/presentation` | false | false |
| 3 | lexicon | `session` | `client/scripts/gameplay/session` | false | false |

### `space-rocks-client-gameplay-reset` / `lexicon-plus-vector`

Which client session owner stops accepting game traffic and clears both the presentation bridge and gameplay composition before replay or route changes?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `_on_gameplay_replay_requested` | `client/scripts/session/gameplay_session_controller.gd` | false | true |
| 2 | lexicon | `test_begin_accepting_gameplay_packets_activates_presentation_bridge` | `client/tests/unit/test_gameplay_session_controller.gd` | false | false |
| 3 | vector | `activate` | `client/tests/unit/test_gameplay_session_controller.gd` | false | false |
| 4 | lexicon | `presentation` | `client/scripts/gameplay/presentation` | false | false |
| 5 | vector | `reset` | `client/scripts/session/gameplay_session_controller.gd` | true | false |
| 6 | vector | `reset` | `client/tests/unit/test_gameplay_session_controller.gd` | false | false |

### `space-rocks-webrtc-channel-recovery` / `lexicon-seeds`

What client path replaces an unreliable transport after one lane closes, bounds the retry window, and emits a terminal recovery outcome instead of reconnecting forever?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `Window` | `tools/parked-plasmic-next-host/src/components/media/CrtMediaFrame.tsx` | false | false |
| 2 | lexicon | `Window` | `web-astro/src/components/media/CrtMediaFrame.tsx` | false | false |
| 3 | lexicon | `path` | `client/addons/gut/collected_script.gd` | false | false |
| 4 | lexicon | `path` | `client/addons/gut/gui/RunResults.gd` | false | false |
| 5 | lexicon | `path` | `client/addons/gut/gui/ShellOutOptions.gd` | false | false |
| 6 | lexicon | `path` | `client/addons/gut/gui/ShortcutDialog.gd` | false | false |

### `space-rocks-webrtc-channel-recovery` / `lexicon-plus-vector`

What client path replaces an unreliable transport after one lane closes, bounds the retry window, and emits a terminal recovery outcome instead of reconnecting forever?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `test_recovery_timeout_closes_replacement_and_emits_failure_once` | `client/tests/unit/networking/test_realtime_transport_session.gd` | false | false |
| 2 | lexicon | `Window` | `tools/parked-plasmic-next-host/src/components/media/CrtMediaFrame.tsx` | false | false |
| 3 | vector | `test_two_sequential_successful_recoveries_are_allowed` | `client/tests/unit/networking/test_realtime_transport_session.gd` | false | false |
| 4 | lexicon | `Window` | `web-astro/src/components/media/CrtMediaFrame.tsx` | false | false |
| 5 | vector | `_fail_recovery` | `client/scripts/networking/webrtc/realtime_transport_session.gd` | false | true |
| 6 | lexicon | `path` | `client/addons/gut/collected_script.gd` | false | false |

### `space-rocks-client-process-shutdown` / `lexicon-seeds`

Where does a window-close request gracefully mark the server connection, stop a hosted local process, and then terminate the scene tree without pretending this is a room leave?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `stop` | `client/scripts/boot/local_server_process.gd` | false | false |
| 2 | lexicon | `process` | `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd` | false | false |
| 3 | lexicon | `process` | `client/scripts/gameplay/presentation/local_player_presentation_controller.gd` | false | false |
| 4 | lexicon | `Close` | `services/diagnostic-aggregator/hosted/service.go` | false | false |
| 5 | lexicon | `Stop` | `services/game-server/internal/game/game.go` | false | false |
| 6 | lexicon | `Close` | `services/game-server/internal/networking/tooling/router.go` | false | false |

### `space-rocks-client-process-shutdown` / `lexicon-plus-vector`

Where does a window-close request gracefully mark the server connection, stop a hosted local process, and then terminate the scene tree without pretending this is a room leave?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `close_gracefully` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 2 | lexicon | `stop` | `client/scripts/boot/local_server_process.gd` | false | false |
| 3 | vector | `request_shutdown` | `client/scripts/session/app_shutdown_controller.gd` | true | false |
| 4 | lexicon | `process` | `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd` | false | false |
| 5 | vector | `leaveDisconnectedRoom` | `services/game-server/internal/networking/websocket_room_exit.go` | false | false |
| 6 | lexicon | `process` | `client/scripts/gameplay/presentation/local_player_presentation_controller.gd` | false | false |

### `space-rocks-account-result-persistence` / `lexicon-seeds`

How does a completed round's per-player outcome cross the in-process persistence boundary, route online accounts over HTTP, and update cumulative Rails statistics exactly once?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `process` | `client/scripts/devtools/context/devtools_overlay_context.gd` | false | false |
| 2 | lexicon | `process` | `client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd` | false | false |
| 3 | lexicon | `process` | `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd` | false | false |
| 4 | lexicon | `process` | `client/scripts/gameplay/presentation/local_player_presentation_controller.gd` | false | false |
| 5 | lexicon | `process` | `client/scripts/devtools/context/devtools_command_context.gd` | false | false |
| 6 | lexicon | `process` | `client/scripts/devtools/context/devtools_hotkey_context.gd` | false | false |

### `space-rocks-account-result-persistence` / `lexicon-plus-vector`

How does a completed round's per-player outcome cross the in-process persistence boundary, route online accounts over HTTP, and update cumulative Rails statistics exactly once?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `show` | `services/api-server/app/controllers/api/player/stats_controller.rb` | false | false |
| 2 | lexicon | `process` | `client/scripts/devtools/context/devtools_overlay_context.gd` | false | false |
| 3 | vector | `SerializeStats` | `services/api-server/app/services/player_stats/serialize_stats.rb` | false | false |
| 4 | lexicon | `process` | `client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd` | false | false |
| 5 | vector | `<block>@61:13` | `services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb` | false | false |
| 6 | lexicon | `process` | `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd` | false | false |

### `space-rocks-observability-contract-fanout` / `lexicon-seeds`

Which source-of-truth fanout validates one telemetry contract and plans deterministic outputs for the Go services, Godot client, Rails API, JSON fixtures, and reference docs?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 2 | lexicon | `json` | `services/api-server/app/lib/observability/emitter.rb` | false | false |
| 3 | lexicon | `json` | `services/api-server/app/services/auth/providers/discord_current_user.rb` | false | false |
| 4 | lexicon | `json` | `services/api-server/app/services/auth/providers/discord_token_exchange.rb` | false | false |
| 5 | lexicon | `Rails` | `services/api-server/config/initializers/filter_parameter_logging.rb` | false | false |
| 6 | lexicon | `Client` | `services/diagnostic-aggregator/internal/diagnosticclient/client.go` | false | false |

### `space-rocks-observability-contract-fanout` / `lexicon-plus-vector`

Which source-of-truth fanout validates one telemetry contract and plans deterministic outputs for the Go services, Godot client, Rails API, JSON fixtures, and reference docs?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `contract_generated.gd` | `client/scripts/generated/observability/contract_generated.gd` | false | false |
| 2 | lexicon | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 3 | vector | `contract_generated` | `client/scripts/generated/observability/contract_generated.gd` | false | false |
| 4 | lexicon | `json` | `services/api-server/app/lib/observability/emitter.rb` | false | false |
| 5 | vector | `test_observability_generators_are_deterministic_and_complete` | `tools/data_sync/tests/test_observability_generators.py` | false | false |
| 6 | lexicon | `json` | `services/api-server/app/services/auth/providers/discord_current_user.rb` | false | false |

### `space-rocks-canonical-event-admission` / `lexicon-seeds`

Where do the Go service layer, Godot client, and Rails worker each enforce the same event admission, redaction, size, identity, serialization, and degraded-write rules before accepting a canonical record?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 2 | lexicon | `Identity` | `services/game-server/internal/authclient/types.go` | false | false |
| 3 | lexicon | `size` | `client/addons/gut/one_to_many.gd` | false | false |
| 4 | lexicon | `record` | `client/scripts/devtools/measurement/lifecycle_counter_window.gd` | false | false |
| 5 | lexicon | `write` | `client/scripts/devtools/measurement/measurement_report_writer.gd` | false | false |
| 6 | lexicon | `record` | `client/scripts/devtools/measurement/timing_summary.gd` | false | false |

### `space-rocks-canonical-event-admission` / `lexicon-plus-vector`

Where do the Go service layer, Godot client, and Rails worker each enforce the same event admission, redaction, size, identity, serialization, and degraded-write rules before accepting a canonical record?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `records.go` | `services/game-server/internal/protocol/realtime/records.go` | false | false |
| 2 | lexicon | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 3 | vector | `quantized_records.go` | `services/game-server/internal/protocol/realtime/quantized_records.go` | false | false |
| 4 | lexicon | `Identity` | `services/game-server/internal/authclient/types.go` | false | false |
| 5 | vector | `BuildWorldDeltaPacket` | `services/game-server/internal/protocol/realtime/delta.go` | false | false |
| 6 | lexicon | `size` | `client/addons/gut/one_to_many.gd` | false | false |

