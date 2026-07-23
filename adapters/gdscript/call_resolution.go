package main

type callResolution struct {
	functionTargets   []string
	constructorOwners []string
	reason            string
	knownReceiver     bool
}

func (m *semanticModel) resolveCall(context analysisContext, call callReference) callResolution {
	if len(call.receiver) > 0 && isCallableInvocation(call.name) {
		if targets := ownerSlice(m.inferExpressionCallables(context, call.receiver)); len(targets) > 0 {
			return callResolution{functionTargets: targets, knownReceiver: true}
		}
	}
	if len(call.receiver) == 0 {
		if call.name == "super" {
			return m.resolveSuperConstructor(context)
		}
		if isBuiltin(call.name) {
			return callResolution{reason: "builtin-target"}
		}
		targets := m.methodTargets(context.ownerID, call.name, false, false)
		if len(targets) > 0 {
			return callResolution{functionTargets: targets, knownReceiver: true}
		}
		if m.hasExternalParent(context.ownerID, make(map[string]bool)) {
			return callResolution{reason: "external-target", knownReceiver: true}
		}
		return callResolution{reason: "dynamic-target"}
	}

	receiverName := simpleIdentifier(call.receiver)
	if receiverName == "self" {
		return m.resolveMethods([]string{context.ownerID}, call.name, false, true)
	}
	if receiverName == "super" {
		return m.resolveMethods(m.facts.parentByOwnerID[context.ownerID], call.name, false, true)
	}
	if receiverName != "" {
		if owners, known := m.preloadAliasOwners(context.file.path, receiverName); known {
			if len(owners) == 0 {
				return callResolution{reason: m.preloadAliasUnresolvedReason(context.file.path, receiverName), knownReceiver: true}
			}
			if len(owners) != 1 {
				return callResolution{reason: "ambiguous-target", knownReceiver: true}
			}
			if call.name == "new" {
				return callResolution{constructorOwners: ownerSlice(owners), knownReceiver: true}
			}
			return m.resolveMethods(ownerSlice(owners), call.name, true, true)
		}
		if owners := m.typeAliasOwners(context.file.path, receiverName); len(owners) > 0 {
			if len(owners) != 1 {
				return callResolution{reason: "ambiguous-target", knownReceiver: true}
			}
			if call.name == "new" {
				return callResolution{constructorOwners: ownerSlice(owners), knownReceiver: true}
			}
			return m.resolveMethods(ownerSlice(owners), call.name, true, true)
		}
		if m.bindingIsBuiltin(context, receiverName) {
			return callResolution{reason: "builtin-target", knownReceiver: true}
		}
		if owners := m.bindingOwners(context, receiverName); len(owners) > 0 {
			return m.resolveMethods(ownerSlice(owners), call.name, false, true)
		}
		if m.bindingDeclared(context, receiverName) {
			return callResolution{reason: "dynamic-target", knownReceiver: true}
		}
		if owners, ambiguous := m.classOwners(context.file.path, receiverName); len(owners) > 0 || ambiguous {
			if ambiguous {
				return callResolution{reason: "ambiguous-target", knownReceiver: true}
			}
			if call.name == "new" {
				return callResolution{constructorOwners: ownerSlice(owners), knownReceiver: true}
			}
			return m.resolveMethods(ownerSlice(owners), call.name, true, true)
		}
		if owner := m.autoloadOwner(context.file.path, receiverName); owner != "" {
			return m.resolveMethods([]string{owner}, call.name, false, true)
		}
		if isBuiltin(receiverName) {
			return callResolution{reason: "builtin-target", knownReceiver: true}
		}
	}

	if call.name == "new" {
		if owners := m.inferTypeReferenceOwners(context, call.receiver); len(owners) > 0 {
			return callResolution{constructorOwners: ownerSlice(owners), knownReceiver: true}
		}
	}
	owners := m.inferExpressionOwners(context, call.receiver)
	if len(owners) > 0 {
		return m.resolveMethods(ownerSlice(owners), call.name, false, true)
	}
	return callResolution{reason: "dynamic-target"}
}

func (m *semanticModel) autoloadOwner(sourcePath, name string) string {
	projectRoot := m.facts.projectRootByFilePath[normalizeSourcePath(sourcePath)]
	return m.facts.autoloadOwnerByProjectName[projectRoot][name]
}

func (m *semanticModel) resolveSuperConstructor(context analysisContext) callResolution {
	parents := m.facts.parentByOwnerID[context.ownerID]
	if len(parents) == 0 {
		if m.facts.externalParentByOwnerID[context.ownerID] {
			return callResolution{reason: "external-target", knownReceiver: true}
		}
		return callResolution{reason: "missing-target", knownReceiver: true}
	}
	var targets []string
	for _, parent := range parents {
		targets = append(targets, m.methodTargets(parent, "_init", false, false)...)
	}
	return callResolution{functionTargets: uniqueSorted(targets), knownReceiver: true}
}

func (m *semanticModel) resolveMethods(owners []string, name string, staticOnly, known bool) callResolution {
	var targets []string
	for _, owner := range owners {
		targets = append(targets, m.methodTargets(owner, name, staticOnly, false)...)
	}
	targets = uniqueSorted(targets)
	if len(targets) == 0 {
		reason := "dynamic-target"
		if known {
			reason = "missing-target"
			if !staticOnly {
				reason = "external-target"
			} else {
				for _, owner := range owners {
					if m.hasExternalParent(owner, make(map[string]bool)) {
						reason = "external-target"
						break
					}
				}
			}
		}
		return callResolution{reason: reason, knownReceiver: known}
	}
	return callResolution{functionTargets: targets, knownReceiver: known}
}

func (m *semanticModel) methodTargets(owner, name string, staticOnly, parentsOnly bool) []string {
	return m.methodTargetsSeen(owner, name, staticOnly, parentsOnly, make(map[string]bool))
}

func (m *semanticModel) methodTargetsSeen(owner, name string, staticOnly, parentsOnly bool, seen map[string]bool) []string {
	if owner == "" || seen[owner] {
		return nil
	}
	seen[owner] = true
	if !parentsOnly {
		methods := m.facts.methodByOwnerID[owner][name]
		if staticOnly {
			methods = m.facts.staticMethodByOwnerID[owner][name]
		}
		if len(methods) > 0 {
			return uniqueSorted(methods)
		}
	}
	var inherited []string
	for _, parent := range m.facts.parentByOwnerID[owner] {
		inherited = append(inherited, m.methodTargetsSeen(parent, name, staticOnly, false, seen)...)
	}
	return uniqueSorted(inherited)
}

func (m *semanticModel) bindingOwners(context analysisContext, name string) ownerSet {
	if context.functionID != "" {
		if values := m.locals[context.functionID][name]; len(values) > 0 {
			return cloneOwners(values)
		}
	}
	return m.memberOwners(context.ownerID, name, make(map[string]bool))
}

func (m *semanticModel) memberOwners(owner, name string, seen map[string]bool) ownerSet {
	if owner == "" || seen[owner] {
		return nil
	}
	seen[owner] = true
	if values := m.members[owner][name]; len(values) > 0 {
		return cloneOwners(values)
	}
	var result ownerSet
	for _, parent := range m.facts.parentByOwnerID[owner] {
		result = unionOwners(result, m.memberOwners(parent, name, seen))
	}
	return result
}
