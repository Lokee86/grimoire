"""Shared call-graph value shapes and ordering helpers."""

from __future__ import annotations

import ast
from dataclasses import dataclass

_SEQUENCE_TYPES = {
    "AsyncIterable",
    "AsyncIterator",
    "Collection",
    "Generator",
    "Iterable",
    "Iterator",
    "List",
    "Sequence",
    "Set",
    "Tuple",
    "list",
    "set",
    "tuple",
    "frozenset",
}
_MAPPING_TYPES = {"Dict", "Mapping", "MutableMapping", "dict"}
_UNION_TYPES = {"Annotated", "Optional", "Union"}
_WRAPPER_TYPES = {"ClassVar", "Final", "Required", "NotRequired", "Type", "type"}
_CONTAINER_ACCESSORS = {"get", "pop", "setdefault"}
_CALLABLE_KINDS = {"function", "method", "type"}
_SEMANTIC_DECORATORS = {
    "abstractmethod",
    "cached_property",
    "classmethod",
    "dataclass",
    "final",
    "overload",
    "override",
    "property",
    "staticmethod",
}


@dataclass(frozen=True)
class TypeShape:
    direct: frozenset[str] = frozenset()
    elements: frozenset[str] = frozenset()
    callables: frozenset[str] = frozenset()
    element_callables: frozenset[str] = frozenset()

    def merge(self, other: "TypeShape") -> "TypeShape":
        return TypeShape(
            self.direct | other.direct,
            self.elements | other.elements,
            self.callables | other.callables,
            self.element_callables | other.element_callables,
        )

    def element_shape(self) -> "TypeShape":
        return TypeShape(direct=self.elements, callables=self.element_callables)


_EMPTY = TypeShape()


def _position(node: ast.AST, *, end: bool = False) -> tuple[int, int]:
    line_name = "end_lineno" if end else "lineno"
    column_name = "end_col_offset" if end else "col_offset"
    return (getattr(node, line_name, 0), getattr(node, column_name, 0))


def _precedes(node: ast.AST, before: ast.AST) -> bool:
    return _position(node, end=True) <= _position(before)


def _merge_shapes(shapes: list[TypeShape]) -> TypeShape:
    result = _EMPTY
    for shape in shapes:
        result = result.merge(shape)
    return result


def _elements_from_shapes(shapes: list[TypeShape]) -> TypeShape:
    return TypeShape(
        elements=frozenset(
            identifier
            for shape in shapes
            for identifier in (*shape.direct, *shape.elements)
        ),
        element_callables=frozenset(
            identifier
            for shape in shapes
            for identifier in (*shape.callables, *shape.element_callables)
        ),
    )


def _return_expressions(
    node: ast.FunctionDef | ast.AsyncFunctionDef | ast.Lambda,
) -> list[ast.expr]:
    if isinstance(node, ast.Lambda):
        return [node.body]
    expressions: list[ast.expr] = []

    class Visitor(ast.NodeVisitor):
        def visit_Return(self, return_node: ast.Return) -> None:
            if return_node.value is not None:
                expressions.append(return_node.value)

        def visit_FunctionDef(self, nested: ast.FunctionDef) -> None:
            if nested is node:
                self.generic_visit(nested)

        def visit_AsyncFunctionDef(self, nested: ast.AsyncFunctionDef) -> None:
            if nested is node:
                self.generic_visit(nested)

        def visit_Lambda(self, nested: ast.Lambda) -> None:
            return

        def visit_ClassDef(self, nested: ast.ClassDef) -> None:
            return

    Visitor().visit(node)
    return expressions
