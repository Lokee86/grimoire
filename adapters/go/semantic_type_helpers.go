package main

import (
	"go/types"
	"strings"
)

func isStandardLibraryNamespace(namespace string) bool {
	if namespace == "" || strings.Contains(namespace, ":") {
		return false
	}
	first := namespace
	if slash := strings.IndexByte(first, '/'); slash >= 0 {
		first = first[:slash]
	}
	return !strings.Contains(first, ".")
}

func isInterfaceType(value types.Type) bool {
	value = types.Unalias(value)
	if pointer, ok := value.(*types.Pointer); ok {
		value = pointer.Elem()
	}
	if named, ok := value.(*types.Named); ok {
		value = named.Underlying()
	}
	_, ok := value.(*types.Interface)
	return ok
}

func receiverTypeName(value types.Type) string {
	value = types.Unalias(value)
	if pointer, ok := value.(*types.Pointer); ok {
		return "*" + receiverTypeName(pointer.Elem())
	}
	if named, ok := value.(*types.Named); ok {
		return named.Obj().Name()
	}
	return types.TypeString(value, func(*types.Package) string { return "" })
}

func objectNamespace(object types.Object) string {
	if object == nil || object.Pkg() == nil {
		return ""
	}
	return object.Pkg().Path()
}
