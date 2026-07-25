#!/usr/bin/env python3
"""Report deterministic semantic-depth metrics for facts-v1 streams."""

from __future__ import annotations

import argparse
import json
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any

CALL_RELATIONS = {"calls", "possible-calls"}
CALLABLE_KINDS = {"function", "method", "constructor"}


def load(path: Path) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines:
        raise ValueError(f"{path}: fact stream is empty")
    records = [json.loads(line) for line in lines]
    header = records[0]
    if header.get("record") != "lexicon" or header.get("schema_version") != 1:
        raise ValueError(f"{path}: first record is not a facts-v1 header")
    if not all(isinstance(record, dict) for record in records):
        raise ValueError(f"{path}: every record must be an object")
    return header, records[1:]


def metric(pairs: set[tuple[str, str]], occurrences: int | None = None) -> dict[str, int]:
    return {
        "occurrences": len(pairs) if occurrences is None else occurrences,
        "unique_source_target_pairs": len(pairs),
    }


def node_metric(occurrences: int) -> dict[str, int]:
    return {"occurrences": occurrences}


def edge_pair(record: dict[str, Any]) -> tuple[str, str]:
    return record.get("source", ""), record.get("target", "")


def summarize(path: Path) -> dict[str, Any]:
    header, records = load(path)
    nodes = {record["id"]: record for record in records if record.get("record") == "node"}
    edges = [record for record in records if record.get("record") == "edge"]
    unresolved = [record for record in records if record.get("record") == "unresolved"]

    callable_ids = {node_id for node_id, node in nodes.items() if node.get("kind") in CALLABLE_KINDS}
    definitions = [
        node for node in nodes.values()
        if node.get("kind") in CALLABLE_KINDS and node.get("attributes", {}).get("definition") is True
    ]
    prototypes = [
        node for node in nodes.values()
        if node.get("kind") in CALLABLE_KINDS and node.get("attributes", {}).get("definition") is not True
    ]
    defined_by_callable = {
        record.get("target")
        for record in edges
        if record.get("relation") == "defines" and record.get("source") in callable_ids
    }
    fields = [node for node in nodes.values() if node.get("kind") == "field"]
    parameters = [node for node in nodes.values() if node.get("kind") == "parameter"]
    locals_ = [
        node for node in nodes.values()
        if node.get("kind") == "variable" and node.get("id") in defined_by_callable
    ]

    relations: dict[str, dict[str, int]] = {}
    relation_pairs: defaultdict[str, set[tuple[str, str]]] = defaultdict(set)
    relation_occurrences: Counter[str] = Counter()
    for record in edges:
        relation = record.get("relation", "")
        relation_occurrences[relation] += 1
        relation_pairs[relation].add(edge_pair(record))
    for relation in sorted(relation_occurrences):
        relations[relation] = metric(relation_pairs[relation], relation_occurrences[relation])

    call_edges = [record for record in edges if record.get("relation") in CALL_RELATIONS]

    def evidence_metric(predicate: Any) -> dict[str, int]:
        selected = [record for record in call_edges if predicate(record)]
        return metric({edge_pair(record) for record in selected}, len(selected))

    evidence_labels: defaultdict[str, set[tuple[str, str]]] = defaultdict(set)
    evidence_label_occurrences: Counter[str] = Counter()
    macro_expansion_depths: Counter[int] = Counter()
    for record in call_edges:
        attributes = record.get("attributes", {})
        for label in attributes.get("evidence", []):
            if not isinstance(label, str):
                continue
            evidence_label_occurrences[label] += 1
            evidence_labels[label].add(edge_pair(record))
        depth = attributes.get("expansion_depth")
        if isinstance(depth, int):
            macro_expansion_depths[depth] += 1

    evidence = {
        "direct_calls": evidence_metric(
            lambda record: record.get("relation") == "calls" and not record.get("attributes", {}).get("indirect")
        ),
        "possible_calls": evidence_metric(lambda record: record.get("relation") == "possible-calls"),
        "indirect_calls": evidence_metric(lambda record: bool(record.get("attributes", {}).get("indirect"))),
        "macro_body_calls": evidence_metric(
            lambda record: "macro-body" in record.get("attributes", {}).get("evidence", [])
        ),
        "macro_mediated_calls": evidence_metric(
            lambda record: record.get("attributes", {}).get("indirect") == "macro"
        ),
        "self_recursion": evidence_metric(
            lambda record: record.get("source") == record.get("target")
        ),
        "passes_to": relations.get("passes-to", metric(set(), 0)),
        "reads": relations.get("reads", metric(set(), 0)),
        "writes": relations.get("writes", metric(set(), 0)),
        "resolution_evidence": {
            label: metric(evidence_labels[label], evidence_label_occurrences[label])
            for label in sorted(evidence_label_occurrences)
        },
        "macro_expansion_depth": {
            "max": max(macro_expansion_depths, default=0),
            "by_depth": {str(depth): count for depth, count in sorted(macro_expansion_depths.items())},
        },
    }

    unresolved_by_reason: Counter[str] = Counter()
    unresolved_by_relation: Counter[str] = Counter()
    unresolved_sources: defaultdict[str, set[str]] = defaultdict(set)
    for record in unresolved:
        relation = record.get("relation", "")
        reason = record.get("reason", "")
        unresolved_by_relation[relation] += 1
        unresolved_by_reason[reason] += 1
        unresolved_sources[reason].add(record.get("source", ""))

    return {
        "schema_version": 1,
        "tool": "c-family-semantic-depth",
        "stream": {
            "path": path.as_posix(),
            "adapter_version": header.get("adapter_version", ""),
            "language": header.get("language", ""),
            "repository": header.get("repository", ""),
        },
        "nodes": {
            "definitions": node_metric(len(definitions)),
            "prototypes": node_metric(len(prototypes)),
            "fields": node_metric(len(fields)),
            "locals": node_metric(len(locals_)),
            "parameters": node_metric(len(parameters)),
        },
        "relations": relations,
        "evidence": evidence,
        "unresolved": {
            "occurrences": len(unresolved),
            "by_relation": dict(sorted(unresolved_by_relation.items())),
            "by_reason": {
                reason: {
                    "occurrences": unresolved_by_reason[reason],
                    "unique_sources": len(unresolved_sources[reason]),
                }
                for reason in sorted(unresolved_by_reason)
            },
        },
    }


