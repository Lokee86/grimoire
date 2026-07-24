"""Assignment and loop-target decomposition helpers."""

from __future__ import annotations

import ast


def target_name(target: ast.expr) -> str | None:
    if isinstance(target, ast.Name):
        return target.id
    if isinstance(target, ast.Attribute):
        parent = target_name(target.value)
        return f"{parent}.{target.attr}" if parent else None
    return None

def target_names(target: ast.expr) -> list[str]:
    return [name for name, _ in target_bindings(target)]

def target_bindings(target: ast.expr) -> list[tuple[str, int | None]]:
    if isinstance(target, ast.Name):
        return [(target.id, None)]
    if isinstance(target, (ast.Tuple, ast.List)):
        return [
            (name, index)
            for index, item in enumerate(target.elts)
            for name, _ in target_bindings(item)
        ]
    return []
