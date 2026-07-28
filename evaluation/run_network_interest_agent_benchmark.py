from __future__ import annotations

"""Compatibility entry point for the grounded architecture benchmark task."""

import sys

from run_agent_benchmark import main


if __name__ == "__main__":
    if len(sys.argv) == 1:
        sys.argv.extend(["--task", "space-rocks-room-scale-architecture"])
    raise SystemExit(main())