def tuple_key(record: dict[str, Any], nodes: dict[str, dict[str, Any]]) -> tuple[Any, ...]:
    def reference(node_id: str) -> tuple[str, str]:
        node = nodes.get(node_id, {})
        return node.get("path", ""), node.get("qualified_name", "")

    source_path, source_name = reference(record.get("source", ""))
    target_path, target_name = reference(record.get("target", ""))
    return record.get("relation", ""), source_path, source_name, target_path, target_name


def rendered_tuple(value: tuple[Any, ...]) -> dict[str, Any]:
    relation, source_path, source_name, target_path, target_name = value
    return {
        "relation": relation,
        "source": {"path": source_path, "qualified_name": source_name},
        "target": {"path": target_path, "qualified_name": target_name},
    }


def compare(left: Path, right: Path) -> dict[str, Any]:
    _, left_records = load(left)
    _, right_records = load(right)
    left_nodes = {record["id"]: record for record in left_records if record.get("record") == "node"}
    right_nodes = {record["id"]: record for record in right_records if record.get("record") == "node"}
    left_keys = {tuple_key(record, left_nodes) for record in left_records if record.get("record") == "edge"}
    right_keys = {tuple_key(record, right_nodes) for record in right_records if record.get("record") == "edge"}

    def group(values: set[tuple[Any, ...]]) -> dict[str, Any]:
        ordered = sorted(values)
        return {"count": len(ordered), "tuples": [rendered_tuple(value) for value in ordered]}

    return {
        "shared": group(left_keys & right_keys),
        "left_only": group(left_keys - right_keys),
        "right_only": group(right_keys - left_keys),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("facts", nargs="+", type=Path, help="one stream, or two streams to compare")
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    if len(args.facts) not in {1, 2}:
        parser.error("provide one facts stream or two streams for comparison")
    try:
        if len(args.facts) == 1:
            result = summarize(args.facts[0])
        else:
            left, right = args.facts
            result = {
                "schema_version": 1,
                "tool": "c-family-semantic-depth",
                "left": summarize(left),
                "right": summarize(right),
                "comparison": compare(left, right),
            }
    except (OSError, ValueError, json.JSONDecodeError) as error:
        parser.error(str(error))
    rendered = json.dumps(result, indent=2, sort_keys=True) + "\n"
    if args.output:
        args.output.write_text(rendered, encoding="utf-8")
    else:
        print(rendered, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
