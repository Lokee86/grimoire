package interstack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLinksHTTPPacketsAndConfigAcrossLanguages(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "client/api.gd", `func auth_me_path():
	return "%s/api/auth/me" % API_BASE

func create_room_packet():
	var packet := {}
	packet[FIELD_TYPE] = "create_room_request"
	return packet

const API_URL_ENV := "API_URL"

func api_url():
	return OS.get_environment(API_URL_ENV)
`)
	writeFixture(t, root, "client/auth_api_client.gd", `func get_current_user():
	return await api_http_client.get_json(ApiConfigScript.auth_me_path(), token, trace_id)
`)
	writeFixture(t, root, "services/api-server/config/routes.rb", `Rails.application.routes.draw do
  namespace :api do
    namespace :auth do
      get "me", to: "me#show"
    end
  end
end
`)
	writeFixture(t, root, "services/api-server/app/controllers/api/auth/me_controller.rb", `module Api
  module Auth
    class MeController
      def show
      end
    end
  end
end
`)
	writeFixture(t, root, "services/game-server/internal/game/packets.go", `package game
const PacketTypeCreateRoomRequest = "create_room_request"
`)
	writeFixture(t, root, "services/game-server/internal/networking/inbound/lobby.go", `package inbound

func HandleLobby(packet Packet) {
	switch packet.Type {
	case game.PacketTypeCreateRoomRequest:
		createRoom()
	}
}
`)

	libraries := []Library{
		{
			Language: "gdscript", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('a'), "client/api.gd"),
				callableNode(testID('b'), "method", "auth_me_path", "client/api.gd", 1),
				callableNode(testID('c'), "method", "create_room_packet", "client/api.gd", 4),
				callableNode(testID('d'), "method", "api_url", "client/api.gd", 11),
				fileNode(testID('4'), "client/auth_api_client.gd"),
				callableNode(testID('5'), "method", "get_current_user", "client/auth_api_client.gd", 1),
			},
		},
		{
			Language: "ruby", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('e'), "services/api-server/config/routes.rb"),
				fileNode(testID('f'), "services/api-server/app/controllers/api/auth/me_controller.rb"),
				{
					ID: testID('1'), Kind: "method", Name: "show",
					Path:          "services/api-server/app/controllers/api/auth/me_controller.rb",
					QualifiedName: "Api::Auth::MeController#show",
					Span:          &Span{Path: "services/api-server/app/controllers/api/auth/me_controller.rb", StartLine: 4, StartColumn: 11, EndLine: 4, EndColumn: 15},
				},
			},
		},
		{
			Language: "go", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('2'), "services/game-server/internal/networking/inbound/lobby.go"),
				callableNode(testID('3'), "function", "HandleLobby", "services/game-server/internal/networking/inbound/lobby.go", 3),
			},
		},
	}

	result, err := Resolve(root, libraries)
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.HTTPContracts != 1 || result.Summary.HTTPLinks != 2 {
		t.Fatalf("unexpected HTTP summary: %+v", result.Summary)
	}
	if result.Summary.MessageChannels != 1 || result.Summary.MessageLinks != 2 {
		t.Fatalf("unexpected message summary: %+v", result.Summary)
	}
	if result.Summary.ConfigKeys != 1 {
		t.Fatalf("unexpected config summary: %+v", result.Summary)
	}

	httpID := nodeIDByKindAndName(t, result.Nodes, "http-endpoint", "GET /api/auth/me")
	messageID := nodeIDByKindAndName(t, result.Nodes, "message-channel", "create_room_request")
	configID := nodeIDByKindAndName(t, result.Nodes, "config-key", "API_URL")
	assertEdge(t, result.Edges, testID('b'), httpID, "calls-endpoint")
	assertEdge(t, result.Edges, testID('5'), httpID, "calls-endpoint")
	assertEdge(t, result.Edges, httpID, testID('1'), "handled-by")
	assertEdge(t, result.Edges, testID('c'), messageID, "publishes")
	assertEdge(t, result.Edges, messageID, testID('3'), "consumes")
	assertEdge(t, result.Edges, testID('d'), configID, "reads-config")
}

func TestResolveLinksGoServeMuxHandleThroughBoundHandlerProvider(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "client/api_config.gd", `func player_data_profile_path():
	return "%s/api/player-data/profile" % player_data_base_url()
`)
	writeFixture(t, root, "client/profile_api_client.gd", `func load_profile():
	return await api_http_client.post_json(ApiConfigScript.player_data_profile_path(), {})
`)
	writeFixture(t, root, "services/game-server/cmd/game-server/main.go", `package main
func runWithContext() {
	playerDataProfileHandler := newPlayerDataProfileHTTPHandler()
	mux.Handle("POST /api/player-data/profile", playerDataProfileHandler)
}
func newPlayerDataProfileHTTPHandler() http.Handler {
	return nil
}
`)

	libraries := []Library{
		{
			Language: "gdscript", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('a'), "client/api_config.gd"),
				callableNode(testID('b'), "method", "player_data_profile_path", "client/api_config.gd", 1),
				fileNode(testID('c'), "client/profile_api_client.gd"),
				callableNode(testID('d'), "method", "load_profile", "client/profile_api_client.gd", 1),
			},
		},
		{
			Language: "go", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('e'), "services/game-server/cmd/game-server/main.go"),
				callableNode(testID('f'), "function", "runWithContext", "services/game-server/cmd/game-server/main.go", 2),
				callableNode(testID('1'), "function", "newPlayerDataProfileHTTPHandler", "services/game-server/cmd/game-server/main.go", 6),
			},
		},
	}

	result, err := Resolve(root, libraries)
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.HTTPContracts != 1 || result.Summary.HTTPLinks != 2 {
		t.Fatalf("unexpected HTTP summary: %+v", result.Summary)
	}
	endpointID := nodeIDByKindAndName(t, result.Nodes, "http-endpoint", "POST /api/player-data/profile")
	assertEdge(t, result.Edges, testID('b'), endpointID, "calls-endpoint")
	assertEdge(t, result.Edges, testID('d'), endpointID, "calls-endpoint")
	assertEdge(t, result.Edges, endpointID, testID('1'), "handled-by")
}

