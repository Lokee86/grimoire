package objectstore

import (
	"fmt"
	"sort"
)

type GCOptions struct {
	KeepSnapshots int
}

type GCPlan struct {
	CurrentSnapshot    string
	PreservedSnapshots []string
	DeleteSnapshots    []string
	PreservedObjects   []string
	DeleteObjects      []string
}

type GCResult struct {
	DryRun           bool
	DeletedSnapshots []string
	DeletedObjects   []string
}

func (s Store) PlanGC(options GCOptions) (GCPlan, error) {
	if options.KeepSnapshots < 0 {
		return GCPlan{}, fmt.Errorf("Lexicon GC snapshot retention cannot be negative: %d", options.KeepSnapshots)
	}
	currentID, current, err := s.Current()
	if err != nil {
		return GCPlan{}, err
	}
	snapshots, err := s.listSnapshots()
	if err != nil {
		return GCPlan{}, err
	}
	pins, err := s.readConsumerPins()
	if err != nil {
		return GCPlan{}, err
	}

	preserved := map[string]struct{}{currentID: {}}
	for index := 0; index < len(snapshots) && index < options.KeepSnapshots; index++ {
		preserved[snapshots[index].id] = struct{}{}
	}
	for _, pin := range pins {
		preserved[pin] = struct{}{}
	}

	preservedIDs := sortedGCIDs(preserved)
	manifests := map[string]Manifest{currentID: current}
	objects := make(map[string]struct{})
	for _, id := range preservedIDs {
		manifest, ok := manifests[id]
		if !ok {
			manifest, err = s.Load(id)
			if err != nil {
				return GCPlan{}, fmt.Errorf("load preserved Lexicon snapshot %s: %w", id, err)
			}
			manifests[id] = manifest
		}
		if err := addManifestObjects(objects, manifest); err != nil {
			return GCPlan{}, fmt.Errorf("snapshot %s: %w", id, err)
		}
	}

	plan := GCPlan{CurrentSnapshot: currentID}
	for _, id := range preservedIDs {
		plan.PreservedSnapshots = append(plan.PreservedSnapshots, id)
	}
	for _, snapshot := range snapshots {
		if _, ok := preserved[snapshot.id]; !ok {
			plan.DeleteSnapshots = append(plan.DeleteSnapshots, snapshot.id)
		}
	}
	for id := range objects {
		plan.PreservedObjects = append(plan.PreservedObjects, id)
	}
	allObjects, err := s.listObjects()
	if err != nil {
		return GCPlan{}, err
	}
	for _, id := range allObjects {
		if _, ok := objects[id]; !ok {
			plan.DeleteObjects = append(plan.DeleteObjects, id)
		}
	}
	sort.Strings(plan.PreservedSnapshots)
	sort.Strings(plan.DeleteSnapshots)
	sort.Strings(plan.PreservedObjects)
	sort.Strings(plan.DeleteObjects)
	return plan, nil
}

func (s Store) GarbageCollect(options GCOptions, dryRun bool) (GCResult, error) {
	plan, err := s.PlanGC(options)
	if err != nil {
		return GCResult{}, err
	}
	return s.ExecuteGC(plan, dryRun)
}
