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

func TestScoreNodeRejectsIsolatedGenericNameInBroadQuery(t *testing.T) {
	query := "repository freshness source fingerprint refresh lock process tree"
	terms := queryTerms(query)
	genericScore, _ := scoreNode(Node{
		Kind:          "method",
		Name:          "source",
		Path:          "arcana/src/repository/compiler.rs",
		QualifiedName: "RepositoryCompileError::std::error::Error::source",
	}, query, terms)
	ownerScore, reasons := scoreNode(Node{
		Kind:          "function",
		Name:          "RepositoryFingerprint",
		Path:          "internal/repostate/fingerprint.go",
		QualifiedName: "repostate.RepositoryFingerprint",
	}, query, terms)
	if genericScore != 0 {
		t.Fatalf("isolated generic source method score = %v, want 0", genericScore)
	}
	if ownerScore <= 0 || len(reasons) == 0 {
		t.Fatalf("ownership symbol score/reasons = %v/%#v", ownerScore, reasons)
	}
}

func TestScoreNodeRetainsExactGenericNameLookup(t *testing.T) {
	score, _ := scoreNode(Node{
		Kind: "method", Name: "source", Path: "internal/errors.go",
	}, "source", queryTerms("source"))
	if score <= 0 {
		t.Fatalf("exact generic symbol lookup score = %v", score)
	}
}

func TestScoreNodeRetainsGenericNameWithStrongContext(t *testing.T) {
	query := "agent query response schema contract"
	score, reasons := scoreNode(Node{
		Kind:          "type",
		Name:          "Response",
		Path:          "internal/agentquery/schema.go",
		QualifiedName: "internal/agentquery/schema.go::Response",
	}, query, queryTerms(query))
	if score <= 0 || len(reasons) == 0 {
		t.Fatalf("contextual generic symbol score/reasons = %v/%#v", score, reasons)
	}
}

func TestRankNodesBroadQueryPrefersMultiTermOwnershipSymbol(t *testing.T) {
	query := "repository freshness source fingerprint refresh lock process tree"
	facts := library{nodes: map[string]Node{
		"repository": {
			ID: "repository", Kind: "module", Name: "repository", Path: "arcana/src/lib.rs",
		},
		"process": {
			ID: "process", Kind: "function", Name: "process", Path: "lexicon/adapters/rust/process.go",
		},
		"fingerprint": {
			ID: "fingerprint", Kind: "function", Name: "Fingerprint", Path: "lexicon/internal/registry.go",
		},
		"owner": {
			ID: "owner", Kind: "function", Name: "RepositoryFingerprint", Path: "internal/repostate/fingerprint.go",
		},
	}}
	ranked := rankNodes(facts, query, queryTerms(query))
	if len(ranked) == 0 || ranked[0].node.ID != "owner" {
		t.Fatalf("broad query ranking = %#v", ranked)
	}
}
