from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
from pathlib import Path
import shutil

from benchmark_grounding import summarize_audit

ROOT = Path(r"C:\!bin\workspace")
GRIMOIRE_REPO = ROOT / "grimoire"
DEFAULT_OUTPUT = GRIMOIRE_REPO / "evaluation" / "results" / "agent-benchmark-v2"
DEFAULT_CHECKOUTS = ROOT / "benchmark-checkouts" / "agent-benchmark-v2"
CONDITIONS = ("plain", "cbm", "grimoire")


def unique_match(root: Path, filename: str) -> Path | None:
    matches = list(root.rglob(filename)) if root.is_dir() else []
    if not matches:
        return None
    if len(matches) != 1:
        raise RuntimeError(f"expected one {filename} below {root}, found {len(matches)}")
    return matches[0]


def main() -> int:
    parser = argparse.ArgumentParser(description="Import an isolated agent benchmark run into the canonical result set.")
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--task", required=True)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--checkout-root", type=Path, default=DEFAULT_CHECKOUTS)
    parser.add_argument("--remove-source", action="store_true")
    args = parser.parse_args()

    source = args.source.resolve()
    output = args.output.resolve()
    checkout_root = args.checkout_root.resolve()
    source_summary_path = source / "summary.json"
    source_task_dir = source / args.task
    if not source_summary_path.is_file() or not source_task_dir.is_dir():
        raise RuntimeError(f"incomplete source run: {source}")

    source_summary = json.loads(source_summary_path.read_text(encoding="utf-8"))
    source_task = source_summary.get("tasks", {}).get(args.task)
    if not isinstance(source_task, dict):
        raise RuntimeError(f"source summary does not contain task {args.task!r}")

    target_task_dir = output / args.task
    if target_task_dir.exists():
        shutil.rmtree(target_task_dir)
    shutil.copytree(source_task_dir, target_task_dir, ignore=shutil.ignore_patterns("cbm-cache"))

    for condition in CONDITIONS:
        checkout = checkout_root / args.task / condition
        usage_source = unique_match(checkout, f"{condition}.usage.json")
        if usage_source is not None:
            usage_target = target_task_dir / f"{condition}.usage.json"
            shutil.copy2(usage_source, usage_target)
            source_task["runs"][condition]["usage"] = json.loads(usage_target.read_text(encoding="utf-8"))
        if condition == "grimoire":
            audit_source = unique_match(checkout, "grimoire.mcp-audit.jsonl")
            if audit_source is not None:
                audit_target = target_task_dir / "grimoire.mcp-audit.jsonl"
                shutil.copy2(audit_source, audit_target)
                source_task["runs"][condition]["discovery_output"] = summarize_audit(audit_target)

    output.mkdir(parents=True, exist_ok=True)
    canonical_path = output / "summary.json"
    if canonical_path.is_file():
        canonical = json.loads(canonical_path.read_text(encoding="utf-8"))
    else:
        canonical = {key: value for key, value in source_summary.items() if key != "tasks"}
        canonical["tasks"] = {}
    canonical.setdefault("tasks", {})[args.task] = source_task
    canonical["completed_at"] = datetime.now(timezone.utc).isoformat()
    canonical_path.write_text(json.dumps(canonical, indent=2) + "\n", encoding="utf-8")

    if args.remove_source:
        shutil.rmtree(source)
    print(json.dumps({
        "imported": args.task,
        "output": str(output),
        "conditions": {
            condition: {
                "usage": source_task["runs"][condition].get("usage") is not None,
                "valid": source_task["runs"][condition].get("valid"),
            }
            for condition in CONDITIONS
        },
    }, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
