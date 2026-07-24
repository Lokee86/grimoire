#!/usr/bin/env python3
"""Measure call-site resolution coverage in a facts-v1 JSONL stream."""

from __future__ import annotations

import argparse
import json
import math
import subprocess
import sys
from collections import Counter
from pathlib import Path
from typing import Any

VALIDATOR_PATH = Path(__file__).with_name("validate_jsonl.py")
CALL_RELATIONS = {"calls", "possible-calls"}
PERCENTILES = (50, 75, 90, 95, 99)


def validate(path: Path) -> None:
    subprocess.run([sys.executable, str(VALIDATOR_PATH), str(path)], check=True)


def span_key(span: dict[str, Any] | None) -> tuple[Any, ...]:
    span = span or {}
    return (
        span.get("path", ""),
        span.get("start_line", 0),
        span.get("start_column", 0),
        span.get("end_line", 0),
        span.get("end_column", 0),
    )


def record_sort_key(record: dict[str, Any]) -> tuple[Any, ...]:
    if record["record"] == "node":
        return (0, record["id"], record["kind"], record["path"], record["qualified_name"])
    if record["record"] == "edge":
        return (1, record["source"], record["target"], record["relation"], *span_key(record.get("span")))
    return (
        2,
        record["source"],
        record["relation"],
        record["expression"],
        record["reason"],
        *span_key(record.get("span")),
    )


def source_path(record: dict[str, Any], nodes: dict[str, dict[str, Any]]) -> str:
    span = record.get("span") or {}
    if span.get("path"):
        return span["path"]
    if record.get("owner"):
        return record["owner"]
    source = nodes.get(record.get("source", "")) or {}
    return source.get("path") or source.get("owner") or ""


def call_site_key(record: dict[str, Any], nodes: dict[str, dict[str, Any]]) -> tuple[Any, ...]:
    span = record.get("span")
    return (source_path(record, nodes), span_key(span))


def percentile(values: list[int], percent: int) -> int | None:
    if not values:
        return None
    ordered = sorted(values)
    rank = max(1, math.ceil(percent * len(ordered) / 100))
    return ordered[rank - 1]


def summarize(path: Path, limit: int = 10) -> dict[str, Any]:
    validate(path)
    records = [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines()]
    header = records[0]
    nodes = {record["id"]: record for record in records[1:] if record["record"] == "node"}
    sites: dict[tuple[Any, ...], dict[str, Any]] = {}

    for record in records[1:]:
        if record["record"] not in {"edge", "unresolved"}:
            continue
        if record["relation"] not in CALL_RELATIONS and not (
            record["record"] == "unresolved" and record["relation"] == "calls"
        ):
            continue
        key = call_site_key(record, nodes)
        site = sites.setdefault(
            key,
            {
                "source_id": record.get("source", ""),
                "source_path": key[0],
                "span": record.get("span"),
                "definite": False,
                "possible_targets": set(),
                "unresolved_reasons": set(),
                "expressions": set(),
            },
        )
        if record["record"] == "edge":
            if record["relation"] == "calls":
                site["definite"] = True
            else:
                site["possible_targets"].add(record["target"])
        else:
            site["unresolved_reasons"].add(record["reason"])
            if record.get("expression"):
                site["expressions"].add(record["expression"])

    counts = Counter()
    unresolved_reasons: Counter[str] = Counter()
    fanout_sites: list[dict[str, Any]] = []
    for site in sites.values():
        definite = site["definite"]
        possible = bool(site["possible_targets"])
        if definite and possible:
            category = "definite_plus_possible"
        elif definite:
            category = "definite_only"
        elif possible:
            category = "possible_only"
        else:
            category = "unresolved_only"
        counts[category] += 1
        for reason in site["unresolved_reasons"]:
            unresolved_reasons[reason] += 1

        normalized = {
            "source_id": site["source_id"],
            "source_path": site["source_path"],
            "span": site["span"],
            "possible_target_count": len(site["possible_targets"]),
            "possible_targets": sorted(site["possible_targets"]),
            "unresolved_reasons": sorted(site["unresolved_reasons"]),
            "expressions": sorted(site["expressions"]),
        }
        if possible:
            fanout_sites.append(normalized)

    fanout_values = [site["possible_target_count"] for site in fanout_sites]
    fanout_sites.sort(
        key=lambda site: (
            -site["possible_target_count"],
            site["source_path"],
            span_key(site["span"]),
            site["source_id"],
        )
    )
    return {
        "schema_version": 1,
        "tool": "c-family-call-resolution-metrics",
        "adapter_version": header["adapter_version"],
        "language": header["language"],
        "repository": header["repository"],
        "call_sites": {
            "total": len(sites),
            "definite_only": counts["definite_only"],
            "possible_only": counts["possible_only"],
            "definite_plus_possible": counts["definite_plus_possible"],
            "unresolved_only": counts["unresolved_only"],
        },
        "unresolved_reason_counts": dict(sorted(unresolved_reasons.items())),
        "possible_target_fanout": {
            "site_count": len(fanout_sites),
            "percentiles": {
                f"p{percent}": percentile(fanout_values, percent) for percent in PERCENTILES
            },
        },
        "highest_fanout_call_sites": fanout_sites[:limit],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("facts", type=Path)
    parser.add_argument("--limit", type=int, default=10, help="maximum highest-fanout sites to emit")
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    if args.limit < 0:
        parser.error("--limit must be non-negative")
    try:
        result = summarize(args.facts, args.limit)
    except (OSError, ValueError, json.JSONDecodeError, subprocess.CalledProcessError) as error:
        parser.error(str(error))
    rendered = json.dumps(result, indent=2, sort_keys=True) + "\n"
    if args.output:
        args.output.write_text(rendered, encoding="utf-8")
    else:
        print(rendered, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
