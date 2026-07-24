from __future__ import annotations

import argparse
import json
from collections import Counter
from pathlib import Path
from typing import Any


def identity(record: dict[str, Any]) -> tuple[Any, ...]:
    kind = record.get("record")
    if kind == "lexicon":
        return (kind,)
    if kind == "node":
        return (kind, record.get("id"))
    if kind == "edge":
        return (
            kind,
            record.get("source"),
            record.get("target"),
            record.get("relation"),
            json.dumps(record.get("span"), sort_keys=True),
        )
    return (
        kind,
        record.get("source"),
        record.get("relation"),
        record.get("expression"),
        record.get("reason"),
        json.dumps(record.get("span"), sort_keys=True),
    )


def load(path: Path) -> list[dict[str, Any]]:
    return [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines()]


def enrich(record: dict[str, Any], nodes: dict[str, dict[str, Any]]) -> dict[str, Any]:
    enriched = dict(record)
    if "source" in record:
        enriched["source_node"] = nodes.get(record["source"])
    if "target" in record:
        enriched["target_node"] = nodes.get(record["target"])
    return enriched


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("left", type=Path)
    parser.add_argument("right", type=Path)
    parser.add_argument("--limit", type=int, default=20)
    args = parser.parse_args()

    left = load(args.left)
    right = load(args.right)
    left_nodes = {record["id"]: record for record in left if record.get("record") == "node"}
    right_nodes = {record["id"]: record for record in right if record.get("record") == "node"}
    left_by_identity = {identity(record): record for record in left}
    right_by_identity = {identity(record): record for record in right}
    left_keys = set(left_by_identity)
    right_keys = set(right_by_identity)
    changed = [
        key
        for key in sorted(left_keys & right_keys, key=repr)
        if left_by_identity[key] != right_by_identity[key]
    ]
    counts = Counter(key[0] for key in left_keys ^ right_keys)
    result = {
        "changed_count": len(changed),
        "changed_samples": [
            {
                "identity": key,
                "left": left_by_identity[key],
                "right": right_by_identity[key],
            }
            for key in changed[: args.limit]
        ],
        "left_only_count": len(left_keys - right_keys),
        "left_only_samples": [enrich(left_by_identity[key], left_nodes) for key in sorted(left_keys - right_keys, key=repr)[: args.limit]],
        "record_count_delta": len(left) - len(right),
        "right_only_count": len(right_keys - left_keys),
        "right_only_samples": [enrich(right_by_identity[key], right_nodes) for key in sorted(right_keys - left_keys, key=repr)[: args.limit]],
        "symmetric_difference_by_record": dict(sorted(counts.items())),
    }
    print(json.dumps(result, indent=2, sort_keys=True))
    return 1 if changed or left_keys != right_keys else 0


if __name__ == "__main__":
    raise SystemExit(main())
