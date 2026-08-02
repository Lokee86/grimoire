package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const language = "java"

type span struct {
	EndColumn   int    `json:"end_column"`
	EndLine     int    `json:"end_line"`
	Path        string `json:"path"`
	StartColumn int    `json:"start_column"`
	StartLine   int    `json:"start_line"`
}

type factSet struct {
	nodes      map[string]map[string]any
	edges      map[string]map[string]any
	unresolved map[string]map[string]any
}

func newFactSet() *factSet {
	return &factSet{
		nodes:      make(map[string]map[string]any),
		edges:      make(map[string]map[string]any),
		unresolved: make(map[string]map[string]any),
	}
}

func nodeID(kind, identity string) string {
	return digest([]byte("lexicon:v1\x00" + language + "\x00" + kind + "\x00" + identity))
}

func contentID(content []byte) string {
	return digest(content)
}

func digest(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func (facts *factSet) addNode(kind, name, path, qualifiedName, identity, owner string, sourceSpan *span, attributes map[string]any, content string) string {
	id := nodeID(kind, identity)
	record := map[string]any{
		"id": id, "kind": kind, "name": name, "path": path,
		"qualified_name": qualifiedName, "record": "node",
	}
	if owner != "" {
		record["owner"] = owner
	}
	if sourceSpan != nil {
		record["span"] = sourceSpan
	}
	if len(attributes) != 0 {
		record["attributes"] = attributes
	}
	if content != "" {
		record["content_id"] = content
	}
	if previous, exists := facts.nodes[id]; !exists || jsonText(record) < jsonText(previous) {
		facts.nodes[id] = record
	}
	return id
}

func (facts *factSet) addEdge(source, target, relation, owner string, sourceSpan *span, attributes map[string]any) {
	record := map[string]any{
		"record": "edge", "relation": relation, "source": source, "target": target,
	}
	if owner != "" {
		record["owner"] = owner
	}
	if sourceSpan != nil {
		record["span"] = sourceSpan
	}
	if len(attributes) != 0 {
		record["attributes"] = attributes
	}
	facts.edges[jsonText(record)] = record
}

func (facts *factSet) addUnresolved(source, relation, expression, reason, owner string, sourceSpan *span, attributes map[string]any) {
	record := map[string]any{
		"expression": expression, "reason": reason, "record": "unresolved",
		"relation": relation, "source": source,
	}
	if owner != "" {
		record["owner"] = owner
	}
	if sourceSpan != nil {
		record["span"] = sourceSpan
	}
	if len(attributes) != 0 {
		record["attributes"] = attributes
	}
	facts.unresolved[jsonText(record)] = record
}

func (facts *factSet) render(repository string) ([]byte, error) {
	nodes := sortedValues(facts.nodes, nodeKey)
	edges := sortedValues(facts.edges, edgeKey)
	unresolved := sortedValues(facts.unresolved, unresolvedKey)

	header := map[string]any{
		"adapter_version": adapterVersion,
		"language":        language,
		"mode":            "full",
		"record":          "lexicon",
		"repository":      repository,
		"schema_version":  1,
	}
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(header); err != nil {
		return nil, err
	}
	for _, group := range [][]map[string]any{nodes, edges, unresolved} {
		for _, record := range group {
			if err := encoder.Encode(record); err != nil {
				return nil, err
			}
		}
	}
	return output.Bytes(), nil
}

type keyedRecord struct {
	key    string
	record map[string]any
}

func sortedValues(records map[string]map[string]any, keyFor func(map[string]any) string) []map[string]any {
	keyed := make([]keyedRecord, 0, len(records))
	for _, record := range records {
		keyed = append(keyed, keyedRecord{key: keyFor(record), record: record})
	}
	sort.Slice(keyed, func(left, right int) bool { return keyed[left].key < keyed[right].key })
	result := make([]map[string]any, len(keyed))
	for index := range keyed {
		result[index] = keyed[index].record
	}
	return result
}

func jsonText(record map[string]any) string {
	data, _ := json.Marshal(record)
	return string(data)
}

func nodeKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record["id"], record["kind"], record["path"], record["qualified_name"], jsonText(record))
}

func edgeKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record["source"], record["target"], record["relation"], spanKey(record), jsonText(record))
}

func unresolvedKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s", record["source"], record["relation"], record["expression"], record["reason"], spanKey(record), jsonText(record))
}

func spanKey(record map[string]any) string {
	value, ok := record["span"].(*span)
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprintf("%s\x00%09d\x00%09d\x00%09d\x00%09d", value.Path, value.StartLine, value.StartColumn, value.EndLine, value.EndColumn)
}
