package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

type factSet struct {
	language   string
	nodes      []map[string]any
	edges      []map[string]any
	unresolved []map[string]any
}

func newFactSet(language string) *factSet {
	return &factSet{language: language}
}

func (facts *factSet) addFile(path string, content []byte) string {
	path = filepath.ToSlash(path)
	fileID := facts.nodeID("file", path)
	moduleID := facts.nodeID("module", path)
	owner := path
	facts.nodes = append(facts.nodes,
		map[string]any{
			"content_id": contentID(content), "id": fileID, "kind": "file", "name": filepath.Base(filepath.FromSlash(path)),
			"owner": owner, "path": path, "qualified_name": path, "record": "node",
		},
		map[string]any{
			"id": moduleID, "kind": "module", "name": strings.TrimSuffix(filepath.Base(filepath.FromSlash(path)), filepath.Ext(path)),
			"owner": owner, "path": path, "qualified_name": path, "record": "node",
		},
	)
	facts.edges = append(facts.edges, map[string]any{
		"owner": owner, "record": "edge", "relation": "contains", "source": fileID, "target": moduleID,
	})
	return moduleID
}

func (facts *factSet) addDeclaration(path, moduleID, kind, name string, lineNumber int, line string) {
	span := lineSpan(path, lineNumber, line)
	canonical := fmt.Sprintf("%s::%s:%s:%d", path, kind, name, lineNumber)
	id := facts.nodeID(kind, canonical)
	facts.nodes = append(facts.nodes, map[string]any{
		"id": id, "kind": kind, "name": name, "owner": path, "path": path,
		"qualified_name": path + "::" + name, "record": "node", "span": span,
	})
	facts.edges = append(facts.edges, map[string]any{
		"owner": path, "record": "edge", "relation": "defines", "source": moduleID, "span": span, "target": id,
	})
}

func (facts *factSet) addImport(path, moduleID, keyword, target string, lineNumber int, line string) {
	span := lineSpan(path, lineNumber, line)
	canonical := fmt.Sprintf("%s::import:%d:%s", path, lineNumber, target)
	id := facts.nodeID("import", canonical)
	facts.nodes = append(facts.nodes, map[string]any{
		"attributes": map[string]any{"expression": strings.TrimSpace(line), "keyword": keyword, "target": target},
		"id":         id, "kind": "import", "name": target, "owner": path, "path": path,
		"qualified_name": canonical, "record": "node", "span": span,
	})
	facts.edges = append(facts.edges, map[string]any{
		"owner": path, "record": "edge", "relation": "defines", "source": moduleID, "span": span, "target": id,
	})
	facts.unresolved = append(facts.unresolved, map[string]any{
		"expression": target, "owner": path, "reason": "external-target", "record": "unresolved",
		"relation": "imports", "source": id, "span": span,
	})
}

func (facts *factSet) render(repository string, changedFiles, removedFiles []string, incremental bool) []byte {
	sort.Slice(facts.nodes, func(left, right int) bool { return nodeSortKey(facts.nodes[left]) < nodeSortKey(facts.nodes[right]) })
	sort.Slice(facts.edges, func(left, right int) bool { return edgeSortKey(facts.edges[left]) < edgeSortKey(facts.edges[right]) })
	sort.Slice(facts.unresolved, func(left, right int) bool {
		return unresolvedSortKey(facts.unresolved[left]) < unresolvedSortKey(facts.unresolved[right])
	})
	header := map[string]any{
		"adapter_version": adapterVersion, "language": facts.language, "record": "lexicon",
		"repository": repository, "schema_version": 1,
	}
	if incremental {
		header["changed_files"] = sortedPaths(changedFiles)
		header["mode"] = "incremental"
		header["removed_files"] = sortedPaths(removedFiles)
		header["shared_complete"] = false
	}
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	_ = encoder.Encode(header)
	for _, records := range [][]map[string]any{facts.nodes, facts.edges, facts.unresolved} {
		for _, record := range records {
			_ = encoder.Encode(record)
		}
	}
	return output.Bytes()
}

func (facts *factSet) nodeID(kind, canonical string) string {
	return digest("lexicon:v1\x00" + facts.language + "\x00" + kind + "\x00" + canonical)
}

func contentID(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(hash[:])
}

func lineSpan(path string, lineNumber int, line string) map[string]any {
	return map[string]any{
		"end_column": utf8.RuneCountInString(line) + 1, "end_line": lineNumber,
		"path": filepath.ToSlash(path), "start_column": 1, "start_line": lineNumber,
	}
}

func sortedPaths(paths []string) []string {
	result := make([]string, len(paths))
	for index, path := range paths {
		result[index] = filepath.ToSlash(path)
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
	return fmt.Sprintf("%s\x00%08d\x00%08d", span["path"], span["start_line"], span["start_column"])
}
