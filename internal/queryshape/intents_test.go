package queryshape

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
)

func TestPlanRetrievalIntentsSplitsCoordinatedRepositoryActions(t *testing.T) {
	query := "For a long repository task that names several files and asks for both implementation context and retrieval-quality evidence, follow how Grimoire plans every query window, searches semantic and lexical candidates, recovers exact path or symbol matches, merges and curates candidates, fits the final package budget, and exposes enough stage data for a mode-by-category baseline report."

	intents := PlanRetrievalIntents(query)
	if len(intents) != maxRetrievalIntentEntries {
		t.Fatalf("intent count = %d, want %d: %+v", len(intents), maxRetrievalIntentEntries, intents)
	}
	if intents[0].Intent != evidence.IntentMixed || intents[0].Weight >= intents[1].Weight {
		t.Fatalf("long mixed query should retain a low-weight context pass: %+v", intents)
	}
	for _, fragments := range [][]string{
		{"plans every query window", "searches semantic and lexical candidates"},
		{"recovers exact path or symbol matches"},
		{"merges and curates candidates"},
		{"fits the final package budget"},
		{"exposes enough stage data"},
	} {
		for _, fragment := range fragments {
			if !intentQueriesContain(intents[1:], fragment) {
				t.Errorf("missing decomposed action %q: %+v", fragment, intents)
			}
		}
	}
}

func TestPlanRetrievalIntentsIgnoresFencedExamplesAndPrioritizesImplementationSections(t *testing.T) {
	query := "## Goal\nDetermine whether context packages are useful.\n\n## Phase 1 — Define the evaluation contract\nAdd a repository-owned evaluation format and corpus model with required evidence.\n```json\n{\"query\": \"Where is vector snapshot freshness validated?\"}\n```\n\n## Phase 3 — Add the evaluation runner\nRun every case in lexical and semantic modes, record retrieval stages, and write a concise summary.\n\n## Phase 4 — Define scoring\nScore required evidence recall, failure stages, irrelevant selections, and aggregate metrics.\n\n## Phase 5 — Establish the baseline\nProduce per-case and mode-by-category reports with classified retrieval failures.\n\n## Phase 7 — Build the Space Rocks snapshot\nBuild a complete external vector snapshot."

	intents := PlanRetrievalIntents(query)
	if len(intents) < 5 || len(intents) > maxRetrievalIntentEntries {
		t.Fatalf("unexpected structured intent count %d: %+v", len(intents), intents)
	}
	joined := strings.ToLower(joinIntentQueries(intents[1:]))
	if strings.Contains(joined, "vector snapshot freshness") {
		t.Fatalf("fenced example leaked into retrieval clauses: %s", joined)
	}
	if strings.Contains(joined, "goal:") || strings.Contains(joined, "space rocks corpus") {
		t.Fatalf("generic or external section consumed a retrieval slot: %s", joined)
	}
	for _, required := range []string{
		"evaluation contract",
		"evaluation runner",
		"define scoring",
		"establish the baseline",
	} {
		if !strings.Contains(joined, required) {
			t.Errorf("missing implementation section %q: %s", required, joined)
		}
	}
}

func TestPlanRetrievalIntentsPreservesFocusedQuery(t *testing.T) {
	query := "Where is vector snapshot freshness validated?"
	intents := PlanRetrievalIntents(query)
	if len(intents) != 1 || intents[0].Intent != evidence.IntentDirectLocation || intents[0].Query != query || intents[0].Weight != 1 {
		t.Fatalf("focused query changed: %+v", intents)
	}
}

func intentQueriesContain(intents []RetrievalIntent, fragment string) bool {
	fragment = strings.ToLower(fragment)
	for _, intent := range intents {
		if strings.Contains(strings.ToLower(intent.Query), fragment) {
			return true
		}
	}
	return false
}

func joinIntentQueries(intents []RetrievalIntent) string {
	parts := make([]string, 0, len(intents))
	for _, intent := range intents {
		parts = append(parts, intent.Query)
	}
	return strings.Join(parts, "\n")
}
