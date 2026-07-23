package objectstore

import (
	"encoding/json"
	"fmt"
	"sort"
)

type dependencyRecord struct {
	Record   string `json:"record"`
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
	Reason   string `json:"reason"`
}

// IncrementalScope loads the current language objects once, decides whether
// direct edits require full analysis, and computes the incremental emission
// and context closures from the same dependency view.
func (s Store) IncrementalScope(language string, roots []string) (bool, []string, []string, error) {
	_, objects, nodeOwners, unresolved, err := s.dependencyData(language)
	if err != nil {
		return true, nil, nil, err
	}
	rootSet := make(map[string]struct{}, len(roots))
	for _, path := range roots {
		rootSet[path] = struct{}{}
	}
	foundRoots := make(map[string]struct{}, len(rootSet))
	fullRequired := false
	reverse := make(map[string]map[string]struct{})
	forward := make(map[string]map[string]struct{})
	for owner, object := range objects {
		_, directRoot := rootSet[owner]
		if directRoot {
			foundRoots[owner] = struct{}{}
		}
		for _, raw := range object.Records {
			var record dependencyRecord
			if err := json.Unmarshal(raw, &record); err != nil {
				return true, nil, nil, err
			}
			if directRoot {
				if record.Record == "unresolved" && repositorySensitiveUnresolved(record.Reason) {
					fullRequired = true
				}
				if record.Record == "edge" && semanticRelation(record.Relation) && nodeOwners[record.Target] != owner {
					fullRequired = true
				}
			}
			if record.Record != "edge" || record.Target == "" {
				continue
			}
			targetOwner := nodeOwners[record.Target]
			if targetOwner == "" || targetOwner == owner {
				continue
			}
			addRelation(reverse, targetOwner, owner)
			addRelation(forward, owner, targetOwner)
		}
	}
	if len(foundRoots) != len(rootSet) {
		fullRequired = true
	}
	emitSeeds := append([]string(nil), roots...)
	for path := range unresolved {
		emitSeeds = append(emitSeeds, path)
	}
	emit := closure(emitSeeds, reverse)
	context := closure(emit, forward)
	return fullRequired, emit, context, nil
}

func (s Store) DependencyScope(language string, roots []string) ([]string, []string, error) {
	_, emit, context, err := s.IncrementalScope(language, roots)
	return emit, context, err
}

func (s Store) ImpactedFiles(language string, roots []string) ([]string, error) {
	emit, _, err := s.DependencyScope(language, roots)
	return emit, err
}

func (s Store) dependencyData(language string) (LanguageEntry, map[string]FactObject, map[string]string, map[string]struct{}, error) {
	_, manifest, err := s.Current()
	if err != nil {
		return LanguageEntry{}, nil, nil, nil, err
	}
	entry, ok := languageEntry(manifest, language)
	if !ok {
		return LanguageEntry{}, nil, nil, nil, fmt.Errorf("snapshot has no %s analysis", language)
	}
	objects := make(map[string]FactObject, len(entry.Files))
	nodeOwners := make(map[string]string)
	unresolved := make(map[string]struct{})
	for _, file := range entry.Files {
		object, err := s.LoadObject(file.ObjectID)
		if err != nil {
			return LanguageEntry{}, nil, nil, nil, err
		}
		objects[file.Path] = object
		for _, raw := range object.Records {
			var record dependencyRecord
			if err := json.Unmarshal(raw, &record); err != nil {
				return LanguageEntry{}, nil, nil, nil, fmt.Errorf("decode %s dependency record: %w", file.Path, err)
			}
			if record.Record == "node" && record.ID != "" {
				nodeOwners[record.ID] = file.Path
			}
			if record.Record == "unresolved" {
				unresolved[file.Path] = struct{}{}
			}
		}
	}
	return entry, objects, nodeOwners, unresolved, nil
}

func addRelation(graph map[string]map[string]struct{}, source, target string) {
	if graph[source] == nil {
		graph[source] = make(map[string]struct{})
	}
	graph[source][target] = struct{}{}
}

func closure(seeds []string, graph map[string]map[string]struct{}) []string {
	selected := make(map[string]struct{})
	queue := append([]string(nil), seeds...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == "" {
			continue
		}
		if _, exists := selected[current]; exists {
			continue
		}
		selected[current] = struct{}{}
		for next := range graph[current] {
			queue = append(queue, next)
		}
	}
	result := make([]string, 0, len(selected))
	for path := range selected {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func languageEntry(manifest Manifest, language string) (LanguageEntry, bool) {
	for _, entry := range manifest.Languages {
		if entry.Language == language {
			return entry, true
		}
	}
	return LanguageEntry{}, false
}
