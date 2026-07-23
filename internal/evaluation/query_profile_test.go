package evaluation

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/queryshape"
)

func TestScoreQueryProfileMatchesExpectation(t *testing.T) {
	entry := Case{ExpectedQueryProfile: &QueryProfileExpectation{
		Scope: queryshape.ScopeFocused, Specificity: queryshape.LevelHigh,
		Breadth: queryshape.LevelLow, Ambiguity: queryshape.LevelLow,
	}}
	run := CaseRun{
		QueryProfile: queryshape.Profile{
			Specificity: queryshape.LevelHigh, Breadth: queryshape.LevelLow, Ambiguity: queryshape.LevelLow,
		},
		RetrievalPolicy: queryshape.RetrievalPolicy{Scope: queryshape.ScopeFocused},
	}
	ScoreQueryProfile(entry, &run)
	if !run.QueryProfileMatched || len(run.QueryProfileMismatches) != 0 {
		t.Fatalf("unexpected profile score: %+v", run)
	}
}

func TestScoreQueryProfileReportsEveryMismatch(t *testing.T) {
	entry := Case{ExpectedQueryProfile: &QueryProfileExpectation{
		Scope: queryshape.ScopeFocused, Specificity: queryshape.LevelHigh,
		Breadth: queryshape.LevelLow, Ambiguity: queryshape.LevelLow,
	}}
	run := CaseRun{
		QueryProfile: queryshape.Profile{
			Specificity: queryshape.LevelLow, Breadth: queryshape.LevelHigh, Ambiguity: queryshape.LevelHigh,
		},
		RetrievalPolicy: queryshape.RetrievalPolicy{Scope: queryshape.ScopeExploratory},
	}
	ScoreQueryProfile(entry, &run)
	if run.QueryProfileMatched || len(run.QueryProfileMismatches) != 4 {
		t.Fatalf("unexpected profile score: %+v", run)
	}
}
