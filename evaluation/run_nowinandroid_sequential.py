from __future__ import annotations

"""Retired compatibility entry point.

The prompt-shaped Now in Android task is preserved in checked-in historical
results. New sequential runs use run_agent_benchmark.py with explicit --task and
--condition selections from agent_benchmark_tasks.v2.json.
"""


def main() -> int:
    raise SystemExit(
        "run_nowinandroid_sequential.py is retired; use "
        "run_agent_benchmark.py --task <task-id> --condition <condition>"
    )


if __name__ == "__main__":
    main()
