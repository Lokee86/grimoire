package interstack

import "testing"

func TestResolveLinksGrimoireLexiconAndArcanaBoundaries(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "internal/arcanagraph/protocol.go", `package arcanagraph
import "os/exec"
func runProtocol(ctx Context, command string) {
	exec.CommandContext(ctx, command, "protocol", "--snapshot", snapshot)
}
`)
	writeFixture(t, root, "internal/lexiconfacts/state.go", `package lexiconfacts
import "os/exec"
func ResolveExport(ctx Context, command string) {
	args := []string{"export", "--snapshot", snapshot}
	exec.CommandContext(ctx, command, args...)
}
`)
	writeFixture(t, root, "lexicon/internal/cli/cli.go", `package cli
func Run(arguments []string) {
	switch arguments[0] {
	case "scan":
		runScan()
	case "export":
		runExport()
	}
}
`)
	writeFixture(t, root, "lexicon/internal/objectstore/store.go", `package objectstore
func (s Store) Publish() { writeAtomic(join(s.Root, "CURRENT")) }
func (s Store) Current() { readFile(join(s.Root, "CURRENT")) }
`)
	writeFixture(t, root, "arcana/src/main.rs", `fn main() {
    match parse() {
        Ok(cli::Command::Sync(command)) => run_sync(command),
        Ok(cli::Command::Protocol(command)) => run_protocol(command),
    }
}
`)
	writeFixture(t, root, "arcana/src/cli_protocol.rs", `pub fn run_protocol(command: &ProtocolCommand) {
    serve_jsonl(command);
}
`)
	writeFixture(t, root, "arcana/src/cli_sync.rs", `pub fn run_sync(command: &SyncCommand) { publish_current(command); }
fn publish_current(state: &Path) { replace_file(state.join("CURRENT")); }
fn read_current(state: &Path) { read_to_string(state.join("CURRENT")); }
`)
	writeFixture(t, root, "arcana/src/lexicon/snapshot.rs", `pub fn current(root: &Path) { read(root.join("CURRENT")); }
pub fn load(root: &Path) { read(root.join("snapshots")); }
`)

	libraries := []Library{
		{Language: "go", Repository: "warlock", Nodes: []Node{
			fileNode(testID('a'), "internal/arcanagraph/protocol.go"),
			callableNode(testID('b'), "function", "runProtocol", "internal/arcanagraph/protocol.go", 3),
			fileNode(testID('c'), "internal/lexiconfacts/state.go"),
			callableNode(testID('d'), "function", "ResolveExport", "internal/lexiconfacts/state.go", 3),
			fileNode(testID('e'), "lexicon/internal/cli/cli.go"),
			callableNode(testID('f'), "function", "Run", "lexicon/internal/cli/cli.go", 2),
			fileNode(testID('1'), "lexicon/internal/objectstore/store.go"),
			callableNode(testID('2'), "method", "Publish", "lexicon/internal/objectstore/store.go", 2),
			callableNode(testID('3'), "method", "Current", "lexicon/internal/objectstore/store.go", 3),
		}},
		{Language: "rust", Repository: "warlock", Nodes: []Node{
			fileNode(testID('4'), "arcana/src/main.rs"),
			callableNode(testID('5'), "function", "main", "arcana/src/main.rs", 1),
			fileNode(testID('6'), "arcana/src/cli_protocol.rs"),
			callableNode(testID('7'), "function", "run_protocol", "arcana/src/cli_protocol.rs", 1),
			fileNode(testID('8'), "arcana/src/cli_sync.rs"),
			callableNode(testID('9'), "function", "run_sync", "arcana/src/cli_sync.rs", 1),
			callableNode(testID('g'), "function", "publish_current", "arcana/src/cli_sync.rs", 2),
			callableNode(testID('h'), "function", "read_current", "arcana/src/cli_sync.rs", 3),
			fileNode(testID('i'), "arcana/src/lexicon/snapshot.rs"),
			callableNode(testID('j'), "function", "current", "arcana/src/lexicon/snapshot.rs", 1),
			callableNode(testID('k'), "function", "load", "arcana/src/lexicon/snapshot.rs", 2),
		}},
	}

	result, err := Resolve(root, libraries)
	if err != nil {
		t.Fatal(err)
	}
	arcanaProcess := nodeIDByKindAndName(t, result.Nodes, "process", "arcana")
	lexiconProcess := nodeIDByKindAndName(t, result.Nodes, "process", "lexicon")
	protocol := nodeIDByKindAndName(t, result.Nodes, "protocol", "arcana.query.v1")
	arcanaProtocol := nodeIDByKindAndName(t, result.Nodes, "cli-command", "arcana protocol")
	lexiconExport := nodeIDByKindAndName(t, result.Nodes, "cli-command", "lexicon export")
	lexiconCurrent := nodeIDByKindAndName(t, result.Nodes, "state-path", ".lexicon/CURRENT")
	arcanaCurrent := nodeIDByKindAndName(t, result.Nodes, "state-path", ".arcana/CURRENT")

	assertEdge(t, result.Edges, testID('b'), arcanaProcess, "invokes-process")
	assertEdge(t, result.Edges, testID('b'), arcanaProtocol, "calls")
	assertEdge(t, result.Edges, testID('d'), lexiconProcess, "invokes-process")
	assertEdge(t, result.Edges, testID('d'), lexiconExport, "calls")
	assertEdge(t, result.Edges, testID('b'), protocol, "produces-message")
	assertEdge(t, result.Edges, protocol, testID('7'), "consumes-message")
	assertEdge(t, result.Edges, testID('2'), lexiconCurrent, "writes")
	assertEdge(t, result.Edges, testID('d'), lexiconCurrent, "reads")
	assertEdge(t, result.Edges, testID('9'), lexiconCurrent, "reads")
	assertEdge(t, result.Edges, testID('g'), arcanaCurrent, "writes")
	assertEdge(t, result.Edges, testID('h'), arcanaCurrent, "reads")
}

func TestResolveRejectsParserTokensAsMessageChannels(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "arcana/src/parser.rs", `pub const PARAMETERS: &str = "parameters";
pub const TYPE_PARAMETERS: &str = "type-parameters";
pub const LANGUAGE: &str = "rust";
fn parse(token: &str) {
    match token {
        PARAMETERS => parse_parameters(),
        TYPE_PARAMETERS => parse_type_parameters(),
        LANGUAGE => parse_language(),
    }
}
`)
	libraries := []Library{{Language: "rust", Repository: "arcana", Nodes: []Node{
		fileNode(testID('a'), "arcana/src/parser.rs"),
		callableNode(testID('b'), "function", "parse", "arcana/src/parser.rs", 4),
	}}}
	result, err := Resolve(root, libraries)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range result.Nodes {
		if node.Kind == "message-channel" {
			t.Fatalf("parser token misclassified as message channel: %+v", node)
		}
	}
}
