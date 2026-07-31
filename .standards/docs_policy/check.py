from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

sys.dont_write_bytecode = True

SCRIPT_DIR = Path(__file__).resolve().parent
if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

from audit import audit  # noqa: E402
from baseline import finding_key, load_baseline, write_baseline  # noqa: E402
from model import load_config  # noqa: E402


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description="Check repository documentation compliance")
    result.add_argument("--repo", default=".", help="repository root")
    result.add_argument("--config", default="docs-standard.json", help="repository-relative configuration path")
    result.add_argument("--changed-from", help="git base revision for change-impact checks")
    result.add_argument("--write-baseline", help="write current findings to this repository-relative baseline path")
    result.add_argument("--json", action="store_true", dest="as_json", help="emit machine-readable output")
    return result


def main() -> int:
    args = parser().parse_args()
    repo = Path(args.repo).resolve()
    try:
        config = load_config(repo, args.config)
        all_findings = audit(repo, config, args.changed_from)
        if args.write_baseline:
            baseline_path = repo / args.write_baseline
            baseline_path.parent.mkdir(parents=True, exist_ok=True)
            write_baseline(baseline_path, all_findings)
            print(f"documentation baseline: wrote {len(all_findings)} finding(s) to {args.write_baseline}")
            return 0

        baseline_keys = load_baseline(repo, config.baseline) if config.baseline else set()
        findings = [finding for finding in all_findings if finding_key(finding) not in baseline_keys]
        suppressed = len(all_findings) - len(findings)
        stale_baseline = len(baseline_keys - {finding_key(finding) for finding in all_findings})
    except (OSError, ValueError) as error:
        print(f"documentation check failed: {error}", file=sys.stderr)
        return 2

    if args.as_json:
        print(
            json.dumps(
                {
                    "schema": "laughing-skull.documentation-check.v1",
                    "repository": str(repo),
                    "profile": config.profile,
                    "passed": not findings,
                    "baselined_findings": suppressed,
                    "stale_baseline_entries": stale_baseline,
                    "findings": [finding.__dict__ for finding in findings],
                },
                indent=2,
                sort_keys=True,
            )
        )
    elif findings:
        for finding in findings:
            print(f"{finding.code}: {finding.path}: {finding.message}")
        suffix = f", {suppressed} baselined" if suppressed else ""
        print(f"documentation check: {len(findings)} new finding(s){suffix}")
    else:
        details: list[str] = []
        if suppressed:
            details.append(f"{suppressed} baselined legacy finding(s)")
        if stale_baseline:
            details.append(f"{stale_baseline} stale baseline entr{'y' if stale_baseline == 1 else 'ies'}")
        suffix = f"; {', '.join(details)}" if details else ""
        print(f"documentation check: passed ({config.profile}{suffix})")
    return 1 if findings else 0


if __name__ == "__main__":
    raise SystemExit(main())
