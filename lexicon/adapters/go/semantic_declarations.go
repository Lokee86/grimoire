package main

import (
	"go/ast"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

func (s *scanner) collectSemanticTargets(pkg *packages.Package, targets *semanticTargets) {
	if pkg.TypesInfo == nil || pkg.Fset == nil {
		return
	}
	for _, file := range pkg.Syntax {
		rel, ok := s.semanticFilePath(pkg.Fset, file)
		if !ok {
			continue
		}
		importPath := s.importPathFor(rel)
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				object, ok := pkg.TypesInfo.Defs[declaration.Name].(*types.Func)
				if !ok {
					continue
				}
				key := declarationKey(importPath, rel, declaration)
				if _, exists := s.nodes[key]; !exists {
					continue
				}
				targets.add(object, semanticFunctionID(object, s.canonicalNamespace(objectNamespace(object))), key)
			case *ast.GenDecl:
				for _, specification := range declaration.Specs {
					typeSpec, ok := specification.(*ast.TypeSpec)
					if !ok {
						continue
					}
					interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok || interfaceType.Methods == nil {
						continue
					}
					for _, field := range interfaceType.Methods.List {
						for _, name := range field.Names {
							object, ok := pkg.TypesInfo.Defs[name].(*types.Func)
							if !ok {
								continue
							}
							key := hashIdentity(interfaceMethodIdentity(importPath, typeSpec.Name.Name, name.Name))
							if _, exists := s.nodes[key]; !exists {
								continue
							}
							targets.add(object, interfaceMethodIdentity(importPath, typeSpec.Name.Name, name.Name), key)
						}
					}
				}
			}
		}
	}
}

func (targets *semanticTargets) add(object *types.Func, id string, key NodeKey) {
	targets.byObject[object] = key
	targets.byID[id] = appendUniqueKey(targets.byID[id], key)
}

func (s *scanner) collectSemanticTypes(packages []*packages.Package, targets *semanticTargets) {
	var concrete []namedTypeTarget
	var interfaces []interfaceTarget
	seenTypes := make(map[NodeKey]bool)
	for _, pkg := range packages {
		if pkg.Types == nil || !s.isInternalNamespace(pkg.Types.Path()) {
			continue
		}
		names := pkg.Types.Scope().Names()
		sort.Strings(names)
		for _, name := range names {
			typeName, ok := pkg.Types.Scope().Lookup(name).(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := types.Unalias(typeName.Type()).(*types.Named)
			if !ok {
				continue
			}
			key := s.ensureTypeNode(named)
			if seenTypes[key] {
				continue
			}
			seenTypes[key] = true
			if iface, ok := named.Underlying().(*types.Interface); ok {
				iface.Complete()
				interfaces = append(interfaces, interfaceTarget{named: named, iface: iface, key: key})
				s.ensureInterfaceMembers(named, iface, key, targets)
				continue
			}
			if structure, ok := named.Underlying().(*types.Struct); ok {
				s.collectEmbeddedTypeRelations(named, structure, key, *targets)
			}
			concrete = append(concrete, namedTypeTarget{named: named, key: key})
		}
	}

	for _, candidate := range concrete {
		for _, contract := range interfaces {
			receiver := types.Type(candidate.named)
			if !types.Implements(receiver, contract.iface) && !types.Implements(types.NewPointer(candidate.named), contract.iface) {
				continue
			}
			s.addEdge(candidate.key, contract.key, RelImplements, nil)
			s.linkImplementedMethods(candidate.named, contract, *targets)
		}
	}
}

func (s *scanner) collectEmbeddedTypeRelations(
	named *types.Named,
	structure *types.Struct,
	ownerKey NodeKey,
	targets semanticTargets,
) {
	for index := 0; index < structure.NumFields(); index++ {
		field := structure.Field(index)
		if !field.Embedded() {
			continue
		}
		embedded := field.Type()
		for {
			pointer, ok := embedded.(*types.Pointer)
			if !ok {
				break
			}
			embedded = pointer.Elem()
		}
		embeddedNamed, ok := types.Unalias(embedded).(*types.Named)
		if !ok {
			continue
		}
		embeddedKey := s.ensureTypeNode(embeddedNamed)
		if embeddedKey != ownerKey {
			s.addEdge(ownerKey, embeddedKey, RelExtends, nil)
		}
		methodSet := types.NewMethodSet(types.NewPointer(embeddedNamed))
		for index := 0; index < named.NumMethods(); index++ {
			method := named.Method(index)
			selection := methodSet.Lookup(method.Pkg(), method.Name())
			if selection == nil {
				continue
			}
			base, ok := selection.Obj().(*types.Func)
			if !ok {
				continue
			}
			overrider := s.internalFunctionCandidates(method, targets)
			baseTarget := s.internalFunctionCandidates(base, targets)
			if len(overrider) == 1 && len(baseTarget) == 1 && overrider[0] != baseTarget[0] {
				s.addEdge(overrider[0], baseTarget[0], RelOverrides, nil)
			}
		}
	}
}

func (s *scanner) ensureInterfaceMembers(
	named *types.Named,
	iface *types.Interface,
	interfaceKey NodeKey,
	targets *semanticTargets,
) {
	namespace := s.canonicalNamespace(objectNamespace(named.Obj()))
	for index := 0; index < iface.NumExplicitMethods(); index++ {
		method := iface.ExplicitMethod(index)
		if _, exists := targets.byObject[method]; exists {
			continue
		}
		key := hashIdentity(interfaceMethodIdentity(namespace, named.Obj().Name(), method.Name()))
		s.addNode(NodeFact{Key: key, Kind: KindMethod, Path: s.pathForNamespace(namespace), Name: method.Name()})
		s.addEdge(interfaceKey, key, RelDefines, nil)
		targets.add(method, interfaceMethodIdentity(namespace, named.Obj().Name(), method.Name()), key)
	}
	for index := 0; index < iface.NumEmbeddeds(); index++ {
		embedded := iface.EmbeddedType(index)
		embeddedKey := s.ensureTypeNode(embedded)
		s.addEdge(interfaceKey, embeddedKey, RelExtends, nil)
	}
}

func (s *scanner) linkImplementedMethods(candidate *types.Named, contract interfaceTarget, targets semanticTargets) {
	receiver := types.Type(candidate)
	if !types.Implements(receiver, contract.iface) {
		receiver = types.NewPointer(candidate)
	}
	methodSet := types.NewMethodSet(receiver)
	for index := 0; index < contract.iface.NumMethods(); index++ {
		interfaceMethod := contract.iface.Method(index)
		selection := methodSet.Lookup(interfaceMethod.Pkg(), interfaceMethod.Name())
		if selection == nil {
			continue
		}
		concreteMethod, ok := selection.Obj().(*types.Func)
		if !ok {
			continue
		}
		concreteCandidates := s.internalFunctionCandidates(concreteMethod, targets)
		interfaceKey, exists := targets.byObject[interfaceMethod]
		if len(concreteCandidates) == 1 && exists && concreteCandidates[0] != interfaceKey {
			s.addEdge(concreteCandidates[0], interfaceKey, RelImplements, nil)
			targets.interfaceImplementations[interfaceKey] = appendUniqueKey(
				targets.interfaceImplementations[interfaceKey], concreteCandidates[0],
			)
		}
	}
}
