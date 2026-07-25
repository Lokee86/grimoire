package arcanagraph

import (
	"slices"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestSourceCandidatesLocalizesGraphEvidenceIntoPreparedChunks(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		preparedFile("handler.go", "handler", 1, 12),
		preparedFile("service.go", "service", 1, 20),
		preparedFile("repo.go", "repo", 1, 16),
		preparedFile("service_test.go", "test", 1, 14),
		preparedFile("broken.go", "broken", 1, 8),
	}}
	service := sourceNode("service", "Service.Run", "service.go", 3, 12)
	handler := sourceNode("handler", "Handler.Serve", "handler.go", 2, 9)
	repository := sourceNode("repository", "Repository.Save", "repo.go", 4, 11)
	testNode := sourceNode("test", "TestServiceRun", "service_test.go", 3, 10)
	serviceIdentity := evidence.RangeIdentity("service.go", 3, 12)
	facts := []structure.Evidence{
		{
			Provider: "arcana", Kind: "operational_role", Node: &service,
			Relationships: []structure.Relationship{
				{Direction: "incoming", Relation: "calls", Certainty: "definite", Node: handler},
				{Direction: "outgoing", Relation: "possible_calls", Certainty: "possible", Node: repository},
			},
			Context: &evidence.Descriptor{
				GroupIDs: []string{"role-group"},
				Links:    []evidence.Link{{Identity: serviceIdentity, Relation: "source"}},
			},
		},
		{
			Provider: "arcana", Kind: "impact", Node: &service,
			Dependents: []structure.DepthNode{{Depth: 2, Node: testNode}},
			Context:    &evidence.Descriptor{GroupIDs: []string{"impact-group"}},
		},
		{
			Provider: "arcana", Kind: "call_chain",
			Chain:   &structure.Path{Depth: 2, Nodes: []structure.Node{handler, service, repository}, Relations: []string{"calls", "calls"}},
			Context: &evidence.Descriptor{GroupIDs: []string{"chain-group"}},
		},
		{
			Provider: "arcana", Kind: "unresolved", Node: &service,
			Unresolved: []structure.Unresolved{{
				Relation: "calls", Expression: "missingTarget()",
				Span: &structure.Span{Path: "broken.go", StartLine: 4, EndLine: 4},
			}},
			Context: &evidence.Descriptor{GroupIDs: []string{"unresolved-group"}},
		},
	}

	candidates := SourceCandidates(snapshot, facts, 10)
	if len(candidates) != 5 {
		t.Fatalf("got %d candidates, want 5: %+v", len(candidates), candidates)
	}
	gotPaths := make([]string, len(candidates))
	for index := range candidates {
		gotPaths[index] = candidates[index].Chunk.Path
		if candidates[index].Source != candidateSource || candidates[index].Rank != index+1 {
			t.Fatalf("candidate %d lost Arcana rank/source: %+v", index, candidates[index])
		}
	}
	wantPaths := []string{"service.go", "handler.go", "repo.go", "service_test.go", "broken.go"}
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("paths = %v, want %v", gotPaths, wantPaths)
	}

	handlerCandidate := candidates[1]
	if !slices.Contains(handlerCandidate.Reasons, "Arcana incoming calls graph neighbor of Service.Run") ||
		!slices.Contains(handlerCandidate.Reasons, "Arcana call-chain node 1 of 3") {
		t.Fatalf("handler graph reasons were not merged: %+v", handlerCandidate.Reasons)
	}
	if handlerCandidate.Context == nil ||
		!slices.Contains(handlerCandidate.Context.Intents, evidence.IntentMechanism) ||
		!slices.Contains(handlerCandidate.Context.Intents, evidence.IntentCallChain) ||
		!slices.Contains(handlerCandidate.Context.GroupIDs, "role-group") ||
		!slices.Contains(handlerCandidate.Context.GroupIDs, "chain-group") {
		t.Fatalf("handler context lost graph provenance: %+v", handlerCandidate.Context)
	}
	if !slices.Contains(handlerCandidate.Context.Links, evidence.Link{Identity: serviceIdentity, Relation: "source"}) {
		t.Fatalf("handler source was not linked to graph subject: %+v", handlerCandidate.Context.Links)
	}

	serviceCandidate := candidates[0]
	if serviceCandidate.Context == nil || slices.Contains(serviceCandidate.Context.Links, evidence.Link{Identity: serviceIdentity, Relation: "source"}) {
		t.Fatalf("service candidate retained a self-link: %+v", serviceCandidate.Context)
	}
}

