# Arcana semantic graph retrieval evaluation: Space Rocks

Generated: 2026-07-26 23:05:44-07:00  
Variant: `hybrid-rerank-graph256`  
Arcana snapshot: `sha256:1e7c96b463135cbcd7cf92860fe18e3a3445129407c94956a2b87cb6f7cf1177`  
Embedding identity: `qwen3-embedding-0.6b-q8_0-512d`  
Corpus cases: `9`

The paired modes use the same prepared source, Lexicon export, and Arcana graph snapshot. `lexicon-seeds` bypasses semantic lookup; `lexicon-plus-vector` adds the existing Arcana vector index before the same deterministic graph expansion.

## Aggregate comparison

| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | p95 latency | Median payload | p95 payload | Provider calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| lexicon-seeds | 0.0% | 0.0% | 0.000 | 0.0% | 13983.4 ms | 22830.4 ms | 13823 B | 61206 B | 2.00 |
| lexicon-plus-vector | 22.2% | 22.2% | 0.074 | 22.2% | 17213.2 ms | 25490.3 ms | 50964 B | 77776 B | 3.00 |

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
| MRR | +0.074 |
| Required structural recall | +22.2 pp |
| Median latency | +3229.8 ms |
| Median payload | +37141 B |
| Mean provider calls | +1.00 |

## Cases

| Case | Mode | Pass | Seed recall | MRR | Structural recall | Latency | Payload | Calls | Error |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `space-rocks-room-player-activation` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 13935.7 ms | 5875 B | 2 |  |
| `space-rocks-room-player-activation` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 19158.7 ms | 59372 B | 3 |  |
| `space-rocks-realtime-lane-commit` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 15919.2 ms | 10511 B | 2 |  |
| `space-rocks-realtime-lane-commit` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 18629.4 ms | 77776 B | 3 |  |
| `space-rocks-client-inbound-binding` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 13972.0 ms | 18507 B | 2 |  |
| `space-rocks-client-inbound-binding` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 16279.8 ms | 62626 B | 3 |  |
| `space-rocks-client-gameplay-reset` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 13983.4 ms | 5600 B | 2 |  |
| `space-rocks-client-gameplay-reset` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 17213.2 ms | 50964 B | 3 |  |
| `space-rocks-webrtc-channel-recovery` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 15821.2 ms | 7836 B | 2 |  |
| `space-rocks-webrtc-channel-recovery` | lexicon-plus-vector | true | 100.0% | 0.500 | 100.0% | 25490.3 ms | 54324 B | 3 |  |
| `space-rocks-client-process-shutdown` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 22830.4 ms | 61206 B | 2 |  |
| `space-rocks-client-process-shutdown` | lexicon-plus-vector | true | 100.0% | 0.167 | 100.0% | 18548.1 ms | 40156 B | 3 |  |
| `space-rocks-account-result-persistence` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 12208.1 ms | 56886 B | 2 |  |
| `space-rocks-account-result-persistence` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 14004.1 ms | 48762 B | 3 |  |
| `space-rocks-observability-contract-fanout` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 12927.2 ms | 13823 B | 2 |  |
| `space-rocks-observability-contract-fanout` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 15916.5 ms | 32338 B | 3 |  |
| `space-rocks-canonical-event-admission` | lexicon-seeds | false | 0.0% | 0.000 | 0.0% | 14829.8 ms | 42603 B | 2 |  |
| `space-rocks-canonical-event-admission` | lexicon-plus-vector | false | 0.0% | 0.000 | 0.0% | 16883.6 ms | 35015 B | 3 |  |

## Per-case seed rankings

### `space-rocks-room-player-activation` / `lexicon-seeds`

Where does a connected lobby roster become live authoritative avatars while preserving per-connection identity and rolling back partial binds?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `back` | `client/addons/gut/printers.gd` | false | false |
| 2 | lexicon | `connected` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 3 | lexicon | `Identity` | `services/game-server/internal/authclient/types.go` | false | false |
| 4 | lexicon | `lobby` | `client/scripts/lobby` | false | false |

### `space-rocks-room-player-activation` / `lexicon-plus-vector`

Where does a connected lobby roster become live authoritative avatars while preserving per-connection identity and rolling back partial binds?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `handleReturnToLobbyRequest` | `services/game-server/internal/networking/room_handlers.go` | false | false |
| 2 | vector | `_connect_lobby_signals` | `client/scripts/lobby/multiplayer_lobby_presenter.gd` | false | false |
| 3 | vector | `ResetToLobby` | `services/game-server/internal/rooms/room_lifecycle.go` | false | false |
| 4 | vector | `apply_lobby_state` | `client/scripts/ui/lobby/multiplayer_lobby.gd` | false | false |
| 5 | vector | `SetReadyInLobby` | `services/game-server/internal/rooms/room_lobby.go` | false | false |
| 6 | vector | `resetToLobbyLocked` | `services/game-server/internal/rooms/room_lifecycle.go` | false | false |

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
| 1 | vector | `UpdateLane` | `services/game-server/internal/protocol/realtime/baseline.go` | false | false |
| 2 | vector | `AdvanceMetadataForSuccessfulWrite` | `services/game-server/internal/protocol/realtime/baseline.go` | false | true |
| 3 | vector | `CommitSuccessfulCandidate` | `services/game-server/internal/protocol/realtime/baseline.go` | false | false |
| 4 | vector | `_ensure_realtime_transport_session` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 5 | vector | `_start_transport` | `client/scripts/networking/webrtc/realtime_transport_session.gd` | false | false |
| 6 | vector | `_is_stale_sequence` | `client/scripts/protocol/realtime/baseline_tracker.gd` | false | false |

