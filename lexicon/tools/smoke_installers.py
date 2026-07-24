from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


def executable_name(name: str) -> str:
    return name + ".exe" if os.name == "nt" else name


def run(command: list[str], cwd: Path, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    completed = subprocess.run(
        command,
        cwd=cwd,
        env=env,
        check=False,
        capture_output=True,
        text=True,
    )
    if completed.returncode != 0:
        raise RuntimeError(completed.stdout + completed.stderr)
    return completed


def smoke_windows(distribution: Path, temporary: Path) -> Path:
    powershell = shutil.which("powershell.exe") or shutil.which("pwsh.exe")
    if powershell is None:
        raise RuntimeError("PowerShell was not found")
    install_dir = temporary / "installed"
    run(
        [
            powershell,
            "-NoProfile",
            "-ExecutionPolicy",
            "Bypass",
            "-File",
            str(distribution / "install.ps1"),
            "-InstallDir",
            str(install_dir),
            "-NoPath",
        ],
        distribution,
    )
    return install_dir


def smoke_unix(distribution: Path, temporary: Path) -> Path:
    install_dir = temporary / "installed"
    bin_dir = temporary / "bin"
    env = os.environ.copy()
    env["LEXICON_BIN_DIR"] = str(bin_dir)
    run(["sh", str(distribution / "install.sh"), str(install_dir)], distribution, env)
    if not (bin_dir / "lexicon").exists():
        raise RuntimeError("installer did not create the command link")
    return install_dir


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="Install and verify an extracted Lexicon release package.")
    parser.add_argument("--distribution", type=Path, required=True)
    parser.add_argument("--version", required=True)
    args = parser.parse_args(argv)
    distribution = args.distribution.resolve()

    with tempfile.TemporaryDirectory(prefix="lexicon-install-") as directory:
        temporary = Path(directory)
        install_dir = smoke_windows(distribution, temporary) if os.name == "nt" else smoke_unix(distribution, temporary)
        binary = install_dir / executable_name("lexicon")
        if not binary.is_file():
            raise RuntimeError("installer did not copy the Lexicon executable")
        if not (install_dir / "adapters").is_dir():
            raise RuntimeError("installer did not copy the adapters directory")
        version = run([str(binary), "version"], install_dir).stdout.strip()
        expected = f"lexicon version {args.version}"
        if version != expected:
            raise RuntimeError(f"expected {expected}, got {version}")

    print(f"Lexicon installer smoke test passed for {args.version}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