func TestResolvePreservesNestedRailsNamespacesAcrossBlankLines(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "services/api-server/config/routes.rb", `Rails.application.routes.draw do
  namespace :api do
    namespace :auth do
      get "me", to: "me#show"
    end

    namespace :internal do
      namespace :player_data, path: "player-data" do
        post "stats", to: "stats#create"
      end
    end
  end

  namespace :internal do
    namespace :player_data, path: "player-data" do
      post "match-results", to: "match_results#create"
    end
  end
end
`)
	writeFixture(t, root, "services/player-data/playerdata/rails_store.go", `package playerdata
func loadStats() {
	request, _ := s.newJSONRequest(http.MethodPost, "/api/internal/player-data/stats", nil)
}
func recordMatchResult() {
	request, _ := s.newJSONRequest(http.MethodPost, "/internal/player-data/match-results", nil)
}
`)

	libraries := []Library{
		{
			Language: "ruby", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('a'), "services/api-server/config/routes.rb"),
				{
					ID: testID('b'), Kind: "method", Name: "create",
					Path:          "services/api-server/app/controllers/api/internal/player_data/stats_controller.rb",
					QualifiedName: "Api::Internal::PlayerData::StatsController#create",
					Span:          &Span{Path: "services/api-server/app/controllers/api/internal/player_data/stats_controller.rb", StartLine: 1, StartColumn: 1, EndLine: 1, EndColumn: 2},
				},
				{
					ID: testID('c'), Kind: "method", Name: "create",
					Path:          "services/api-server/app/controllers/internal/player_data/match_results_controller.rb",
					QualifiedName: "Internal::PlayerData::MatchResultsController#create",
					Span:          &Span{Path: "services/api-server/app/controllers/internal/player_data/match_results_controller.rb", StartLine: 1, StartColumn: 1, EndLine: 1, EndColumn: 2},
				},
			},
		},
		{
			Language: "go", Repository: "space-rocks",
			Nodes: []Node{
				fileNode(testID('d'), "services/player-data/playerdata/rails_store.go"),
				callableNode(testID('e'), "function", "loadStats", "services/player-data/playerdata/rails_store.go", 2),
				callableNode(testID('f'), "function", "recordMatchResult", "services/player-data/playerdata/rails_store.go", 5),
			},
		},
	}

	result, err := Resolve(root, libraries)
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.HTTPContracts != 3 || result.Summary.HTTPLinks != 2 {
		t.Fatalf("unexpected HTTP summary: %+v", result.Summary)
	}
	statsEndpointID := nodeIDByKindAndName(t, result.Nodes, "http-endpoint", "POST /api/internal/player-data/stats")
	matchEndpointID := nodeIDByKindAndName(t, result.Nodes, "http-endpoint", "POST /internal/player-data/match-results")
	assertEdge(t, result.Edges, testID('e'), statsEndpointID, "calls-endpoint")
	assertEdge(t, result.Edges, statsEndpointID, testID('b'), "handled-by")
	assertEdge(t, result.Edges, testID('f'), matchEndpointID, "calls-endpoint")
	assertEdge(t, result.Edges, matchEndpointID, testID('c'), "handled-by")
}

func TestEncodeIsDeterministicAndFactsV1Readable(t *testing.T) {
	result := Result{
		Repository: "example",
		Nodes: []Node{{
			ID: stableNodeID("config-key", "config\x00TOKEN"), Kind: "config-key",
			Name: "TOKEN", Path: "@interstack/config/token", QualifiedName: "config:TOKEN",
		}},
	}
	first, err := Encode(result)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Encode(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("interstack output is not deterministic")
	}
	if !strings.Contains(string(first), `"language":"interstack"`) {
		t.Fatalf("missing interstack header: %s", first)
	}
	library, err := ParseLibrary(first)
	if err != nil {
		t.Fatal(err)
	}
	if library.Language != Language || len(library.Nodes) != 1 {
		t.Fatalf("unexpected parsed library: %+v", library)
	}
}

func writeFixture(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testID(character byte) string {
	return "sha256:" + strings.Repeat(string(character), 64)
}

func fileNode(id, path string) Node {
	return Node{ID: id, Kind: "file", Name: filepath.Base(path), Path: path, QualifiedName: path}
}

func callableNode(id, kind, name, path string, line uint32) Node {
	return Node{
		ID: id, Kind: kind, Name: name, Path: path, QualifiedName: name,
		Span: &Span{Path: path, StartLine: line, StartColumn: 1, EndLine: line, EndColumn: 2},
	}
}

func nodeIDByKindAndName(t *testing.T, nodes []Node, kind, name string) string {
	t.Helper()
	for _, node := range nodes {
		if node.Kind == kind && node.Name == name {
			return node.ID
		}
	}
	t.Fatalf("missing %s node %q", kind, name)
	return ""
}

func assertEdge(t *testing.T, edges []factEdge, source, target, relation string) {
	t.Helper()
	for _, edge := range edges {
		if edge.Source == source && edge.Target == target && edge.Relation == relation {
			return
		}
	}
	t.Fatalf("missing %s edge %s -> %s", relation, source, target)
}