### `space-rocks-client-inbound-binding` / `lexicon-seeds`

Where are decoded control signals and unreliable gameplay families connected to their application and presentation consumers without making the network facade own packet application?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `decode` | `client/scripts/networking/packets/packet_codec.gd` | false | false |
| 2 | lexicon | `connected` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 3 | lexicon | `presentation` | `client/scripts/gameplay/presentation` | false | false |

### `space-rocks-client-inbound-binding` / `lexicon-plus-vector`

Where are decoded control signals and unreliable gameplay families connected to their application and presentation consumers without making the network facade own packet application?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `handle_gameplay_packet` | `client/scripts/protocol/realtime/presentation_bridge.gd` | false | false |
| 2 | vector | `connect_gameplay_signals` | `client/scripts/session/session_network_controller.gd` | false | false |
| 3 | vector | `PacketFamily` | `services/game-server/internal/protocol/realtime/candidate_types.go` | false | false |
| 4 | vector | `HandleGameplayPacket` | `services/game-server/internal/networking/inbound/gameplay.go` | false | false |
| 5 | vector | `decode` | `client/scripts/networking/packets/packet_codec.gd` | false | false |
| 6 | vector | `PacketFamily` | `services/game-server/internal/protocol/realtime/payload_world.go` | false | false |

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
| 2 | vector | `GameplayComposition` | `client/scripts/gameplay/gameplay_composition.gd` | false | false |
| 3 | vector | `begin_accepting_gameplay_packets` | `client/scripts/session/gameplay_session_controller.gd` | false | false |
| 4 | vector | `GameplaySessionController` | `client/scripts/session/gameplay_session_controller.gd` | false | false |
| 5 | vector | `PresentationBridge` | `client/scripts/protocol/realtime/presentation_bridge.gd` | false | false |
| 6 | vector | `configure` | `client/scripts/session/gameplay_session_controller.gd` | false | false |

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
| 1 | vector | `_on_recovery_failed` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 2 | vector | `_replace_transport_for_recovery` | `client/scripts/networking/webrtc/realtime_transport_session.gd` | true | false |
| 3 | vector | `_fail_recovery` | `client/scripts/networking/webrtc/realtime_transport_session.gd` | false | true |
| 4 | vector | `_on_recovery_succeeded` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 5 | vector | `_on_recovery_started` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 6 | vector | `close` | `client/scripts/networking/webrtc/webrtc_transport.gd` | false | false |

### `space-rocks-client-process-shutdown` / `lexicon-seeds`

Where does a window-close request gracefully mark the server connection, stop a hosted local process, and then terminate the scene tree without pretending this is a room leave?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `stop` | `client/scripts/boot/local_server_process.gd` | false | false |
| 2 | lexicon | `Stop` | `services/game-server/internal/game/game.go` | false | false |
| 3 | lexicon | `Close` | `services/diagnostic-aggregator/hosted/service.go` | false | false |
| 4 | lexicon | `Close` | `services/game-server/internal/networking/tooling/router.go` | false | false |
| 5 | lexicon | `process` | `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd` | false | false |
| 6 | lexicon | `process` | `client/scripts/gameplay/presentation/local_player_presentation_controller.gd` | false | false |

### `space-rocks-client-process-shutdown` / `lexicon-plus-vector`

Where does a window-close request gracefully mark the server connection, stop a hosted local process, and then terminate the scene tree without pretending this is a room leave?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `stop` | `client/scripts/boot/local_server_process.gd` | false | false |
| 2 | vector | `close_gracefully` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 3 | vector | `send_leave_room_request` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 4 | vector | `leaveRoom` | `services/game-server/internal/networking/websocket_room_exit.go` | false | false |
| 5 | vector | `begin_graceful_close` | `client/scripts/networking/client_connection_service.gd` | false | false |
| 6 | vector | `request_shutdown` | `client/scripts/session/app_shutdown_controller.gd` | true | false |

### `space-rocks-account-result-persistence` / `lexicon-seeds`

