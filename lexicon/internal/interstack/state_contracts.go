package interstack

import "strings"

type stateRole struct {
	Callable string
	Path     string
	Relation string
}

func (r *resolver) detectStateContracts(file sourceFile) {
	path := strings.ToLower(normalizeSourcePath(file.Path))
	var roles []stateRole
	switch {
	case strings.HasSuffix(path, "lexicon/internal/objectstore/store.go"):
		roles = []stateRole{
			{Callable: "Publish", Path: ".lexicon", Relation: "writes"},
			{Callable: "Publish", Path: ".lexicon/CURRENT", Relation: "writes"},
			{Callable: "Publish", Path: ".lexicon/snapshots", Relation: "writes"},
			{Callable: "Current", Path: ".lexicon/CURRENT", Relation: "reads"},
			{Callable: "Current", Path: ".lexicon/snapshots", Relation: "reads"},
		}
	case strings.HasSuffix(path, "internal/lexiconfacts/state.go"):
		roles = []stateRole{
			{Callable: "ResolveExport", Path: ".lexicon", Relation: "reads"},
			{Callable: "ResolveExport", Path: ".lexicon/CURRENT", Relation: "reads"},
			{Callable: "ResolveExport", Path: ".lexicon/snapshots", Relation: "reads"},
		}
	case strings.HasSuffix(path, "internal/arcanagraph/state.go"):
		roles = []stateRole{
			{Callable: "ResolveSnapshot", Path: ".lexicon/CURRENT", Relation: "reads"},
			{Callable: "ResolveSnapshot", Path: ".arcana", Relation: "reads"},
			{Callable: "ResolveSnapshot", Path: ".arcana/CURRENT", Relation: "reads"},
			{Callable: "ResolveSnapshot", Path: ".arcana/snapshots", Relation: "reads"},
		}
	case strings.HasSuffix(path, "arcana/src/lexicon/snapshot.rs"):
		roles = []stateRole{
			{Callable: "current", Path: ".lexicon/CURRENT", Relation: "reads"},
			{Callable: "load", Path: ".lexicon/snapshots", Relation: "reads"},
		}
	case strings.HasSuffix(path, "arcana/src/cli_sync.rs"):
		roles = []stateRole{
			{Callable: "run_sync", Path: ".lexicon", Relation: "reads"},
			{Callable: "run_sync", Path: ".lexicon/CURRENT", Relation: "reads"},
			{Callable: "run_sync", Path: ".lexicon/snapshots", Relation: "reads"},
			{Callable: "run_sync", Path: ".arcana", Relation: "writes"},
			{Callable: "run_sync", Path: ".arcana/snapshots", Relation: "writes"},
			{Callable: "publish_current", Path: ".arcana/CURRENT", Relation: "writes"},
			{Callable: "read_current", Path: ".arcana/CURRENT", Relation: "reads"},
		}
	default:
		return
	}
	for _, role := range roles {
		owner, ok := r.index.callableByName(role.Callable, file.Path)
		if !ok {
			continue
		}
		node := r.addStatePath(role.Path)
		r.addEdge(factEdge{
			Source: owner.ID, Target: node.ID, Relation: role.Relation,
			Attributes: map[string]any{"confidence": 1.0, "transport": "filesystem-state"},
		})
		r.result.Summary.StateLinks++
	}
}

func (r *resolver) addStatePath(path string) Node {
	identity := "state-path\x00" + path
	id := stableNodeID("state-path", identity)
	if node, ok := r.nodes[id]; ok {
		return node
	}
	node := Node{
		ID: id, Kind: "state-path", Name: path,
		Path: syntheticPath("state", identity), QualifiedName: "state:" + path,
		Attributes: map[string]any{"path": path, "transport": "filesystem-state"},
	}
	r.addNode(node)
	r.result.Summary.StatePaths++
	return node
}
