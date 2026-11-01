#!/usr/bin/env python3
"""Fail-closed link checker for published markdown.

Checks README, claims/readiness/threat-model docs, docs/**, and site/docs/**.
Fails on file:/// schemes, host-specific /home/ paths, broken relative links,
and external links returning HTTP 4xx/5xx (after retries).
"""

from __future__ import annotations

import concurrent.futures
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
EXCEPTIONS = ROOT / "docs" / "launch" / "LINK_EXCEPTIONS.txt"

TARGETS = [
    ROOT / "README.md",
    ROOT / "CLAIMS.md",
    ROOT / "LAUNCH_READINESS.md",
    ROOT / "THREAT_MODEL.md",
]
TARGETS += sorted((ROOT / "docs").rglob("*.md"))
TARGETS += sorted((ROOT / "site" / "docs").rglob("*.mdx"))

MD_LINK = re.compile(r"\[[^\]]*\]\(([^)\s]+)(?:\s+\"[^\"]*\")?\)")
HTML_ATTR = re.compile(r'(?:href|src)="([^"]+)"')
INLINE_CODE = re.compile(r"`[^`]*`")

SKIP_PREFIXES = ("mailto:", "tel:", "data:", "javascript:", "#", "file:///")
SKIP_HOSTS = ("http://localhost", "http://127.0.0.1", "https://localhost", "https://127.0.0.1")
USER_AGENT = "exist.io-link-check/1.0 (+https://github.com/RavaniRoshan/exist.io)"


def load_exceptions() -> list[str]:
    if not EXCEPTIONS.exists():
        return []
    entries = []
    for line in EXCEPTIONS.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        entries.append(line.split()[0])
    return entries


def links_in_file(path: Path):
    in_fence = False
    for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if line.strip().startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        clean = INLINE_CODE.sub("", line)
        for match in MD_LINK.finditer(clean):
            yield number, match.group(1)
        for match in HTML_ATTR.finditer(clean):
            yield number, match.group(1)


def split_link(link: str) -> str:
    return link.split("#", 1)[0].split("?", 1)[0]


def external_status(url: str) -> tuple[bool, str]:
    last = ""
    for attempt in range(3):
        for method in ("HEAD", "GET"):
            request = urllib.request.Request(url, method=method, headers={"User-Agent": USER_AGENT})
            try:
                with urllib.request.urlopen(request, timeout=15) as response:
                    if 200 <= response.status < 400:
                        return True, str(response.status)
                    last = f"HTTP {response.status}"
            except urllib.error.HTTPError as error:
                last = f"HTTP {error.code}"
                if error.code in (405, 501) and method == "HEAD":
                    continue
                if error.code < 500 and error.code != 429:
                    return False, last
            except Exception as error:  # noqa: BLE001 - network failures are reported verbatim
                last = f"{type(error).__name__}: {error}"
        time.sleep(0.5 * (attempt + 1))
    return False, last


def main() -> int:
    exceptions = load_exceptions()
    failures: list[str] = []
    checked: dict[str, tuple[bool, str]] = {}
    local_checks: list[tuple[Path, int, str]] = []

    for path in TARGETS:
        if not path.exists():
            failures.append(f"{path}: missing target file")
            continue
        raw_lines = path.read_text(encoding="utf-8").splitlines()
        for number, line in enumerate(raw_lines, 1):
            if "file:///" in line:
                failures.append(f"{path.relative_to(ROOT)}:{number}: file:/// scheme is forbidden")
            if "/home/" in line:
                failures.append(f"{path.relative_to(ROOT)}:{number}: host-specific /home/ path")

        for number, link in links_in_file(path):
            if not link or link.startswith(SKIP_PREFIXES):
                continue
            if link.startswith(SKIP_HOSTS) or link.startswith("<"):
                continue
            clean = split_link(link)
            if not clean:
                continue
            if clean.startswith(("http://", "https://")):
                if any(clean.startswith(prefix) for prefix in exceptions):
                    continue
                local_checks.append((path, number, clean))
                continue
            if clean.startswith("/"):
                continue  # site-absolute route, validated by the docs build
            target = (path.parent / clean).resolve()
            if not target.exists():
                failures.append(f"{path.relative_to(ROOT)}:{number}: missing local target {clean}")

    urls = sorted({url for _, _, url in local_checks})
    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:
        futures = {pool.submit(external_status, url): url for url in urls}
        for future in concurrent.futures.as_completed(futures):
            url = futures[future]
            ok, detail = future.result()
            checked[url] = (ok, detail)

    for path, number, url in local_checks:
        ok, detail = checked[url]
        if not ok:
            failures.append(f"{path.relative_to(ROOT)}:{number}: {url} -> {detail}")

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        print(f"\nlink check failed: {len(failures)} issue(s)", file=sys.stderr)
        return 1

    print(f"link check passed: {len(urls)} external URL(s), {len(TARGETS)} file(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
