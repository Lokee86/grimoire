from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path

import run_multirepo_agent_benchmark as benchmark

TASK = next(task for task in benchmark.TASKS if task["id"] == "nowinandroid-topic-notification-muting")
SUMMARY_PATH = benchmark.OUT / "summary.partial.json"
CONDITIONS = ("plain", "cbm", "grimoire")


def load_summary() -> dict:
    if SUMMARY_PATH.is_file():
        return json.loads(SUMMARY_PATH.read_text(encoding="utf-8"))
    return {
        "model": benchmark.MODEL,
        "provider": benchmark.PROVIDER,
        "started_at": datetime.now(timezone.utc).isoformat(),
        "tasks": {},
    }


def save_result(condition: str, profile: str, preparation: dict, run: dict) -> None:
    summary = load_summary()
    summary["parallel_within_task"] = False
    summary["sequential_tasks"] = True
    summary["nowinandroid_resumed_sequentially"] = True
    entry = summary.setdefault("tasks", {}).setdefault(
        TASK["id"],
        {
            "repository": str(TASK["repo"]),
            "commit": TASK["commit"],
            "execution": "sequential conditions",
            "profiles": {},
            "runs": {},
        },
    )
    entry["profiles"][condition] = profile
    entry["runs"][condition] = run
    entry.update(preparation)
    entry["updated_at"] = datetime.now(timezone.utc).isoformat()
    SUMMARY_PATH.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")


def run_condition(condition: str) -> dict:
    task_out = benchmark.OUT / TASK["id"]
    task_out.mkdir(parents=True, exist_ok=True)
    if condition == "grimoire":
        checkout = benchmark.ROOT / "nia-g"
        if not checkout.is_dir():
            benchmark.prepare_worktree(TASK["repo"], checkout, TASK["commit"])
    else:
        checkout = benchmark.CHECKOUT_ROOT / TASK["id"] / condition
        benchmark.prepare_worktree(TASK["repo"], checkout, TASK["commit"])

    preparation: dict = {}
    cbm_project = None
    cbm_cache = task_out / "cbm-cache"
    if condition == "cbm":
        cbm_project, elapsed = benchmark.prewarm_cbm(checkout, cbm_cache, task_out)
        preparation["cbm_project"] = cbm_project
        preparation["cbm_index_seconds"] = round(elapsed, 3)
    elif condition == "grimoire":
        elapsed = benchmark.prewarm_grimoire(checkout, task_out)
        preparation["grimoire_prewarm_seconds"] = round(elapsed, 3)

    profile = f"bench-nowinandroid-seq-{condition}"
    benchmark.prepare_profile(profile, condition, checkout, cbm_cache)
    active = {
        condition: benchmark.launch_run(
            TASK,
            condition,
            checkout,
            task_out,
            profile,
            cbm_project,
        )
    }
    run = benchmark.collect_wave(active)[condition]
    save_result(condition, profile, preparation, run)
    payload = {
        "task": TASK["id"],
        "condition": condition,
        "preparation": preparation,
        "run": run,
    }
    print(json.dumps(payload, indent=2))
    return payload


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("condition", choices=CONDITIONS)
    args = parser.parse_args()
    run_condition(args.condition)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
