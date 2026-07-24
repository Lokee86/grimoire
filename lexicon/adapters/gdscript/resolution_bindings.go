package main

import (
	"path/filepath"
	"sort"
	"strings"
)

func resolveClassID(facts *factSet, sourcePath, className string) (string, bool) {
	sameFile := facts.classByFileAndName[normalizeSourcePath(sourcePath)][className]
	if len(sameFile) == 1 {
		return sameFile[0], true
	}
	if len(sameFile) > 1 {
		return "", false
	}
	classIDs := facts.classByName[className]
	if len(classIDs) != 1 {
		return "", false
	}
	return classIDs[0], true
}

func (m *semanticModel) classOwners(path, name string) (ownerSet, bool) {
	path = normalizeSourcePath(path)
	if ids := m.facts.classByFileAndName[path][name]; len(ids) > 0 {
		if len(ids) == 1 {
			return ownerSet{ids[0]: {}}, false
		}
		return nil, true
	}
	ids := m.facts.classByName[name]
	if len(ids) == 1 {
		return ownerSet{ids[0]: {}}, false
	}
	return nil, len(ids) > 1
}

func (m *semanticModel) preloadAliasUnresolvedReason(path, name string) string {
	paths := m.facts.preloadAliasByFileAndName[normalizeSourcePath(path)][name]
	for _, targetPath := range paths {
		if strings.EqualFold(filepath.Ext(targetPath), ".gd") {
			return "missing-target"
		}
	}
	return "external-target"
}

func (m *semanticModel) preloadAliasOwners(path, name string) (ownerSet, bool) {
	paths := m.facts.preloadAliasByFileAndName[normalizeSourcePath(path)][name]
	if len(paths) == 0 {
		return nil, false
	}
	owners := make(ownerSet)
	for _, targetPath := range paths {
		if candidates := m.facts.scriptOwnerCandidatesByPath[targetPath]; len(candidates) > 0 {
			for _, owner := range candidates {
				owners[owner] = struct{}{}
			}
			continue
		}
		if owner := m.facts.scriptOwnerByPath[targetPath]; owner != "" {
			owners[owner] = struct{}{}
		}
	}
	return owners, true
}

func (m *semanticModel) hasExternalParent(owner string, seen map[string]bool) bool {
	if owner == "" || seen[owner] {
		return false
	}
	seen[owner] = true
	if m.facts.externalParentByOwnerID[owner] {
		return true
	}
	for _, parent := range m.facts.parentByOwnerID[owner] {
		if m.hasExternalParent(parent, seen) {
			return true
		}
	}
	return false
}

func (m *semanticModel) bindingIsBuiltin(context analysisContext, name string) bool {
	if context.functionID != "" {
		if decl := m.facts.declarationByID[context.functionID]; decl != nil {
			if isBuiltinType(decl.parameterTypes[name]) {
				return true
			}
		}
	}
	for i := range context.file.declarations {
		decl := &context.file.declarations[i]
		if decl.name == name && decl.typeName != "" && isBuiltinType(decl.typeName) {
			return true
		}
	}
	return false
}

func isCallableInvocation(name string) bool {
	switch name {
	case "call", "callv", "call_deferred":
		return true
	default:
		return false
	}
}

func isBuiltinType(name string) bool {
	name = strings.TrimSpace(name)
	if bracket := strings.Index(name, "["); bracket >= 0 {
		name = name[:bracket]
	}
	switch name {
	case "void", "Variant", "bool", "int", "float", "String", "StringName", "NodePath", "Array", "Dictionary", "PackedByteArray", "PackedInt32Array", "PackedInt64Array", "PackedFloat32Array", "PackedFloat64Array", "PackedStringArray", "Vector2", "Vector2i", "Vector3", "Vector3i", "Vector4", "Vector4i", "Rect2", "Rect2i", "Transform2D", "Transform3D", "Basis", "Quaternion", "Plane", "Projection", "AABB", "Color", "RID", "Callable", "Signal":
		return true
	default:
		return isBuiltin(name)
	}
}

func ownerSlice(values ownerSet) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
