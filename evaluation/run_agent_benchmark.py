from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import time

from benchmark_runner import BenchmarkEnvironment, collect_runs, prepare_worktree
from benchmark_tasks import load_task_suite, validate_evidence_prefixes

ROOT = Path(r"C:\!bin\workspace")
GRIMOIRE_REPO = ROOT / "grimoire"
DEFAULT_TASKS = GRIMOIRE_REPO / "evaluation" / "agent_benchmark_tasks.v2.json"
DEFAULT_BUILD = Path(os.environ.get("GRIMOIRE_BENCH_BUILD", GRIMOIRE_REPO / "build"))
DEFAULT_OUTPUT = GRIMOIRE_REPO / "evaluation" / "results" / "agent-benchmark-v2"
DEFAULT_CHECKOUTS = ROOT / "benchmark-checkouts" / "agent-benchmark-v2"
MODEL = "gpt-5.6-sol"
PROVIDER = "openai-codex"
CONDITIONS = ("plain", "cbm", "grimoire")


def select_tasks(suite: dict, selected: list[str]) -> list[dict]:
    tasks = suite["tasks"]
    if not selected:
        return tasks
    wanted = set(selected)
    chosen = [task for task in tasks if task["id"] in wanted]
    missing = wanted - {task["id"] for task in chosen}
    if missing:
        raise ValueError(f"unknown task ids: {', '.join(sorted(missing))}")
    return chosen


def main() -> int:
    parser = argparse.ArgumentParser(description="Run the version 2 grounded agent benchmark suite.")
    parser.add_argument("--tasks", type=Path, default=DEFAULT_TASKS)
    parser.add_argument("--build", type=Path, default=DEFAULT_BUILD)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--checkout-root", type=Path, default=DEFAULT_CHECKOUTS)
    parser.add_argument("--task", action="append", default=[])
    parser.add_argument("--condition", action="append", choices=CONDITIONS, default=[])
    parser.add_argument("--model", default=MODEL)
    parser.add_argument("--provider", default=PROVIDER)
    parser.add_argument("--check", action="store_true", help="validate tasks and dependencies without running benchmarks")
    args = parser.parse_args()

    suite = load_task_suite(args.tasks.resolve(), ROOT)
    tasks = select_tasks(suite, args.task)
    conditions = tuple(dict.fromkeys(args.condition or CONDITIONS))
    environment = BenchmarkEnvironment(ROOT, args.build.resolve(), args.model, args.provider)
    environment.require_dependencies(conditions)
    if args.check:
        print(json.dumps({
            "ready": True,
            "tasks": [task["id"] for task in tasks],
            "conditions": conditions,
            "build": str(args.build.resolve()),
        }, indent=2))
        return 0
    args.output.mkdir(parents=True, exist_ok=True)
    args.checkout_root.mkdir(parents=True, exist_ok=True)

    summary = {
        "schema": "grimoire.agent-benchmark.v2",
        "task_suite": str(args.tasks.resolve()),
        "model": args.model,
        "provider": args.provider,
        "conditions": conditions,
        "parallel_within_task": len(conditions) > 1,
        "sequential_tasks": True,
        "started_at": datetime.now(timezone.utc).isoformat(),
        "tasks": {},
    }
    for task in tasks:
        task_output = args.output / task["id"]
        task_output.mkdir(parents=True, exist_ok=True)
        checkouts: dict[str, Path] = {}
        commits: set[str] = set()
        for condition in conditions:
            checkout = args.checkout_root / task["id"] / condition
            commits.add(prepare_worktree(task["repo"], checkout, task["revision"]))
            validate_evidence_prefixes(task, checkout)
            checkouts[condition] = checkout
        if len(commits) != 1:
            raise RuntimeError(f"conditions resolved different commits: {sorted(commits)}")
        commit = next(iter(commits))

        preparation: dict[str, object] = {}
        cbm_project = None
        cbm_cache = task_output / "cbm-cache"
        if "cbm" in conditions:
            preparation["cbm"] = environment.prewarm_cbm(checkouts["cbm"], cbm_cache, task_output)
            cbm_project = preparation["cbm"]["project"]
        if "grimoire" in conditions:
            preparation["grimoire"] = environment.prewarm_grimoire(checkouts["grimoire"], task_output)

        profiles: dict[str, str] = {}
        for condition in conditions:
            profile = f"bench-v2-{task['id'][:20]}-{condition}"
            audit = task_output / "grimoire.mcp-audit.jsonl" if condition == "grimoire" else None
            if audit and audit.exists():
                audit.unlink()
            environment.prepare_profile(
                profile,
                condition,
                checkouts[condition],
                cbm_cache=cbm_cache if condition == "cbm" else None,
                audit_log=audit,
            )
            profiles[condition] = profile

        active = {
            condition: environment.launch(
                task,
                condition,
                checkouts[condition],
                task_output,
                profiles[condition],
                cbm_project,
            )
            for condition in conditions
        }
        runs = collect_runs(
            active,
            checkout_by_condition=checkouts,
            output=task_output,
            expected_sections=suite["evidence_sections"],
            required_path_prefixes=[
                item["path_prefix"] for item in task["rubric"]["required_evidence"]
            ],
        )
        summary["tasks"][task["id"]] = {
            "category": task["category"],
            "repository": str(task["repo"]),
            "commit": commit,
            "rubric": task["rubric"],
            "preparation": preparation,
            "profiles": profiles,
            "runs": runs,
        }
        (args.output / "summary.partial.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
        time.sleep(3)

    summary["completed_at"] = datetime.now(timezone.utc).isoformat()
    (args.output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
