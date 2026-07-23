package queryshape

import "testing"

func TestAutomaticBudgetsFollowScope(t *testing.T) {
	tests := []struct {
		scope           Scope
		target, maximum int
	}{
		{ScopeFocused, FocusedTargetTokens, FocusedMaximumTokens},
		{ScopeBounded, BoundedTargetTokens, BoundedMaximumTokens},
		{ScopeExploratory, ExploratoryTargetTokens, ExploratoryMaximumTokens},
	}
	for _, test := range tests {
		policy := RetrievalPolicy{Scope: test.scope, Shadow: true, BudgetMode: "automatic-shadow"}
		applyAutomaticBudget(&policy)
		if policy.TargetTokens != test.target || policy.MaximumTokens != test.maximum {
			t.Fatalf("scope %s: unexpected budget policy: %+v", test.scope, policy)
		}
		active := Activate(policy)
		if active.Shadow || active.BudgetMode != "automatic" {
			t.Fatalf("scope %s: policy did not activate: %+v", test.scope, active)
		}
	}
}
