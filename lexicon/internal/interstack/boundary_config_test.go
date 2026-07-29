package interstack

import "testing"

func TestResolveLinksBoundaryEnvironmentConfiguration(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "internal/app/engine_specs.go", `package app
var lexiconEngine = engineSpec{commandEnv: "GRIMOIRE_LEXICON_COMMAND"}
var arcanaEngine = engineSpec{commandEnv: "GRIMOIRE_ARCANA_COMMAND"}
`)
	writeFixture(t, root, "lexicon/internal/config/config.go", `package config
import "os"
func StateRoot() string { return os.Getenv("LEXICON_STATE_DIR") }
`)
	libraries := []Library{{Language: "go", Repository: "warlock", Nodes: []Node{
		fileNode(testID('a'), "internal/app/engine_specs.go"),
		fileNode(testID('b'), "lexicon/internal/config/config.go"),
		callableNode(testID('c'), "function", "StateRoot", "lexicon/internal/config/config.go", 3),
	}}}
	result, err := Resolve(root, libraries)
	if err != nil {
		t.Fatal(err)
	}
	lexiconCommand := nodeIDByKindAndName(t, result.Nodes, "config-key", "GRIMOIRE_LEXICON_COMMAND")
	arcanaCommand := nodeIDByKindAndName(t, result.Nodes, "config-key", "GRIMOIRE_ARCANA_COMMAND")
	state := nodeIDByKindAndName(t, result.Nodes, "config-key", "LEXICON_STATE_DIR")
	assertEdge(t, result.Edges, testID('a'), lexiconCommand, "reads-config")
	assertEdge(t, result.Edges, testID('a'), arcanaCommand, "reads-config")
	assertEdge(t, result.Edges, testID('c'), state, "reads-config")
}
