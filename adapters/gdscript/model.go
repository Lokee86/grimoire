package main

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

const (
	adapterVersion = "0.2.0"
	language       = "gdscript"
)

type sourceSpan = map[string]any

type factSet struct {
	nodes                     []map[string]any
	edges                     []map[string]any
	unresolved                []map[string]any
	nodeByID                  map[string]map[string]any
	edgeKeys                  map[string]struct{}
	edgeOrderKeys             []string
	unresolvedKeys            map[string]struct{}
	unresolvedOrderKeys       []string
	moduleByPath              map[string]string
	classByName               map[string][]string
	classByFileAndName        map[string]map[string][]string
	methodByClassName         map[string]map[string][]string
	methodByClassID           map[string]map[string][]string
	preloadAliasByFileAndName map[string]map[string][]string
	staticMethodByModulePath  map[string]map[string][]string
	fileByPath                map[string]string
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
	if f.edgeKeys == nil {
		f.edgeKeys = make(map[string]struct{})
	}
	key := edgeSortKey(record)
	if _, exists := f.edgeKeys[key]; exists {
		return
	}
	f.edgeKeys[key] = struct{}{}
	f.edges = append(f.edges, record)
	f.edgeOrderKeys = append(f.edgeOrderKeys, key)
	f.indexClassMethod(record)
	f.indexStaticFunction(record)
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
	classID := owner["id"].(string)
	if f.methodByClassID == nil {
		f.methodByClassID = make(map[string]map[string][]string)
	}
	if f.methodByClassID[classID] == nil {
		f.methodByClassID[classID] = make(map[string][]string)
	}
	methodName, _ := method["name"].(string)
	f.methodByClassID[classID][methodName] = append(f.methodByClassID[classID][methodName], method["id"].(string))
	if len(f.classByName[className]) == 1 {
		if f.methodByClassName[className] == nil {
			f.methodByClassName[className] = make(map[string][]string)
		}
		f.methodByClassName[className][methodName] = append(f.methodByClassName[className][methodName], method["id"].(string))
	}
}

func (f *factSet) indexClassDeclaration(path, name, id string) {
	if f.classByFileAndName == nil {
		f.classByFileAndName = make(map[string]map[string][]string)
	}
	path = normalizeSourcePath(path)
	if f.classByFileAndName[path] == nil {
		f.classByFileAndName[path] = make(map[string][]string)
	}
	f.classByFileAndName[path][name] = append(f.classByFileAndName[path][name], id)
}

func (f *factSet) indexPreloadAlias(path, name, targetPath string) {
	if f.preloadAliasByFileAndName == nil {
		f.preloadAliasByFileAndName = make(map[string]map[string][]string)
	}
	path = normalizeSourcePath(path)
	if f.preloadAliasByFileAndName[path] == nil {
		f.preloadAliasByFileAndName[path] = make(map[string][]string)
	}
	f.preloadAliasByFileAndName[path][name] = append(f.preloadAliasByFileAndName[path][name], normalizeSourcePath(targetPath))
}

func (f *factSet) indexStaticFunction(record map[string]any) {
	if record["relation"] != "defines" {
		return
	}
	method, ok := f.nodeByID[record["target"].(string)]
	if !ok || method["kind"] != "function" {
		return
	}
	attrs, _ := method["attributes"].(map[string]any)
	if attrs == nil || attrs["static"] != true {
		return
	}
	span, _ := method["span"].(map[string]any)
	if spanInt(span, "start_column") != 1 {
		return
	}
	path, _ := method["path"].(string)
	if f.staticMethodByModulePath == nil {
		f.staticMethodByModulePath = make(map[string]map[string][]string)
	}
	path = normalizeSourcePath(path)
	if f.staticMethodByModulePath[path] == nil {
		f.staticMethodByModulePath[path] = make(map[string][]string)
	}
	name, _ := method["name"].(string)
	f.staticMethodByModulePath[path][name] = append(f.staticMethodByModulePath[path][name], method["id"].(string))
}

func (f *factSet) addUnresolved(record map[string]any) {
	if f.unresolvedKeys == nil {
		f.unresolvedKeys = make(map[string]struct{})
	}
	key := unresolvedSortKey(record)
	if _, exists := f.unresolvedKeys[key]; exists {
		return
	}
	f.unresolvedKeys[key] = struct{}{}
	f.unresolved = append(f.unresolved, record)
	f.unresolvedOrderKeys = append(f.unresolvedOrderKeys, key)
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

func normalizeSourcePath(path string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
}
