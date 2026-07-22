package main

import "fmt"

type declaration struct {
	keyword           string
	kind              string
	name              string
	nameIndex         int
	indent            int
	span              sourceSpan
	extends           string
	attributes        map[string]any
	parameters        []string
	parameterNames    []string
	parameterTypes    map[string]string
	parameterDefaults map[string][]token
	returnType        string
	typeName          string
	initializer       []token
	preloadPath       string
	static            bool
	async             bool
	nodeID            string
	key               string
	ownerID           string
	ownerClassID      string
	ownerFunction     string
}

type parsedFile struct {
	path          string
	content       []byte
	statements    []statement
	declarations  []declaration
	imports       []importReference
	calls         []callReference
	moduleID      string
	classID       string
	scriptOwnerID string
}

type scope struct {
	indent   int
	id       string
	key      string
	kind     string
	classID  string
	function string
}

func processDeclarations(facts *factSet, pf *parsedFile) {
	prepareScriptClass(facts, pf)
	if facts.scriptOwnerByPath == nil {
		facts.scriptOwnerByPath = make(map[string]string)
		facts.scriptOwnerCandidatesByPath = make(map[string][]string)
	}
	facts.scriptOwnerByPath[normalizeSourcePath(pf.path)] = pf.scriptOwnerID
	occurrences := make(map[string]int)
	var scopes []scope
	for i := range pf.declarations {
		decl := &pf.declarations[i]
		for len(scopes) > 0 && decl.indent <= scopes[len(scopes)-1].indent {
			scopes = scopes[:len(scopes)-1]
		}
		parentID, parentKey := pf.scriptOwnerID, pf.path
		ownerClass, ownerFunction := pf.scriptOwnerID, ""
		if len(scopes) > 0 {
			current := scopes[len(scopes)-1]
			parentID, parentKey = current.id, current.key
			ownerClass, ownerFunction = current.classID, current.function
		}
		if decl.keyword == "class_name" {
			parentID, parentKey = pf.moduleID, pf.path
			ownerClass = decl.nodeID
		}
		decl.ownerID, decl.ownerClassID, decl.ownerFunction = parentID, ownerClass, ownerFunction
		if decl.kind == "extends" {
			continue
		}
		if decl.kind == "type" && decl.nodeID != "" {
			base := parentKey + "::type::" + decl.name
			if occurrences[base] == 0 {
				occurrences[base] = 1
			}
		}
		if decl.kind == "type" && decl.nodeID == "" {
			decl.key = nextDeclarationKey(occurrences, parentKey+"::type::"+decl.name)
			decl.nodeID = nodeID("type", decl.key)
			facts.classByName[decl.name] = append(facts.classByName[decl.name], decl.nodeID)
			facts.indexClassDeclaration(pf.path, decl.name, decl.nodeID)
		}
		if decl.kind == "type" && decl.keyword == "class_name" && decl.indent == 0 {
			path := normalizeSourcePath(pf.path)
			facts.scriptOwnerCandidatesByPath[path] = append(facts.scriptOwnerCandidatesByPath[path], decl.nodeID)
		}
		if decl.nodeID == "" {
			decl.key = nextDeclarationKey(occurrences, parentKey+"::"+decl.kind+"::"+decl.name)
			decl.nodeID = nodeID(decl.kind, decl.key)
		}
		if decl.preloadPath != "" {
			facts.indexPreloadAlias(pf.path, decl.name, decl.preloadPath)
		}
		attrs := declarationAttributes(decl)
		facts.addNode(node(decl.kind, decl.name, pf.path, qualifiedDeclaration(pf.path, parentKey, decl.name), decl.nodeID, decl.span, "", attrs))
		facts.addEdge(edge(parentID, decl.nodeID, "contains", decl.span))
		facts.addEdge(edge(parentID, decl.nodeID, "defines", decl.span))
		facts.indexDeclaration(pf, decl)
		if decl.kind == "type" && decl.keyword != "class_name" {
			scopes = append(scopes, scope{indent: decl.indent, id: decl.nodeID, key: decl.key, kind: "type", classID: decl.nodeID, function: ownerFunction})
		} else if decl.kind == "function" {
			scopes = append(scopes, scope{indent: decl.indent, id: decl.nodeID, key: decl.key, kind: "function", classID: ownerClass, function: decl.nodeID})
		}
	}
}

func prepareScriptClass(facts *factSet, pf *parsedFile) {
	pf.scriptOwnerID = pf.moduleID
	for i := range pf.declarations {
		decl := &pf.declarations[i]
		if decl.kind != "type" || decl.keyword != "class_name" || decl.indent != 0 {
			continue
		}
		decl.key = pf.path + "::type::" + decl.name
		decl.nodeID = nodeID("type", decl.key)
		pf.classID, pf.scriptOwnerID = decl.nodeID, decl.nodeID
		facts.classByName[decl.name] = append(facts.classByName[decl.name], decl.nodeID)
		facts.indexClassDeclaration(pf.path, decl.name, decl.nodeID)
		return
	}
}

func nextDeclarationKey(occurrences map[string]int, base string) string {
	occurrence := occurrences[base]
	occurrences[base] = occurrence + 1
	if occurrence == 0 {
		return base
	}
	return fmt.Sprintf("%s#%d", base, occurrence+1)
}

func declarationAttributes(decl *declaration) map[string]any {
	attrs := cloneMap(decl.attributes)
	if decl.kind == "function" {
		attrs["parameters"] = append([]string(nil), decl.parameters...)
		if decl.keyword == "lambda" {
			attrs["anonymous"] = true
		}
		if decl.returnType != "" {
			attrs["return_type"] = decl.returnType
		}
		if decl.static {
			attrs["static"] = true
		}
		if decl.async {
			attrs["async"] = true
		}
	}
	if decl.extends != "" {
		attrs["extends"] = decl.extends
	}
	return attrs
}

func qualifiedDeclaration(path, parentKey, name string) string {
	if parentKey == path {
		return path + "::" + name
	}
	return parentKey + "::" + name
}
