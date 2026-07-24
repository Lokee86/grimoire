from __future__ import annotations

import json
import subprocess
from pathlib import Path

DOC_LEDGER_COMMIT = "7fbdb81307ff7c6b1ba28886c3a40b8837ebf785"
LEXICANTER_URL = "https://github.com/Saturnine-Softworks/Lexicanter.git"
LEXICANTER_COMMIT = "eac754788b3cf18a930c085c1c49f8f353e18107"


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


def ensure_doc_ledger(workspace: Path, corpus_root: Path) -> Path:
    source = workspace / "demon-docs"
    target = corpus_root / "doc-ledger-python"
    if target.exists():
        return target
    run(
        "git",
        "-C",
        str(source),
        "worktree",
        "add",
        "--detach",
        str(target),
        DOC_LEDGER_COMMIT,
    )
    return target


def ensure_lexicanter(corpus_root: Path) -> Path:
    target = corpus_root / "lexicanter"
    if not target.exists():
        run("git", "clone", "--filter=blob:none", LEXICANTER_URL, str(target))
    if revision(target) != LEXICANTER_COMMIT:
        run("git", "fetch", "origin", LEXICANTER_COMMIT, cwd=target)
        run("git", "checkout", "--detach", LEXICANTER_COMMIT, cwd=target)
    return target


def revision(repository: Path) -> str:
    return run("git", "rev-parse", "HEAD", cwd=repository)


def main() -> None:
    workspace = Path(__file__).resolve().parents[2]
    corpus_root = workspace / "corpus"
    corpus_root.mkdir(parents=True, exist_ok=True)

    doc_ledger = ensure_doc_ledger(workspace, corpus_root)
    lexicanter = ensure_lexicanter(corpus_root)
    state = {
        "doc-ledger-python": {
            "path": doc_ledger.as_posix(),
            "revision": revision(doc_ledger),
        },
        "lexicanter": {
            "path": lexicanter.as_posix(),
            "revision": revision(lexicanter),
        },
    }
    state_path = Path(__file__).with_name("corpus_state.json")
    state_path.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(state, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
