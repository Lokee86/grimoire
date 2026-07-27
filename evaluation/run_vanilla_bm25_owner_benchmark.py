from __future__ import annotations

import argparse
import json
import math
import os
import pathlib
import re
import statistics
import time
from collections import Counter, defaultdict
from dataclasses import dataclass

SOURCE_EXTENSIONS = {
    ".c", ".cc", ".cpp", ".cs", ".go", ".gd", ".java", ".js", ".jsx",
    ".kt", ".kts", ".py", ".rb", ".rs", ".svelte", ".ts", ".tsx",
}
IGNORED_DIRS = {
    ".arcana", ".astro", ".git", ".godot", ".grimoire", ".lexicon", ".next",
    ".pytest_cache", ".worktrees", ".workingtrees", "__pycache__", "bin", "coverage",
    "dist", "evaluation", "logs", "node_modules", "target", "tmp", "vendor",
}
TOKEN_PATTERN = re.compile(r"[A-Za-z][A-Za-z0-9_]*")


@dataclass(frozen=True)
class Document:
    path: str
    start_line: int
    end_line: int
    text: str
    terms: tuple[str, ...]


def tokenize(text: str) -> tuple[str, ...]:
    terms: list[str] = []
    for raw in TOKEN_PATTERN.findall(text):
        expanded = re.sub(r"([a-z0-9])([A-Z])", r"\1 \2", raw).replace("_", " ")
        terms.extend(part.lower() for part in expanded.split() if len(part) > 1)
    return tuple(terms)


def source_files(root: pathlib.Path, exclude_tests: bool):
    for directory, subdirectories, filenames in os.walk(root):
        subdirectories[:] = [name for name in subdirectories if name not in IGNORED_DIRS]
        base = pathlib.Path(directory)
        for filename in filenames:
            path = base / filename
            if path.suffix.lower() not in SOURCE_EXTENSIONS:
                continue
            relative = path.relative_to(root)
            lowered = relative.as_posix().lower()
            if exclude_tests and any(re.search(r"(^|[_.-])(tests?|specs?)([_.-]|$)", part.lower()) for part in relative.parts):
                continue
            yield path


def build_documents(root: pathlib.Path, window: int, stride: int, exclude_tests: bool) -> list[Document]:
    documents: list[Document] = []
    for path in source_files(root, exclude_tests):
        try:
            lines = path.read_text(encoding="utf-8", errors="ignore").splitlines()
        except OSError:
            continue
        relative = path.relative_to(root).as_posix()
        if not lines:
            continue
        if window <= 0:
            text = relative + "\n" + "\n".join(lines)
            documents.append(Document(relative, 1, len(lines), text, tokenize(text)))
            continue
        for offset in range(0, len(lines), stride):
            chunk = lines[offset : offset + window]
            if not chunk:
                break
            text = relative + "\n" + "\n".join(chunk)
            documents.append(Document(relative, offset + 1, offset + len(chunk), text, tokenize(text)))
            if offset + window >= len(lines):
                break
    return documents


def bm25_rank(documents: list[Document], query: str, limit: int) -> list[tuple[float, Document]]:
    query_terms = tokenize(query)
    if not query_terms or not documents:
        return []
    frequencies = [Counter(document.terms) for document in documents]
    document_frequency: dict[str, int] = defaultdict(int)
    for frequency in frequencies:
        for term in set(query_terms) & frequency.keys():
            document_frequency[term] += 1
    average_length = statistics.fmean(len(document.terms) for document in documents) or 1.0
    scores: list[tuple[float, Document]] = []
    for document, frequency in zip(documents, frequencies):
        score = 0.0
        length = len(document.terms)
        for term in query_terms:
            tf = frequency.get(term, 0)
            if tf == 0:
                continue
            df = document_frequency.get(term, 0)
            idf = math.log(1.0 + (len(documents) - df + 0.5) / (df + 0.5))
            denominator = tf + 1.2 * (1.0 - 0.75 + 0.75 * length / average_length)
            score += idf * (tf * 2.2) / denominator
        if score > 0:
            scores.append((score, document))
    scores.sort(key=lambda item: (-item[0], item[1].path, item[1].start_line))
    return scores[:limit]


def required_rank(results: list[tuple[float, Document]], seed: dict) -> int:
    expected_path = seed["path"].replace("\\", "/").lower()
    expected_name = seed["name"].lower()
    for rank, (_, document) in enumerate(results, 1):
        if document.path.lower() == expected_path and expected_name in document.text.lower():
            return rank
    return 0


def evaluate(root: pathlib.Path, corpus: dict, window: int, stride: int, exclude_tests: bool) -> dict:
    started = time.perf_counter()
    documents = build_documents(root, window, stride, exclude_tests)
    index_ms = (time.perf_counter() - started) * 1000
    measured = []
    found = total = passed = 0
    reciprocal_ranks: list[float] = []
    latencies: list[float] = []
    for case in corpus["cases"]:
        query_started = time.perf_counter()
        results = bm25_rank(documents, case["query"], corpus["top_k"])
        latency_ms = (time.perf_counter() - query_started) * 1000
        ranks = [required_rank(results, seed) for seed in case["required_seeds"]]
        case_found = sum(rank > 0 for rank in ranks)
        best = min((rank for rank in ranks if rank > 0), default=0)
        found += case_found
        total += len(ranks)
        passed += int(case_found == len(ranks))
        reciprocal_ranks.append(1 / best if best else 0)
        latencies.append(latency_ms)
        measured.append({
            "id": case["id"],
            "required_ranks": ranks,
            "latency_ms": latency_ms,
            "results": [
                {"score": score, "path": doc.path, "start_line": doc.start_line, "end_line": doc.end_line}
                for score, doc in results
            ],
        })
    return {
        "repository": corpus["repository"],
        "documents": len(documents),
        "window_lines": window,
        "stride_lines": stride,
        "exclude_tests": exclude_tests,
        "index_ms": index_ms,
        "pass_rate": passed / len(corpus["cases"]) if corpus["cases"] else 0,
        "required_seed_recall": found / total if total else 0,
        "mrr": statistics.fmean(reciprocal_ranks) if reciprocal_ranks else 0,
        "median_query_ms": statistics.median(latencies) if latencies else 0,
        "cases": measured,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="Vanilla BM25 owner benchmark")
    parser.add_argument("--root", required=True)
    parser.add_argument("--cases", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--window", type=int, default=40)
    parser.add_argument("--stride", type=int, default=20)
    parser.add_argument("--exclude-tests", action="store_true")
    args = parser.parse_args()
    corpus = json.loads(pathlib.Path(args.cases).read_text(encoding="utf-8"))
    report = evaluate(pathlib.Path(args.root), corpus, args.window, args.stride, args.exclude_tests)
    pathlib.Path(args.output).write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    print(
        f"{report['repository']}: documents={report['documents']} "
        f"pass={report['pass_rate']:.1%} recall={report['required_seed_recall']:.1%} "
        f"mrr={report['mrr']:.3f} median_ms={report['median_query_ms']:.1f} "
        f"index_ms={report['index_ms']:.1f}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
