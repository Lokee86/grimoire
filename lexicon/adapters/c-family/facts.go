package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
)

type factSet struct {
	repository   string
	incremental  bool
	changedFiles []string
	removedFiles []string
	changed      map[string]struct{}
	nodes        map[string]map[string]any
	edges        map[string]map[string]any
	unresolved   map[string]map[string]any
}

func newFactSet(repository string, changedFiles, removedFiles []string, incremental bool) *factSet {
	changed := make(map[string]struct{}, len(changedFiles))
	for _, path := range changedFiles {
		changed[filepath.ToSlash(path)] = struct{}{}
	}
	return &factSet{
		repository: repository, incremental: incremental,
		changedFiles: sortedPaths(changedFiles), removedFiles: sortedPaths(removedFiles), changed: changed,
		nodes: map[string]map[string]any{}, edges: map[string]map[string]any{}, unresolved: map[string]map[string]any{},
	}
}

func (facts *factSet) emits(owner string) bool {
	if !facts.incremental {
		return true
	}
	_, ok := facts.changed[filepath.ToSlash(owner)]
	return ok
}

func (facts *factSet) addNode(owner string, record map[string]any) {
	if !facts.emits(owner) {
		return
	}
	facts.nodes[record["id"].(string)] = record
}

func (facts *factSet) addEdge(owner string, record map[string]any) {
	if !facts.emits(owner) {
		return
	}
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["source"], record["target"], record["relation"], spanSortKey(record))
	facts.edges[key] = record
}

func (facts *factSet) addUnresolved(owner string, record map[string]any) {
	if !facts.emits(owner) {
		return
	}
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record["source"], record["relation"], record["expression"], record["reason"], spanSortKey(record))
	facts.unresolved[key] = record
}

func (facts *factSet) render() ([]byte, error) {
	header := map[string]any{
		"adapter_version": adapterVersion,
		"language":        streamLanguage,
		"mode":            "full",
		"record":          "lexicon",
		"repository":      facts.repository,
		"schema_version":  1,
	}
	if facts.incremental {
		header["changed_files"] = facts.changedFiles
		header["mode"] = "incremental"
		header["removed_files"] = facts.removedFiles
		header["shared_complete"] = false
	}

	nodes := mapValues(facts.nodes)
	edges := mapValues(facts.edges)
	unresolved := mapValues(facts.unresolved)
	sort.Slice(nodes, func(i, j int) bool { return nodeSortKey(nodes[i]) < nodeSortKey(nodes[j]) })
	sort.Slice(edges, func(i, j int) bool { return edgeSortKey(edges[i]) < edgeSortKey(edges[j]) })
	sort.Slice(unresolved, func(i, j int) bool { return unresolvedSortKey(unresolved[i]) < unresolvedSortKey(unresolved[j]) })

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	if err := encoder.Encode(header); err != nil {
		return nil, err
	}
	for _, records := range [][]map[string]any{nodes, edges, unresolved} {
		for _, record := range records {
			if err := encoder.Encode(record); err != nil {
				return nil, err
			}
		}
	}
	return output.Bytes(), nil
}

func mapValues(values map[string]map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func nodeID(kind, canonical string) string {
	return digest("lexicon:v1\x00" + streamLanguage + "\x00" + kind + "\x00" + canonical)
}

func contentID(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(hash[:])
}

func sortedPaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func nodeSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["id"], record["kind"], record["path"], record["qualified_name"])
}

func edgeSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["source"], record["target"], record["relation"], spanSortKey(record))
}

func unresolvedSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record["source"], record["relation"], record["expression"], record["reason"], spanSortKey(record))
}

func spanSortKey(record map[string]any) string {
	span, _ := record["span"].(map[string]any)
	return fmt.Sprintf("%s\x00%08v\x00%08v\x00%08v\x00%08v", span["path"], span["start_line"], span["start_column"], span["end_line"], span["end_column"])
}
