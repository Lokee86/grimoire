#!/usr/bin/env python3
"""Install a combined Grimoire bundle produced by scripts/workflow.py."""

from __future__ import annotations

import argparse
import platform
import shutil
from pathlib import Path


def executable_name(name: str) -> str:
    return name + ".exe" if platform.system().lower() == "windows" else name


def native_library_name() -> str:
    system = platform.system().lower()
    if system == "windows":
        return "grimoire_vector_ffi.dll"
    if system == "darwin":
        return "libgrimoire_vector_ffi.dylib"
    return "libgrimoire_vector_ffi.so"


def copy_file(source: Path, destination: Path) -> None:
    if not source.is_file():
        raise FileNotFoundError(source)
    destination.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source, destination)


def install(source: Path, bin_dir: Path) -> None:
    source = source.resolve()
    bin_dir = bin_dir.resolve()
    source_bin = source / "bin"
    source_native = source / "native"
    names = [executable_name(name) for name in ("grimoire", "lexicon", "arcana")]
    for name in names:
        if not (source_bin / name).is_file():
            raise FileNotFoundError(f"combined bundle is missing {source_bin / name}")
    library = source_native / native_library_name()
    if not library.is_file():
        raise FileNotFoundError(f"combined bundle is missing {library}")
    bin_dir.mkdir(parents=True, exist_ok=True)
    for name in names:
        copy_file(source_bin / name, bin_dir / name)
    # Keep the native library beside grimoire so normal discovery works on Windows
    # without requiring a GRIMOIRE_VECTOR_ENGINE environment variable.
    copy_file(library, bin_dir / library.name)
    print(f"installed Grimoire tools to {bin_dir}")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, default=Path(__file__).resolve().parent, help="combined build or extracted bundle; defaults to this bundle")
    parser.add_argument("--bin-dir", type=Path, required=True, help="directory receiving the commands")
    args = parser.parse_args()
    try:
        install(args.source, args.bin_dir)
    except OSError as error:
        parser.error(str(error))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
