from __future__ import annotations

import importlib.util
import io
import json
import os
import subprocess
import sys
import tempfile
import unittest
from contextlib import contextmanager
from datetime import datetime, timezone
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = ROOT.parents[1]


def load_module(name: str, path: Path):
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


render_templates = load_module("render_templates", ROOT / "scripts" / "render_templates.py")
fetch_cloudflare_ips = load_module("fetch_cloudflare_ips", ROOT / "scripts" / "fetch_cloudflare_ips.py")


class RenderTemplatesTests(unittest.TestCase):
    def test_render_tree_sets_exact_thirty_day_sunset(self):
        cutover = render_templates.parse_cutover("2026-07-19T14:30:00+08:00")
        replacements, sunset = render_templates.build_replacements(
            cutover, 30, "miraiapi.cloud", "akimirai.xyz"
        )
        self.assertEqual(sunset.isoformat(), "2026-08-18T14:30:00+08:00")
        self.assertEqual(replacements["{{SUNSET_HTTP_DATE}}"], "Tue, 18 Aug 2026 06:30:00 GMT")
        self.assertEqual(replacements["{{RETIRE_AT_UTC}}"], "2026-08-18 06:30:00 UTC")

        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary)
            rendered = render_templates.render_tree(ROOT, output, replacements)
            self.assertTrue(rendered)
            for path in rendered:
                self.assertNotIn("{{", path.read_text(encoding="utf-8"))
            compat = (output / "nginx" / "akimirai.xyz-compat.conf").read_text(encoding="utf-8")
            headers = (output / "nginx" / "miraiapi-legacy-headers.conf").read_text(encoding="utf-8")
            self.assertIn("Sunset \"Tue, 18 Aug 2026 06:30:00 GMT\"", headers)
            self.assertIn("return 308 https://miraiapi.cloud$request_uri", compat)

    def test_cutover_requires_offset(self):
        with self.assertRaisesRegex(ValueError, "UTC offset"):
            render_templates.parse_cutover("2026-07-19T14:30:00")


class CloudflareIPTests(unittest.TestCase):
    @contextmanager
    def response(self, body: str):
        yield io.BytesIO(body.encode("ascii"))

    def test_fetch_validate_and_render(self):
        def opener(request, timeout: int):
            self.assertEqual(timeout, 30)
            self.assertEqual(request.headers["User-agent"], "MiraiAPI-domain-migration/1.0")
            body = (
                "173.245.48.0/20\n103.21.244.0/22\n"
                if request.full_url.endswith("ips-v4")
                else "2400:cb00::/32\n"
            )
            return self.response(body)

        ipv4 = fetch_cloudflare_ips.validate_networks(
            fetch_cloudflare_ips.fetch_lines(fetch_cloudflare_ips.IPV4_URL, opener), 4
        )
        ipv6 = fetch_cloudflare_ips.validate_networks(
            fetch_cloudflare_ips.fetch_lines(fetch_cloudflare_ips.IPV6_URL, opener), 6
        )
        config = fetch_cloudflare_ips.render_config(
            ipv4, ipv6, datetime(2026, 7, 19, tzinfo=timezone.utc)
        )
        self.assertIn("set_real_ip_from 173.245.48.0/20;", config)
        self.assertIn("set_real_ip_from 2400:cb00::/32;", config)
        self.assertIn("real_ip_header CF-Connecting-IP;", config)
        self.assertIn("real_ip_recursive on;", config)

    def test_rejects_wrong_address_family(self):
        with self.assertRaisesRegex(ValueError, "expected IPv4"):
            fetch_cloudflare_ips.validate_networks(["2400:cb00::/32"], 4)


