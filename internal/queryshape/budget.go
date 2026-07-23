package queryshape

const (
	FocusedTargetTokens      = 3000
	FocusedMaximumTokens     = 4000
	BoundedTargetTokens      = 6000
	BoundedMaximumTokens     = 8000
	ExploratoryTargetTokens  = 12000
	ExploratoryMaximumTokens = 16000
)

// Activate marks a shadow recommendation as authoritative for a context request.
func Activate(policy RetrievalPolicy) RetrievalPolicy {
	policy.Shadow = false
	if policy.BudgetMode == "automatic-shadow" {
		policy.BudgetMode = "automatic"
	}
	return policy
}

func applyAutomaticBudget(policy *RetrievalPolicy) {
	switch policy.Scope {
	case ScopeFocused:
		policy.TargetTokens = FocusedTargetTokens
		policy.MaximumTokens = FocusedMaximumTokens
	case ScopeExploratory:
		policy.TargetTokens = ExploratoryTargetTokens
		policy.MaximumTokens = ExploratoryMaximumTokens
	default:
		policy.TargetTokens = BoundedTargetTokens
		policy.MaximumTokens = BoundedMaximumTokens
	}
}
