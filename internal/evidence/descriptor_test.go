package evidence

import "testing"

func TestRangeIdentityNormalizesPath(t *testing.T) {
	got := RangeIdentity(`internal\\app\\context.go`, 10, 20)
	want := "range:internal/app/context.go:10:20"
	if got != want {
		t.Fatalf("RangeIdentity() = %q, want %q", got, want)
	}
}

func TestStableIDIsDeterministicAndOrdered(t *testing.T) {
	first := StableID("chain", "caller", "callee")
	second := StableID("chain", "caller", "callee")
	reversed := StableID("chain", "callee", "caller")
	if first != second {
		t.Fatalf("StableID() was not deterministic: %q != %q", first, second)
	}
	if first == reversed {
		t.Fatalf("StableID() ignored ordered parts: %q", first)
	}
}

func TestMergeRetainsProviderContributions(t *testing.T) {
	left := Descriptor{
		Identity:           "range:a.go:1:5",
		Intents:            []Intent{IntentCallChain},
		Roles:              []Role{RolePrimary},
		GroupIDs:           []string{"chain:one"},
		ExactMatchStrength: 0.5,
		EstimatedTokens:    50,
		RedundancyKey:      "a.go:symbol",
		Links: []Link{{
			Identity: "range:b.go:1:5",
			Relation: "calls",
		}},
	}
	right := Descriptor{
		Identity:           "ignored-conflict",
		Intents:            []Intent{IntentCallChain, IntentMechanism},
		Roles:              []Role{RoleSupporting},
		GroupIDs:           []string{"chain:one", "mechanism:one"},
		ExactMatchStrength: 0.9,
		EstimatedTokens:    70,
		RedundancyKey:      "ignored-conflict",
		Links: []Link{
			{Identity: "range:b.go:1:5", Relation: "calls"},
			{Identity: "range:c.go:1:5", Relation: "persists", Required: true},
		},
	}

	got := Merge(left, right)
	if got.Identity != left.Identity {
		t.Fatalf("Merge() identity = %q, want %q", got.Identity, left.Identity)
	}
	if got.ExactMatchStrength != 0.9 {
		t.Fatalf("Merge() exact strength = %v, want 0.9", got.ExactMatchStrength)
	}
	if got.EstimatedTokens != 70 {
		t.Fatalf("Merge() estimated tokens = %d, want 70", got.EstimatedTokens)
	}
	if got.RedundancyKey != left.RedundancyKey {
		t.Fatalf("Merge() redundancy key = %q, want %q", got.RedundancyKey, left.RedundancyKey)
	}
	if len(got.Intents) != 2 || len(got.Roles) != 2 || len(got.GroupIDs) != 2 || len(got.Links) != 2 {
		t.Fatalf("Merge() failed to retain unique metadata: %+v", got)
	}
}
