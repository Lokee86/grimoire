package objectstore

import (
	"fmt"
	"sort"
)

func addManifestObjects(objects map[string]struct{}, manifest Manifest) error {
	for _, language := range manifest.Languages {
		if language.SharedObjectID != "" {
			if !validID(language.SharedObjectID) {
				return fmt.Errorf("invalid shared_object_id %q", language.SharedObjectID)
			}
			objects[language.SharedObjectID] = struct{}{}
		}
		for _, file := range language.Files {
			if !validID(file.ObjectID) {
				return fmt.Errorf("invalid object_id %q", file.ObjectID)
			}
			objects[file.ObjectID] = struct{}{}
		}
	}
	return nil
}

func validateGCPlan(plan GCPlan) error {
	if !validID(plan.CurrentSnapshot) {
		return fmt.Errorf("invalid Lexicon GC current snapshot %q", plan.CurrentSnapshot)
	}
	preservedSnapshots, err := validGCIDs("snapshot", plan.PreservedSnapshots)
	if err != nil {
		return err
	}
	if _, ok := preservedSnapshots[plan.CurrentSnapshot]; !ok {
		return fmt.Errorf("Lexicon GC plan does not preserve CURRENT snapshot %s", plan.CurrentSnapshot)
	}
	if err := rejectOverlap("snapshot", preservedSnapshots, plan.DeleteSnapshots); err != nil {
		return err
	}
	preservedObjects, err := validGCIDs("object", plan.PreservedObjects)
	if err != nil {
		return err
	}
	return rejectOverlap("object", preservedObjects, plan.DeleteObjects)
}

func validGCIDs(kind string, ids []string) (map[string]struct{}, error) {
	result := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if !validID(id) {
			return nil, fmt.Errorf("invalid Lexicon GC %s ID %q", kind, id)
		}
		if _, exists := result[id]; exists {
			return nil, fmt.Errorf("duplicate Lexicon GC %s ID %s", kind, id)
		}
		result[id] = struct{}{}
	}
	return result, nil
}

func rejectOverlap(kind string, preserved map[string]struct{}, deleted []string) error {
	deletedIDs, err := validGCIDs(kind, deleted)
	if err != nil {
		return err
	}
	for id := range deletedIDs {
		if _, ok := preserved[id]; ok {
			return fmt.Errorf("Lexicon GC plan deletes preserved %s %s", kind, id)
		}
	}
	return nil
}

func sortedGCIDs(ids map[string]struct{}) []string {
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}
