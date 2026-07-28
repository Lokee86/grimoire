from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--binary", required=True)
    parser.add_argument("--cwd", required=True)
    parser.add_argument("--cache-dir", required=True)
    return parser.parse_args()


def emit(message: dict[str, Any]) -> None:
    sys.stdout.write(json.dumps(message, separators=(",", ":")) + "\n")
    sys.stdout.flush()


def main() -> int:
    args = parse_args()
    environment = os.environ.copy()
    environment["CBM_CACHE_DIR"] = str(Path(args.cache_dir))
    process = subprocess.Popen(
        [args.binary],
        cwd=args.cwd,
        env=environment,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=sys.stderr,
        text=True,
        bufsize=1,
    )
    if process.stdin is None or process.stdout is None:
        raise RuntimeError("failed to open CBM stdio")

    internal_id = 10_000_000
    for line in sys.stdin:
        try:
            request = json.loads(line)
        except json.JSONDecodeError:
            process.stdin.write(line)
            process.stdin.flush()
            continue

        if request.get("method") != "tools/list" or "id" not in request:
            process.stdin.write(json.dumps(request, separators=(",", ":")) + "\n")
            process.stdin.flush()
            if "id" in request:
                response_line = process.stdout.readline()
                if not response_line:
                    return process.wait()
                sys.stdout.write(response_line)
                sys.stdout.flush()
            continue

        tools: list[Any] = []
        cursor: str | None = None
        while True:
            internal_id += 1
            params: dict[str, Any] = {}
            if cursor is not None:
                params["cursor"] = cursor
            child_request = {
                "jsonrpc": "2.0",
                "id": internal_id,
                "method": "tools/list",
                "params": params,
            }
            process.stdin.write(json.dumps(child_request, separators=(",", ":")) + "\n")
            process.stdin.flush()
            response_line = process.stdout.readline()
            if not response_line:
                return process.wait()
            child_response = json.loads(response_line)
            if "error" in child_response:
                emit({"jsonrpc": "2.0", "id": request["id"], "error": child_response["error"]})
                break
            result = child_response.get("result") or {}
            tools.extend(result.get("tools") or [])
            next_cursor = result.get("nextCursor")
            if not next_cursor:
                emit({"jsonrpc": "2.0", "id": request["id"], "result": {"tools": tools}})
                break
            cursor = str(next_cursor)

    process.terminate()
    try:
        return process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        process.kill()
        return process.wait()


if __name__ == "__main__":
    raise SystemExit(main())
