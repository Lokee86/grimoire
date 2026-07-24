from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path

from run_validation import executable, npm_command


def run_test(label: str, command: list[str], cwd: Path, env: dict[str, str] | None = None) -> dict[str, object]:
    print(f"[{label}] {' '.join(command)}", flush=True)
    completed = subprocess.run(
        command,
        cwd=cwd,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    if completed.stdout:
        print(completed.stdout.rstrip(), flush=True)
    if completed.returncode != 0:
        raise RuntimeError(f"{label} failed with exit code {completed.returncode}")
    return {"label": label, "status": "passed"}


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    go = executable("go.exe", Path("C:/Program Files/Go/bin/go.exe"))
    ruby = executable("ruby.exe", Path("C:/Ruby34-x64/bin/ruby.exe"))
    cargo = Path(executable("cargo.exe", Path.home() / ".cargo" / "bin" / "cargo.exe"))
    cargo_env = os.environ.copy()
    cargo_env["PATH"] = str(cargo.parent) + os.pathsep + cargo_env.get("PATH", "")

    tests = [
        ("application-go", [go, "test", "./..."], root, None),
        ("adapter-c-family", [go, "test", "./..."], root / "adapters" / "c-family", None),
        ("adapter-go", [go, "test", "./..."], root / "adapters" / "go", None),
        ("adapter-gdscript", [go, "test", "./..."], root / "adapters" / "gdscript", None),
        ("adapter-generic", [go, "test", "./..."], root / "adapters" / "generic", None),
        ("adapter-python", [sys.executable, "-m", "pytest"], root / "adapters" / "python", None),
        ("adapter-ruby", [ruby, "test/test_adapter.rb"], root / "adapters" / "ruby", None),
        ("adapter-typescript", npm_command("test"), root / "adapters" / "typescript", None),
        (
            "adapter-rust",
            [str(cargo), "test", "--manifest-path", "adapters/rust/Cargo.toml"],
            root,
            cargo_env,
        ),
    ]

    results = [run_test(label, command, cwd, env) for label, command, cwd, env in tests]
    print(json.dumps({"results": results}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