class StaticContractTests(unittest.TestCase):
    def test_nginx_contract_covers_streaming_and_compatibility(self):
        proxy = (ROOT / "nginx" / "miraiapi-proxy.conf").read_text(encoding="utf-8")
        compat = (ROOT / "nginx" / "akimirai.xyz-compat.conf.template").read_text(encoding="utf-8")
        primary = (ROOT / "nginx" / "miraiapi.cloud.conf.template").read_text(encoding="utf-8")
        for directive in (
            "proxy_http_version 1.1",
            "proxy_set_header Upgrade $http_upgrade",
            "proxy_read_timeout 300s",
            "proxy_send_timeout 300s",
        ):
            self.assertIn(directive, proxy)
        self.assertIn("client_max_body_size 256m", primary)
        self.assertIn("(?:v1|v1beta|backend-api|antigravity|api|setup)", compat)
        self.assertIn("location = /health", compat)
        self.assertIn("location ^~ /rent-ledger/", compat)

    def test_legacy_compatibility_proxies_bare_gateway_aliases(self):
        compat = (ROOT / "nginx" / "akimirai.xyz-compat.conf.template").read_text(encoding="utf-8")
        for route_contract in (
            "(?:v1|v1beta|backend-api|antigravity|api|setup)",
            "(?:responses|videos)",
            "chat/completions|embeddings|alpha/search",
            "images/(?:generations|edits)",
        ):
            self.assertIn(route_contract, compat)

    def test_retirement_script_has_validation_and_restore_path(self):
        script = (ROOT / "scripts" / "retire_miraiapi_legacy.sh").read_text(encoding="utf-8")
        self.assertIn('if ! "$nginx_bin" -t', script)
        self.assertIn("restore_legacy", script)
        self.assertIn('if ! "$systemctl_bin" reload nginx', script)

    def test_public_defaults_use_new_domain(self):
        xianyu_script = (REPO_ROOT / "tools" / "xianyu_auto_fulfill.mjs").read_text(encoding="utf-8")
        xianyu_doc = (REPO_ROOT / "tools" / "xianyu_auto_fulfill.md").read_text(encoding="utf-8")
        rent_doc = (REPO_ROOT / "deploy" / "RENT_LEDGER_UPDATE.md").read_text(encoding="utf-8")
        self.assertIn("const defaultAPIBase = 'https://miraiapi.cloud'", xianyu_script)
        self.assertIn("https://miraiapi.cloud/redeem", xianyu_doc)
        self.assertIn("https://miraiapi.cloud/rent-ledger/latest.json", rent_doc)

    def test_retired_cn_domain_is_absent_from_maintained_assets(self):
        maintained = [
            *ROOT.rglob("*.py"),
            *ROOT.rglob("*.sh"),
            *ROOT.rglob("*.template"),
            ROOT / "README.md",
            REPO_ROOT / "deploy" / "RENT_LEDGER_UPDATE.md",
            REPO_ROOT / "deploy" / "rent-ledger-update.Caddyfile.example",
            REPO_ROOT / "tools" / "xianyu_auto_fulfill.mjs",
            REPO_ROOT / "tools" / "xianyu_auto_fulfill.md",
        ]
        retired_domain = "miraiapi" + ".cn"
        offenders = [
            str(path.relative_to(REPO_ROOT))
            for path in maintained
            if path.is_file() and retired_domain in path.read_text(encoding="utf-8")
        ]
        self.assertEqual(offenders, [])

    def test_retirement_script_parses_as_bash_when_available(self):
        if os.name == "nt":
            self.skipTest("run bash syntax validation on the Linux target")
        try:
            result = subprocess.run(
                ["bash", "-n", str(ROOT / "scripts" / "retire_miraiapi_legacy.sh")],
                check=False,
                capture_output=True,
                text=True,
            )
        except FileNotFoundError:
            self.skipTest("bash is not installed")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_retirement_integration_when_linux_bash_is_available(self):
        if os.name == "nt":
            self.skipTest("run retirement integration on the Linux target")
        try:
            result = subprocess.run(
                ["bash", str(ROOT / "tests" / "test_retire_script.sh")],
                check=False,
                capture_output=True,
                text=True,
            )
        except FileNotFoundError:
            self.skipTest("bash is not installed")
        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
