package main

import "sort"

type visibilityIndex struct {
	includeDistance  map[string]map[string]int
	translationRoots map[string]map[string]struct{}
}

func buildVisibilityIndex(files []*sourceFile, index fileIndex) visibilityIndex {
	direct := make(map[string][]string, len(files))
	for _, file := range files {
		seen := map[string]struct{}{}
		for _, include := range file.Includes {
			target := localIncludeTarget(index, include)
			if target == nil {
				continue
			}
			if _, duplicate := seen[target.Path]; duplicate {
				continue
			}
			seen[target.Path] = struct{}{}
			direct[file.Path] = append(direct[file.Path], target.Path)
		}
		sort.Strings(direct[file.Path])
	}

	visibility := visibilityIndex{
		includeDistance:  make(map[string]map[string]int, len(files)),
		translationRoots: make(map[string]map[string]struct{}, len(files)),
	}
	for _, file := range files {
		visibility.includeDistance[file.Path] = reachableIncludeDistances(file.Path, direct)
	}
	for _, file := range files {
		if isHeaderPath(file.Path) {
			continue
		}
		members := visibility.includeDistance[file.Path]
		members[file.Path] = 0
		for member := range members {
			if visibility.translationRoots[member] == nil {
				visibility.translationRoots[member] = map[string]struct{}{}
			}
			visibility.translationRoots[member][file.Path] = struct{}{}
		}
	}
	return visibility
}

func reachableIncludeDistances(source string, direct map[string][]string) map[string]int {
	distances := map[string]int{}
	type queuedPath struct {
		path     string
		distance int
	}
	queue := []queuedPath{{path: source, distance: 0}}
	visited := map[string]struct{}{source: {}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, target := range direct[current.path] {
			distance := current.distance + 1
			if existing, ok := distances[target]; !ok || distance < existing {
				distances[target] = distance
			}
			if _, ok := visited[target]; ok {
				continue
			}
			visited[target] = struct{}{}
			queue = append(queue, queuedPath{path: target, distance: distance})
		}
	}
	return distances
}

func (index visibilityIndex) fileLocalVisible(sourcePath, declarationPath string) bool {
	if sourcePath == declarationPath {
		return true
	}
	sourceRoots := index.translationRoots[sourcePath]
	declarationRoots := index.translationRoots[declarationPath]
	for root := range sourceRoots {
		if _, ok := declarationRoots[root]; ok {
			return true
		}
	}
	return false
}

func (index visibilityIndex) includeRank(sourcePath, declarationPath string) (int, bool) {
	if sourcePath == declarationPath {
		return 0, true
	}
	distance, ok := index.includeDistance[sourcePath][declarationPath]
	return distance, ok
}
