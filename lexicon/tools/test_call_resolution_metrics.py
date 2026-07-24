from __future__ import annotations

import importlib.util
import json
import subprocess
import sys
from pathlib import Path
from typing import Any

MODULE_PATH = Path(__file__).with_name("call_resolution_metrics.py")
SPEC = importlib.util.spec_from_file_location("call_resolution_metrics", MODULE_PATH)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


def identity(char: str) -> str:
    return "sha256:" + char * 64


def span(path: str, line: int) -> dict[str, Any]:
    return {
        "end_column": 15,
        "end_line": line,
        "path": path,
        "start_column": 3,
        "start_line": line,
    }


def write_stream(path: Path) -> None:
    sources = {
        char: {"id": identity(char), "kind": "function", "name": f"run_{char}", "path": f"src/{char}.c", "qualified_name": f"run_{char}", "record": "node"}
        for char in "abcde"
    }
    targets = {
        char: {"id": identity(char), "kind": "function", "name": f"target_{char}", "path": f"lib/{char}.c", "qualified_name": f"target_{char}", "record": "node"}
        for char in "1234"
    }
    records: list[dict[str, Any]] = [
        {"adapter_version": "0.3.0", "language": "c-family", "record": "lexicon", "repository": "example/repo", "schema_version": 1},
        *sources.values(),
        *targets.values(),
    ]

    def edge(source: str, target: str, relation: str, line: int) -> dict[str, Any]:
        return {"owner": f"src/{source}.c", "record": "edge", "relation": relation, "source": identity(source), "span": span(f"src/{source}.c", line), "target": identity(target)}

    records += [
        edge("a", "1", "calls", 10),
        edge("b", "1", "possible-calls", 20),
        edge("b", "2", "possible-calls", 20),
        edge("c", "1", "possible-calls", 30),
        edge("c", "2", "calls", 30),
        edge("e", "1", "possible-calls", 50),
        edge("e", "2", "possible-calls", 50),
        edge("e", "3", "possible-calls", 50),
        edge("e", "4", "possible-calls", 50),
        {"expression": "ambiguous()", "owner": "src/b.c", "reason": "ambiguous-target", "record": "unresolved", "relation": "calls", "source": identity("b"), "span": span("src/b.c", 20)},
        {"expression": "dynamic()", "owner": "src/c.c", "reason": "dynamic-target", "record": "unresolved", "relation": "calls", "source": identity("c"), "span": span("src/c.c", 30)},
        {"expression": "missing()", "owner": "src/d.c", "reason": "missing-target", "record": "unresolved", "relation": "calls", "source": identity("d"), "span": span("src/d.c", 40)},
    ]
    node_records = records[1 : 1 + len(sources) + len(targets)]
    fact_records = records[1 + len(node_records) :]
    fact_records.sort(key=lambda record: MODULE.record_sort_key(record))
    path.write_text("".join(json.dumps(record, sort_keys=True, separators=(",", ":")) + "\n" for record in [records[0], *sorted(node_records, key=MODULE.record_sort_key), *fact_records]), encoding="utf-8")


def test_aggregates_call_sites_and_fanout_by_span(tmp_path: Path) -> None:
    facts = tmp_path / "facts.jsonl"
    write_stream(facts)

    report = MODULE.summarize(facts)

    assert report["call_sites"] == {
        "total": 5,
        "definite_only": 1,
        "possible_only": 2,
        "definite_plus_possible": 1,
        "unresolved_only": 1,
    }
    assert report["unresolved_reason_counts"] == {
        "ambiguous-target": 1,
        "dynamic-target": 1,
        "missing-target": 1,
    }
    assert report["possible_target_fanout"] == {
        "site_count": 3,
        "percentiles": {"p50": 2, "p75": 4, "p90": 4, "p95": 4, "p99": 4},
    }
    highest = report["highest_fanout_call_sites"]
    assert [site["possible_target_count"] for site in highest] == [4, 2, 1]
    assert highest[0]["source_path"] == "src/e.c"
    assert highest[0]["span"] == span("src/e.c", 50)


def test_cli_output_is_stable_and_limit_is_applied(tmp_path: Path) -> None:
    facts = tmp_path / "facts.jsonl"
    write_stream(facts)
    command = [sys.executable, str(MODULE_PATH), str(facts), "--limit", "1"]
    first = subprocess.run(command, check=True, capture_output=True, text=True).stdout
    second = subprocess.run(command, check=True, capture_output=True, text=True).stdout

    assert first == second
    result = json.loads(first)
    assert list(result) == sorted(result)
    assert len(result["highest_fanout_call_sites"]) == 1