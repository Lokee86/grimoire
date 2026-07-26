package knowledgevector

import (
	"testing"
	"time"

	"github.com/Lokee86/grimoire/internal/knowledge"
)

func TestIndexIdentityIgnoresRepositoryMetadata(t *testing.T) {
	now := time.Now().UTC()
	base := knowledge.Index{
		Version: knowledge.FormatVersion,
		Root:    `C:\first\repo`,
		Documents: []knowledge.Document{{
			Path: "docs/design.md", Kind: knowledge.KindArchitecture, Hash: "document-hash",
			Sections: []knowledge.Section{{ID: "section-1", Hash: "section-hash"}},
		}},
	}
	changedMetadata := base
	changedMetadata.Root = `D:\moved\repo`
	changedMetadata.GitCommit = "different-commit"
	changedMetadata.GitTime = &now
	changedMetadata.SourceFingerprint = "different-source-fingerprint"

	if IndexIdentity(base) != IndexIdentity(changedMetadata) {
		t.Fatal("repository metadata changed the documentation vector identity")
	}
}

func TestIndexIdentityChangesWithDocumentationContent(t *testing.T) {
	base := knowledge.Index{
		Version: knowledge.FormatVersion,
		Documents: []knowledge.Document{{
			Path: "docs/design.md", Kind: knowledge.KindArchitecture, Hash: "document-hash",
			Sections: []knowledge.Section{{ID: "section-1", Hash: "section-hash"}},
		}},
	}
	changed := base
	changed.Documents = append([]knowledge.Document(nil), base.Documents...)
	changed.Documents[0].Sections = append([]knowledge.Section(nil), base.Documents[0].Sections...)
	changed.Documents[0].Sections[0].Hash = "changed-section-hash"

	if IndexIdentity(base) == IndexIdentity(changed) {
		t.Fatal("documentation content change did not change the vector identity")
	}
}
