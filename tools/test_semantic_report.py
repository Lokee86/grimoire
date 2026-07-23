from __future__ import annotations

import importlib.util
import json
from pathlib import Path

MODULE_PATH = Path(__file__).with_name("semantic_report.py")
SPEC = importlib.util.spec_from_file_location("semantic_report", MODULE_PATH)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


def identity(char: str) -> str:
    return "sha256:" + char * 64


def write_stream(path: Path) -> None:
    source = identity("a")
    target = identity("b")
    data = identity("c")
    records = [
        {"adapter_version": "0.1.0", "language": "go", "record": "lexicon", "repository": "example/repo", "schema_version": 1},
        {"id": source, "kind": "function", "name": "run", "path": "main.go", "qualified_name": "run", "record": "node"},
        {"id": target, "kind": "function", "name": "work", "path": "main.go", "qualified_name": "work", "record": "node"},
        {"id": data, "kind": "variable", "name": "value", "path": "main.go", "qualified_name": "value", "record": "node"},
        {"record": "edge", "relation": "calls", "source": source, "target": target},
        {"record": "edge", "relation": "reads", "source": source, "target": data},
        {"expression": "dynamic()", "reason": "dynamic-target", "record": "unresolved", "relation": "calls", "source": source},
    ]
    path.write_text("".join(json.dumps(record, sort_keys=True, separators=(",", ":")) + "\n" for record in records), encoding="utf-8")


def test_summarizes_semantic_capabilities(tmp_path: Path) -> None:
    facts = tmp_path / "facts.jsonl"
    write_stream(facts)
    report = MODULE.summarize(facts)
    assert report["edge_relations"] == {"calls": 1, "reads": 1}
    assert report["call_quality"] == {"definite_fraction": 0.5, "resolved_fraction": 0.5, "total": 2}
    assert report["capabilities"]["dataflow"] is True
    assert report["capabilities"]["dependencies"] is False
