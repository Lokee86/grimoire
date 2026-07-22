package main

import "strings"

func (m *semanticModel) typeOwners(path, name string) ownerSet {
	name = strings.TrimSpace(name)
	if name == "" || isBuiltinType(name) {
		return nil
	}
	if bracket := strings.Index(name, "["); bracket >= 0 {
		name = strings.TrimSpace(name[:bracket])
	}
	if dot := strings.LastIndex(name, "."); dot >= 0 {
		prefix, nested := strings.TrimSpace(name[:dot]), strings.TrimSpace(name[dot+1:])
		if owners, known := m.preloadAliasOwners(path, prefix); known {
			var targets []string
			for owner := range owners {
				targets = append(targets, m.facts.typeByOwnerID[owner][nested]...)
			}
			if targets = uniqueSorted(targets); len(targets) == 1 {
				return ownerSet{targets[0]: {}}
			}
		}
		name = nested
	}
	if id, ok := resolveClassID(m.facts, path, name); ok {
		return ownerSet{id: {}}
	}
	return nil
}

func (m *semanticModel) addMember(owner, name string, values ownerSet) bool {
	if owner == "" || name == "" || len(values) == 0 {
		return false
	}
	if m.members[owner] == nil {
		m.members[owner] = make(map[string]ownerSet)
	}
	return mergeOwners(m.members[owner], name, values)
}

func (m *semanticModel) addLocal(function, name string, values ownerSet) bool {
	if function == "" || name == "" || len(values) == 0 {
		return false
	}
	if m.locals[function] == nil {
		m.locals[function] = make(map[string]ownerSet)
	}
	return mergeOwners(m.locals[function], name, values)
}

func (m *semanticModel) addReturn(function string, values ownerSet) bool {
	if function == "" || len(values) == 0 {
		return false
	}
	if m.returns[function] == nil {
		m.returns[function] = make(ownerSet)
	}
	changed := false
	for value := range values {
		if _, exists := m.returns[function][value]; !exists {
			m.returns[function][value] = struct{}{}
			changed = true
		}
	}
	return changed
}

func mergeOwners(bindings map[string]ownerSet, name string, values ownerSet) bool {
	if bindings[name] == nil {
		bindings[name] = make(ownerSet)
	}
	changed := false
	for value := range values {
		if _, exists := bindings[name][value]; !exists {
			bindings[name][value] = struct{}{}
			changed = true
		}
	}
	return changed
}

func (m *semanticModel) hasLocal(function, name string) bool {
	return len(m.locals[function][name]) > 0
}

func (m *semanticModel) hasMember(owner, name string) bool {
	return len(m.members[owner][name]) > 0
}

func (m *semanticModel) localDeclared(function, name string) bool {
	return m.facts.declaredLocalByFunction[function][name]
}

func (m *semanticModel) memberDeclared(owner, name string) bool {
	if m.facts.declaredMemberByOwner[owner][name] {
		return true
	}
	for _, parent := range m.facts.parentByOwnerID[owner] {
		if m.memberDeclared(parent, name) {
			return true
		}
	}
	return false
}

func cloneOwners(values ownerSet) ownerSet {
	if len(values) == 0 {
		return nil
	}
	result := make(ownerSet, len(values))
	for value := range values {
		result[value] = struct{}{}
	}
	return result
}

func unionOwners(sets ...ownerSet) ownerSet {
	result := make(ownerSet)
	for _, values := range sets {
		for value := range values {
			result[value] = struct{}{}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
