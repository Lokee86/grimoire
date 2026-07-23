package evaluation

import (
	"fmt"

	"github.com/Lokee86/grimoire/internal/queryshape"
)

func ScoreQueryProfile(entry Case, run *CaseRun) {
	run.ExpectedQueryProfile = entry.ExpectedQueryProfile
	if entry.ExpectedQueryProfile == nil {
		return
	}
	expected := *entry.ExpectedQueryProfile
	actual := run.QueryProfile
	policy := run.RetrievalPolicy
	if policy.Scope != expected.Scope {
		run.QueryProfileMismatches = append(run.QueryProfileMismatches,
			fmt.Sprintf("scope: expected %s, got %s", expected.Scope, policy.Scope))
	}
	compareProfileLevel(&run.QueryProfileMismatches, "specificity", expected.Specificity, actual.Specificity)
	compareProfileLevel(&run.QueryProfileMismatches, "breadth", expected.Breadth, actual.Breadth)
	compareProfileLevel(&run.QueryProfileMismatches, "ambiguity", expected.Ambiguity, actual.Ambiguity)
	run.QueryProfileMatched = len(run.QueryProfileMismatches) == 0
}

func compareProfileLevel(mismatches *[]string, label string, expected, actual queryshape.Level) {
	if actual == expected {
		return
	}
	*mismatches = append(*mismatches, fmt.Sprintf("%s: expected %s, got %s", label, expected, actual))
}
