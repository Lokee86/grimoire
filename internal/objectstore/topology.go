package objectstore

import (
	"encoding/json"
	"fmt"
)

func (s Store) DirectChangesRequireFull(language string, roots []string) (bool, error) {
	_, objects, nodeOwners, _, err := s.dependencyData(language)
	if err != nil {
		return true, err
	}
	for _, path := range roots {
		object, ok := objects[path]
		if !ok {
			return true, nil
		}
		for _, raw := range object.Records {
			var record dependencyRecord
			if err := json.Unmarshal(raw, &record); err != nil {
				return true, err
			}
			if record.Record == "unresolved" && repositorySensitiveUnresolved(record.Reason) {
				return true, nil
			}
			if record.Record == "edge" && semanticRelation(record.Relation) && nodeOwners[record.Target] != path {
				return true, nil
			}
		}
	}
	return false, nil
}

func repositorySensitiveUnresolved(reason string) bool {
	switch reason {
	case "missing-target", "ambiguous-target", "generated-target":
		return true
	default:
		return false
	}
}

func semanticRelation(relation string) bool {
	switch relation {
	case "contains", "defines":
		return false
	default:
		return true
	}
}

type topologyRecord struct {
	Record        string `json:"record"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Relation      string `json:"relation"`
	Expression    string `json:"expression"`
	Reason        string `json:"reason"`
	CandidateName string `json:"candidate_name"`
}

func (s Store) RequiresFullAnalysis(language string, changedFiles []string, incrementalPath string) (bool, error) {
	_, objects, previousOwners, _, err := s.dependencyData(language)
	if err != nil {
		return true, err
	}
	selected := make(map[string]struct{}, len(changedFiles))
	previous := make(map[string]map[string]struct{}, len(changedFiles))
	for _, path := range changedFiles {
		selected[path] = struct{}{}
		object, ok := objects[path]
		if !ok {
			return true, nil
		}
		previous[path] = relationKeys(object.Records)
	}
	_, records, err := parseOutput(incrementalPath)
	if err != nil {
		return true, err
	}
	owners := previousOwners
	for id, owner := range nodeOwners(records) {
		owners[id] = owner
	}
	for _, record := range records {
		key, relationship, err := relationKey(record.raw)
		if err != nil {
			return true, err
		}
		if !relationship {
			continue
		}
		owner := recordOwner(record.value, owners)
		if _, ok := selected[owner]; !ok {
			continue
		}
		if _, existed := previous[owner][key]; !existed {
			return true, nil
		}
	}
	return false, nil
}

func relationKeys(records []json.RawMessage) map[string]struct{} {
	result := make(map[string]struct{})
	for _, raw := range records {
		key, relationship, err := relationKey(raw)
		if err == nil && relationship {
			result[key] = struct{}{}
		}
	}
	return result
}

func relationKey(raw json.RawMessage) (string, bool, error) {
	var record topologyRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return "", false, fmt.Errorf("decode relationship topology: %w", err)
	}
	var values []string
	switch record.Record {
	case "edge":
		values = []string{"edge", record.Source, record.Target, record.Relation}
	case "unresolved":
		values = []string{"unresolved", record.Source, record.Relation, record.Expression, record.Reason, record.CandidateName}
	default:
		return "", false, nil
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "", false, err
	}
	return string(encoded), true, nil
}
