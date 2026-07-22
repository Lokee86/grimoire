"""AST declaration and relationship evidence collection."""

from __future__ import annotations

import ast

from .contract import expression_text, span
from .model import CallInfo, Facts, FileContext, ImportInfo, InheritanceInfo


class DeclarationVisitor(ast.NodeVisitor):
    def __init__(self, facts: Facts, context: FileContext) -> None:
        self.facts = facts
        self.context = context
        self.class_stack: list[tuple[str, str]] = []
        self.function_stack: list[tuple[str, str]] = []
        self.owner_stack: list[str] = [context.module_id]
        self.import_index = 0

    @property
    def owner_id(self) -> str:
        return self.owner_stack[-1]

    @property
    def class_qname(self) -> str | None:
        return self.class_stack[-1][0] if self.class_stack else None

    def _attributes(self, node: ast.AST) -> dict[str, object]:
        decorators = sorted(expression_text(item, self.context.source) for item in getattr(node, "decorator_list", []))
        attributes: dict[str, object] = {}
        if decorators:
            attributes["decorators"] = decorators
        if isinstance(node, ast.AsyncFunctionDef):
            attributes["async"] = True
        return attributes

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        nested_names = [name for _, name in self.class_stack + self.function_stack]
        nested_names.append(node.name)
        qname = f"{self.context.module_name}.{'.'.join(nested_names)}"
        identifier = self.facts.add_node(
            "type",
            node.name,
            self.context.relative_path,
            qname,
            record_span=span(node, self.context.relative_path, self.context.lines),
            attributes={
                **self._attributes(node),
                **({"bases": sorted(expression_text(base, self.context.source) for base in node.bases)} if node.bases else {}),
            },
        )
        self.facts.symbols[qname] = identifier
        self.facts.symbol_kinds[qname] = "type"
        self.facts.add_edge(
            self.owner_id,
            identifier,
            "defines",
            record_span=span(node, self.context.relative_path, self.context.lines),
        )
        self.class_stack.append((qname, node.name))
        self.owner_stack.append(identifier)
        for base in node.bases:
            self._inheritance(identifier, qname, base)
        for statement in node.body:
            self.visit(statement)
        self.owner_stack.pop()
        self.class_stack.pop()

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        self._visit_function(node)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        self._visit_function(node)

    def _visit_function(self, node: ast.FunctionDef | ast.AsyncFunctionDef) -> None:
        prefix = [self.context.module_name]
        prefix.extend(name for _, name in self.class_stack)
        prefix.extend(name for _, name in self.function_stack)
        prefix.append(node.name)
        qname = ".".join(prefix)
        kind = "method" if self.class_stack and not self.function_stack else "function"
        identifier = self.facts.add_node(
            kind,
            node.name,
            self.context.relative_path,
            qname,
            record_span=span(node, self.context.relative_path, self.context.lines),
            attributes=self._attributes(node),
        )
        self.facts.symbols[qname] = identifier
        self.facts.symbol_kinds[qname] = kind
        self.facts.add_edge(
            self.owner_id,
            identifier,
            "defines",
            record_span=span(node, self.context.relative_path, self.context.lines),
        )
        self.function_stack.append((qname, node.name))
        self.owner_stack.append(identifier)
        for statement in node.body:
            self.visit(statement)
        self.owner_stack.pop()
        self.function_stack.pop()

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            self._add_import(node, alias.name, alias.asname or alias.name.split(".")[0], 0, None, False)

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        module = node.module or ""
        for alias in node.names:
            binding = None if alias.name == "*" else (alias.asname or alias.name)
            self._add_import(node, module, binding, node.level, alias.name, alias.name == "*")

    def _add_import(
        self,
        statement: ast.AST,
        target_module: str,
        binding: str | None,
        relative_level: int,
        target_name: str | None,
        star: bool,
    ) -> None:
        self.import_index += 1
        statement_span = span(statement, self.context.relative_path, self.context.lines)
        expression = expression_text(statement, self.context.source)
        if target_name and target_name != "*":
            expression = f"{expression} [{target_name}]"
        qualified_name = (
            f"{self.context.module_name}::import:{statement_span.get('start_line', 0) if statement_span else 0}:"
            f"{self.import_index}:{binding or target_module}"
        )
        identifier = self.facts.add_node(
            "import",
            binding or target_name or target_module,
            self.context.relative_path,
            qualified_name,
            identity=qualified_name,
            record_span=statement_span,
            attributes={"expression": expression},
        )
        self.facts.add_edge(self.owner_id, identifier, "defines", record_span=statement_span)
        self.facts.imports.append(
            ImportInfo(
                module_name=self.context.module_name,
                owner_id=self.owner_id,
                node_id=identifier,
                statement=statement,
                expression=expression,
                binding=binding,
                target_module=target_module,
                target_name=target_name,
                relative_level=relative_level,
                star=star,
                is_package=self.context.relative_path.endswith("/__init__.py")
                or self.context.relative_path == "__init__.py",
            )
        )
        if self.owner_id == self.context.module_id and binding:
            self.facts.module_bindings[(self.context.module_name, binding)] = (None, "unresolved")

    def visit_Call(self, node: ast.Call) -> None:
        self.facts.calls.append(
            CallInfo(
                module_name=self.context.module_name,
                owner_id=self.owner_id,
                class_qname=self.class_qname,
                expression_node=node,
                callee=node.func,
            )
        )
        self.generic_visit(node)

    def _inheritance(self, source_id: str, class_qname: str, base: ast.expr) -> None:
        self.facts.inheritances.append(
            InheritanceInfo(
                source_id=source_id,
                module_name=self.context.module_name,
                class_qname=class_qname,
                base=base,
                source=self.context.source,
                path=self.context.relative_path,
                lines=self.context.lines,
            )
        )
