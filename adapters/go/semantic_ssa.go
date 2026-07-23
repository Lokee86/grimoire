package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"

	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func (s *scanner) collectSSACalls(roots []*packages.Package, targets semanticTargets) error {
	program, _ := ssautil.AllPackages(roots, ssa.InstantiateGenerics)
	program.Build()
	functions := ssautil.AllFunctions(program)
	s.collectSSACaptures(functions, program.Fset)
	graph := vta.CallGraph(functions, nil)
	outcomes := make(map[string]*ssaOutcome)
	for _, node := range graph.Nodes {
		for _, edge := range node.Out {
			if edge.Site == nil || edge.Callee == nil || edge.Caller == nil {
				continue
			}
			source, ok := s.ssaSourceKey(edge.Caller.Func, targets, program.Fset)
			if !ok {
				continue
			}
			position := program.Fset.PositionFor(edge.Site.Pos(), false)
			rel, err := s.relative(position.Filename)
			if err != nil {
				continue
			}
			key, exists := s.callsiteKeys[callsiteStartKey(source, rel, position)]
			if !exists {
				continue
			}
			target, internal, ok := s.ssaTargetKey(edge.Callee.Func, targets, program.Fset)
			if !ok {
				continue
			}
			common := edge.Site.Common()
			outcome := outcomes[key]
			if outcome == nil {
				outcome = &ssaOutcome{targets: make(map[NodeKey]bool)}
				outcomes[key] = outcome
			}
			outcome.invoke = outcome.invoke || common.IsInvoke()
			if !common.IsInvoke() || internal {
				outcome.targets[target] = true
			}
		}
	}

	keys := make([]string, 0, len(outcomes))
	for key := range outcomes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		outcome := outcomes[key]
		resolvedTargets := make([]NodeKey, 0, len(outcome.targets))
		for target := range outcome.targets {
			resolvedTargets = append(resolvedTargets, target)
		}
		sort.Slice(resolvedTargets, func(i, j int) bool { return resolvedTargets[i] < resolvedTargets[j] })
		if len(resolvedTargets) == 0 {
			continue
		}
		existing := s.semanticCalls[key]
		if !outcome.invoke && existing.resolved {
			continue
		}
		if outcome.invoke {
			edges := make([]semanticEdge, 0, len(resolvedTargets))
			relation := RelPossibleCalls
			if len(resolvedTargets) == 1 {
				relation = RelCalls
			}
			for _, target := range resolvedTargets {
				edges = append(edges, semanticEdge{target: target, relation: relation})
			}
			s.mergeSemanticCall(key, semanticCall{edges: edges, resolved: true, class: callClassInterface})
			continue
		}
		if len(resolvedTargets) == 1 {
			s.mergeSemanticCall(key, semanticCall{
				edges:    []semanticEdge{{target: resolvedTargets[0], relation: RelCalls}},
				resolved: true,
				class:    callClassDynamic,
			})
			continue
		}
		edges := make([]semanticEdge, 0, len(resolvedTargets))
		for _, target := range resolvedTargets {
			edges = append(edges, semanticEdge{target: target, relation: RelPossibleCalls})
		}
		s.mergeSemanticCall(key, semanticCall{edges: edges, resolved: true, class: callClassDynamic})
	}
	return nil
}

func (s *scanner) collectSSACaptures(functions map[*ssa.Function]bool, set *token.FileSet) {
	ordered := make([]*ssa.Function, 0, len(functions))
	for function := range functions {
		ordered = append(ordered, function)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].String() < ordered[j].String() })
	for _, function := range ordered {
		literal, ok := function.Syntax().(*ast.FuncLit)
		if !ok || len(function.FreeVars) == 0 {
			continue
		}
		position := set.PositionFor(literal.Pos(), false)
		rel, err := s.relative(position.Filename)
		if err != nil {
			continue
		}
		closure, exists := s.closureKeys[closurePositionKey(rel, position)]
		if !exists {
			continue
		}
		for index, variable := range function.FreeVars {
			variablePosition := set.PositionFor(variable.Pos(), false)
			variablePath := rel
			identity := fmt.Sprintf("capture:%s:%d:%s", closure, index, variable.Name())
			var span *SourceSpan
			if variablePosition.IsValid() {
				if candidate, pathErr := s.relative(variablePosition.Filename); pathErr == nil {
					variablePath = candidate
				}
				identity = fmt.Sprintf(
					"variable:%s:%s:%d:%d:%s",
					s.importPathFor(variablePath), variablePath, variablePosition.Line, variablePosition.Column, variable.Name(),
				)
				span = &SourceSpan{
					Path:        variablePath,
					StartLine:   uint32(variablePosition.Line),
					StartColumn: uint32(variablePosition.Column),
					EndLine:     uint32(variablePosition.Line),
					EndColumn:   uint32(variablePosition.Column),
				}
			}
			key := hashIdentity(identity)
			s.addNode(NodeFact{Key: key, Kind: KindVariable, Path: variablePath, Name: variable.Name(), Span: span})
			s.addEdge(closure, key, RelReferences, span)
			s.summary.Captures++
		}
	}
}

func (s *scanner) ssaSourceKey(function *ssa.Function, targets semanticTargets, set *token.FileSet) (NodeKey, bool) {
	if function == nil {
		return "", false
	}
	if object := function.Object(); object != nil {
		typed, ok := object.(*types.Func)
		if !ok || !s.isInternalNamespace(objectNamespace(typed)) {
			return "", false
		}
		key, _, unambiguous := s.ensureFunctionNode(typed, targets, set)
		return key, unambiguous
	}
	literal, ok := function.Syntax().(*ast.FuncLit)
	if !ok {
		return "", false
	}
	position := set.PositionFor(literal.Pos(), false)
	rel, err := s.relative(position.Filename)
	if err != nil {
		return "", false
	}
	key, exists := s.closureKeys[closurePositionKey(rel, position)]
	return key, exists
}

func (s *scanner) ssaTargetKey(function *ssa.Function, targets semanticTargets, set *token.FileSet) (NodeKey, bool, bool) {
	if function == nil {
		return "", false, false
	}
	if object := function.Object(); object != nil {
		typed, ok := object.(*types.Func)
		if !ok {
			return "", false, false
		}
		key, internal, unambiguous := s.ensureFunctionNode(typed, targets, set)
		return key, internal, unambiguous
	}
	literal, ok := function.Syntax().(*ast.FuncLit)
	if !ok {
		return s.ensureSyntheticSSAFunction(function, set)
	}
	position := set.PositionFor(literal.Pos(), false)
	rel, err := s.relative(position.Filename)
	if err != nil {
		return s.ensureSyntheticSSAFunction(function, set)
	}
	key, exists := s.closureKeys[closurePositionKey(rel, position)]
	if !exists {
		return s.ensureSyntheticSSAFunction(function, set)
	}
	return key, true, true
}
