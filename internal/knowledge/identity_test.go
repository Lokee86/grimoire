package knowledge

import "testing"

func TestIdentityIgnoresRepositoryMetadata(t *testing.T) {
	base := Index{
		Version: FormatVersion,
		Root:    `C:\first\repo`,
		Documents: []Document{{
			Path: "docs/design.md", Kind: KindArchitecture, Hash: "document-hash",
			Sections: []Section{{ID: "section-1", Hash: "section-hash"}},
		}},
	}
	changed := base
	changed.Root = `D:\moved\repo`
	changed.GitCommit = "another-commit"
	if Identity(base) != Identity(changed) {
		t.Fatal("repository metadata changed knowledge identity")
	}
}

func TestIdentityChangesWithDocumentationAndCodeLinks(t *testing.T) {
	base := Index{
		Version: FormatVersion,
		Documents: []Document{{
			Path: "docs/design.md", Kind: KindArchitecture, Hash: "document-hash",
			Sections: []Section{{ID: "section-1", Hash: "section-hash"}},
		}},
	}
	changedDocument := base
	changedDocument.Documents = append([]Document(nil), base.Documents...)
	changedDocument.Documents[0].Sections = append([]Section(nil), base.Documents[0].Sections...)
	changedDocument.Documents[0].Sections[0].Hash = "changed-section"
	if Identity(base) == Identity(changedDocument) {
		t.Fatal("documentation change did not change knowledge identity")
	}

	changedLinks := base
	changedLinks.Documents = append([]Document(nil), base.Documents...)
	changedLinks.Documents[0].Sections = append([]Section(nil), base.Documents[0].Sections...)
	changedLinks.Documents[0].Sections[0].CodeLinks = []CodeLink{{Kind: "symbol", Value: "Resolve", SourcePath: "internal/resolve.go", Evidence: "exact"}}
	if Identity(base) == Identity(changedLinks) {
		t.Fatal("code-link change did not change knowledge identity")
	}
}