func TestSourceCandidatesUsesFirstPreparedChunkWhenGraphNodeHasOnlyPath(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "service.go",
		Chunks: []index.Chunk{
			{ID: "first", Path: "service.go", StartLine: 1, EndLine: 10, TokenCount: 10},
			{ID: "second", Path: "service.go", StartLine: 11, EndLine: 20, TokenCount: 10},
		},
	}}}
	node := structure.Node{Name: "Service", Path: "service.go"}
	facts := []structure.Evidence{{Provider: "arcana", Kind: "operational_role", Node: &node}}

	candidates := SourceCandidates(snapshot, facts, 4)
	if len(candidates) != 1 || candidates[0].Chunk.ID != "first" {
		t.Fatalf("path-only graph node candidates = %+v, want first prepared chunk", candidates)
	}
}

func TestSourceCandidatesKeepsImpactOwnerWithoutOperationalRoleEvidence(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		preparedFile("service.go", "service", 1, 12),
		preparedFile("caller.go", "caller", 1, 12),
	}}
	service := sourceNode("service", "Service.Run", "service.go", 2, 8)
	caller := sourceNode("caller", "Caller.Run", "caller.go", 2, 8)
	facts := []structure.Evidence{{
		Provider: "arcana", Kind: "impact", Node: &service,
		Dependents: []structure.DepthNode{{Depth: 1, Node: caller}},
	}}

	candidates := SourceCandidates(snapshot, facts, 4)
	if len(candidates) != 2 || candidates[0].Chunk.Path != "service.go" || candidates[1].Chunk.Path != "caller.go" {
		t.Fatalf("impact-only candidates = %+v", candidates)
	}
	if !slices.Contains(candidates[0].Reasons, "Arcana impact subject Service.Run") {
		t.Fatalf("impact owner reason missing: %+v", candidates[0].Reasons)
	}
}

func TestSourceCandidatesUsesNearestChunkForStaleSpan(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "service.go",
		Chunks: []index.Chunk{
			{ID: "first", Path: "service.go", StartLine: 1, EndLine: 10},
			{ID: "second", Path: "service.go", StartLine: 21, EndLine: 30},
		},
	}}}
	node := sourceNode("service", "Service.Run", "service.go", 18, 18)
	facts := []structure.Evidence{{Provider: "arcana", Kind: "operational_role", Node: &node}}

	candidates := SourceCandidates(snapshot, facts, 4)
	if len(candidates) != 1 || candidates[0].Chunk.ID != "second" {
		t.Fatalf("stale-span candidates = %+v, want nearest prepared chunk", candidates)
	}
}

func TestSourceCandidatesHonorsLimitAfterRanking(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		preparedFile("one.go", "one", 1, 8),
		preparedFile("two.go", "two", 1, 8),
		preparedFile("three.go", "three", 1, 8),
	}}
	subject := sourceNode("subject", "Subject", "one.go", 1, 4)
	facts := []structure.Evidence{{
		Provider: "arcana", Kind: "operational_role", Node: &subject,
		Relationships: []structure.Relationship{
			{Direction: "incoming", Relation: "calls", Node: sourceNode("two", "Two", "two.go", 1, 4)},
			{Direction: "outgoing", Relation: "calls", Node: sourceNode("three", "Three", "three.go", 1, 4)},
		},
	}}

	candidates := SourceCandidates(snapshot, facts, 2)
	if len(candidates) != 2 || candidates[0].Chunk.Path != "one.go" || candidates[1].Chunk.Path != "two.go" {
		t.Fatalf("limited candidates = %+v", candidates)
	}
}

func preparedFile(path, id string, start, end int) index.FileRecord {
	return index.FileRecord{
		Path: path,
		Chunks: []index.Chunk{{
			ID: id, Path: path, StartLine: start, EndLine: end,
			TokenCount: end - start + 1, Text: "source for " + path,
		}},
	}
}

func sourceNode(identity, name, path string, start, end int) structure.Node {
	return structure.Node{
		Identity: identity, Name: name, Path: path,
		Span: &structure.Span{Path: path, StartLine: start, EndLine: end},
	}
}
