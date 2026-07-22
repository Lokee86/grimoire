package main

import "go/ast"

func (s *scanner) addCallEdges() {
	for _, function := range s.callables {
		targets := s.targets[function.packageKey]
		ast.Inspect(function.body, func(node ast.Node) bool {
			if node != function.body {
				if _, nested := node.(*ast.FuncLit); nested {
					return false
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			s.summary.CallExpressions++
			span := s.span(call.Pos(), call.End(), function.path)
			if resolution, exists := s.semanticCalls[callsiteKey(function.source, span)]; exists {
				if resolution.resolved {
					s.addResolvedCall(function.source, span, resolution)
					return true
				}
				if resolution.reason == ReasonAmbiguousTarget && s.isInternalNamespace(resolution.namespace) && resolution.name != "" {
					target := s.ensureNamedFunction(resolution.namespace, resolution.name)
					s.addResolvedCall(function.source, span, semanticCall{
						edges:    []semanticEdge{{target: target, relation: RelCalls}},
						resolved: true, class: callClassInternal,
					})
					return true
				}
				s.addUnresolved(function, call, resolution.reason, resolution.namespace, resolution.name)
				return true
			}

			if literal := calledFuncLiteral(call.Fun); literal != nil {
				position := s.set.PositionFor(literal.Pos(), false)
				if closure, exists := s.closureKeys[closurePositionKey(function.path, position)]; exists {
					s.addResolvedCall(function.source, span, semanticCall{
						edges:    []semanticEdge{{target: closure, relation: RelCalls}},
						resolved: true, class: callClassDynamic,
					})
					return true
				}
			}
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				receiver := expressionName(selector.X)
				if identifier, ok := selector.X.(*ast.Ident); ok {
					if namespace := s.fileImports[function.path][identifier.Name]; namespace != "" {
						class := callClassExternal
						if s.isInternalNamespace(namespace) {
							class = callClassInternal
						}
						target := s.ensureNamedFunction(namespace, selector.Sel.Name)
						s.addResolvedCall(function.source, span, semanticCall{
							edges:    []semanticEdge{{target: target, relation: RelCalls}},
							resolved: true, class: class,
						})
						return true
					}
				}
				target := s.ensureUnknownMethod(receiver, selector.Sel.Name)
				s.addResolvedCall(function.source, span, semanticCall{
					edges:    []semanticEdge{{target: target, relation: RelCalls}},
					resolved: true, class: callClassDynamic,
				})
				return true
			}
			identifier, ok := call.Fun.(*ast.Ident)
			if !ok {
				reason, namespace, name := classifyNonIdentifier(call.Fun)
				s.addUnresolved(function, call, reason, namespace, name)
				return true
			}
			candidates := targets[identifier.Name]
			if len(candidates) == 0 {
				if isBuiltin(identifier.Name) {
					target := s.ensureBuiltinNode(identifier.Name)
					s.addResolvedCall(function.source, span, semanticCall{
						edges:    []semanticEdge{{target: target, relation: RelCalls}},
						resolved: true, class: callClassBuiltin,
					})
					return true
				}
				s.addUnresolved(function, call, ReasonMissingTarget, function.namespace, identifier.Name)
				return true
			}
			if len(candidates) > 1 {
				s.addUnresolved(function, call, ReasonAmbiguousTarget, function.namespace, identifier.Name)
				return true
			}
			s.addEdge(function.source, candidates[0], RelCalls, span)
			s.summary.DirectCalls++
			return true
		})
	}
}

func (s *scanner) addResolvedCall(source NodeKey, span *SourceSpan, resolution semanticCall) {
	hasDefiniteCall := false
	for _, edge := range resolution.edges {
		s.addEdge(source, edge.target, edge.relation, span)
		switch edge.relation {
		case RelCalls:
			hasDefiniteCall = true
		case RelPossibleCalls:
			s.summary.PossibleCallTargets++
		}
	}
	if hasDefiniteCall {
		s.summary.DirectCalls++
	}
	switch resolution.class {
	case callClassBuiltin:
		s.summary.BuiltinCalls++
	case callClassConversion:
		s.summary.ConversionCalls++
	case callClassExternal:
		s.summary.ExternalCalls++
	case callClassDynamic:
		s.summary.DynamicCalls++
	case callClassInterface:
		s.summary.InterfaceCalls++
	}
}

func (s *scanner) addUnresolved(
	function callable,
	call *ast.CallExpr,
	reason UnresolvedReason,
	namespace string,
	name string,
) {
	s.facts.Unresolved = append(s.facts.Unresolved, UnresolvedReferenceFact{
		Source: function.source, Relation: RelCalls, Expression: expressionText(s.set, call.Fun),
		CandidateNamespace: namespace, CandidateName: name, Reason: reason,
		Span: s.span(call.Pos(), call.End(), function.path),
	})
	s.summary.UnresolvedCalls++
	s.trackUnresolvedReason(reason)
}

func (s *scanner) trackUnresolvedReason(reason UnresolvedReason) {
	switch reason {
	case ReasonBuiltinTarget:
		s.summary.BuiltinCalls++
	case ReasonTypeConversion:
		s.summary.ConversionCalls++
	case ReasonExternalTarget:
		s.summary.ExternalCalls++
	case ReasonDynamicTarget:
		s.summary.DynamicCalls++
	}
}
