"""Canonical Lexicon record ordering and JSONL emission."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

from .contract import LANGUAGE, SCHEMA_VERSION
from .model import Facts


def _span_key(record: dict[str, Any]) -> tuple[Any, ...]:
    value = record.get("span") or {}
    return (
        value.get("path", ""),
        value.get("start_line", 0),
        value.get("start_column", 0),
        value.get("end_line", 0),
        value.get("end_column", 0),
    )


def _record_sort_key(record: dict[str, Any]) -> tuple[Any, ...]:
    kind = record["record"]
    if kind == "node":
        return (0, record["id"], record["kind"], record["path"], record["qualified_name"])
    if kind == "edge":
        return (1, record["source"], record["target"], record["relation"], *_span_key(record))
    return (
        2,
        record["source"],
        record["relation"],
        record["expression"],
        record["reason"],
        *_span_key(record),
    )


def emit_records(facts: Facts, adapter_version: str) -> list[dict[str, Any]]:
    header = {
        "record": "lexicon",
        "schema_version": SCHEMA_VERSION,
        "adapter_version": adapter_version,
        "language": LANGUAGE,
        "repository": facts.repository,
    }
    records = [header, *sorted(facts.nodes.values(), key=_record_sort_key)]
    records.extend(sorted(facts.edges.values(), key=_record_sort_key))
    records.extend(sorted(facts.unresolved.values(), key=_record_sort_key))
    return records


def write_records(records: list[dict[str, Any]], output: Path) -> None:
    lines = [json.dumps(record, ensure_ascii=False, sort_keys=True, separators=(",", ":")) for record in records]
    if str(output) == "-":
        sys.stdout.write("\n".join(lines) + "\n")
        return
    destination = output.expanduser()
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_text("\n".join(lines) + "\n", encoding="utf-8", newline="\n")
