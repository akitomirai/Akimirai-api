#!/usr/bin/env python3
"""Generate an Nginx real-IP include from Cloudflare's official IP lists."""

from __future__ import annotations

import argparse
import ipaddress
import os
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from urllib.request import Request, urlopen


IPV4_URL = "https://www.cloudflare.com/ips-v4"
IPV6_URL = "https://www.cloudflare.com/ips-v6"


def fetch_lines(url: str, opener=urlopen) -> list[str]:
    request = Request(url, headers={"User-Agent": "MiraiAPI-domain-migration/1.0"})
    with opener(request, timeout=30) as response:
        body = response.read().decode("ascii")
    return [line.strip() for line in body.splitlines() if line.strip()]


def validate_networks(lines: list[str], expected_version: int) -> list[str]:
    networks: set[str] = set()
    for line in lines:
        network = ipaddress.ip_network(line, strict=True)
        if network.version != expected_version:
            raise ValueError(f"expected IPv{expected_version} network, got {line}")
        networks.add(str(network))
    if not networks:
        raise ValueError(f"Cloudflare IPv{expected_version} list is empty")
    return sorted(networks, key=lambda value: (ipaddress.ip_network(value).version, int(ipaddress.ip_network(value).network_address)))


def render_config(ipv4: list[str], ipv6: list[str], generated_at: datetime | None = None) -> str:
    generated = (generated_at or datetime.now(timezone.utc)).astimezone(timezone.utc)
    lines = [
        "# Generated from Cloudflare's official IP lists.",
        f"# Source: {IPV4_URL}",
        f"# Source: {IPV6_URL}",
        f"# Generated: {generated.isoformat()}",
    ]
    lines.extend(f"set_real_ip_from {network};" for network in [*ipv4, *ipv6])
    lines.extend(["real_ip_header CF-Connecting-IP;", "real_ip_recursive on;", ""])
    return "\n".join(lines)


def atomic_write(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    handle, temporary = tempfile.mkstemp(prefix=f".{path.name}.", dir=path.parent, text=True)
    try:
        with os.fdopen(handle, "w", encoding="utf-8", newline="\n") as stream:
            stream.write(content)
        os.replace(temporary, path)
    except Exception:
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass
        raise


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    ipv4 = validate_networks(fetch_lines(IPV4_URL), 4)
    ipv6 = validate_networks(fetch_lines(IPV6_URL), 6)
    atomic_write(args.output, render_config(ipv4, ipv6))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
