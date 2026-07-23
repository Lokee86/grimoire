package main

import (
	"fmt"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

type semanticCall struct {
	edges     []semanticEdge
	resolved  bool
	reason    UnresolvedReason
	namespace string
	name      string
	class     callClass
	contract  NodeKey
}

type semanticTargets struct {
	byObject                 map[*types.Func]NodeKey
	byID                     map[string][]NodeKey
	interfaceImplementations map[NodeKey][]NodeKey
}

type namedTypeTarget struct {
	named *types.Named
	key   NodeKey
}

type interfaceTarget struct {
	named *types.Named
	iface *types.Interface
	key   NodeKey
}

type ssaOutcome struct {
	invoke  bool
	targets map[NodeKey]bool
}

func (s *scanner) loadSemanticCalls() error {
	config := &packages.Config{
		Mode:  packages.LoadAllSyntax | packages.NeedModule,
		Dir:   s.root,
		Tests: true,
	}
	roots, err := packages.Load(config, "./...")
	if err != nil {
		return fmt.Errorf("load Go packages: %w", err)
	}
	loaded := flattenPackages(roots)
	for _, pkg := range loaded {
		s.summary.SemanticErrors += len(pkg.Errors)
	}

	targets := semanticTargets{
		byObject:                 make(map[*types.Func]NodeKey),
		byID:                     make(map[string][]NodeKey),
		interfaceImplementations: make(map[NodeKey][]NodeKey),
	}
	for _, pkg := range loaded {
		s.collectSemanticTargets(pkg, &targets)
	}
	s.collectSemanticTypes(loaded, &targets)
	for _, pkg := range loaded {
		s.collectSemanticCalls(pkg, targets)
	}
	if err := s.collectSSACalls(roots, targets); err != nil {
		return err
	}
	return nil
}

func flattenPackages(roots []*packages.Package) []*packages.Package {
	seen := make(map[string]*packages.Package)
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if pkg == nil || seen[pkg.ID] != nil {
			return
		}
		seen[pkg.ID] = pkg
		paths := make([]string, 0, len(pkg.Imports))
		for path := range pkg.Imports {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		for _, path := range paths {
			visit(pkg.Imports[path])
		}
	}
	for _, root := range roots {
		visit(root)
	}
	result := make([]*packages.Package, 0, len(seen))
	for _, pkg := range seen {
		result = append(result, pkg)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
