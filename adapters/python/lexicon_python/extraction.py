"""AST declaration and relationship evidence collection."""

from __future__ import annotations

import ast

from .contract import expression_text
from .extraction_declarations import DeclarationFlow
from .extraction_flow import LocalFlow
from .extraction_imports import ImportFlow
from .model import Facts, FileContext


class DeclarationVisitor(
    DeclarationFlow,
    ImportFlow,
    LocalFlow,
    ast.NodeVisitor,
):
    def __init__(self, facts: Facts, context: FileContext) -> None:
        self.facts = facts
        self.context = context
        self.class_stack: list[tuple[str, str]] = []
        self.function_stack: list[tuple[str, str]] = []
        self.lexical_stack: list[tuple[str, str]] = []
        self.owner_stack: list[str] = [context.module_id]
        self.import_index = 0
        self.control_flow_depth = 0

    @property
    def owner_id(self) -> str:
        return self.owner_stack[-1]

    @property
    def class_qname(self) -> str | None:
        return self.class_stack[-1][0] if self.class_stack else None

    @property
    def scope_id(self) -> str:
        return self.owner_id

    def _attributes(self, node: ast.AST) -> dict[str, object]:
        decorators = sorted(expression_text(item, self.context.source) for item in getattr(node, "decorator_list", []))
        attributes: dict[str, object] = {}
        if decorators:
            attributes["decorators"] = decorators
        if isinstance(node, ast.AsyncFunctionDef):
            attributes["async"] = True
        return attributes
