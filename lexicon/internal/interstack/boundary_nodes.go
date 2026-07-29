package interstack

import (
	"path/filepath"
	"strings"
	"unicode"
)

func (r *resolver) detectArcanaProtocol(file sourceFile) {
	path := strings.ToLower(filepath.ToSlash(file.Path))
	producer := strings.HasSuffix(path, "internal/arcanagraph/protocol.go")
	consumer := strings.HasSuffix(path, "arcana/src/cli_protocol.rs")
	if !producer && !consumer {
		return
	}
	protocol := r.addProtocol("arcana.query.v1")
	process := r.addProcess("arcana")
	command := r.addCLICommand("arcana", "protocol")
	if producer {
		if owner, ok := r.index.callableByName("runProtocol", file.Path); ok {
			r.addEdge(factEdge{Source: owner.ID, Target: protocol.ID, Relation: "produces-message"})
			r.addEdge(factEdge{Source: owner.ID, Target: command.ID, Relation: "calls"})
			r.result.Summary.ProtocolLinks++
		}
	} else if owner, ok := r.index.callableByName("run_protocol", file.Path); ok {
		r.addEdge(factEdge{Source: protocol.ID, Target: owner.ID, Relation: "consumes-message"})
		r.addEdge(factEdge{Source: owner.ID, Target: command.ID, Relation: "defines"})
		r.result.Summary.ProtocolLinks++
	}
	r.addEdge(factEdge{Source: process.ID, Target: command.ID, Relation: "contains"})
	r.addEdge(factEdge{Source: command.ID, Target: protocol.ID, Relation: "contains"})
}

func (r *resolver) addProcess(name string) Node {
	identity := "process\x00" + name
	id := stableNodeID("process", identity)
	if node, ok := r.nodes[id]; ok {
		return node
	}
	node := Node{
		ID: id, Kind: "process", Name: name,
		Path: syntheticPath("processes", identity), QualifiedName: "process:" + name,
	}
	r.addNode(node)
	r.result.Summary.Processes++
	return node
}

func (r *resolver) addCLICommand(process, command string) Node {
	identity := "cli-command\x00" + process + "\x00" + command
	id := stableNodeID("cli-command", identity)
	if node, ok := r.nodes[id]; ok {
		return node
	}
	name := process + " " + command
	node := Node{
		ID: id, Kind: "cli-command", Name: name,
		Path: syntheticPath("commands", identity), QualifiedName: "cli:" + name,
	}
	r.addNode(node)
	r.result.Summary.Commands++
	return node
}

func (r *resolver) addProtocol(name string) Node {
	identity := "protocol\x00" + name
	id := stableNodeID("protocol", identity)
	if node, ok := r.nodes[id]; ok {
		return node
	}
	node := Node{
		ID: id, Kind: "protocol", Name: name,
		Path: syntheticPath("protocols", identity), QualifiedName: "protocol:" + name,
	}
	r.addNode(node)
	r.result.Summary.Protocols++
	return node
}

func camelToKebab(value string) string {
	var output strings.Builder
	for index, character := range value {
		if unicode.IsUpper(character) && index > 0 {
			output.WriteByte('-')
		}
		output.WriteRune(unicode.ToLower(character))
	}
	return output.String()
}
