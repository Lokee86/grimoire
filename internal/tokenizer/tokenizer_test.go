package tokenizer

import "testing"

func TestCountUsesO200kBase(t *testing.T) {
	if Name != "o200k_base" {
		t.Fatalf("unexpected tokenizer name %q", Name)
	}

	count, err := Count("Hello, world!")
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("expected 4 tokens, got %d", count)
	}
}
