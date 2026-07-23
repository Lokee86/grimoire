from __future__ import annotations

import configparser
import json
import os
from pathlib import Path

LANGUAGE_EXTENSIONS = {
    ".py": "python",
    ".rb": "ruby",
    ".ts": "typescript",
    ".tsx": "typescript",
    ".mts": "typescript",
    ".cts": "typescript",
    ".js": "javascript",
    ".jsx": "javascript",
    ".mjs": "javascript",
    ".cjs": "javascript",
    ".svelte": "svelte",
    ".gd": "gdscript",
    ".rs": "rust",
    ".go": "go",
}

IGNORED_DIRECTORIES = {
    ".git",
    ".worktrees",
    ".workingtrees",
    ".ddocs",
    ".lexicon",
    ".arcana",
    ".grimoire",
    ".pitlord",
    ".cantrip",
    ".homunculus",
    ".incubus",
    ".ritual",
    ".warlock",
    ".godot",
    ".import",
    ".venv",
    "venv",
    "__pycache__",
    "node_modules",
    "target",
    "vendor",
    "dist",
    "build",
}

MANIFESTS = (
    "pyproject.toml",
    "requirements.txt",
    "Gemfile",
    "package.json",
    "project.godot",
    "Cargo.toml",
    "go.mod",
)


def origin_url(repository: Path) -> str:
    config_path = repository / ".git" / "config"
    if not config_path.is_file():
        return ""
    parser = configparser.ConfigParser()
    parser.read(config_path, encoding="utf-8")
    return parser.get('remote "origin"', "url", fallback="")


def inventory(repository: Path) -> dict[str, object]:
    languages: dict[str, int] = {}
    for directory, child_directories, file_names in os.walk(repository):
        child_directories[:] = sorted(
            child for child in child_directories if child not in IGNORED_DIRECTORIES
        )
        for file_name in file_names:
            language = LANGUAGE_EXTENSIONS.get(Path(file_name).suffix.lower())
            if language:
                languages[language] = languages.get(language, 0) + 1
    return {
        "repository": repository.name,
        "path": repository.as_posix(),
        "origin": origin_url(repository),
        "manifests": [name for name in MANIFESTS if (repository / name).is_file()],
        "languages": dict(sorted(languages.items())),
    }


def main() -> None:
    workspace = Path(__file__).resolve().parents[2]
    repositories = sorted(
        (path for path in workspace.iterdir() if path.is_dir() and (path / ".git").exists()),
        key=lambda path: path.name.lower(),
    )
    print(json.dumps([inventory(repository) for repository in repositories], indent=2))


if __name__ == "__main__":
    main()
