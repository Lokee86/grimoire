package repostate

import "path/filepath"

func reusablePreparedFingerprint(location paths, repository RepositoryStatus, quickFingerprint string) string {
	for _, path := range preparedMarkerPaths(location) {
		var marker stateMarker
		if err := readJSON(path, &marker); err != nil || marker.SourceFingerprint == "" {
			continue
		}
		if repository.GitAvailable {
			if repository.GitDirty || repository.GitHead == "" || marker.GitHead != repository.GitHead {
				continue
			}
			return marker.SourceFingerprint
		}
		if quickFingerprint != "" && marker.QuickFingerprint == quickFingerprint {
			return marker.SourceFingerprint
		}
	}
	return ""
}

func preparedSourceChanged(location paths, repository RepositoryStatus, quickFingerprint, sourceFingerprint string) bool {
	for _, path := range preparedMarkerPaths(location) {
		var marker stateMarker
		if err := readJSON(path, &marker); err != nil {
			continue
		}
		if repository.GitAvailable {
			return repository.GitDirty || marker.GitHead != repository.GitHead
		}
		if marker.QuickFingerprint != "" && quickFingerprint != "" {
			return marker.QuickFingerprint != quickFingerprint
		}
		if marker.SourceFingerprint != "" {
			return marker.SourceFingerprint != sourceFingerprint
		}
	}
	return false
}

func preparedMarkersTrackRepository(location paths, repository RepositoryStatus, quickFingerprint string) bool {
	for _, path := range preparedMarkerPaths(location) {
		var marker stateMarker
		if err := readJSON(path, &marker); err != nil {
			return false
		}
		if repository.GitAvailable {
			if repository.GitDirty || repository.GitHead == "" || marker.GitHead != repository.GitHead {
				return false
			}
			continue
		}
		if quickFingerprint == "" || marker.QuickFingerprint != quickFingerprint {
			return false
		}
	}
	return true
}

func preparedMarkerPaths(location paths) []string {
	return []string{
		filepath.Join(location.grimoire, ".repostate.json"),
		filepath.Join(location.lexicon, ".repostate.json"),
	}
}
