"""Static import, inheritance, and call resolution."""

from __future__ import annotations

from .bindings import BindingResolver, dotted, resolve_relative_module
from .callgraph import CallGraphResolver
from .contract import expression_text, span
from .model import CallInfo, Facts, FileContext


def _import_span(facts: Facts, node_id: str) -> dict[str, object] | None:
    return facts.nodes.get(node_id, {}).get("span")


def _context_for_call(call: CallInfo, contexts: list[FileContext]) -> FileContext | None:
    return next((context for context in contexts if context.module_name == call.module_name), None)


def _call_span(call: CallInfo, contexts: list[FileContext]) -> dict[str, object] | None:
    context = _context_for_call(call, contexts)
    return span(call.expression_node, context.relative_path, context.lines) if context else None


def _call_expression(call: CallInfo, contexts: list[FileContext]) -> str:
    context = _context_for_call(call, contexts)
    return expression_text(call.expression_node, context.source if context else "")


def resolve_facts(facts: Facts, contexts: list[FileContext]) -> None:
    bindings = BindingResolver(facts)
    import_results = bindings.resolve_imports()
    for info, (target_id, reason) in zip(facts.imports, import_results):
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
        target_name = dotted(inheritance.base)
        target_id, reason = bindings.resolve_reference(
            inheritance.module_name,
            inheritance.class_qname,
            target_name,
        )
        base_span = span(inheritance.base, inheritance.path, inheritance.lines)
        if target_id:
            relation = "implements" if facts.nodes.get(target_id, {}).get("kind") in {"interface", "trait"} else "extends"
            facts.add_edge(inheritance.source_id, target_id, relation, record_span=base_span)
        else:
            facts.add_unresolved(
                inheritance.source_id,
                "extends",
                expression_text(inheritance.base, inheritance.source),
                reason,
                record_span=base_span,
                candidate_name=target_name,
            )

    _emit_overrides(facts, bindings)

    callgraph = CallGraphResolver(facts, bindings)
    for call in facts.calls:
        target_ids, reason = callgraph.resolve_call(call)
        call_span = _call_span(call, contexts)
        if len(target_ids) == 1:
            facts.add_edge(call.owner_id, target_ids[0], "calls", record_span=call_span)
        elif target_ids:
            for target_id in target_ids:
                facts.add_edge(
                    call.owner_id,
                    target_id,
                    "possible-calls",
                    record_span=call_span,
                    attributes={"candidate_count": len(target_ids)},
                )
        else:
            facts.add_unresolved(
                call.owner_id,
                "calls",
                _call_expression(call, contexts),
                reason,
                record_span=call_span,
                candidate_name=dotted(call.callee),
            )


def _emit_overrides(facts: Facts, bindings: BindingResolver) -> None:
    for function in facts.functions.values():
        if not function.class_qname:
            continue
        method_name = function.qname.rsplit(".", 1)[-1]
        for ancestor in _ancestor_qnames(facts, bindings, function.module_name, function.class_qname):
            target = facts.symbols.get(f"{ancestor}.{method_name}")
            if target and facts.nodes.get(target, {}).get("kind") == "method" and target != function.node_id:
                facts.add_edge(function.node_id, target, "overrides")


def _ancestor_qnames(
    facts: Facts,
    bindings: BindingResolver,
    module_name: str,
    class_qname: str,
    seen: set[str] | None = None,
) -> tuple[str, ...]:
    seen = set() if seen is None else seen
    if class_qname in seen:
        return ()
    seen.add(class_qname)
    info = facts.classes.get(class_qname)
    if info is None:
        return ()
    result: list[str] = []
    for base in info.node.bases:
        target_id, _ = bindings.resolve_reference(module_name, class_qname, dotted(base))
        target_qname = facts.node_qnames.get(target_id) if target_id else None
        if not target_qname:
            continue
        result.append(target_qname)
        result.extend(_ancestor_qnames(facts, bindings, module_name, target_qname, seen))
    return tuple(dict.fromkeys(result))
