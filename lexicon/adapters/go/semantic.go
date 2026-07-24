package main

import (
	"fmt"
	"go/types"
	"sort"
	"time"

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

func (s *scanner) loadSemanticCalls(options ScanOptions) error {
	for _, module := range s.modules {
		packageLoadStarted := time.Now()
		config := &packages.Config{
			Mode:  packages.LoadAllSyntax | packages.NeedModule,
			Dir:   module.Root,
			Tests: true,
		}
		roots, err := packages.Load(config, "./...")
		if err != nil {
			return fmt.Errorf("load Go module %s: %w", module.Path, err)
		}
		loaded := flattenPackages(roots)
		s.summary.PackageLoad += time.Since(packageLoadStarted)
		for _, pkg := range loaded {
			s.summary.SemanticErrors += len(pkg.Errors)
		}

		semanticIndexStarted := time.Now()
		targets := semanticTargets{
			byObject:                 make(map[*types.Func]NodeKey),
			byID:                     make(map[string][]NodeKey),
			interfaceImplementations: make(map[NodeKey][]NodeKey),
		}
		for _, pkg := range loaded {
			s.collectSemanticTargets(pkg, &targets)
		}
		s.collectSemanticTypes(loaded, &targets)
		s.summary.SemanticIndex += time.Since(semanticIndexStarted)

		typedResolutionStarted := time.Now()
		if err := s.collectSemanticCallsParallel(loaded, targets, options); err != nil {
			return fmt.Errorf("resolve Go module %s: %w", module.Path, err)
		}
		s.summary.TypedResolution += time.Since(typedResolutionStarted)

		ssaResolutionStarted := time.Now()
		if err := s.collectSSACalls(roots, targets); err != nil {
			return fmt.Errorf("analyze Go module %s: %w", module.Path, err)
		}
		s.summary.SSAResolution += time.Since(ssaResolutionStarted)
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
