package main

import (
	"sort"
	"strings"
)

type typeDeclaration struct {
	declarationKind string
	id              string
}

type resolutionContext struct {
	explicitImports []string
	packageName     string
	wildcardImports []string
}

type relationshipEvidence struct {
	annotation   bool
	attributes   map[string]any
	context      resolutionContext
	expression   string
	lexicalOwner string
	lookupName   string
	owner        string
	relation     string
	source       string
	span         *span
}

func (parser *javaParser) queueRelationship(source, relation, expression, lookupName, lexicalOwner string, evidenceSpan *span, annotation bool, attributes map[string]any) {
	context := parser.resolution
	context.explicitImports = append([]string(nil), context.explicitImports...)
	context.wildcardImports = append([]string(nil), context.wildcardImports...)
	parser.state.relationships = append(parser.state.relationships, relationshipEvidence{
		annotation: annotation, attributes: attributes, context: context, expression: expression,
		lexicalOwner: lexicalOwner, lookupName: lookupName, owner: parser.path,
		relation: relation, source: source, span: evidenceSpan,
	})
}

func (state *analysisState) resolveRelationships() {
	sort.Slice(state.relationships, func(left, right int) bool {
		one, two := state.relationships[left], state.relationships[right]
		return relationshipKey(one) < relationshipKey(two)
	})
	for _, evidence := range state.relationships {
		candidates := state.relationshipCandidates(evidence)
		switch len(candidates) {
		case 0:
			state.facts.addUnresolved(evidence.source, evidence.relation, evidence.expression, "external-target", evidence.owner, evidence.span, evidence.attributes)
		case 1:
			state.facts.addEdge(evidence.source, candidates[0], evidence.relation, evidence.owner, evidence.span, evidence.attributes)
		default:
			attributes := cloneAttributes(evidence.attributes)
			attributes["candidate_count"] = len(candidates)
			state.facts.addUnresolved(evidence.source, evidence.relation, evidence.expression, "ambiguous-target", evidence.owner, evidence.span, attributes)
		}
	}
}

func relationshipKey(evidence relationshipEvidence) string {
	position := ""
	if evidence.span != nil {
		position = spanKey(map[string]any{"span": evidence.span})
	}
	return evidence.owner + "\x00" + position + "\x00" + evidence.source + "\x00" + evidence.relation + "\x00" + evidence.expression
}

func (state *analysisState) relationshipCandidates(evidence relationshipEvidence) []string {
	qualifiedNames := candidateQualifiedNames(evidence.lookupName, evidence.lexicalOwner, evidence.context)
	if state.exactQualifiedReference(evidence.lookupName) {
		qualifiedNames = appendUnique(qualifiedNames, evidence.lookupName)
	}
	var result []string
	for _, qualifiedName := range qualifiedNames {
		for _, declaration := range state.types[qualifiedName] {
			if !evidence.annotation || declaration.declarationKind == "annotation" {
				result = append(result, declaration.id)
			}
		}
	}
	sort.Strings(result)
	return result
}

func (state *analysisState) exactQualifiedReference(name string) bool {
	if !strings.Contains(name, ".") {
		return false
	}
	for packageName := range state.namespaces {
		if packageName != "<default>" && strings.HasPrefix(name, packageName+".") {
			return true
		}
	}
	return false
}

func candidateQualifiedNames(name, lexicalOwner string, context resolutionContext) []string {
	candidates := make(map[string]struct{})
	add := func(candidate string) {
		if candidate != "" {
			candidates[candidate] = struct{}{}
		}
	}
	if context.packageName == "" {
		add(name)
	} else {
		add(context.packageName + "." + name)
	}
	for scope := lexicalOwner; scope != "" && scope != context.packageName; {
		add(scope + "." + name)
		cut := strings.LastIndex(scope, ".")
		if cut < 0 {
			break
		}
		scope = scope[:cut]
	}
	first, suffix := name, ""
	if dot := strings.Index(name, "."); dot >= 0 {
		first, suffix = name[:dot], name[dot:]
	}
	for _, imported := range context.explicitImports {
		if simpleName(imported) == first {
			add(imported + suffix)
		}
	}
	for _, imported := range context.wildcardImports {
		add(imported + "." + name)
	}
	result := make([]string, 0, len(candidates))
	for candidate := range candidates {
		result = append(result, candidate)
	}
	sort.Strings(result)
	return result
}

func simpleName(qualifiedName string) string {
	if dot := strings.LastIndex(qualifiedName, "."); dot >= 0 {
		return qualifiedName[dot+1:]
	}
	return qualifiedName
}

func cloneAttributes(attributes map[string]any) map[string]any {
	result := make(map[string]any, len(attributes)+1)
	for key, value := range attributes {
		result[key] = value
	}
	return result
}
