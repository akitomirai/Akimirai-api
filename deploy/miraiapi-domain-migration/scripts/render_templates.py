#!/usr/bin/env python3
"""Render domain-migration templates from an offset-aware cutover timestamp."""

from __future__ import annotations

import argparse
import json
from datetime import datetime, timedelta, timezone
from email.utils import format_datetime
from pathlib import Path


DEFAULT_COMPATIBILITY_DAYS = 30


def parse_cutover(value: str) -> datetime:
    normalized = value.replace("Z", "+00:00")
    parsed = datetime.fromisoformat(normalized)
    if parsed.tzinfo is None or parsed.utcoffset() is None:
        raise ValueError("cutover timestamp must include a UTC offset")
    return parsed


def build_replacements(
    cutover: datetime,
    compatibility_days: int,
    primary_domain: str,
    legacy_domain: str,
) -> tuple[dict[str, str], datetime]:
    sunset = cutover + timedelta(days=compatibility_days)
    sunset_utc = sunset.astimezone(timezone.utc)
    replacements = {
        "{{PRIMARY_DOMAIN}}": primary_domain,
        "{{PRIMARY_WWW_DOMAIN}}": f"www.{primary_domain}",
        "{{LEGACY_DOMAIN}}": legacy_domain,
        "{{LEGACY_WWW_DOMAIN}}": f"www.{legacy_domain}",
        "{{CUTOVER_ISO8601}}": cutover.isoformat(),
        "{{SUNSET_HTTP_DATE}}": format_datetime(sunset_utc, usegmt=True),
        "{{RETIRE_AT_UTC}}": sunset_utc.strftime("%Y-%m-%d %H:%M:%S UTC"),
    }
    return replacements, sunset


def render_tree(source_dir: Path, output_dir: Path, replacements: dict[str, str]) -> list[Path]:
    rendered: list[Path] = []
    for source in sorted(source_dir.rglob("*.template")):
        relative = source.relative_to(source_dir)
        destination = output_dir / relative.with_suffix("")
        text = source.read_text(encoding="utf-8")
        for token, value in replacements.items():
            text = text.replace(token, value)
        unresolved = sorted({part.split("}}", 1)[0] + "}}" for part in text.split("{{")[1:] if "}}" in part})
        if unresolved:
            raise ValueError(f"unresolved template tokens in {source}: {', '.join(unresolved)}")
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_text(text, encoding="utf-8", newline="\n")
        rendered.append(destination)
    return rendered


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--cutover", required=True, help="ISO-8601 timestamp with UTC offset")
    parser.add_argument("--source-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--compatibility-days", type=int, default=DEFAULT_COMPATIBILITY_DAYS)
    parser.add_argument("--primary-domain", default="miraiapi.cn")
    parser.add_argument("--legacy-domain", default="akimirai.xyz")
    args = parser.parse_args()

    if args.compatibility_days < 1:
        parser.error("--compatibility-days must be positive")

    try:
        cutover = parse_cutover(args.cutover)
    except ValueError as exc:
        parser.error(str(exc))

    replacements, sunset = build_replacements(
        cutover,
        args.compatibility_days,
        args.primary_domain,
        args.legacy_domain,
    )
    rendered = render_tree(args.source_dir, args.output_dir, replacements)
    manifest = {
        "primary_domain": args.primary_domain,
        "legacy_domain": args.legacy_domain,
        "cutover": cutover.isoformat(),
        "compatibility_days": args.compatibility_days,
        "sunset": sunset.isoformat(),
        "rendered_files": [str(path.relative_to(args.output_dir)) for path in rendered],
    }
    (args.output_dir / "migration-manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
        newline="\n",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
