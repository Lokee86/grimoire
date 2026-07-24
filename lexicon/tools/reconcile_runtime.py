#!/usr/bin/env python3
"""Validate and reconcile Lexicon runtime evidence with a static facts stream."""

from __future__ import annotations

import argparse
import json
import re
from collections import Counter
from pathlib import Path, PurePosixPath
from typing import Any

ID_PATTERN = re.compile(r"^sha256:[0-9a-f]{64}$")
RUNTIME_RELATIONS = {"calls", "reads", "writes"}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def read_jsonl(path: Path) -> list[dict[str, Any]]:
    lines = path.read_text(encoding="utf-8").splitlines()
    require(bool(lines), f"{path}: stream is empty")
    records = [json.loads(line) for line in lines]
    require(all(isinstance(record, dict) for record in records), f"{path}: every line must be an object")
    return records


def validate_id(value: Any, label: str) -> str:
    require(isinstance(value, str) and ID_PATTERN.fullmatch(value) is not None, f"{label}: invalid Lexicon ID")
    return value


def validate_owner(value: Any, label: str) -> None:
    if value is None:
        return
    require(isinstance(value, str) and bool(value), f"{label}: owner must be a non-empty path")
    require("\\" not in value, f"{label}: owner must use forward slashes")
    path = PurePosixPath(value)
    require(not path.is_absolute() and "." not in path.parts and ".." not in path.parts, f"{label}: owner is not normalized")


def stable_value(value: Any, label: str) -> None:
    if isinstance(value, (str, int, float, bool)) or value is None:
        return
    require(isinstance(value, list), f"{label}: attributes must contain scalars or scalar arrays")
    require(all(isinstance(item, (str, int, float, bool)) or item is None for item in value), f"{label}: attribute array must contain scalars")
    require(value == sorted(value, key=lambda item: (type(item).__name__, repr(item))), f"{label}: attribute arrays must be sorted")


def validate_runtime(records: list[dict[str, Any]]) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    header = records[0]
    require(header.get("record") == "lexicon-runtime", "runtime header.record must be lexicon-runtime")
    require(header.get("schema_version") == 1, "runtime schema_version must be 1")
    for field in ("repository", "run_id"):
        require(isinstance(header.get(field), str) and bool(header[field]), f"runtime header.{field} is required")
    snapshot = header.get("static_snapshot")
    if snapshot is not None:
        validate_id(snapshot, "runtime header.static_snapshot")

    observations = records[1:]
    previous_key: tuple[str, str, str, str, str, str] | None = None
    seen_aggregates: set[tuple[str, str, str, str, str, str]] = set()
    for line, record in enumerate(observations, start=2):
        require(record.get("record") == "observation", f"runtime line {line}: invalid record")
        relation = record.get("relation")
        require(relation in RUNTIME_RELATIONS, f"runtime line {line}: invalid relation")
        source = validate_id(record.get("source"), f"runtime line {line}.source")
        count = record.get("count")
        require(isinstance(count, int) and not isinstance(count, bool) and count > 0, f"runtime line {line}: count must be positive")
        target = record.get("target")
        external = record.get("external_target")
        require((target is None) != (external is None), f"runtime line {line}: exactly one target form is required")
        target_key = ""
        if target is not None:
            target_key = validate_id(target, f"runtime line {line}.target")
        else:
            require(isinstance(external, str) and bool(external), f"runtime line {line}: external_target must be non-empty")
        first_seen = record.get("first_seen_ns")
        last_seen = record.get("last_seen_ns")
        for name, value in (("first_seen_ns", first_seen), ("last_seen_ns", last_seen)):
            require(value is None or (isinstance(value, int) and not isinstance(value, bool) and value >= 0), f"runtime line {line}: {name} must be non-negative")
        if first_seen is not None and last_seen is not None:
            require(last_seen >= first_seen, f"runtime line {line}: last_seen_ns precedes first_seen_ns")
        thread = record.get("thread", "")
        require(isinstance(thread, str), f"runtime line {line}: thread must be a string")
        validate_owner(record.get("owner"), f"runtime line {line}")
        attributes = record.get("attributes", {})
        require(isinstance(attributes, dict), f"runtime line {line}: attributes must be an object")
        for name, value in attributes.items():
            require(isinstance(name, str), f"runtime line {line}: attribute names must be strings")
            stable_value(value, f"runtime line {line}.attributes.{name}")
        attributes_key = json.dumps(attributes, sort_keys=True, separators=(",", ":"))
        key = (relation, source, target_key, str(external or ""), thread, attributes_key)
        require(previous_key is None or key >= previous_key, "runtime observations are not canonically sorted")
        previous_key = key
        require(key not in seen_aggregates, f"runtime line {line}: duplicate observation aggregate")
        seen_aggregates.add(key)
    return header, observations


def reconcile(static_records: list[dict[str, Any]], runtime_records: list[dict[str, Any]]) -> dict[str, Any]:
    static_header = static_records[0]
    require(static_header.get("record") == "lexicon", "static first record must be Lexicon header")
    runtime_header, observations = validate_runtime(runtime_records)
    require(static_header.get("repository") == runtime_header["repository"], "static and runtime repository identities differ")

    nodes = {record["id"] for record in static_records[1:] if record.get("record") == "node" and isinstance(record.get("id"), str)}
    edges = {
        (record.get("source"), record.get("target"), record.get("relation"))
        for record in static_records[1:]
        if record.get("record") == "edge"
    }
    categories: Counter[str] = Counter()
    details: list[dict[str, Any]] = []

    for observation in observations:
        relation = observation["relation"]
        source = observation["source"]
        target = observation.get("target")
        if source not in nodes or (target is not None and target not in nodes):
            category = "unknown-static-id"
        elif target is None:
            category = "external-runtime-target"
        elif relation == "calls" and (source, target, "calls") in edges:
            category = "confirmed-definite"
        elif relation == "calls" and (source, target, "possible-calls") in edges:
            category = "confirmed-possible"
        elif (source, target, relation) in edges:
            category = "confirmed-definite"
        else:
            category = "unmodeled-static-target"
        categories[category] += observation["count"]
        detail = {
            "category": category,
            "count": observation["count"],
            "relation": relation,
            "source": source,
        }
        if target is not None:
            detail["target"] = target
        else:
            detail["external_target"] = observation["external_target"]
        details.append(detail)

    return {
        "repository": runtime_header["repository"],
        "run_id": runtime_header["run_id"],
        "observation_counts": dict(sorted(categories.items())),
        "observations": details,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("static_facts", type=Path)
    parser.add_argument("runtime_evidence", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    try:
        report = reconcile(read_jsonl(args.static_facts), read_jsonl(args.runtime_evidence))
    except (OSError, ValueError, json.JSONDecodeError) as error:
        parser.error(str(error))
    rendered = json.dumps(report, indent=2, sort_keys=True) + "\n"
    if args.output:
        args.output.write_text(rendered, encoding="utf-8")
    else:
        print(rendered, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
