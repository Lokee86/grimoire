from __future__ import annotations

"""Compatibility entry point for the version 2 grounded benchmark runner.

The July 27 HikariCP/Detekt/Now in Android runner and its prompt-shaped tasks are
retained in Git history and checked-in result artifacts. New runs use the hidden-
rubric task suite in run_agent_benchmark.py.
"""

from run_agent_benchmark import main


if __name__ == "__main__":
    raise SystemExit(main())
