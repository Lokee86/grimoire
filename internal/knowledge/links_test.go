package knowledge

import (
	"fmt"
	"testing"
)

func TestLinksForUsesIndexedExactReferences(t *testing.T) {
	catalog := &codeCatalog{
		paths: map[string]struct{}{
			"services/game-server/main.go": {},
		},
		values: map[string][]catalogEntry{
			"SubmitLogin":     {{kind: "symbol", value: "SubmitLogin", path: "client/login.gd"}},
			"/session/start":  {{kind: "endpoint", value: "/session/start", path: "api/routes.rb"}},
			"session.timeout": {{kind: "config-contract", value: "session.timeout", path: "config/session.toml"}},
		},
	}
	links := catalog.linksFor("SubmitLoginHandler is unrelated. SubmitLogin calls POST /session/start?source=test. See services/game-server/main.go. Configure session.timeout.")
	if len(links) != 4 {
		t.Fatalf("links = %#v", links)
	}
	found := make(map[string]bool)
	for _, link := range links {
		found[link.Kind+":"+link.Value+":"+link.SourcePath] = true
	}
	for _, expected := range []string{
		"path:services/game-server/main.go:services/game-server/main.go",
		"symbol:SubmitLogin:client/login.gd",
		"endpoint:/session/start:api/routes.rb",
		"config-contract:session.timeout:config/session.toml",
	} {
		if !found[expected] {
			t.Fatalf("missing %s in %#v", expected, links)
		}
	}
}

func TestLinksForDropsAmbiguousContractsAndCapsResults(t *testing.T) {
	catalog := &codeCatalog{paths: make(map[string]struct{}), values: make(map[string][]catalogEntry)}
	for index := 0; index < 20; index++ {
		catalog.values["world"] = append(catalog.values["world"], catalogEntry{
			kind: "contract", value: "world", path: fmt.Sprintf("internal/world%02d.go", index),
		})
	}
	for index := 0; index < 60; index++ {
		value := fmt.Sprintf("Unique%02d", index)
		catalog.values[value] = []catalogEntry{{kind: "symbol", value: value, path: fmt.Sprintf("internal/unique%02d.go", index)}}
	}
	text := "world"
	for index := 0; index < 60; index++ {
		text += " " + fmt.Sprintf("Unique%02d", index)
	}
	links := catalog.linksFor(text)
	if len(links) != maxCodeLinks {
		t.Fatalf("link count = %d, want %d", len(links), maxCodeLinks)
	}
	for _, link := range links {
		if link.Value == "world" {
			t.Fatalf("ambiguous contract was retained: %#v", link)
		}
	}
}

func BenchmarkLinksForIndexedCatalog(b *testing.B) {
	catalog := &codeCatalog{paths: make(map[string]struct{}), values: make(map[string][]catalogEntry)}
	for index := 0; index < 50000; index++ {
		value := fmt.Sprintf("Symbol%05d", index)
		catalog.values[value] = []catalogEntry{{kind: "symbol", value: value, path: fmt.Sprintf("internal/file%05d.go", index)}}
	}
	text := "Symbol00010 Symbol01000 Symbol49999"
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = catalog.linksFor(text)
	}
}
