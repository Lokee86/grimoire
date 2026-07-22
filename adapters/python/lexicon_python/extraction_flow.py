"""Python extraction visitor layer."""

from __future__ import annotations

import ast

from .model import CallInfo, LocalAssignmentInfo, LoopBindingInfo
from .extraction_targets import target_bindings, target_name


class LocalFlow:
    def _record_local_write(
        self,
        target: ast.expr,
        node: ast.AST,
        value: ast.expr | None = None,
        annotation: ast.expr | None = None,
    ) -> None:
        if self.scope_id is None:
            return
        name = target_name(target)
        if name is None:
            return
        self.facts.local_assignments.append(
            LocalAssignmentInfo(
                module_name=self.context.module_name,
                scope_id=self.scope_id,
                class_qname=self.class_qname,
                name=name,
                assignment_node=node,
                value=value,
                annotation=annotation,
                branch_dependent=self.control_flow_depth > 0,
            )
        )

    def visit_Assign(self, node: ast.Assign) -> None:
        for target in node.targets:
            self._record_local_write(target, node, node.value)
        self.generic_visit(node)

    def visit_AnnAssign(self, node: ast.AnnAssign) -> None:
        self._record_local_write(node.target, node, node.value, node.annotation)
        self.generic_visit(node)

    def visit_AugAssign(self, node: ast.AugAssign) -> None:
        self._record_local_write(node.target, node)
        self.generic_visit(node)

    def visit_NamedExpr(self, node: ast.NamedExpr) -> None:
        self._record_local_write(node.target, node, node.value)
        self.generic_visit(node)

    def _record_loop_targets(self, node: ast.For | ast.AsyncFor, branch_dependent: bool) -> None:
        if self.scope_id is None:
            return
        for name, element_index in target_bindings(node.target):
            self.facts.loop_bindings.append(
                LoopBindingInfo(
                    self.context.module_name,
                    self.scope_id,
                    self.class_qname,
                    name,
                    node,
                    node.iter,
                    branch_dependent,
                    element_index,
                )
            )

    def _visit_branch(self, nodes: list[ast.stmt]) -> None:
        self.control_flow_depth += 1
        for statement in nodes:
            self.visit(statement)
        self.control_flow_depth -= 1

    def visit_If(self, node: ast.If) -> None:
        self.visit(node.test)
        self._visit_branch(node.body)
        self._visit_branch(node.orelse)

    def visit_For(self, node: ast.For) -> None:
        self.visit(node.iter)
        self._record_loop_targets(node, self.control_flow_depth > 0)
        self._visit_branch(node.body)
        self._visit_branch(node.orelse)

    def visit_AsyncFor(self, node: ast.AsyncFor) -> None:
        self.visit(node.iter)
        self._record_loop_targets(node, self.control_flow_depth > 0)
        self._visit_branch(node.body)
        self._visit_branch(node.orelse)

    def visit_While(self, node: ast.While) -> None:
        self.visit(node.test)
        self._visit_branch(node.body)
        self._visit_branch(node.orelse)

    def visit_Try(self, node: ast.Try) -> None:
        self._visit_branch(node.body)
        for handler in node.handlers:
            self._visit_branch([handler])
        self._visit_branch(node.orelse)
        self._visit_branch(node.finalbody)

    def visit_Match(self, node: ast.Match) -> None:
        self.visit(node.subject)
        for case in node.cases:
            if case.guard is not None:
                self.visit(case.guard)
            self._visit_branch(case.body)

    def _visit_comprehension(
        self,
        node: ast.ListComp | ast.SetComp | ast.GeneratorExp | ast.DictComp,
        values: list[ast.expr],
    ) -> None:
        if self.scope_id is not None:
            for generator in node.generators:
                for name, element_index in target_bindings(generator.target):
                    self.facts.loop_bindings.append(
                        LoopBindingInfo(
                            self.context.module_name,
                            self.scope_id,
                            self.class_qname,
                            name,
                            node,
                            generator.iter,
                            self.control_flow_depth > 0,
                            element_index,
                        )
                    )
        for generator in node.generators:
            self.visit(generator.iter)
            for condition in generator.ifs:
                self.visit(condition)
        for value in values:
            self.visit(value)

    def visit_ListComp(self, node: ast.ListComp) -> None:
        self._visit_comprehension(node, [node.elt])

    def visit_SetComp(self, node: ast.SetComp) -> None:
        self._visit_comprehension(node, [node.elt])

    def visit_GeneratorExp(self, node: ast.GeneratorExp) -> None:
        self._visit_comprehension(node, [node.elt])

    def visit_DictComp(self, node: ast.DictComp) -> None:
        self._visit_comprehension(node, [node.key, node.value])

    def visit_Call(self, node: ast.Call) -> None:
        self.facts.calls.append(
            CallInfo(
                module_name=self.context.module_name,
                owner_id=self.owner_id,
                class_qname=self.class_qname,
                scope_id=self.scope_id,
                expression_node=node,
                callee=node.func,
            )
        )
        self.generic_visit(node)
