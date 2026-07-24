from __future__ import annotations

import json
from collections import Counter
from pathlib import Path
from typing import Any

SAMPLED_RELATIONS = (
    "calls",
    "possible-calls",
    "reads",
    "writes",
    "depends-on",
    "extends",
    "implements",
    "overrides",
)


def describe_node(node: dict[str, Any] | None) -> dict[str, Any] | None:
    if node is None:
        return None
    return {
        "kind": node.get("kind"),
        "name": node.get("name"),
        "owner": node.get("owner"),
        "path": node.get("path"),
        "qualified_name": node.get("qualified_name"),
        "span": node.get("span"),
    }


def node_matches(node: dict[str, Any] | None, expected: dict[str, Any]) -> bool:
    if node is None:
        return False
    for key in ("kind", "name", "path", "qualified_name"):
        if key in expected and node.get(key) != expected[key]:
            return False
    for key, value in expected.get("attributes", {}).items():
        if node.get("attributes", {}).get(key) != value:
            return False
    return True


def edge_matches(edge: dict[str, Any], nodes: dict[str, dict[str, Any]], expected: dict[str, Any]) -> bool:
    if edge.get("relation") != expected.get("relation"):
        return False
    if "owner" in expected and edge.get("owner") != expected["owner"]:
        return False
    if not node_matches(nodes.get(edge.get("source")), expected.get("source", {})):
        return False
    if not node_matches(nodes.get(edge.get("target")), expected.get("target", {})):
        return False
    for key, value in expected.get("attributes", {}).items():
        if edge.get("attributes", {}).get(key) != value:
            return False
    return True


def summarize(path: Path, case: dict[str, Any]) -> dict[str, Any]:
    records = [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines()]
    header = records[0]
    nodes = {record["id"]: record for record in records[1:] if record["record"] == "node"}
    node_counts: Counter[str] = Counter()
    edge_counts: Counter[str] = Counter()
    unresolved_counts: Counter[str] = Counter()
    unresolved_reasons: Counter[str] = Counter()
    unresolved_call_reasons: Counter[str] = Counter()
    samples: dict[str, list[dict[str, Any]]] = {relation: [] for relation in SAMPLED_RELATIONS}
    samples["unresolved-calls"] = []

    for record in records[1:]:
        if record["record"] == "node":
            node_counts[record["kind"]] += 1
        elif record["record"] == "edge":
            relation = record["relation"]
            edge_counts[relation] += 1
            if relation in samples and len(samples[relation]) < 12:
                samples[relation].append(
                    {
                        "owner": record.get("owner"),
                        "source": describe_node(nodes.get(record["source"])),
                        "span": record.get("span"),
                        "target": describe_node(nodes.get(record["target"])),
                    }
                )
        else:
            relation = record["relation"]
            unresolved_counts[relation] += 1
            unresolved_reasons[record["reason"]] += 1
            if relation == "calls":
                unresolved_call_reasons[record["reason"]] += 1
                if len(samples["unresolved-calls"]) < 20:
                    samples["unresolved-calls"].append(
                        {
                            "candidate_name": record.get("candidate_name"),
                            "expression": record.get("expression"),
                            "owner": record.get("owner"),
                            "reason": record.get("reason"),
                            "source": describe_node(nodes.get(record["source"])),
                            "span": record.get("span"),
                        }
                    )

    node_records = list(nodes.values())
    edge_records = [record for record in records[1:] if record["record"] == "edge"]
    missing_required_nodes = [
        expected
        for expected in case.get("required_nodes", [])
        if not any(node_matches(node, expected) for node in node_records)
    ]
    missing_required_edges = [
        expected
        for expected in case.get("required_edges", [])
        if not any(edge_matches(edge, nodes, expected) for edge in edge_records)
    ]

    return {
        "adapter_version": header["adapter_version"],
        "edge_relations": dict(sorted(edge_counts.items())),
        "language": header["language"],
        "missing_required_edges": missing_required_edges,
        "missing_required_nodes": missing_required_nodes,
        "node_kinds": dict(sorted(node_counts.items())),
        "repository": header["repository"],
        "samples": samples,
        "unresolved_call_reasons": dict(sorted(unresolved_call_reasons.items())),
        "unresolved_reasons": dict(sorted(unresolved_reasons.items())),
        "unresolved_relations": dict(sorted(unresolved_counts.items())),
    }
