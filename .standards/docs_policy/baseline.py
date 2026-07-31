from __future__ import annotations

import json
from pathlib import Path

from audit import Finding


FindingKey = tuple[str, str, str]


def finding_key(finding: Finding) -> FindingKey:
    return finding.code, finding.path, finding.message


def load_baseline(repo: Path, relative_path: str) -> set[FindingKey]:
    path = repo / relative_path
    try:
        raw = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError as error:
        raise ValueError(f"missing documentation baseline: {relative_path}") from error
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid JSON in documentation baseline {relative_path}: {error}") from error
    if not isinstance(raw, dict) or raw.get("version") != 1:
        raise ValueError(f"unsupported documentation baseline: {relative_path}")
    items = raw.get("findings")
    if not isinstance(items, list):
        raise ValueError(f"documentation baseline findings must be an array: {relative_path}")
    result: set[FindingKey] = set()
    for index, item in enumerate(items, start=1):
        if not isinstance(item, dict):
            raise ValueError(f"documentation baseline finding {index} must be an object")
        values = item.get("code"), item.get("path"), item.get("message")
        if not all(isinstance(value, str) and value for value in values):
            raise ValueError(f"documentation baseline finding {index} is incomplete")
        result.add(values)
    return result


def write_baseline(path: Path, findings: list[Finding]) -> None:
    path.write_text(
        json.dumps(
            {
                "version": 1,
                "findings": [finding.__dict__ for finding in sorted(findings)],
            },
            indent=2,
            sort_keys=True,
        )
        + "\n",
        encoding="utf-8",
    )
