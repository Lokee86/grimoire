#!/usr/bin/env python3
"""Download a verified portable Temurin JDK for Lexicon development."""

from __future__ import annotations

import argparse
import hashlib
import json
import shutil
import tempfile
import urllib.request
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ASSETS_URL = (
    "https://api.adoptium.net/v3/assets/latest/{version}/hotspot"
    "?architecture=x64&image_type=jdk&os=windows&vendor=eclipse"
)


def fetch_json(url: str) -> object:
    request = urllib.request.Request(url, headers={"User-Agent": "Lexicon-JDK-Bootstrap/1"})
    with urllib.request.urlopen(request, timeout=60) as response:
        return json.load(response)


def download(url: str, destination: Path) -> None:
    request = urllib.request.Request(url, headers={"User-Agent": "Lexicon-JDK-Bootstrap/1"})
    with urllib.request.urlopen(request, timeout=120) as response, destination.open("wb") as output:
        shutil.copyfileobj(response, output)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def resolve_package(version: int) -> tuple[str, str]:
    assets = fetch_json(ASSETS_URL.format(version=version))
    if not isinstance(assets, list) or not assets:
        raise RuntimeError(f"Adoptium returned no Windows x64 JDK {version} assets")
    package = assets[0]["binary"]["package"]
    return package["link"], package["checksum"]


def install(version: int, destination: Path) -> None:
    link, expected_checksum = resolve_package(version)
    with tempfile.TemporaryDirectory(prefix="lexicon-jdk-") as temporary:
        archive = Path(temporary) / "jdk.zip"
        extracted = Path(temporary) / "extracted"
        print(f"Downloading Temurin JDK {version}...")
        download(link, archive)
        actual_checksum = sha256(archive)
        if actual_checksum.lower() != expected_checksum.lower():
            raise RuntimeError(
                f"JDK checksum mismatch: expected {expected_checksum}, got {actual_checksum}"
            )
        with zipfile.ZipFile(archive) as package:
            package.extractall(extracted)
        roots = [path for path in extracted.iterdir() if path.is_dir()]
        if len(roots) != 1 or not (roots[0] / "bin" / "javac.exe").is_file():
            raise RuntimeError("downloaded archive does not contain one valid JDK root")
        if destination.exists():
            shutil.rmtree(destination)
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(roots[0]), destination)
    print(f"Portable JDK installed at {destination}")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", type=int, default=21)
    parser.add_argument("--output", type=Path, default=ROOT / ".tools" / "jdk")
    args = parser.parse_args()
    install(args.version, args.output.resolve())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
