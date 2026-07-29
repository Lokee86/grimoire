from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
from pathlib import Path

from benchmark_grounding import validate_answer
from benchmark_runner import prepare_worktree, run_checked
from benchmark_tasks import load_task_suite

ROOT = Path(r"C:\!bin\workspace")
GRIMOIRE_REPO = ROOT / "grimoire"
DEFAULT_TASKS = GRIMOIRE_REPO / "evaluation" / "agent_benchmark_tasks.v2.json"
DEFAULT_OUTPUT = GRIMOIRE_REPO / "evaluation" / "results" / "agent-benchmark-v2"
DEFAULT_CHECKOUTS = ROOT / "benchmark-checkouts" / "agent-benchmark-v2"


def main() -> int:
    parser = argparse.ArgumentParser(description="Revalidate saved benchmark answers without rerunning agents.")
    parser.add_argument("--tasks", type=Path, default=DEFAULT_TASKS)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--checkout-root", type=Path, default=DEFAULT_CHECKOUTS)
    parser.add_argument("--task", action="append", default=[])
    args = parser.parse_args()

    suite = load_task_suite(args.tasks.resolve(), ROOT)
    tasks = suite["tasks"]
    if args.task:
        selected = set(args.task)
        tasks = [task for task in tasks if task["id"] in selected]
        missing = selected - {task["id"] for task in tasks}
        if missing:
            raise ValueError(f"unknown task ids: {', '.join(sorted(missing))}")

    summary_path = args.output / "summary.json"
    summary = json.loads(summary_path.read_text(encoding="utf-8"))
    results: dict[str, dict[str, bool]] = {}
    for task in tasks:
        task_id = task["id"]
        task_summary = summary["tasks"].get(task_id)
        if task_summary is None:
            continue
        task_output = args.output / task_id
        required_prefixes = [item["path_prefix"] for item in task["rubric"]["required_evidence"]]
        task_results: dict[str, bool] = {}
        created_checkouts: list[Path] = []
        try:
            for condition, run in task_summary["runs"].items():
                answer_path = task_output / f"{condition}.stdout.txt"
                checkout = args.checkout_root / task_id / condition
                if not checkout.is_dir():
                    prepare_worktree(task["repo"], checkout, task_summary["commit"])
                    created_checkouts.append(checkout)
                audit = task_output / "grimoire.mcp-audit.jsonl" if condition == "grimoire" else None
                report = validate_answer(
                    checkout,
                    answer_path.read_text(encoding="utf-8"),
                    exit_code=int(run["exit_code"]),
                    expected_sections=suite["evidence_sections"],
                    audit_log=audit,
                    required_path_prefixes=required_prefixes,
                )
                payload = report.to_dict()
                (task_output / f"{condition}.grounding.json").write_text(
                    json.dumps(payload, indent=2) + "\n",
                    encoding="utf-8",
                )
                run["grounding"] = payload
                run["execution_valid"] = int(run["exit_code"]) == 0
                run["grounding_valid"] = report.valid
                run["eligible_for_scoring"] = int(run["exit_code"]) == 0 and report.valid
                run["quality_assessed"] = False
                run.pop("valid", None)
                task_results[condition] = run["eligible_for_scoring"]
        finally:
            for checkout in created_checkouts:
                run_checked(["git", "worktree", "remove", "--force", str(checkout)], cwd=task["repo"], timeout=300)
            if created_checkouts:
                run_checked(["git", "worktree", "prune"], cwd=task["repo"], timeout=300)
        results[task_id] = task_results

    summary["revalidated_at"] = datetime.now(timezone.utc).isoformat()
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"revalidated": results}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
