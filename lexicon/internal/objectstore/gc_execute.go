package objectstore

import (
	"fmt"
	"os"
	"sort"
)

func (s Store) ExecuteGC(plan GCPlan, dryRun bool) (GCResult, error) {
	plan, err := canonicalGCPlan(plan)
	if err != nil {
		return GCResult{}, err
	}
	currentID, _, err := s.Current()
	if err != nil {
		return GCResult{}, err
	}
	if currentID != plan.CurrentSnapshot {
		return GCResult{}, fmt.Errorf("Lexicon CURRENT changed during GC: planned %s, found %s", plan.CurrentSnapshot, currentID)
	}

	result := GCResult{DryRun: dryRun}
	if dryRun {
		result.DeletedSnapshots = append([]string(nil), plan.DeleteSnapshots...)
		result.DeletedObjects = append([]string(nil), plan.DeleteObjects...)
		return result, nil
	}
	for _, id := range plan.DeleteSnapshots {
		if err := os.Remove(s.snapshotPath(id)); err != nil {
			return result, fmt.Errorf("delete Lexicon snapshot %s: %w", id, err)
		}
		result.DeletedSnapshots = append(result.DeletedSnapshots, id)
	}
	for _, id := range plan.DeleteObjects {
		if err := os.Remove(s.objectPath(id)); err != nil {
			return result, fmt.Errorf("delete Lexicon object %s: %w", id, err)
		}
		result.DeletedObjects = append(result.DeletedObjects, id)
	}
	return result, nil
}

func canonicalGCPlan(plan GCPlan) (GCPlan, error) {
	if err := validateGCPlan(plan); err != nil {
		return GCPlan{}, err
	}
	plan.PreservedSnapshots = sortedGCList(plan.PreservedSnapshots)
	plan.DeleteSnapshots = sortedGCList(plan.DeleteSnapshots)
	plan.PreservedObjects = sortedGCList(plan.PreservedObjects)
	plan.DeleteObjects = sortedGCList(plan.DeleteObjects)
	return plan, nil
}

func sortedGCList(ids []string) []string {
	result := append([]string(nil), ids...)
	sort.Strings(result)
	return result
}
