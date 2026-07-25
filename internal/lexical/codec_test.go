package lexical

import (
	"reflect"
	"testing"
)

func TestCodecRoundTripPreservesPostingState(t *testing.T) {
	original := Build([]Input{
		{Key: "a", Path: "alpha.go", Text: "func ResolveDamage() {}"},
		{Key: "b", Path: "beta.go", Text: "func CompileSnapshot() {}"},
	})
	data, err := Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Documents(), original.Documents()) {
		t.Fatalf("documents changed across codec round trip:\n got: %+v\nwant: %+v", decoded.Documents(), original.Documents())
	}
	postings := decoded.Posting("damage")
	if len(postings) != 1 || postings[0].Document != 0 || postings[0].Frequency != 1 {
		t.Fatalf("unexpected damage postings: %+v", postings)
	}
	if documents := decoded.CandidateDocuments([]string{"snapshot"}); len(documents) != 1 || documents[0] != 1 {
		t.Fatalf("unexpected snapshot candidates: %v", documents)
	}
	if documents := decoded.CandidateDocumentsAll([]string{"resolve", "damage"}); len(documents) != 1 || documents[0] != 0 {
		t.Fatalf("unexpected intersected candidates: %v", documents)
	}
	if documents := decoded.CandidateDocumentsAll([]string{"resolve", "snapshot"}); len(documents) != 0 {
		t.Fatalf("unrelated terms unexpectedly intersected: %v", documents)
	}
}
