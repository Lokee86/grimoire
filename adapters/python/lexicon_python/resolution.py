"""Static import, inheritance, and call resolution."""

from __future__ import annotations

import ast

from .contract import BUILTINS, expression_text, span
from .model import CallInfo, Facts, FileContext, ImportInfo


def _dotted(node: ast.AST) -> str | None:
    if isinstance(node, ast.Name):
        return node.id
    if isinstance(node, ast.Attribute):
        parent = _dotted(node.value)
        return f"{parent}.{node.attr}" if parent else None
    return None


def resolve_relative_module(info: ImportInfo) -> str:
    if info.relative_level == 0:
        return info.target_module or ""
    current_parts = info.module_name.split(".") if info.module_name else []
    package_parts = current_parts if info.is_package else current_parts[:-1]
    remove = max(info.relative_level - 1, 0)
    if remove:
        package_parts = package_parts[:-remove] if remove <= len(package_parts) else []
    suffix = (info.target_module or "").split(".") if info.target_module else []
    return ".".join(package_parts + suffix)


def resolve_reference(
    facts: Facts,
    module_name: str,
    class_qname: str | None,
    reference: str | None,
) -> tuple[str | None, str]:
    if not reference:
        return None, "unsupported-form"
    if reference in BUILTINS:
        return None, "builtin-target"
    parts = reference.split(".")
    if class_qname and parts[0] in {"self", "cls"} and len(parts) == 2:
        candidate = f"{class_qname}.{parts[1]}"
        if candidate in facts.symbols:
            return facts.symbols[candidate], ""
    binding = facts.module_bindings.get((module_name, parts[0]))
    if binding and binding[0]:
        base_qname = next((qname for qname, value in facts.symbols.items() if value == binding[0]), None)
        if base_qname:
            candidate = ".".join([base_qname, *parts[1:]])
            if candidate in facts.symbols:
                return facts.symbols[candidate], ""
        if len(parts) == 1:
            return binding[0], ""
    candidates = [".".join([module_name, *parts]), reference]
    if class_qname and len(parts) == 1:
        candidates.insert(0, f"{class_qname}.{reference}")
    for candidate in candidates:
        if candidate in facts.symbols:
            return facts.symbols[candidate], ""
        if candidate in facts.modules:
            return facts.modules[candidate], ""
    if binding and binding[1] == "external":
        return None, "external-target"
    return None, "missing-target"


def resolve_import(facts: Facts, info: ImportInfo) -> tuple[str | None, str]:
    if info.star:
        return None, "unsupported-form"
    module_name = resolve_relative_module(info)
    if info.target_name is None:
        target = facts.modules.get(module_name)
        return (target, "") if target else (None, "external-target")
    symbol = f"{module_name}.{info.target_name}" if module_name else info.target_name
    if symbol in facts.symbols:
        return facts.symbols[symbol], ""
    if symbol in facts.modules:
        return facts.modules[symbol], ""
    if module_name in facts.modules:
        return None, "missing-target"
    return None, "external-target"


def _call_resolution(facts: Facts, call: CallInfo) -> tuple[str | None, str]:
    reference = _dotted(call.callee)
    if reference is None:
        return None, "dynamic-target"
    if reference in {"importlib.import_module", "import_module", "__import__"}:
        return None, "dynamic-target"
    return resolve_reference(facts, call.module_name, call.class_qname, reference)


def _import_span(facts: Facts, node_id: str) -> dict[str, object] | None:
    return facts.nodes.get(node_id, {}).get("span")


def _source_for_call(call: CallInfo, contexts: list[FileContext]) -> str:
    for context in contexts:
        if context.module_name == call.module_name:
            return context.source
    return ""


def _call_span(call: CallInfo, contexts: list[FileContext]) -> dict[str, object] | None:
    for context in contexts:
        if context.module_name == call.module_name:
            return span(call.expression_node, context.relative_path, context.lines)
    return None


def resolve_facts(facts: Facts, contexts: list[FileContext]) -> None:
    for info in facts.imports:
        target_id, reason = resolve_import(facts, info)
        if info.owner_id == facts.modules.get(info.module_name) and info.binding:
            facts.module_bindings[(info.module_name, info.binding)] = (
                target_id,
                "external" if target_id is None and reason == "external-target" else reason,
            )
        import_span = _import_span(facts, info.node_id)
        if target_id:
            facts.add_edge(info.owner_id, target_id, "imports", record_span=import_span)
        else:
            facts.add_unresolved(
                info.owner_id,
                "imports",
                info.expression,
                reason,
                record_span=import_span,
                candidate_name=resolve_relative_module(info) or info.target_module,
            )

    for inheritance in facts.inheritances:
        target_name = _dotted(inheritance.base)
        target_id, reason = resolve_reference(
            facts,
            inheritance.module_name,
            inheritance.class_qname,
            target_name,
        )
        base_span = span(inheritance.base, inheritance.path, inheritance.lines)
        if target_id:
            facts.add_edge(inheritance.source_id, target_id, "extends", record_span=base_span)
        else:
            facts.add_unresolved(
                inheritance.source_id,
                "extends",
                expression_text(inheritance.base, inheritance.source),
                reason,
                record_span=base_span,
                candidate_name=target_name,
            )

    for call in facts.calls:
        target_id, reason = _call_resolution(facts, call)
        call_span = _call_span(call, contexts)
        if target_id:
            facts.add_edge(call.owner_id, target_id, "calls", record_span=call_span)
        else:
            facts.add_unresolved(
                call.owner_id,
                "calls",
                expression_text(call.expression_node, _source_for_call(call, contexts)),
                reason,
                record_span=call_span,
                candidate_name=_dotted(call.callee),
            )