How does a completed round's per-player outcome cross the in-process persistence boundary, route online accounts over HTTP, and update cumulative Rails statistics exactly once?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `process` | `client/scripts/devtools/context/devtools_overlay_context.gd` | false | false |
| 2 | lexicon | `process` | `client/scripts/devtools/context/devtools_hotkey_context.gd` | false | false |
| 3 | lexicon | `process` | `client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd` | false | false |
| 4 | lexicon | `process` | `client/scripts/devtools/context/devtools_command_context.gd` | false | false |
| 5 | lexicon | `process` | `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd` | false | false |
| 6 | lexicon | `process` | `client/scripts/gameplay/presentation/local_player_presentation_controller.gd` | false | false |

### `space-rocks-account-result-persistence` / `lexicon-plus-vector`

How does a completed round's per-player outcome cross the in-process persistence boundary, route online accounts over HTTP, and update cumulative Rails statistics exactly once?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `LoadStats` | `services/player-data/playerdata/rails_store.go` | false | false |
| 2 | vector | `player_stat` | `services/api-server/app/services/player_stats/apply_match_result.rb` | false | false |
| 3 | vector | `NewRailsStore` | `services/player-data/playerdata/rails_store.go` | false | false |
| 4 | vector | `player_stat=` | `services/api-server/app/services/player_stats/apply_match_result.rb` | false | false |
| 5 | vector | `show` | `services/api-server/app/controllers/api/player/stats_controller.rb` | false | false |
| 6 | vector | `create` | `services/api-server/app/controllers/internal/player_data/match_results_controller.rb` | false | false |

### `space-rocks-observability-contract-fanout` / `lexicon-seeds`

Which source-of-truth fanout validates one telemetry contract and plans deterministic outputs for the Go services, Godot client, Rails API, JSON fixtures, and reference docs?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 2 | lexicon | `Client` | `services/diagnostic-aggregator/internal/diagnosticclient/client.go` | false | false |
| 3 | lexicon | `json` | `services/api-server/app/lib/observability/emitter.rb` | false | false |
| 4 | lexicon | `Rails` | `services/api-server/config/initializers/filter_parameter_logging.rb` | false | false |
| 5 | lexicon | `json` | `services/api-server/app/services/auth/providers/discord_current_user.rb` | false | false |
| 6 | lexicon | `json` | `services/api-server/app/services/auth/providers/discord_token_exchange.rb` | false | false |

### `space-rocks-observability-contract-fanout` / `lexicon-plus-vector`

Which source-of-truth fanout validates one telemetry contract and plans deterministic outputs for the Go services, Godot client, Rails API, JSON fixtures, and reference docs?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `event_definition` | `services/api-server/app/lib/observability/contract_generated.rb` | false | false |
| 2 | vector | `field_definition` | `services/api-server/app/lib/observability/contract_generated.rb` | false | false |
| 3 | vector | `service_definition` | `services/api-server/app/lib/observability/contract_generated.rb` | false | false |
| 4 | vector | `generate_observability_go` | `tools/data_sync/data_sync/generators/observability_go.py` | false | true |
| 5 | vector | `known_event?` | `services/api-server/app/lib/observability/contract_generated.rb` | false | false |
| 6 | vector | `generate_observability_gds` | `tools/data_sync/data_sync/generators/observability_gds.py` | false | true |

### `space-rocks-canonical-event-admission` / `lexicon-seeds`

Where do the Go service layer, Godot client, and Rails worker each enforce the same event admission, redaction, size, identity, serialization, and degraded-write rules before accepting a canonical record?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | lexicon | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 2 | lexicon | `size` | `client/addons/gut/one_to_many.gd` | false | false |
| 3 | lexicon | `record` | `client/scripts/devtools/measurement/timing_summary.gd` | false | false |
| 4 | lexicon | `record` | `client/scripts/devtools/measurement/lifecycle_counter_window.gd` | false | false |
| 5 | lexicon | `write` | `client/scripts/devtools/measurement/measurement_report_writer.gd` | false | false |
| 6 | lexicon | `Identity` | `services/game-server/internal/authclient/types.go` | false | false |

### `space-rocks-canonical-event-admission` / `lexicon-plus-vector`

Where do the Go service layer, Godot client, and Rails worker each enforce the same event admission, redaction, size, identity, serialization, and degraded-write rules before accepting a canonical record?

| Rank | Source | Node | Path | Required | Supporting |
| ---: | --- | --- | --- | ---: | ---: |
| 1 | vector | `WriteRecord` | `shared/go/servicelog/logger.go` | false | false |
| 2 | vector | `WriteRecord` | `shared/go/observabilityevent/types.go` | false | false |
| 3 | vector | `ObservabilityContractGenerated` | `client/scripts/generated/observability/contract_generated.gd` | false | false |
| 4 | vector | `client` | `services/player-data/playerdata/rails_store.go` | false | false |
| 5 | vector | `event_definition` | `services/api-server/app/lib/observability/contract_generated.rb` | false | false |
| 6 | vector | `BuildWorldDeltaPacket` | `services/game-server/internal/protocol/realtime/delta.go` | false | false |

