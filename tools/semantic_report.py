#!/usr/bin/env python3
"""Summarize semantic relationship coverage in one or more Lexicon fact streams."""

from __future__ import annotations

import argparse
import importlib.util
import json
from collections import Counter
from pathlib import Path
from typing import Any

VALIDATOR_PATH = Path(__file__).with_name("validate_jsonl.py")
SPEC = importlib.util.spec_from_file_location("validate_jsonl", VALIDATOR_PATH)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError("cannot load Lexicon validator")
VALIDATOR = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VALIDATOR)

DISPATCH_RELATIONS = {"extends", "implements", "uses-trait", "includes", "overrides"}


def load_stream(path: Path) -> list[dict[str, Any]]:
    VALIDATOR.validate(path)
    return [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines()]


def ratio(numerator: int, denominator: int) -> float | None:
    if denominator == 0:
        return None
    return round(numerator / denominator, 6)


def summarize(path: Path) -> dict[str, Any]:
    records = load_stream(path)
    header = records[0]
    nodes = Counter()
    edges = Counter()
    unresolved = Counter()
    for record in records[1:]:
        kind = record["record"]
        if kind == "node":
            nodes[record["kind"]] += 1
        elif kind == "edge":
            edges[record["relation"]] += 1
        else:
            unresolved[record["relation"]] += 1

    call_definite = edges["calls"]
    call_possible = edges["possible-calls"]
    call_unresolved = unresolved["calls"]
    call_total = call_definite + call_possible + call_unresolved
    dataflow_edges = edges["reads"] + edges["writes"]
    dispatch_edges = sum(edges[relation] for relation in DISPATCH_RELATIONS)

    return {
        "adapter_version": header["adapter_version"],
        "call_quality": {
            "definite_fraction": ratio(call_definite, call_total),
            "resolved_fraction": ratio(call_definite + call_possible, call_total),
            "total": call_total,
        },
        "capabilities": {
            "dataflow": dataflow_edges > 0,
            "dependencies": edges["depends-on"] > 0,
            "dispatch_relationships": dispatch_edges > 0,
            "runtime_evidence": False,
        },
        "edge_relations": dict(sorted(edges.items())),
        "language": header["language"],
        "node_kinds": dict(sorted(nodes.items())),
        "path": path.as_posix(),
        "repository": header["repository"],
        "unresolved_relations": dict(sorted(unresolved.items())),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("facts", nargs="+", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    try:
        reports = [summarize(path) for path in args.facts]
    except (OSError, ValueError, json.JSONDecodeError) as error:
        parser.error(str(error))
    result = {"streams": sorted(reports, key=lambda item: (item["language"], item["repository"], item["path"]))}
    rendered = json.dumps(result, indent=2, sort_keys=True) + "\n"
    if args.output:
        args.output.write_text(rendered, encoding="utf-8")
    else:
        print(rendered, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
