package objectstore

import (
	"fmt"
	"sort"
)

// Analysis is one validated adapter stream. JSONL is confined to this
// transport boundary; callers apply the parsed records directly to immutable
// snapshot objects.
type Analysis struct {
	Header  Header
	records []parsedRecord
}

func ReadAnalysis(path, language string) (*Analysis, error) {
	header, records, err := parseOutput(path)
	if err != nil {
		return nil, err
	}
	if err := validateHeader(header, language, path); err != nil {
		return nil, err
	}
	if header.Mode == "incremental" {
		if header.ChangedFiles == nil || header.RemovedFiles == nil {
			return nil, fmt.Errorf("incremental adapter output must declare changed_files and removed_files")
		}
		if header.SharedComplete == nil {
			return nil, fmt.Errorf("incremental adapter output must declare shared_complete")
		}
	}
	return &Analysis{Header: header, records: records}, nil
}

func (a *Analysis) IsIncremental() bool {
	return a != nil && a.Header.Mode == "incremental"
}

func (a *Analysis) groups(allowedOwners map[string]struct{}) (map[string]typedRecords, typedRecords) {
	owners := nodeOwners(a.records)
	groups := make(map[string]typedRecords)
	shared := typedRecords{}
	for _, record := range a.records {
		owner := recordOwner(record.value, owners)
		if owner == "" {
			shared.append(record.typed)
			continue
		}
		if allowedOwners != nil {
			if _, allowed := allowedOwners[owner]; !allowed {
				shared.append(record.typed)
				continue
			}
		}
		group := groups[owner]
		group.append(record.typed)
		groups[owner] = group
	}
	return groups, shared
}

func samePaths(left, right []string) bool {
	left = normalizedPaths(left)
	right = normalizedPaths(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func normalizedPaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		path = normalizeOwner(path)
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}
