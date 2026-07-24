from __future__ import annotations

import argparse
import os
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def run(binary: Path, *arguments: str, cwd: Path | None = None) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        [str(binary), *arguments],
        cwd=cwd,
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stdout + result.stderr)
    return result


def build_binary(directory: Path) -> Path:
    binary = directory / executable_name("lexicon")
    result = subprocess.run(
        ["go", "build", "-o", str(binary), "./cmd/lexicon"],
        cwd=ROOT,
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stdout + result.stderr)
    return binary


def executable_name(name: str) -> str:
    return name + ".exe" if os.name == "nt" else name


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="Exercise Lexicon operator commands end to end.")
    parser.add_argument("--distribution", type=Path, help="use an existing packaged distribution")
    args = parser.parse_args(argv)

    with tempfile.TemporaryDirectory(prefix="lexicon-operations-") as directory:
        temporary = Path(directory)
        if args.distribution:
            distribution = args.distribution.resolve()
            binary = distribution / executable_name("lexicon")
            adapters = distribution / "adapters"
        else:
            binary = build_binary(temporary)
            adapters = ROOT / "adapters"

        repository = temporary / "repository"
        repository.mkdir()
        (repository / "main.py").write_text("def answer():\n    return 42\n", encoding="utf-8")
        (repository / "main.go").write_text("package smoke\n\nfunc Answer() int { return 42 }\n", encoding="utf-8")
        (repository / "go.mod").write_text("module example.com/smoke\n\ngo 1.26\n", encoding="utf-8")

        initialized = run(
            binary,
            "init",
            "--repo",
            str(repository),
            "--adapters",
            str(adapters),
            "--languages",
            "python,go",
        )
        if "libraries: go, python" not in initialized.stdout:
            raise RuntimeError(f"unexpected init output: {initialized.stdout}")

        nested = repository / "nested" / "working"
        nested.mkdir(parents=True)
        status = run(binary, "status", cwd=nested)
        if "enabled languages: go, python" not in status.stdout:
            raise RuntimeError(f"unexpected status output: {status.stdout}")

        run(binary, "doctor", cwd=nested)
        run(binary, "rebuild", "--languages", "go,python", cwd=nested)
        run(binary, "languages", "set", "--languages", "go,python", cwd=nested)

        export = temporary / "export"
        run(binary, "export", "--output", str(export), cwd=nested)
        for language in ("go", "python"):
            if not (export / f"{language}.jsonl").is_file():
                raise RuntimeError(f"export did not write {language}.jsonl")

        consumer_code = "import os; assert os.environ['LEXICON_SNAPSHOT_ID'].startswith('sha256:')"
        run(
            binary,
            "consumer",
            "add",
            "--name",
            "smoke",
            "--command",
            sys.executable,
            "--arg=-c",
            "--arg",
            consumer_code,
            "--timeout",
            "10s",
            cwd=nested,
        )
        listed = run(binary, "consumer", "list", cwd=nested)
        if listed.stdout.strip() != "smoke":
            raise RuntimeError(f"unexpected consumer list: {listed.stdout}")
        run(binary, "consumer", "run", "--name", "smoke", cwd=nested)
        run(binary, "gc", "--retain", "2", "--dry-run", cwd=nested)
        run(binary, "consumer", "remove", "--name", "smoke", cwd=nested)

    print("Lexicon operational smoke test passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
