from __future__ import annotations

import json
import subprocess
from pathlib import Path

DOC_LEDGER_COMMIT = "7fbdb81307ff7c6b1ba28886c3a40b8837ebf785"
CODEBASE_MEMORY_COMMIT = "97ce23f9827177fff3858831156e9795c6832b18"
LEXICANTER_URL = "https://github.com/Saturnine-Softworks/Lexicanter.git"
LEXICANTER_COMMIT = "eac754788b3cf18a930c085c1c49f8f353e18107"
GIT_URL = "https://github.com/git/git.git"
GIT_COMMIT = "9a0c4701dcd5725c4184599322b52933ff5005ca"
LEVELDB_URL = "https://github.com/google/leveldb.git"
LEVELDB_COMMIT = "99b3c03b3284f5886f9ef9a4ef703d57373e61be"
FMT_URL = "https://github.com/fmtlib/fmt.git"
FMT_COMMIT = "407c905e45ad75fc29bf0f9bb7c5c2fd3475976f"
CATCH2_URL = "https://github.com/catchorg/Catch2.git"
CATCH2_COMMIT = "191fa38c9b1596cd2576ab531d4ab4d5e8e05190"
NLOHMANN_JSON_URL = "https://github.com/nlohmann/json.git"
NLOHMANN_JSON_COMMIT = "55f93686c01528224f448c19128836e7df245f72"


def run(*args: str, cwd: Path | None = None) -> str:
    completed = subprocess.run(
        args,
        cwd=cwd,
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    return completed.stdout.strip()


def revision(repository: Path) -> str:
    return run("git", "rev-parse", "HEAD", cwd=repository)


def ensure_local_worktree(source: Path, target: Path, commit: str) -> Path:
    if not target.exists():
        run(
            "git",
            "-C",
            str(source),
            "worktree",
            "add",
            "--detach",
            str(target),
            commit,
        )
    elif revision(target) != commit:
        run("git", "checkout", "--detach", commit, cwd=target)
    return target


def ensure_clone(target: Path, url: str, commit: str) -> Path:
    if not target.exists():
        run("git", "clone", "--filter=blob:none", url, str(target))
    if revision(target) != commit:
        run("git", "fetch", "origin", commit, cwd=target)
        run("git", "checkout", "--detach", commit, cwd=target)
    return target


def workspace_root() -> Path:
    common_dir = Path(
        run(
            "git",
            "rev-parse",
            "--path-format=absolute",
            "--git-common-dir",
            cwd=Path(__file__).resolve().parent,
        )
    )
    return common_dir.parent.parent


def main() -> None:
    workspace = workspace_root()
    corpus_root = workspace / "corpus"
    corpus_root.mkdir(parents=True, exist_ok=True)

    repositories = {
        "doc-ledger-python": ensure_local_worktree(
            workspace / "demon-docs",
            corpus_root / "doc-ledger-python",
            DOC_LEDGER_COMMIT,
        ),
        "codebase-memory-mcp": ensure_local_worktree(
            workspace / "codebase-memory-mcp",
            corpus_root / "codebase-memory-mcp",
            CODEBASE_MEMORY_COMMIT,
        ),
        "lexicanter": ensure_clone(
            corpus_root / "lexicanter",
            LEXICANTER_URL,
            LEXICANTER_COMMIT,
        ),
        "git": ensure_clone(corpus_root / "git", GIT_URL, GIT_COMMIT),
        "leveldb": ensure_clone(corpus_root / "leveldb", LEVELDB_URL, LEVELDB_COMMIT),
        "fmt": ensure_clone(corpus_root / "fmt", FMT_URL, FMT_COMMIT),
        "catch2": ensure_clone(corpus_root / "catch2", CATCH2_URL, CATCH2_COMMIT),
        "nlohmann-json": ensure_clone(
            corpus_root / "nlohmann-json",
            NLOHMANN_JSON_URL,
            NLOHMANN_JSON_COMMIT,
        ),
    }
    state = {
        name: {"path": path.as_posix(), "revision": revision(path)}
        for name, path in repositories.items()
    }
    state_path = Path(__file__).with_name("corpus_state.json")
    state_path.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(state, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
