from __future__ import annotations

import importlib.util
import json
from pathlib import Path
from typing import Any

MODULE_PATH = Path(__file__).with_name("semantic_depth.py")
SPEC = importlib.util.spec_from_file_location("semantic_depth", MODULE_PATH)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


def identity(char: str) -> str:
    return "sha256:" + char * 64


def node(char: str, kind: str, name: str, path: str, qualified: str, **attributes: Any) -> dict[str, Any]:
    record: dict[str, Any] = {
        "id": identity(char),
        "kind": kind,
        "name": name,
        "path": path,
        "qualified_name": qualified,
        "record": "node",
    }
    if attributes:
        record["attributes"] = attributes
    return record


def edge(source: str, target: str, relation: str, **attributes: Any) -> dict[str, Any]:
    record: dict[str, Any] = {"record": "edge", "relation": relation, "source": identity(source), "target": identity(target)}
    if attributes:
        record["attributes"] = attributes
    return record


def write_stream(path: Path, suffix: str = "") -> None:
    source = node("a", "function", "run", "src.c", "run", definition=True)
    target = node("b", "function", "work", "src.c", "work", definition=True)
    prototype = node("c", "function", "declared", "api.h", "declared", definition=False)
    field = node("d", "field", "field", "src.c", "Thing::field")
    local = node("e", "variable", "local", "src.c", "run::local")
    parameter = node("f", "parameter", "value", "src.c", "run::value")
    nodes = [source, target, prototype, field, local, parameter]
    records: list[dict[str, Any]] = [
        {"adapter_version": "0.4.0", "language": "c-family", "record": "lexicon", "repository": "fixture", "schema_version": 1},
        *nodes,
        edge("a", "e", "defines"),
        edge("a", "f", "defines"),
        edge("a", "b", "calls"),
        edge("a", "a", "possible-calls", indirect="function-pointer"),
        edge("a", "b", "calls", indirect="macro"),
        edge("a", "f", "passes-to"),
        edge("a", "d", "reads"),
        edge("a", "d", "writes"),
        {"record": "unresolved", "relation": "calls", "source": identity("a"), "expression": "dynamic", "reason": "dynamic-target"},
        {"record": "unresolved", "relation": "calls", "source": identity("a"), "expression": "missing", "reason": "missing-target"},
    ]
    if suffix:
        for record in nodes:
            if record["record"] == "node":
                record["path"] = record["path"].replace("src", suffix)
                record["qualified_name"] = record["qualified_name"].replace("work", "other")
    path.write_text("".join(json.dumps(record, sort_keys=True) + "\n" for record in records), encoding="utf-8")


def test_reports_c_family_semantic_depth(tmp_path: Path) -> None:
    facts = tmp_path / "facts.jsonl"
    write_stream(facts)
    report = MODULE.summarize(facts)

    assert report["nodes"] == {
        "definitions": {"occurrences": 2},
        "prototypes": {"occurrences": 1},
        "fields": {"occurrences": 1},
        "locals": {"occurrences": 1},
        "parameters": {"occurrences": 1},
    }
    assert report["relations"]["calls"] == {"occurrences": 2, "unique_source_target_pairs": 1}
    assert report["relations"]["passes-to"] == {"occurrences": 1, "unique_source_target_pairs": 1}
    assert report["evidence"]["direct_calls"]["occurrences"] == 1
    assert report["evidence"]["possible_calls"]["occurrences"] == 1
    assert report["evidence"]["indirect_calls"]["occurrences"] == 2
    assert report["evidence"]["macro_body_calls"]["occurrences"] == 1
    assert report["evidence"]["self_recursion"]["occurrences"] == 1
    assert report["unresolved"]["by_reason"] == {
        "dynamic-target": {"occurrences": 1, "unique_sources": 1},
        "missing-target": {"occurrences": 1, "unique_sources": 1},
    }


def test_comparison_uses_source_and_target_names(tmp_path: Path) -> None:
    left = tmp_path / "left.jsonl"
    right = tmp_path / "right.jsonl"
    write_stream(left)
    write_stream(right, "other")
    result = MODULE.compare(left, right)

    assert result["shared"]["count"] == 0
    assert result["left_only"]["count"] > 0
    tuple_value = result["left_only"]["tuples"][0]
    assert "sha256:" not in json.dumps(tuple_value)
    assert set(tuple_value["source"]) == {"path", "qualified_name"}
    assert set(tuple_value["target"]) == {"path", "qualified_name"}
