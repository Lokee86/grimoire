package main

import "strings"

type relationshipTarget struct {
	form string
	id   string
	kind string
}

type pendingRelationship struct {
	attributes   map[string]any
	expression   string
	file         *parsedKotlinFile
	kind         string
	lexicalOwner string
	span         sourceSpan
	source       string
	targetName   string
}

func (state *analysis) indexRelationshipTarget(id, qualifiedName, kind, form string) {
	if kind != "type" && kind != "interface" {
		return
	}
	target := relationshipTarget{form: form, id: id, kind: kind}
	state.relationshipByQN[qualifiedName] = append(state.relationshipByQN[qualifiedName], target)
}

func (state *analysis) queueSupertypes(file *parsedKotlinFile, source, lexicalOwner string, supertypes []supertypeDecl) {
	for _, supertype := range supertypes {
		attributes := map[string]any(nil)
		if supertype.delegated {
			attributes = map[string]any{
				"delegate_expression": supertype.delegateExpression,
				"delegated":           true,
			}
		}
		state.pendingRelations = append(state.pendingRelations, pendingRelationship{
			attributes: attributes, expression: supertype.expression, file: file, kind: "supertype",
			lexicalOwner: lexicalOwner, source: source, span: supertype.span, targetName: supertype.targetName,
		})
	}
}

func (state *analysis) queueAnnotations(file *parsedKotlinFile, source, lexicalOwner string, annotations []string, span sourceSpan) {
	for _, annotation := range annotations {
		state.pendingRelations = append(state.pendingRelations, pendingRelationship{
			expression: annotation, file: file, kind: "annotation", lexicalOwner: lexicalOwner,
			source: source, span: span, targetName: annotationReference(annotation),
		})
	}
}

func (state *analysis) emitRelationships() {
	for _, pending := range state.pendingRelations {
		targets, reason := state.resolveRelationship(pending)
		relation := "annotates"
		if pending.kind == "supertype" {
			relation = "extends"
		}
		if len(targets) == 1 {
			if pending.kind == "supertype" && targets[0].kind == "interface" {
				relation = "implements"
			}
			state.facts.addEdge(pending.source, targets[0].id, relation, pending.file.path, &pending.span, pending.attributes)
			continue
		}
		state.facts.addUnresolved(
			pending.source, relation, pending.expression, reason, pending.file.path, &pending.span, pending.attributes,
		)
	}
}

func annotationReference(annotation string) string {
	if colon := strings.LastIndexByte(annotation, ':'); colon >= 0 {
		return annotation[colon+1:]
	}
	return annotation
}
