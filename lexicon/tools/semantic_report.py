#!/usr/bin/env python3
"""Summarize semantic relationship coverage in Lexicon fact streams."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from collections import Counter
from pathlib import Path
from typing import Any

VALIDATOR_PATH = Path(__file__).with_name("validate_jsonl.py")
DISPATCH_RELATIONS = {"extends", "implements", "uses-trait", "includes", "overrides"}
INTERNAL_CALL_REASONS = {"ambiguous-target", "missing-target", "unsupported-form"}


def ratio(numerator: int, denominator: int) -> float | None:
    if denominator == 0:
        return None
    return round(numerator / denominator, 6)


def validate(path: Path) -> None:
    subprocess.run([sys.executable, str(VALIDATOR_PATH), str(path)], check=True)


def call_site_key(record: dict[str, Any]) -> tuple[Any, ...]:
    span = record.get("span") or {}
    return (
        record.get("source", ""),
        span.get("path", record.get("owner", "")),
        span.get("start_line", 0),
        span.get("start_column", 0),
        span.get("end_line", 0),
        span.get("end_column", 0),
    )


def summarize(path: Path) -> dict[str, Any]:
    validate(path)
    nodes: Counter[str] = Counter()
    edges: Counter[str] = Counter()
    unresolved: Counter[str] = Counter()
    unresolved_reasons: Counter[str] = Counter()
    call_sites: dict[tuple[Any, ...], dict[str, Any]] = {}

    with path.open(encoding="utf-8") as handle:
        header = json.loads(next(handle))
        for line in handle:
            record = json.loads(line)
            kind = record["record"]
            if kind == "node":
                nodes[record["kind"]] += 1
                continue
            if kind == "edge":
                relation = record["relation"]
                edges[relation] += 1
                attributes = record.get("attributes", {})
                if relation in {"calls", "possible-calls"}:
                    site = call_sites.setdefault(call_site_key(record), {})
                    site["definite" if relation == "calls" else "possible"] = True
                elif relation == "references" and attributes.get("role") == "macro-expansion":
                    call_sites.setdefault(call_site_key(record), {})["macro"] = True
                continue

            relation = record["relation"]
            unresolved[relation] += 1
            if relation == "calls":
                reason = record["reason"]
                unresolved_reasons[reason] += 1
                site = call_sites.setdefault(call_site_key(record), {})
                site.setdefault("reasons", set()).add(reason)

    definite_sites = 0
    possible_sites = 0
    macro_only_sites = 0
    unresolved_sites = 0
    site_reasons: Counter[str] = Counter()
    internal_unresolved_sites = 0
    for site in call_sites.values():
        if site.get("definite"):
            definite_sites += 1
        elif site.get("possible"):
            possible_sites += 1
        elif site.get("macro"):
            macro_only_sites += 1
        else:
            unresolved_sites += 1
        reasons = site.get("reasons", set())
        for reason in reasons:
            site_reasons[reason] += 1
        if not site.get("definite") and not site.get("possible") and reasons & INTERNAL_CALL_REASONS:
            internal_unresolved_sites += 1

    call_definite = edges["calls"]
    call_possible = edges["possible-calls"]
    call_unresolved = unresolved["calls"]
    call_edge_total = call_definite + call_possible + call_unresolved
    repository_site_total = definite_sites + possible_sites + internal_unresolved_sites
    dataflow_edges = edges["reads"] + edges["writes"]
    dispatch_edges = sum(edges[relation] for relation in DISPATCH_RELATIONS)

    return {
        "adapter_version": header["adapter_version"],
        "call_quality": {
            "definite_fraction": ratio(call_definite, call_edge_total),
            "resolved_fraction": ratio(call_definite + call_possible, call_edge_total),
            "total": call_edge_total,
        },
        "call_site_quality": {
            "definite": definite_sites,
            "macro_only": macro_only_sites,
            "possible": possible_sites,
            "definite_repository_fraction": ratio(definite_sites, repository_site_total),
            "repository_target_fraction": ratio(definite_sites + possible_sites, repository_site_total),
            "repository_target_total": repository_site_total,
            "total_observed": len(call_sites),
            "unresolved": unresolved_sites,
            "unresolved_reasons": dict(sorted(site_reasons.items())),
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
        "unresolved_call_reasons": dict(sorted(unresolved_reasons.items())),
        "unresolved_relations": dict(sorted(unresolved.items())),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("facts", nargs="+", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    try:
        reports = [summarize(path) for path in args.facts]
    except (OSError, ValueError, json.JSONDecodeError, subprocess.CalledProcessError) as error:
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
