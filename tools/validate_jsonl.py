#!/usr/bin/env python3
"""Validate the structural invariants of Lexicon facts v1."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from typing import Any

ID_PATTERN = re.compile(r"^sha256:[0-9a-f]{64}$")
RECORD_ORDER = {"node": 0, "edge": 1, "unresolved": 2}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def span_key(record: dict[str, Any]) -> tuple[Any, ...]:
    span = record.get("span") or {}
    return (
        span.get("path", ""),
        span.get("start_line", 0),
        span.get("start_column", 0),
        span.get("end_line", 0),
        span.get("end_column", 0),
    )


def sort_key(record: dict[str, Any]) -> tuple[Any, ...]:
    kind = record["record"]
    if kind == "node":
        return (0, record["id"], record["kind"], record["path"], record["qualified_name"])
    if kind == "edge":
        return (1, record["source"], record["target"], record["relation"], *span_key(record))
    return (
        2,
        record["source"],
        record["relation"],
        record["expression"],
        record["reason"],
        *span_key(record),
    )


def validate(path: Path) -> None:
    lines = path.read_text(encoding="utf-8").splitlines()
    require(bool(lines), "fact stream is empty")
    records = [json.loads(line) for line in lines]
    header = records[0]
    require(header.get("record") == "lexicon", "first record must be the Lexicon header")
    require(header.get("schema_version") == 1, "schema_version must be 1")
    for field in ("adapter_version", "language", "repository"):
        require(isinstance(header.get(field), str) and bool(header[field]), f"header.{field} is required")

    facts = records[1:]
    seen_nodes: set[str] = set()
    for index, record in enumerate(facts, start=2):
        record_kind = record.get("record")
        require(record_kind in RECORD_ORDER, f"line {index}: invalid record kind")
        if record_kind == "node":
            for field in ("id", "kind", "name", "path", "qualified_name"):
                require(field in record, f"line {index}: node.{field} is required")
            require(ID_PATTERN.match(record["id"]) is not None, f"line {index}: invalid node id")
            require(record["id"] not in seen_nodes, f"line {index}: duplicate node id")
            seen_nodes.add(record["id"])
            content_id = record.get("content_id")
            require(content_id is None or ID_PATTERN.match(content_id) is not None, f"line {index}: invalid content id")
        elif record_kind == "edge":
            for field in ("source", "target", "relation"):
                require(field in record, f"line {index}: edge.{field} is required")
            require(ID_PATTERN.match(record["source"]) is not None, f"line {index}: invalid edge source")
            require(ID_PATTERN.match(record["target"]) is not None, f"line {index}: invalid edge target")
        else:
            for field in ("source", "relation", "expression", "reason"):
                require(field in record, f"line {index}: unresolved.{field} is required")
            require(ID_PATTERN.match(record["source"]) is not None, f"line {index}: invalid unresolved source")

    require(facts == sorted(facts, key=sort_key), "fact records are not canonically sorted")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("facts", type=Path)
    args = parser.parse_args()
    try:
        validate(args.facts)
    except (OSError, ValueError, json.JSONDecodeError) as error:
        parser.error(str(error))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
