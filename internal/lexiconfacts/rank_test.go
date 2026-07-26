package lexiconfacts

import "testing"

func TestScoreNodePrefersImplementationOverCommonParameter(t *testing.T) {
	query := "player movement realtime snapshot lifecycle"
	terms := queryTerms(query)
	parameterScore, _ := scoreNode(Node{
		Kind: "parameter", Name: "player",
		Path:          "client/legacy/player_render/player_sync_lifecycle.gd",
		QualifiedName: "client/legacy/player_render/player_sync_lifecycle.gd::function::configure::player",
	}, query, terms)
	functionScore, reasons := scoreNode(Node{
		Kind: "function", Name: "apply_realtime_player_snapshot",
		Path:          "client/scripts/networking/realtime/apply_realtime_player_snapshot.gd",
		QualifiedName: "apply_realtime_player_snapshot",
	}, query, terms)
	if parameterScore != 0 {
		t.Fatalf("common legacy parameter score = %v, want 0", parameterScore)
	}
	if functionScore <= 0 || len(reasons) == 0 {
		t.Fatalf("implementation score/reasons = %v/%#v", functionScore, reasons)
	}
}

func TestRankNodesDoesNotSeedUnrelatedInterstackNodes(t *testing.T) {
	facts := library{nodes: map[string]Node{
		"sender":   {ID: "sender", Kind: "method", Name: "SubmitLogin", Path: "client/login.gd", QualifiedName: "client/login.gd::SubmitLogin"},
		"endpoint": {ID: "endpoint", Kind: "http-endpoint", Name: "POST /session/start", Path: "@interstack/http/session", QualifiedName: "http:server:POST /session/start"},
		"handler":  {ID: "handler", Kind: "method", Name: "create", Path: "server/sessions_controller.rb", QualifiedName: "SessionsController#create"},
	}}
	ranked := rankNodes(facts, "Trace SubmitLogin", queryTerms("Trace SubmitLogin"))
	if len(ranked) != 1 || ranked[0].node.ID != "sender" {
		t.Fatalf("ranked nodes = %#v", ranked)
	}
}

func TestScoreNodeRetainsExactParameterLookup(t *testing.T) {
	score, _ := scoreNode(Node{
		Kind: "parameter", Name: "player", Path: "client/scripts/player.gd",
	}, "player", queryTerms("player"))
	if score <= 0 {
		t.Fatalf("exact parameter lookup score = %v", score)
	}
}
