package library

import "fmt"

func Merge(fullPath, incrementalPath, destination string) error {
	fullHeader, fullRecords, err := readStream(fullPath)
	if err != nil {
		return fmt.Errorf("read full library: %w", err)
	}
	incrementalHeader, incrementalRecords, err := readStream(incrementalPath)
	if err != nil {
		return fmt.Errorf("read incremental library: %w", err)
	}
	changed, changedPresent := stringSet(incrementalHeader, "changed_files")
	removed, removedPresent := stringSet(incrementalHeader, "removed_files")
	if incrementalHeader["mode"] != "incremental" || !changedPresent || !removedPresent {
		return fmt.Errorf("adapter output is not a complete incremental stream")
	}
	sharedComplete, sharedPresent := incrementalHeader["shared_complete"].(bool)
	if !sharedPresent {
		return fmt.Errorf("incremental adapter output must declare shared_complete")
	}
	if fullHeader["language"] != incrementalHeader["language"] {
		return fmt.Errorf("incremental language does not match full library")
	}

	owners := nodeOwners(fullRecords)
	for id, owner := range nodeOwners(incrementalRecords) {
		owners[id] = owner
	}
	merged := make([]record, 0, len(fullRecords)+len(incrementalRecords))
	for _, item := range fullRecords {
		owner := recordOwner(item, owners)
		if owner == "" {
			if !sharedComplete {
				merged = append(merged, item)
			}
			continue
		}
		if !changed[owner] && !removed[owner] {
			merged = append(merged, item)
		}
	}
	for _, item := range incrementalRecords {
		owner := recordOwner(item, owners)
		if owner == "" {
			if sharedComplete {
				merged = append(merged, item)
			}
			continue
		}
		if !changed[owner] {
			return fmt.Errorf("incremental record is owned by undeclared file %q", owner)
		}
		if removed[owner] {
			return fmt.Errorf("incremental record is owned by removed file %q", owner)
		}
		merged = append(merged, item)
	}

	header := cloneRecord(incrementalHeader)
	header["mode"] = "full"
	delete(header, "changed_files")
	delete(header, "removed_files")
	delete(header, "shared_complete")
	sortRecords(merged)
	return writeStream(destination, header, merged)
}
