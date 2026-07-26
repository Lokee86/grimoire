package scan

import (
	"testing"

	"github.com/Lokee86/lexicon/internal/interstack"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func TestInterstackDrifted(t *testing.T) {
	tests := []struct {
		name     string
		manifest objectstore.Manifest
		want     bool
	}{
		{name: "empty", manifest: objectstore.Manifest{}, want: false},
		{
			name: "orphan derived library",
			manifest: objectstore.Manifest{Languages: []objectstore.LanguageEntry{{
				Language: interstack.Language, AdapterVersion: interstack.AdapterVersion,
				AdapterFingerprint: interstack.AdapterFingerprint(), SchemaVersion: 1,
			}}},
			want: true,
		},
		{
			name:     "ordinary language without derived library",
			manifest: objectstore.Manifest{Languages: []objectstore.LanguageEntry{{Language: "go"}}},
			want:     true,
		},
		{
			name: "current derived library",
			manifest: objectstore.Manifest{Languages: []objectstore.LanguageEntry{
				{Language: "go"},
				{Language: interstack.Language, AdapterVersion: interstack.AdapterVersion,
					AdapterFingerprint: interstack.AdapterFingerprint(), SchemaVersion: 1},
			}},
			want: false,
		},
		{
			name: "old derived library",
			manifest: objectstore.Manifest{Languages: []objectstore.LanguageEntry{
				{Language: "go"},
				{Language: interstack.Language, AdapterVersion: "0.0.1", SchemaVersion: 1},
			}},
			want: true,
		},
		{
			name: "wrong derived fingerprint",
			manifest: objectstore.Manifest{Languages: []objectstore.LanguageEntry{
				{Language: "go"},
				{Language: interstack.Language, AdapterVersion: interstack.AdapterVersion,
					AdapterFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SchemaVersion: 1},
			}},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := interstackDrifted(test.manifest); got != test.want {
				t.Fatalf("interstackDrifted() = %v, want %v", got, test.want)
			}
		})
	}
}
