#!/usr/bin/env python3
"""Build a reproducible Lexicon distribution directory."""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
EXCLUDED_SOURCE_PARTS = {
    ".git", ".pytest_cache", "__pycache__", "build", "dist", "node_modules",
    "target", "test", "tests", "fixtures", "vendor",
}


def run(command: list[str], cwd: Path) -> None:
    print("+", " ".join(command))
    subprocess.run(command, cwd=cwd, check=True)


def go_build_command(output: Path, package: str, version: str | None = None) -> list[str]:
    command = ["go", "build", "-trimpath", "-buildvcs=false"]
    if version:
        command.extend(["-ldflags", f"-X github.com/Lokee86/lexicon/internal/cli.version={version}"])
    command.extend(["-o", str(output), package])
    return command


def npm_executable() -> str:
    return "npm.cmd" if os.name == "nt" else "npm"


def cargo_executable() -> str:
    found = shutil.which("cargo")
    if found:
        return found
    name = "cargo.exe" if os.name == "nt" else "cargo"
    candidate = Path.home() / ".cargo" / "bin" / name
    if candidate.is_file():
        return str(candidate)
    raise FileNotFoundError("cargo executable not found")


def copy_file(source: Path, destination: Path) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, destination)
    shutil.copymode(source, destination)


def copy_sources(source: Path, destination: Path, suffix: str) -> None:
    for path in sorted(source.rglob(f"*{suffix}")):
        relative = path.relative_to(source)
        if any(part.lower() in EXCLUDED_SOURCE_PARTS for part in relative.parts):
            continue
        copy_file(path, destination / relative)


def executable_name(name: str) -> str:
    return name + ".exe" if os.name == "nt" else name


def build_go_adapter(repo: Path, output: Path, language: str) -> None:
    adapter = repo / "adapters" / language
    output.parent.mkdir(parents=True, exist_ok=True)
    run(["go", "build", "-trimpath", "-buildvcs=false", "-o", str(output), "."], adapter)


def build_typescript(repo: Path, output: Path) -> None:
    source = repo / "adapters" / "typescript"
    with tempfile.TemporaryDirectory(prefix="lexicon-typescript-") as temporary:
        work = Path(temporary)
        for name in ("package.json", "package-lock.json", "tsconfig.json"):
            copy_file(source / name, work / name)
        shutil.copytree(source / "src", work / "src")
        run([npm_executable(), "ci", "--silent", "--ignore-scripts"], work)
        run([npm_executable(), "run", "build", "--silent"], work)
        run([npm_executable(), "prune", "--omit=dev", "--ignore-scripts"], work)
        copy_file(work / "package.json", output / "package.json")
        copy_file(work / "package-lock.json", output / "package-lock.json")
        shutil.copytree(work / "dist", output / "dist")
        shutil.copytree(work / "node_modules", output / "node_modules")


def build_distribution(repo: Path, output: Path, version: str | None = None) -> None:
    if output == repo or repo / "adapters" in output.parents or repo / "tools" in output.parents:
        raise ValueError("output must not replace repository sources")
    if output.exists():
        shutil.rmtree(output)
    adapters = output / "adapters"
    adapters.mkdir(parents=True)

    lexicon = output / executable_name("lexicon")
    run(go_build_command(lexicon, "./cmd/lexicon", version), repo)
    for language in ("go", "gdscript"):
        build_go_adapter(repo, adapters / language / executable_name("lexicon-" + language), language)

    rust_output = adapters / "rust" / executable_name("lexicon-rust")
    rust = repo / "adapters" / "rust"
    run([cargo_executable(), "build", "--release", "--locked", "--manifest-path", str(rust / "Cargo.toml")], repo)
    copy_file(rust / "target" / "release" / executable_name("lexicon-rust-adapter"), rust_output)

    build_typescript(repo, adapters / "typescript")
    copy_sources(repo / "adapters" / "python", adapters / "python", ".py")
    copy_sources(repo / "adapters" / "ruby", adapters / "ruby", ".rb")


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output_path", nargs="?", type=Path, help="distribution directory")
    parser.add_argument("--output", dest="output_option", type=Path, help="distribution directory")
    parser.add_argument("--repo", type=Path, default=ROOT, help="Lexicon repository root")
    parser.add_argument("--version", help="version embedded in the Lexicon executable")
    args = parser.parse_args(argv)
    output = args.output_option or args.output_path
    if output is None:
        parser.error("a distribution directory is required (use --output PATH)")
    try:
        build_distribution(args.repo.resolve(), output.resolve(), args.version)
    except (OSError, subprocess.CalledProcessError, ValueError) as error:
        print(f"package release: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
