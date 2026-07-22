package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	adapterVersion = "0.1.0"
	language       = "gdscript"
)

type sourceSpan = map[string]any

type factSet struct {
	nodes             []map[string]any
	edges             []map[string]any
	unresolved        []map[string]any
	nodeByID          map[string]map[string]any
	moduleByPath      map[string]string
	classByName       map[string][]string
	methodByClassName map[string]map[string][]string
	fileByPath        map[string]string
}

func node(kind, name, path, qualified, id string, span sourceSpan, content string, attributes ...map[string]any) map[string]any {
	record := map[string]any{"id": id, "kind": kind, "name": name, "path": path, "qualified_name": qualified, "record": "node"}
	if content != "" {
		record["content_id"] = content
	}
	if span != nil {
		record["span"] = span
	}
	if len(attributes) > 0 && len(attributes[0]) > 0 {
		record["attributes"] = attributes[0]
	}
	return record
}

func edge(source, target, relation string, span sourceSpan) map[string]any {
	record := map[string]any{"record": "edge", "relation": relation, "source": source, "target": target}
	if span != nil {
		record["span"] = span
	}
	return record
}

func unresolved(source, relation, expression, reason string, span sourceSpan) map[string]any {
	record := map[string]any{"expression": expression, "reason": reason, "record": "unresolved", "relation": relation, "source": source}
	if span != nil {
		record["span"] = span
	}
	return record
}

func (f *factSet) addNode(record map[string]any) {
	id := record["id"].(string)
	if _, exists := f.nodeByID[id]; exists {
		return
	}
	f.nodeByID[id] = record
	f.nodes = append(f.nodes, record)
}

func (f *factSet) addEdge(record map[string]any) {
	for _, existing := range f.edges {
		if recordKey(existing) == recordKey(record) {
			return
		}
	}
	f.edges = append(f.edges, record)
	f.indexClassMethod(record)
}

func (f *factSet) indexClassMethod(record map[string]any) {
	if record["relation"] != "defines" {
		return
	}
	owner, ownerOK := f.nodeByID[record["source"].(string)]
	method, methodOK := f.nodeByID[record["target"].(string)]
	if !ownerOK || !methodOK || owner["kind"] != "type" || method["kind"] != "function" {
		return
	}
	className, _ := owner["name"].(string)
	if len(f.classByName[className]) != 1 {
		return
	}
	if f.methodByClassName[className] == nil {
		f.methodByClassName[className] = make(map[string][]string)
	}
	methodName, _ := method["name"].(string)
	f.methodByClassName[className][methodName] = append(f.methodByClassName[className][methodName], method["id"].(string))
}

func (f *factSet) addUnresolved(record map[string]any) {
	for _, existing := range f.unresolved {
		if recordKey(existing) == recordKey(record) {
			return
		}
	}
	f.unresolved = append(f.unresolved, record)
}

func recordKey(record map[string]any) string {
	data, _ := json.Marshal(record)
	return string(data)
}

func nodeID(kind, canonical string) string {
	return digest("lexicon:v1\x00" + language + "\x00" + kind + "\x00" + canonical)
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(hash[:])
}

func contentID(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func spanString(span map[string]any, key string) string {
	value, _ := span[key].(string)
	return value
}

func spanInt(span map[string]any, key string) int {
	value, _ := span[key].(int)
	return value
}
